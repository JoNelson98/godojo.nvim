local M = {}

M.buf = nil
M.split = nil

function M.get_description_lines(challenge)
  local session = require("godojo.session")
  local state = require("godojo.state")
  
  -- Resolve pretty titles for chapters
  local chapter_titles = {
    http = "net/http fundamentals",
    json = "encoding/json API DTOs",
    testing = "httptest Handler Testing",
    boss = "POST /companies Boss Mission",
    arrays = "Arrays & Hashing",
    pointers = "Two Pointers",
    window = "Sliding Window",
  }
  local chapter_title = chapter_titles[challenge.chapter] or (challenge.chapter or "General")

  local progress_prefix = session.get_progress_prefix()
  if progress_prefix == "" then
    progress_prefix = "Mission: 1/1"
  else
    -- Convert "MISSION 1/8" to "Mission: 1/8"
    progress_prefix = "Mission: " .. progress_prefix:sub(9)
  end

  local dojo_name = "HTTP BACKEND DEVELOPMENT"
  if state.active_dojo == "leetcode" then
    dojo_name = "LEETCODE DSA DOJO"
  end

  local lines = {
    " PATH    : " .. dojo_name,
    " CHAPTER : " .. string.upper(chapter_title),
    " " .. string.upper(progress_prefix) .. " · " .. string.upper(challenge.title),
    "=========================================",
    "",
    " Difficulty : " .. string.rep("★", challenge.difficulty) .. string.rep("☆", 5 - challenge.difficulty),
    " Pattern    : " .. (challenge.pattern_id or "N/A"),
    " Type       : " .. (challenge.type or "full_recall"),
    "",
    " OBJECTIVE",
    "-----------",
  }

  -- Split prompt by newline
  for _, line in ipairs(vim.split(challenge.prompt, "\n")) do
    table.insert(lines, " " .. line)
  end

  table.insert(lines, "")
  table.insert(lines, " HINTS")
  table.insert(lines, "-------")
  if challenge.hints and type(challenge.hints) == "table" and #challenge.hints > 0 then
    for i, hint in ipairs(challenge.hints) do
      table.insert(lines, string.format(" %d. %s", i, hint))
    end
  else
    table.insert(lines, " No hints available. (Press <C-h> inside editor to trigger hint systems)")
  end

  table.insert(lines, "")
  table.insert(lines, " KEYMAPS")
  table.insert(lines, "---------")
  table.insert(lines, " <CR>   Submit Solution & Test")
  table.insert(lines, " <C-h>  Get Dynamic Hints")
  table.insert(lines, " <C-r>  Reset Solution block")
  table.insert(lines, " q      Save & Exit to Dashboard")

  return lines
end

return M
