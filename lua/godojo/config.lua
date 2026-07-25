local M = {}

M.defaults = {
  bin_path = nil, -- Resolved dynamically if nil
  session = {
    mode = "standard",
    target_minutes = 8,
    new_patterns_per_session = 1,
  },
  ui = {
    border = "rounded",
    show_timer = false,
    show_progress = true,
  },
  grading = {
    timeout_ms = 10000,
    show_raw_compiler_errors = true,
  },
  storage = {
    path = nil,
  },
}

M.options = {}

function M.setup(opts)
  M.options = vim.tbl_deep_extend("force", M.defaults, opts or {})
end

return M
