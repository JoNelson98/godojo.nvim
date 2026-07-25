-- plugin/godojo.lua
-- Initialize the GoDojo Neovim plugin.

if vim.g.loaded_godojo == 1 then
  return
end
vim.g.loaded_godojo = 1

-- Set up default commands automatically
require("godojo.commands").setup()
