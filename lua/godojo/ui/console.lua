local Popup = require("nui.popup")
local Layout = require("nui.layout")

local M = {}

M.layout = nil
M.left_popup = nil
M.right_popup = nil
M.grade_data = nil
M.prev_win = nil

-- Safely splits multiline text by newline and appends each line with a prefix.
local function append_lines(target_table, prefix, text)
  if not text or text == "" then
    return
  end
  for _, line in ipairs(vim.split(text, "\n")) do
    table.insert(target_table, prefix .. line)
  end
end

function M.mount(grade_data)
  if M.layout then
    M.unmount()
  end

  M.grade_data = grade_data
  M.prev_win = vim.api.nvim_get_current_win()

  -- Left Panel (Testcases)
  M.left_popup = Popup({
    enter = false,
    buf_options = {
      buftype = "nofile",
      bufhidden = "wipe",
    },
    win_options = {
      wrap = true,
      number = false,
      relativenumber = false,
    },
    border = {
      style = "rounded",
      text = {
        top = " TEST CASES ",
        top_align = "left",
      },
    },
  })

  -- Right Panel (Results)
  M.right_popup = Popup({
    enter = true,
    buf_options = {
      buftype = "nofile",
      bufhidden = "wipe",
    },
    win_options = {
      wrap = true,
      number = false,
      relativenumber = false,
    },
    border = {
      style = "rounded",
      text = {
        top = " RESULTS & FEEDBACK ",
        top_align = "left",
      },
    },
  })

  -- Combine into centered Layout (90% width, 78% height) - FIXED position to 50%
  M.layout = Layout(
    {
      relative = "editor",
      position = "50%",
      size = {
        width = "90%",
        height = "78%",
      },
    },
    Layout.Box({
      Layout.Box(M.left_popup, { size = "35%" }),
      Layout.Box(M.right_popup, { size = "65%" }),
    }, { direction = "row" })
  )

  M.layout:mount()

  -- Set keymaps inside both popups
  local function setup_keys(popup_obj)
    local bufnr = popup_obj.bufnr
    local function map(key, fn)
      vim.keymap.set("n", key, fn, { buffer = bufnr, silent = true, nowait = true })
    end

    map("q", function()
      M.unmount()
    end)

    map("H", function()
      vim.api.nvim_set_current_win(M.left_popup.winid)
    end)

    map("L", function()
      vim.api.nvim_set_current_win(M.right_popup.winid)
    end)

    -- Map Enter inside console to proceed to next challenge if correct!
    map("<CR>", function()
      if M.grade_data and M.grade_data.correct then
        M.unmount()
        require("godojo.session").advance()
      else
        M.unmount()
      end
    end)
  end

  setup_keys(M.left_popup)
  setup_keys(M.right_popup)

  M.render()

  -- Focus on the results pane automatically
  vim.api.nvim_set_current_win(M.right_popup.winid)
end

function M.render()
  if not M.left_popup or not M.right_popup then
    return
  end

  local left_lines = {}
  local right_lines = {}

  -- Render Left Panel (Test cases list)
  table.insert(left_lines, "  Test Case Runs:")
  table.insert(left_lines, "  -----------------------")
  if M.grade_data.compile_ok and M.grade_data.tests then
    for i, t in ipairs(M.grade_data.tests) do
      local icon = t.passed and "✓" or "✗"
      table.insert(left_lines, string.format("  [%d] %s %s", i, icon, t.name))
    end
  else
    table.insert(left_lines, "  [✗] Compilation Failed")
  end

  -- Render Right Panel (Grading Results)
  table.insert(right_lines, "  STATUS: " .. (M.grade_data.correct and "✓ PASSED" or "✗ FAILED"))
  table.insert(right_lines, "  ==================================================")
  table.insert(right_lines, "")
  table.insert(right_lines, "  Summary:")
  table.insert(right_lines, "  • Compilation        : " .. (M.grade_data.compile_ok and "✓ OK" or "✗ FAILED"))

  if M.grade_data.compile_ok then
    local passed_count = 0
    for _, t in ipairs(M.grade_data.tests) do
      if t.passed then passed_count = passed_count + 1 end
    end
    table.insert(right_lines, string.format("  • Behavioral Tests   : %d/%d passed", passed_count, #M.grade_data.tests))
  end

  if M.grade_data.compile_ok and M.grade_data.ast_checks then
    local ast_passed = 0
    for _, a in ipairs(M.grade_data.ast_checks) do
      if a.passed then ast_passed = ast_passed + 1 end
    end
    table.insert(right_lines, string.format("  • AST & Idiom Checks : %d/%d passed", ast_passed, #M.grade_data.ast_checks))
  end

  table.insert(right_lines, "")
  table.insert(right_lines, "  Feedback details:")
  table.insert(right_lines, "  ------------------")

  if not M.grade_data.compile_ok then
    -- Show compiler output safely split by newlines
    for _, fb in ipairs(M.grade_data.feedback or {}) do
      append_lines(right_lines, "  ", fb.message)
    end
  else
    -- Show failing test error messages safely split by newlines
    local has_fails = false
    for _, t in ipairs(M.grade_data.tests) do
      if not t.passed then
        has_fails = true
        table.insert(right_lines, "  • " .. t.name .. " failed:")
        append_lines(right_lines, "    ", t.message)
      end
    end

    -- Render AST Check Details safely
    if M.grade_data.ast_checks and #M.grade_data.ast_checks > 0 then
      table.insert(right_lines, "")
      table.insert(right_lines, "  AST & Idiom Verifications:")
      table.insert(right_lines, "  ----------------------------")
      for _, a in ipairs(M.grade_data.ast_checks) do
        local icon = a.passed and "  [✓] Passed" or "  [✗] FAILED"
        table.insert(right_lines, string.format("    %-12s : %s", icon, a.name))
        if not a.passed then
          has_fails = true
        end
      end
    end

    if not has_fails then
      table.insert(right_lines, "")
      table.insert(right_lines, "  All checks passed successfully! Fantastic job!")
      table.insert(right_lines, "  [Press <Enter> inside this console to advance]")
    end
  end

  -- Write lines to buffers
  vim.api.nvim_buf_set_lines(M.left_popup.bufnr, 0, -1, false, left_lines)
  vim.api.nvim_buf_set_lines(M.right_popup.bufnr, 0, -1, false, right_lines)
end

function M.unmount()
  if M.layout then
    M.unmount_internal()
  end
end

function M.unmount_internal()
  M.layout:unmount()
  M.layout = nil
  M.left_popup = nil
  M.right_popup = nil
  M.grade_data = nil

  -- Restore focus to editor window
  if M.prev_win and vim.api.nvim_win_is_valid(M.prev_win) then
    vim.api.nvim_set_current_win(M.prev_win)
  end
  M.prev_win = nil
end

return M
