local godojo = require("godojo")
local state = require("godojo.state")
local M = {}

M.buf = nil
M.win = nil
M.tabpage = nil
M.prev_tabpage = nil

local function get_ascii_dashboard(stats, engine_status, active_chapter)
  stats = stats or {
    total_attempts = 0,
    reviews_due = 0,
    stable_mastery = 0.0,
    weak_patterns = {}
  }
  engine_status = engine_status or "Connected successfully"
  active_chapter = active_chapter or "net/http fundamentals"

  local pct = math.floor((stats.stable_mastery or 0.0) * 100)
  local filled = math.floor((stats.stable_mastery or 0.0) * 20)
  if filled > 20 then filled = 20 end
  local bar = string.rep("█", filled) .. string.rep("░", 20 - filled)

  local dojo_display = "HTTP Backend Dojo"
  if state.active_dojo == "leetcode" then
    dojo_display = "LeetCode DSA Dojo"
  end

  local lines = {
    "================================================================================",
    "                                 GO DOJO (v0.1)                                 ",
    "================================================================================",
    "",
    "  Dojo              :  " .. dojo_display,
    "  Current Chapter   :  " .. active_chapter,
    "  Session Estimate  :  7 minutes",
    string.format("  Reviews Due       :  %d challenges", stats.reviews_due or 0),
    "  New Patterns      :  1 (json.decode_request)",
    string.format("  Stable Mastery    :  [%s] %d%%", bar, pct),
    "",
    "  Weak Patterns",
    "  ---------------",
  }

  if stats.weak_patterns and type(stats.weak_patterns) == "table" and #stats.weak_patterns > 0 then
    for _, pat in ipairs(stats.weak_patterns) do
      table.insert(lines, "  • " .. pat)
    end
  else
    table.insert(lines, "  • No weak patterns! Excellent!")
  end

  local next_lines = {
    "",
    "  Recent Progress",
    "  ----------------",
    string.format("  • Total recorded attempts: %d", stats.total_attempts or 0),
    "",
    "--------------------------------------------------------------------------------",
    "  [Enter] Start Training Session      [c] Select & View Learning Paths",
    "  [r]     Reviews Only (due items)    [s] View Statistics Popup",
    "  [w]     Weak-Pattern Training       [d] Switch Active Dojo",
    "  [p]     Select & Load Problem       [q] Exit GoDojo",
    "================================================================================",
    "  Go Engine Connection: " .. engine_status,
    "================================================================================",
  }

  for _, l in ipairs(next_lines) do
    table.insert(lines, l)
  end

  return lines
end

