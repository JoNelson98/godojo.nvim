local Split = require("nui.split")
local Layout = require("nui.layout")
local desc_renderer = require("godojo.ui.description")
local console = require("godojo.ui.console")
local book = require("godojo.ui.book")
local overlay = require("godojo.ui.overlay")

local M = {}

M.layout = nil
M.desc_split = nil
M.editor_split = nil
M.challenge = nil
M.tmp_file = nil
M.prev_tabpage = nil
M.tabpage = nil

function M.mount(challenge)
  local keep_tab = false
  if M.layout then
    keep_tab = true
    M.unmount(keep_tab)
  end

  M.challenge = challenge
  
  if not keep_tab then
    M.prev_tabpage = vim.api.nvim_get_current_tabpage()
    -- Create a dedicated new tab
    vim.cmd("tabnew")
    M.tabpage = vim.api.nvim_get_current_tabpage()
  end

  -- Write starter code to a temp file
  M.tmp_file = vim.fn.tempname() .. "_challenge.go"
  
  -- Explicitly create missing parent directories so macOS doesn't fail
  pcall(vim.fn.mkdir, vim.fs.dirname(M.tmp_file), "p")

  local f = io.open(M.tmp_file, "w")
  if f then
    f:write(challenge.starter)
    f:close()
  end

  -- Left split (Description)
  M.desc_split = Split({
    position = "left",
    size = "38%",
    buf_options = {
      buftype = "nofile",
      bufhidden = "wipe",
      swapfile = false,
      filetype = "markdown",
    },
    win_options = {
      wrap = true,
      number = false,
      relativenumber = false,
      signcolumn = "no",
    },
  })

  -- Right split (Editor, file-backed Go buffer)
  local editor_buf = vim.fn.bufadd(M.tmp_file)
  vim.fn.bufload(editor_buf)
  
  -- Force populate the buffer with starter lines directly in memory
  local starter_lines = vim.split(challenge.starter, "\n")
  vim.api.nvim_buf_set_lines(editor_buf, 0, -1, false, starter_lines)

  -- Mount standard splits first
  M.editor_split = Split({
    position = "right",
    size = "62%",
    win_options = {
      number = true,
      relativenumber = false,
    },
  })

  -- Combine into Layout (options as first arg, box as second arg)
  M.layout = Layout(
    {
      position = "left",
      size = "100%",
    },
    Layout.Box({
      Layout.Box(M.desc_split, { size = "38%" }),
      Layout.Box(M.editor_split, { size = "62%" }),
    }, { direction = "row" })
  )

  M.layout:mount()

  -- DIRECT NATIVE BINDING: Force bind our file-backed buffer to the split window!
  vim.api.nvim_win_set_buf(M.editor_split.winid, editor_buf)

  -- Apply description content
  M.render_description()

  -- Configure editor buffer options and keymaps
  vim.api.nvim_set_option_value("buflisted", false, { buf = editor_buf })
  vim.api.nvim_set_option_value("filetype", "go", { buf = editor_buf })

  local function map(key, fn)
    vim.keymap.set("n", key, fn, { buffer = editor_buf, silent = true, nowait = true })
  end

  map("q", function()
    M.unmount()
  end)

  map("<C-r>", function()
    M.reset_code()
  end)

  map("<CR>", function()
    M.submit_solution()
  end)

  -- Ctrl + h triggers dynamic Hint display and tracks stats!
  map("<C-h>", function()
    if M.challenge and M.challenge.hints and type(M.challenge.hints) == "table" and #M.challenge.hints > 0 then
      -- Track hint usage for session metrics
      local session = require("godojo.session")
      session.hints_used = session.hints_used + 1

      local hint_text = table.concat(M.challenge.hints, "\n• ")
      vim.notify("GoDojo Hints:\n• " .. hint_text, vim.log.levels.INFO)
    else
      vim.notify("GoDojo: No hints available for this challenge.", vim.log.levels.INFO)
    end
  end)

  -- LeetCode Dojo Keymaps (customizable or standard defaults)
  map("<leader>db", function()
    book.open_book()
  end)

  map("<leader>dp", function()
    overlay.open_overlay(M.challenge)
  end)

  -- Focus on editor window
  vim.api.nvim_set_current_win(M.editor_split.winid)
