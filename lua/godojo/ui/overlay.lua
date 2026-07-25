local Popup = require("nui.popup")
local godojo = require("godojo")

local M = {}

M.popup = nil

-- Appends multiline text safely into a table splitting by newlines.
local function append_text(tbl, prefix, text)
  if not text or text == "" then
    return
  end
  for _, line in ipairs(vim.split(text, "\n")) do
    table.insert(tbl, prefix .. line)
  end
end

-- Mounts the progressive disclosure Pattern Overlay floating window.
function M.open_overlay(challenge)
  if M.popup then
    M.popup:unmount()
  end

  local pattern_id = challenge.pattern_id
  if not pattern_id or pattern_id == "" or pattern_id == "N/A" then
    vim.notify("GoDojo: No design pattern associated with this challenge.", vim.log.levels.INFO)
    return
  end

  -- Query pattern data from engine
  godojo.call_engine("patterns", {}, function(err, resp)
    if err or not resp or resp.status ~= "ok" or not resp.patterns then
      vim.notify("GoDojo: Failed to load pattern card.", vim.log.levels.ERROR)
      return
    end

    local card = nil
    for _, p in ipairs(resp.patterns) do
      if p.id == pattern_id then
        card = p
        break
      end
    end

    if not card then
      vim.notify("GoDojo: Pattern card not found: " .. pattern_id, vim.log.levels.WARN)
      return
    end

    local lines = {
      "",
      " ### RECOGNITION CLUES",
      "-----------------------",
    }
    for _, clue in ipairs(card.clues or {}) do
      table.insert(lines, " * " .. clue)
    end

    table.insert(lines, "")
    table.insert(lines, " ### CORE INVARIANT")
    table.insert(lines, "--------------------")
    for _, line in ipairs(vim.split(card.invariant or "", "\n")) do
      table.insert(lines, " " .. line)
    end

    table.insert(lines, "")
    table.insert(lines, "=========================================================")
    table.insert(lines, "  [Tab] Go Skeleton      [h] Problem Hint")
    table.insert(lines, "  [q / <Esc>] Close")

    -- DYNAMIC WIDTH AND HEIGHT RESPONSIVENESS:
    -- Prevent cut-offs on different size monitors.
    local width = math.min(85, vim.o.columns - 10)
    local height = math.min(28, #lines + 2)

    M.popup = Popup({
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
      border = {
        style = "rounded",
        text = {
          top = " " .. string.upper(card.name) .. " OVERLAY ",
          top_align = "center",
        },
      },
      win_options = {
        wrap = true,
        number = false,
        relativenumber = false,
        foldenable = false,
      },
    })

    M.popup:mount()
    vim.api.nvim_buf_set_lines(M.popup.bufnr, 0, -1, false, lines)

    local function close()
      M.popup:unmount()
      M.popup = nil
    end

    local current_hint_idx = 1

    -- Key: Tab to reveal Generic Go skeleton (with full Treesitter syntax colors!)
    vim.keymap.set("n", "<Tab>", function()
      local skel_lines = {
        "",
        " ### GENERIC GO SKELETON",
        "-------------------------",
        "```go",
      }
      append_text(skel_lines, "", card.skeleton)
      table.insert(skel_lines, "```")
      
      table.insert(skel_lines, "")
      table.insert(skel_lines, "=========================================================")
      table.insert(skel_lines, "  [h] Problem Hint      [q / <Esc>] Close")

      local new_height = math.min(30, #skel_lines + 2)
      M.popup:update_layout({
        size = {
          width = width,
          height = new_height,
        }
      })
      vim.api.nvim_buf_set_lines(M.popup.bufnr, 0, -1, false, skel_lines)
    end, { buffer = M.popup.bufnr, silent = true, nowait = true })

    -- Key: h to progressive disclose problem hints
    vim.keymap.set("n", "h", function()
      if not challenge.hints or type(challenge.hints) ~= "table" or #challenge.hints == 0 then
        vim.notify("GoDojo: No hints available for this challenge.", vim.log.levels.INFO)
        return
      end
      if current_hint_idx > #challenge.hints then
        vim.notify("GoDojo: No further hints available.", vim.log.levels.INFO)
        return
      end

      local hint = challenge.hints[current_hint_idx]
      vim.notify(string.format("Hint %d/%d:\n• %s", current_hint_idx, #challenge.hints, hint), vim.log.levels.INFO)
      current_hint_idx = current_hint_idx + 1
    end, { buffer = M.popup.bufnr, silent = true, nowait = true })

    vim.keymap.set("n", "q", close, { buffer = M.popup.bufnr, silent = true, nowait = true })
    vim.keymap.set("n", "<Esc>", close, { buffer = M.popup.bufnr, silent = true, nowait = true })
  end)
end

return M
