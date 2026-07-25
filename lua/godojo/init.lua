local config = require("godojo.config")
local M = {}

function M.setup(opts)
  config.setup(opts)
end

-- Resolves the absolute path to the compiled Go engine binary.
function M.get_bin_path()
  if config.options.bin_path and config.options.bin_path ~= "" then
    return config.options.bin_path
  end

  -- Detect relative to current file's directory:
  -- debug.getinfo(1).source gives something like "@.../lua/godojo/init.lua"
  local source = debug.getinfo(1).source
  if source:sub(1, 1) == "@" then
    local current_file = source:sub(2)
    -- init.lua is in lua/godojo, so we go up 3 directories to find plugin root
    local root = vim.fs.dirname(vim.fs.dirname(vim.fs.dirname(current_file)))
    local bin = root .. "/bin/godojo"
    if vim.fn.executable(bin) == 1 then
      return bin
    end
  end

  -- Fallback to workspace/PATH search
  return "godojo"
end

-- Asynchronously communicates with the Go engine using JSON.
-- @param action string: the engine action to perform (e.g., "ping")
-- @param payload table: any additional keys to send in the JSON payload
-- @param callback fun(err: string|nil, response: table|nil): callback invoked on process completion
function M.call_engine(action, payload, callback)
  local bin = M.get_bin_path()
  if vim.fn.executable(bin) ~= 1 then
    vim.schedule(function()
      callback("GoDojo engine binary not found or not executable at: " .. bin, nil)
    end)
    return
  end

  local req = vim.tbl_deep_extend("force", { action = action }, payload or {})
  local ok, req_json = pcall(vim.json.encode, req)
  if not ok then
    vim.schedule(function()
      callback("Failed to encode request payload: " .. tostring(req_json), nil)
    end)
    return
  end

  -- Use vim.system if available (Neovim >= 0.10.0), otherwise fallback to vim.loop
  if vim.system then
    vim.system({ bin }, { stdin = req_json }, function(obj)
      if obj.code ~= 0 then
        local err_msg = string.format(
          "Engine process exited with code %d. Stderr: %s",
          obj.code,
          obj.stderr or ""
        )
        vim.schedule(function()
          callback(err_msg, nil)
        end)
        return
      end

      local success, resp = pcall(vim.json.decode, obj.stdout)
      if not success then
        local err_msg = string.format(
          "Failed to decode engine JSON response: %s. Raw output: %s",
          tostring(resp),
          obj.stdout or ""
        )
        vim.schedule(function()
          callback(err_msg, nil)
        end)
        return
      end

      vim.schedule(function()
        callback(nil, resp)
      end)
    end)
  else
    -- Fallback for older Neovim versions
    vim.schedule(function()
      callback("GoDojo requires Neovim >= 0.10.0 (supporting vim.system)", nil)
    end)
  end
end

return M
