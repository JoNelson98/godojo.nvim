local dashboard = require("godojo.ui.dashboard")
local session = require("godojo.session")
local M = {}

local subcommands = {
  menu = function()
    dashboard.mount()
  end,
  start = function(mode)
    session.start_session(mode or "standard")
  end,
  run = function()
    local workspace = require("godojo.ui.workspace")
    workspace.submit_solution()
  end,
  reset = function()
    local workspace = require("godojo.ui.workspace")
    workspace.reset_code()
  end,
  desc = function()
    local workspace = require("godojo.ui.workspace")
    if workspace.desc_split then
      workspace.desc_split:toggle()
    else
      vim.notify("GoDojo: Workspace is not active.", vim.log.levels.WARN)
    end
  end,
  console = function()
    local console = require("godojo.ui.console")
    if console.layout then
      console.unmount()
    else
      vim.notify("GoDojo: Test console is not open.", vim.log.levels.INFO)
    end
  end,
  exit = function()
    dashboard.unmount()
    require("godojo.ui.workspace").unmount()
  end,
}

local function get_subcommands_keys()
  local keys = {}
  for k, _ in pairs(subcommands) do
    table.insert(keys, k)
  end
  table.sort(keys)
  return keys
end

function M.setup()
  vim.api.nvim_create_user_command("GoDojo", function(opts)
    local args = vim.split(vim.trim(opts.args or ""), "%s+")
    local sub = args[1]

    if not sub or sub == "" then
      -- Default to menu/dashboard
      dashboard.mount()
      return
    end

    local fn = subcommands[sub]
    if fn then
      fn(args[2])
    else
      vim.notify("GoDojo: Unknown subcommand '" .. sub .. "'", vim.log.levels.ERROR)
    end
  end, {
    nargs = "?",
    complete = function(arg_lead, cmd_line, cursor_pos)
      local keys = get_subcommands_keys()
      return vim.tbl_filter(function(key)
        return string.find(key, arg_lead, 1, true) == 1
      end, keys)
    end,
  })
end

return M
