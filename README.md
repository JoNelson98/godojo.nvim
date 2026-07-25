# godojo.nvim

A gamified, bite-sized training environment inside Neovim for memorizing Go syntax, idioms, standard packages, and HTTP web patterns. Built specifically for active recall and ADHD-friendly development sessions.

## 🚀 Product Overview

`godojo.nvim` helps you drill Go HTTP patterns until they are coded in your muscle memory. Instead of watching lectures or answering multiple-choice questions, you repeatedly type, repair, complete, and reconstruct real Go code.

### The Learning Loop

1. **Read Objective**: Read a short, highly focused objective.
2. **Type Code**: Write 1 to 8 lines of Go in a standard file-backed editor split.
3. **Run Validation**: Press `<CR>` to trigger formatting, parsing (AST inspection), compilation, and testing.
4. **Get Feedback**: Receive precise feedback pointing to the exact line/column with virtual text annotations.
5. **Repeat & Solidify**: Strengthen patterns over time with our built-in spaced repetition scheduler.

---

## 🎨 UI & Layout Mockups

### 1. Dashboard (`:GoDojo`)
A full-tab interactive screen displaying your progress, mastery, and pending reviews.

```text
================================================================================
                                 GO DOJO (v0.1)                                 
================================================================================

  Path              :  HTTP Backend Development
  Current Chapter   :  Chapter 6: JSON APIs
  Session Estimate  :  7 minutes
  Reviews Due       :  4 challenges
  New Patterns      :  1 (json.decode_request)
  Stable Mastery    :  [████████░░░░░░░░░░░░] 38%

  Weak Patterns
  ---------------
  • pointer.receiver         - 20% mastery (Requires 3 more clean recalls)
  ...
```

### 2. Challenge Workspace
Dedicated tab page with a 38% description split on the left and a 62% file-backed Go buffer on the right.

```text
========================================|=======================================
 MISSION 3/8 · JSON DECODING            | package challenge
----------------------------            | 
 Decode the request body into input.    | func decodeRequest(r *http.Request) (
                                        |     CreateCompanyRequest, error,
 Target Pattern : json.Decoder          | ) {
 ...                                    |     // godojo:start
                                        |     // Write your solution here.
                                        |     // godojo:end
========================================|=======================================
```

---

## 📋 Requirements

- **Neovim** >= 0.10.0 (supporting `vim.system`)
- **Go** >= 1.22 (to compile the training engine)
- **nui.nvim** (for layout splits and test console overlays)

---

## 📦 Installation

### Using [lazy.nvim](https://github.com/folke/lazy.nvim)

```lua
{
  "godojo/godojo.nvim",
  dependencies = {
    "MunifTanjim/nui.nvim",
  },
  build = "make build",
  config = function()
    require("godojo").setup({
      session = {
        mode = "standard",
        target_minutes = 8,
      }
    })
  end,
}
```

### Development Setup

Clone the repository locally:

```bash
git clone https://github.com/godojo/godojo.nvim.git ~/code/godojo.nvim
cd ~/code/godojo.nvim
make build
```

Then configure lazy to point to your local directory:

```lua
{
  dir = "~/code/godojo.nvim",
  dependencies = {
    "MunifTanjim/nui.nvim",
  },
  config = function()
    require("godojo").setup()
  end,
}
```

---

## 🛠️ Commands & Mappings

### Neovim Commands

| Command | Action |
| :--- | :--- |
| `:GoDojo` | Opens the main menu/dashboard tab |
| `:GoDojo start` | Begins a new training session |
| `:GoDojo run` | Opens the test console for the current challenge |
| `:GoDojo reset` | Resets the editable solution code blocks |
| `:GoDojo exit` | Exits GoDojo and cleans up all active buffers |

### Default Mappings (Dashboard)

- `<CR>`: Start training / Test connection
- `q`: Close/Exit GoDojo

### Default Mappings (Workspace Editor)

- `<CR>`: Submit/validate code
- `<C-h>`: Toggle hint
- `<C-t>`: Toggle test console overlay

---

## ⚙️ Configuration

```lua
require("godojo").setup({
  session = {
    mode = "standard", -- "quick", "standard", "hyperfocus"
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
  }
})
```

---

## 🧑‍💻 How to Run Tests

### Go Engine Tests

```bash
make test
```

---

## 🗺️ Roadmap

- [x] **Milestone 1**: Repository Bootstrap & Ping Protocol
- [ ] **Milestone 2**: Challenge Loading & Validation YAML Schema
- [ ] **Milestone 3**: Dual Split Scratch buffer Workspace
- [ ] **Milestone 4**: AST Validation Engine & test compiler
- [ ] **Milestone 5**: SQLite Progress Tracking
- [ ] **Milestone 6**: Spaced Review Scheduler & Queue
- [ ] **Milestone 7**: Curated 30 challenge curriculum
- [ ] **Milestone 8**: POST `/companies` Boss Mission