function M.mount()
  if M.tabpage and vim.api.nvim_tabpage_is_valid(M.tabpage) then
    M.focus()
    return
  end

  M.prev_tabpage = vim.api.nvim_get_current_tabpage()

  -- Create dedicated new tab
  vim.cmd("tabnew")
  M.tabpage = vim.api.nvim_get_current_tabpage()
  M.win = vim.api.nvim_get_current_win()

  -- Create scratch buffer
  M.buf = vim.api.nvim_create_buf(false, true)
  vim.api.nvim_buf_set_name(M.buf, "GoDojo Dashboard")
  vim.api.nvim_win_set_buf(M.win, M.buf)

  -- Configure buffer options
  vim.api.nvim_set_option_value("buftype", "nofile", { buf = M.buf })
  vim.api.nvim_set_option_value("bufhidden", "wipe", { buf = M.buf })
  vim.api.nvim_set_option_value("swapfile", false, { buf = M.buf })
  vim.api.nvim_set_option_value("buflisted", false, { buf = M.buf })
  vim.api.nvim_set_option_value("filetype", "godojo-dashboard", { buf = M.buf })

  -- Setup keymaps
  local function map(key, fn)
    vim.keymap.set("n", key, fn, { buffer = M.buf, silent = true, nowait = true })
  end

  map("q", function()
    M.unmount()
  end)

  map("<CR>", function()
    M.unmount()
    vim.cmd("GoDojo start")
  end)

  map("r", function()
    M.unmount()
    vim.cmd("GoDojo start reviews")
  end)

  map("w", function()
    M.unmount()
    vim.cmd("GoDojo start standard")
  end)

  -- Interactive Dojo switcher!
  map("d", function()
    local options = {
      "HTTP Backend Dojo",
      "LeetCode DSA Dojo",
    }
    vim.ui.select(options, {
      prompt = "Switch Active GoDojo Track:",
    }, function(choice)
      if choice == "HTTP Backend Dojo" then
        state.active_dojo = "http"
        M.load_and_render()
      elseif choice == "LeetCode DSA Dojo" then
        state.active_dojo = "leetcode"
        M.load_and_render()
      end
    end)
  end)

  -- Select and load an individual challenge directly!
  map("p", function()
    godojo.call_engine("list_challenges", { chapter = state.active_dojo }, function(err, resp)
      if err or not resp or resp.status ~= "ok" or not resp.queue or #resp.queue == 0 then
        vim.notify("GoDojo: Failed to load challenges list.", vim.log.levels.ERROR)
        return
      end

      local options = {}
      local challenge_map = {}

      for _, c in ipairs(resp.queue) do
        local display = string.format("[%s] %s (%s)", string.upper(c.chapter), c.title, string.rep("★", c.difficulty))
        table.insert(options, display)
        challenge_map[display] = c
      end

      vim.ui.select(options, {
        prompt = "Select specific GoDojo challenge to solve:",
      }, function(choice)
        if choice then
          local selected_c = challenge_map[choice]
          M.unmount()
          require("godojo.ui.workspace").mount(selected_c)
        end
      end)
    end)
  end)

  -- The updated interactive c key selection path popup picker!
  map("c", function()
    godojo.call_engine("curriculum", { chapter = state.active_dojo }, function(err, resp)
      if err or not resp or resp.status ~= "ok" or not resp.curriculum then
        vim.notify("GoDojo: Failed to load curriculum.", vim.log.levels.ERROR)
        return
      end

      local options = {}
      local chapter_map = {}

      for _, path in ipairs(resp.curriculum) do
        local pct = math.floor((path.percent or 0.0) * 100)
        local filled = math.floor((path.percent or 0.0) * 10)
        local bar = string.rep("█", filled) .. string.rep("░", 10 - filled)
        local display = string.format("[%s] %3d%%  %s (%d/%d done)", bar, pct, path.title, path.completed, path.total)
        
        table.insert(options, display)
        chapter_map[display] = path.chapter
      end

      vim.ui.select(options, {
        prompt = "Select GoDojo Learning Path to train:",
      }, function(choice)
        if choice then
          local chapter = chapter_map[choice]
          M.unmount()
          -- Launches session focused specifically on this chapter!
          require("godojo.session").start_session("standard", chapter)
        end
      end)
    end)
  end)

  map("s", function()
    local Popup = require("nui.popup")
    godojo.call_engine("stats", {}, function(err, resp)
      local stats_lines = {}
      local height = 8
      if resp and resp.status == "ok" and resp.stats then
        height = 11
        stats_lines = {
          "",
          "  GoDojo Learning Statistics:",
          "  -----------------------------",
          string.format("  • Total recorded attempts : %d", resp.stats.total_attempts),
          string.format("  • Correct solves          : %d", resp.stats.correct_attempts),
          string.format("  • Overall success rate    : %.1f%%", resp.stats.success_rate * 100),
          string.format("  • Reviews due today       : %d", resp.stats.reviews_due),
          string.format("  • Stable mastered patterns: %d%%", math.floor(resp.stats.stable_mastery * 100)),
          "",
          "  [q / <Esc>] Close Statistics"
        }
      else
        stats_lines = {
          "",
          "  No statistics recorded yet. Start training!",
          "",
          "  [q / <Esc>] Close"
        }
      end

      local popup = Popup({
        relative = "editor",
        position = "50%",
        size = {
          width = 50,
          height = height,
        },
        enter = true,
        focusable = true,
        buf_options = {
          buftype = "nofile",
          bufhidden = "wipe",
        },
        border = {
          style = "rounded",
          text = {
            top = " LEARNER STATISTICS ",
            top_align = "center",
          },
        },
      })
      popup:mount()
      vim.api.nvim_buf_set_lines(popup.bufnr, 0, -1, false, stats_lines)

      local function close()
        popup:unmount()
      end
      vim.keymap.set("n", "q", close, { buffer = popup.bufnr, silent = true, nowait = true })
      vim.keymap.set("n", "<Esc>", close, { buffer = popup.bufnr, silent = true, nowait = true })
    end)
  end)

  -- Load stats and render
  M.load_and_render()