end

function M.render_description()
  if not M.desc_split or not M.desc_split.bufnr then
    return
  end

  local lines = desc_renderer.get_description_lines(M.challenge)
  vim.api.nvim_set_option_value("modifiable", true, { buf = M.desc_split.bufnr })
  vim.api.nvim_buf_set_lines(M.desc_split.bufnr, 0, -1, false, lines)
  vim.api.nvim_set_option_value("modifiable", false, { buf = M.desc_split.bufnr })
end

function M.reset_code()
  if not M.editor_split or not M.editor_split.winid then
    return
  end

  local bufnr = vim.api.nvim_win_get_buf(M.editor_split.winid)

  -- Read starter code lines and extract godojo:start to godojo:end section
  local starter_lines = vim.split(M.challenge.starter, "\n")
  local start_idx, end_idx = nil, nil
  local original_editable_lines = {}

  for idx, line in ipairs(starter_lines) do
    if string.find(line, "godojo:start") then
      start_idx = idx
    elseif string.find(line, "godojo:end") then
      end_idx = idx
      break
    end
  end

  if start_idx and end_idx then
    for i = start_idx + 1, end_idx - 1 do
      table.insert(original_editable_lines, starter_lines[i])
    end
  else
    original_editable_lines = starter_lines
  end

  -- Find the range in current editor buffer
  local buf_lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
  local buf_start, buf_end = nil, nil

  for idx, line in ipairs(buf_lines) do
    if string.find(line, "godojo:start") then
      buf_start = idx
    elseif string.find(line, "godojo:end") then
      buf_end = idx
      break
    end
  end

  if buf_start and buf_end then
    -- Overwrite the editable section
    vim.api.nvim_buf_set_lines(bufnr, buf_start, buf_end - 1, false, original_editable_lines)
    vim.notify("GoDojo: Code section reset to starter.", vim.log.levels.INFO)
  else
    -- Fallback: Reset whole buffer
    vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, starter_lines)
    vim.notify("GoDojo: Entire code buffer reset to starter.", vim.log.levels.INFO)
  end
end

function M.unmount(keep_tab)
  M.unmount_internal(keep_tab)
end

function M.unmount_internal(keep_tab)
  if not M.layout then
    return
  end

  M.layout:unmount()

  if not keep_tab then
    if M.tabpage and vim.api.nvim_tabpage_is_valid(M.tabpage) then
      pcall(vim.cmd, "tabclose")
    end
  end

  -- Delete temp file safely
  if M.tmp_file then
    os.remove(M.tmp_file)
  end

  M.layout = nil
  M.desc_split = nil
  M.editor_split = nil
  M.challenge = nil
  M.tmp_file = nil
  
  if not keep_tab then
    M.tabpage = nil
    -- Restore original tab page
    if M.prev_tabpage and vim.api.nvim_prev_tabpage ~= M.tabpage and vim.api.nvim_tabpage_is_valid(M.prev_tabpage) then
      vim.api.nvim_set_current_tabpage(M.prev_tabpage)
    end
    M.prev_tabpage = nil

    -- Seamlessly open Home Dashboard when exiting a challenge!
    vim.schedule(function()
      require("godojo.ui.dashboard").mount()
    end)
  end
end

function M.submit_solution()
  if not M.editor_split or not M.editor_split.winid or not M.challenge then
    return
  end

  local bufnr = vim.api.nvim_win_get_buf(M.editor_split.winid)
  local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
  local submission = table.concat(lines, "\n")

  vim.notify("GoDojo: Submitting solution for grading...", vim.log.levels.INFO)

  -- Get any active hints used this challenge
  local session = require("godojo.session")
  local hints_used = session.hints_used

  local godojo = require("godojo")
  godojo.call_engine("grade", {
    challenge_id = M.challenge.id,
    submission = submission,
    hints_used = hints_used,
  }, function(err, resp)
    if err then
      vim.notify("GoDojo error: " .. tostring(err), vim.log.levels.ERROR)
    elseif resp and resp.status == "ok" and resp.grade then
      -- Open the Test Console overlay with results
      console.mount(resp.grade)
    else
      vim.notify("GoDojo: Unexpected grading response.", vim.log.levels.ERROR)
    end
  end)
end

return M
