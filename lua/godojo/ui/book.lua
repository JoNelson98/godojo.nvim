local godojo = require("godojo")
local Popup = require("nui.popup")
local M = {}

-- Appends multiline text safely into a table splitting by newlines.
local function append_text(tbl, prefix, text)
  if not text or text == "" then
    return
  end
  for _, line in ipairs(vim.split(text, "\n")) do
    table.insert(tbl, prefix .. line)
  end
end

-- Opens the searchable Book of Patterns using vim.ui.select.
function M.open_book()
  godojo.call_engine("patterns", {}, function(err, resp)
    if err or not resp or resp.status ~= "ok" or not resp.patterns then
      vim.notify("GoDojo: Failed to load pattern cards.", vim.log.levels.ERROR)
      return
    end

    local options = {}
    local card_map = {}

    for idx, card in ipairs(resp.patterns) do
      local display = string.format("%d. %s (%s)", idx, card.name, card.chapter)
      table.insert(options, display)
      card_map[display] = card
    end

    vim.ui.select(options, {
      prompt = "Search & Select DSA Pattern:",
    }, function(choice)
      if choice then
        local card = card_map[choice]
        M.show_card_popup(card)
      end
    end)
  end)
end

-- Opens a centered rounded Popup showing the full details of a PatternCard.
function M.show_card_popup(card)
  local lines = {
    "",
    " # PATTERN : " .. string.upper(card.name),
    " ## CHAPTER : " .. string.upper(card.chapter),
    "===========================================================",
    "",
    " ### WHAT IT SOLVES",
    "--------------------",
  }

  append_text(lines, " ", card.solves)

  table.insert(lines, "")
  table.insert(lines, " ### RECOGNITION CLUES")
  table.insert(lines, "-----------------------")
  for _, clue in ipairs(card.clues or {}) do
    table.insert(lines, " * " .. clue)
  end

  table.insert(lines, "")
  table.insert(lines, " ### CORE INVARIANT")
  table.insert(lines, "--------------------")
  append_text(lines, " ", card.invariant)

  table.insert(lines, "")
  table.insert(lines, " ### GENERIC GO SKELETON")
  table.insert(lines, "-------------------------")
  table.insert(lines, "```go")
  append_text(lines, "", card.skeleton)
  table.insert(lines, "```")

  table.insert(lines, "")
  table.insert(lines, " ### COMPLEXITY")
  table.insert(lines, "----------------")
  table.insert(lines, " * **Time** : " .. (card.time_complexity or "O(n)"))
  table.insert(lines, " * **Space**: " .. (card.space_complexity or "O(1)"))

  table.insert(lines, "")
  table.insert(lines, " ### COMMON MISTAKES")
  table.insert(lines, "---------------------")
  for _, mistake in ipairs(card.mistakes or {}) do
    table.insert(lines, " * " .. mistake)
  end

  if card.related and #card.related > 0 then
    table.insert(lines, "")
    table.insert(lines, " ### RELATED PATTERNS")
    table.insert(lines, "----------------------")
    table.insert(lines, " * " .. table.concat(card.related, ", "))
  end

  table.insert(lines, "")
  table.insert(lines, "===========================================================")
  table.insert(lines, "  [q / <Esc>] Close Card      [b] Back to Searchable Book")

  -- DYNAMIC WIDTH AND HEIGHT RESPONSIVENESS:
  -- Prevent cut-offs on different size monitors.
  local width = math.min(90, vim.o.columns - 10)
  local height = math.min(32, #lines + 2)

  local popup = Popup({
    relative = "editor",
    position = "50%",
    size = {
      width = width,
      height = height,
    },
    enter = true,
    focusable = true,
    buf_options = {
      buftype = "nofile",
      bufhidden = "wipe",
      filetype = "markdown", -- Enables native markdown treesitter syntax coloring!
    },
    win_options = {
      wrap = true,
      number = false,
      relativenumber = false,
      foldenable = false,
    },
    border = {
      style = "rounded",
      text = {
        top = " " .. string.upper(card.name) .. " CARD ",
        top_align = "center",
      },
    },
  })

  popup:mount()
  vim.api.nvim_buf_set_lines(popup.bufnr, 0, -1, false, lines)

  local function close()
    popup:unmount()
  end

  -- Setup navigation keys
  vim.keymap.set("n", "q", close, { buffer = popup.bufnr, silent = true, nowait = true })
  vim.keymap.set("n", "<Esc>", close, { buffer = popup.bufnr, silent = true, nowait = true })
  
  vim.keymap.set("n", "b", function()
    popup:unmount()
    vim.schedule(M.open_book)
  end, { buffer = popup.bufnr, silent = true, nowait = true })
end

return M