end

function M.show()
  if M.tabpage and vim.api.nvim_tabpage_is_valid(M.tabpage) then
    M.focus()
  end
end

function M.hide()
  M.unmount()
end

function M.toggle()
  if M.tabpage and vim.api.nvim_tabpage_is_valid(M.tabpage) then
    M.unmount()
  else
    M.mount()
  end
end

function M.focus()
  if M.win and vim.api.nvim_win_is_valid(M.win) then
    vim.api.nvim_set_current_win(M.win)
  end
end

function M.load_and_render()
  if not M.buf or not vim.api.nvim_buf_is_valid(M.buf) then
    return
  end

  M.render(nil, "Connecting to SQLite database...")

  -- First, query curriculum to find the active uncompleted chapter for this Dojo
  godojo.call_engine("curriculum", { chapter = state.active_dojo }, function(err, resp)
    if err then
      M.render(nil, "Failed: " .. tostring(err), "Error")
      return
    end

    local active_chapter_title = "net/http fundamentals"
    if state.active_dojo == "leetcode" then
      active_chapter_title = "Arrays & Hashing"
    end

    if resp and resp.status == "ok" and resp.curriculum then
      for _, path in ipairs(resp.curriculum) do
        if path.percent < 1.0 then
          active_chapter_title = path.title
          break
        end
      end
    end

    -- Next, query statistics and render
    godojo.call_engine("stats", {}, function(err2, resp2)
      if err2 then
        M.render(nil, "Failed: " .. tostring(err2), active_chapter_title)
        return
      end

      if resp2 and resp2.status == "ok" and resp2.stats then
        M.render(resp2.stats, "Connected successfully", active_chapter_title)
      else
        M.render(nil, "Connected successfully", active_chapter_title)
      end
    end)
  end)
end

function M.render(stats, engine_status, active_chapter)
  if not M.buf or not vim.api.nvim_buf_is_valid(M.buf) then
    return
  end

  local lines = get_ascii_dashboard(stats, engine_status, active_chapter)

  -- Temporarily make modifiable to write lines
  vim.api.nvim_set_option_value("modifiable", true, { buf = M.buf })
  vim.api.nvim_buf_set_lines(M.buf, 0, -1, false, lines)
  vim.api.nvim_set_option_value("modifiable", false, { buf = M.buf })
end

function M.clear()
  if M.buf and vim.api.nvim_buf_is_valid(M.buf) then
    vim.api.nvim_set_option_value("modifiable", true, { buf = M.buf })
    vim.api.nvim_buf_set_lines(M.buf, 0, -1, false, {})
    vim.api.nvim_set_option_value("modifiable", false, { buf = M.buf })
  end
end

function M.unmount()
  if M.buf and vim.api.nvim_buf_is_valid(M.buf) then
    M.buf_deleted = true
    vim.api.nvim_buf_delete(M.buf, { force = true })
  end

  if M.tabpage and vim.api.nvim_tabpage_is_valid(M.tabpage) then
    pcall(vim.cmd, "tabclose")
  end

  M.buf = nil
  M.win = nil
  M.tabpage = nil

  -- Restore previous tab page
  if M.prev_tabpage and vim.api.nvim_prev_tabpage ~= M.tabpage and vim.api.nvim_tabpage_is_valid(M.prev_tabpage) then
    vim.api.nvim_set_current_tabpage(M.prev_tabpage)
  end
  M.prev_tabpage = nil
end

return M
