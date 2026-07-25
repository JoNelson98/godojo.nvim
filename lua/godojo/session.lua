local godojo = require("godojo")
local workspace = require("godojo.ui.workspace")
local dashboard = require("godojo.ui.dashboard")

local M = {}

M.active = false
M.queue = {}
M.current_index = 0
M.start_time = 0
M.hints_used = 0
M.attempts_this_session = 0

function M.start_session(mode, chapter)
  local state = require("godojo.state")
  mode = mode or "standard"
  chapter = chapter or state.active_dojo -- Defaults to active dojo (http or leetcode) if nil!

  vim.notify("GoDojo: Querying spaced queue from scheduler...", vim.log.levels.INFO)

  godojo.call_engine("start_session", { mode = mode, chapter = chapter }, function(err, resp)
    if err then
      vim.notify("GoDojo error: " .. tostring(err), vim.log.levels.ERROR)
      return
    end

    if resp and resp.status == "ok" and resp.queue and #resp.queue > 0 then
      M.active = true
      M.queue = resp.queue
      M.current_index = 1
      M.start_time = os.time()
      M.hints_used = 0
      M.attempts_this_session = 0

      -- Mount the workspace with the first challenge
      workspace.mount(M.queue[1])
    else
      vim.notify("GoDojo: No reviews or new challenges available right now in this path. You are fully caught up!", vim.log.levels.INFO)
    end
  end)
end

function M.get_progress_prefix()
  if not M.active then
    return ""
  end
  return string.format("MISSION %d/%d", M.current_index, #M.queue)
end

function M.advance()
  if not M.active then
    return
  end

  M.current_index = M.current_index + 1
  if M.current_index <= #M.queue then
    -- Clean old workspace and load next challenge
    workspace.mount(M.queue[M.current_index])
  else
    M.complete_session()
  end
end

function M.complete_session()
  M.active = false
  local duration = os.time() - M.start_time
  local minutes = math.floor(duration / 60)
  local seconds = duration % 60

  -- Close workspace tab
  workspace.unmount()

  -- Render beautiful completion dashboard
  local stats = {
    completed = #M.queue,
    total = #M.queue,
    duration_str = string.format("%d mins %d secs", minutes, seconds),
    hints = M.hints_used,
  }

  -- Mount Dashboard tab page and render the session-complete screen
  dashboard.mount()

  -- Overwrite dashboard content with the Session Completed screen
  local lines = {
    "================================================================================",
    "                            SESSION COMPLETED! 🎉                               ",
    "================================================================================",
    "",
    "  Awesome job! You finished today's training session.",
    "",
    "  Performance Summary",
    "  --------------------",
    string.format("  • Challenges completed : %d / %d", stats.completed, stats.total),
    string.format("  • Hints used           : %d", stats.hints),
    string.format("  • Approximate time     : %s", stats.duration_str),
    "",
    "  Mastery Gains",
    "  --------------",
    "  All completed patterns have been updated and scheduled for spaced review.",
    "  Keep up the momentum! Your ADHD brain thrives on repeated, daily recalls.",
    "",
    "--------------------------------------------------------------------------------",
    "  [Enter] Back to Dashboard      [q] Close GoDojo",
    "================================================================================",
  }

  local buf = dashboard.buf
  if buf and vim.api.nvim_buf_is_valid(buf) then
    vim.api.nvim_set_option_value("modifiable", true, { buf = buf })
    vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)
    vim.api.nvim_set_option_value("modifiable", false, { buf = buf })

    -- Map Enter inside completion screen to reload normal stats and render home dashboard
    vim.keymap.set("n", "<CR>", function()
      dashboard.load_and_render()
    end, { buffer = buf, silent = true, nowait = true })
  end
end

return M
