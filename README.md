# 🥷 GoDojo.nvim (v0.1)

A gamified, active-recall training environment inside Neovim for drilling **Go Syntax, Web HTTP Handlers, JSON DTOs, and LeetCode DSA Patterns**.

GoDojo doesn't teach through passive reading. It forces you to write real Go code, formatting, compiling, and running sandboxed tests natively inside Neovim while enforcing structural best practices using **Go AST parsing** and gamifying your memory consolidation using a local **SQLite Spaced Repetition engine**.

---

## 🚀 Key Features

* **Dual Learning Dojos**: 
  * **HTTP Backend Dojo**: Master standard library HTTP handler signatures, redirection, custom JSON decoders, header validation, and `httptest` suites.
  * **LeetCode DSA Dojo**: Drill core Data Structures & Algorithms patterns in Go across **Arrays & Hashing**, **Two Pointers**, and **Sliding Window** tracks.
* **Searchable Book of Patterns (`<leader>db` / `:GoDojo book`)**: Search and open beautifully structured, syntax-colored cards explaining recognition clues, core invariants, complexes, common mistakes, and compilable generic Go templates for key patterns.
* **Progressive Disclosure Overlay (`<leader>dp` / `:GoDojo overlay`)**: Toggle a floating panel while solving a problem to show recognition clues and invariants, press `<Tab>` to reveal the Go skeleton, or hit `h` for progressive hints.
* **Double-Pane Test Console**: Displays formatting warnings, real compiler warnings, Go test-suite logs, and Go AST code verifications side-by-side with color-coded layouts.
* **SQLite Spaced Repetition**: Schedules reviews exponentially (`1d -> 3d -> 7d -> 14d -> 30d`) on clean recalls, decays mastery score by 50% on failures, and tracks weak spots dynamically.

---

## 🎨 Interactive Dashboards

### Main Dashboard (`:GoDojo`)

Switches dynamically between tracks with the **`d` key**:

```text
================================================================================
                                 GO DOJO (v0.1)                                 
================================================================================

  Dojo              :  LeetCode DSA Dojo
  Current Chapter   :  Sliding Window
  Session Estimate  :  7 minutes
  Reviews Due       :  2 challenges
  New Patterns      :  1 (variable-sliding-window)
  Stable Mastery    :  [████████░░░░░░░░░░░░] 40%

  Weak Patterns
  ---------------
  • arrays-and-hashing
  • two-pointers

  Recent Progress
  ----------------
  • Total recorded attempts: 41

--------------------------------------------------------------------------------
  [Enter] Start Training Session      [c] Select & View Learning Paths
  [r]     Reviews Only (due items)    [s] View Statistics Popup
  [w]     Weak-Pattern Training       [d] Switch Active Dojo
  [p]     Select & Load Problem       [q] Exit GoDojo
================================================================================
  Go Engine Connection: Connected successfully
================================================================================
```

---

## 📦 Installation & Setup

GoDojo relies on a compiled Go backend. Setup is incredibly simple using **Lazy.nvim** with compile-on-install.

### Prerequisites
* Go compiler (1.20+) installed on your machine (`go version`)
* SQLite development headers (native on macOS, `sqlite3` on Linux/WSL)
* Neovim (0.9.0+)

### Lazy.nvim Plugin Configuration

Have your friend add this block to their Neovim plugin specs:

```lua
return {
  "JoNelson98/godojo.nvim",
  dependencies = { 
    "MunifTanjim/nui.nvim" -- Required for Dashboard & Split windows
  },
  -- Automatically builds/compiles the Go background engine upon install and updates!
  build = "make build", 
  keys = {
    { "<leader>gd", "<cmd>GoDojo<cr>", desc = "Open GoDojo Dashboard" }
  },
  config = function()
    -- Optional: your configuration overrides go here
  end,
}
```

---

## 🎹 Suggested Keymaps

| Keymap | Location | Action |
|:---:|:---:|:---|
| `<leader>gd` | Normal mode | Toggle main GoDojo Dashboard |
| `<CR>` | Dashboard | Start Scheduled Training Session |
| `c` | Dashboard | View dynamic progress bars & launch targeted chapters |
| `p` | Dashboard | Open a searchable fuzzy picker to run any individual problem |
| `d` | Dashboard | Swap Dojo tracks (HTTP Backend ↔ LeetCode DSA) |
| `r` | Dashboard | Reviews Only mode (pulls due spaced-repetition cards) |
| `w` | Dashboard | Drill your lowest-mastered Weak Patterns |
| `s` | Dashboard | Open Learner Statistics overlay |
| **`q`** | Dashboard | Exit GoDojo & restore active Neovim workspaces |
| **`<CR>`** | Editor Workspace | Submit current solution & compile/grade |
| **`q`** | Editor Workspace | Save code & return directly to Home Dashboard |
| **`<C-h>`** | Editor Workspace | Request problem hints |
| **`<C-r>`** | Editor Workspace | Reset editable code section back to starter |
| **`<leader>db`** | Editor Workspace | Search and browse the **Book of Patterns** |
| **`<leader>dp`** | Editor Workspace | Open the **Progressive Disclosure Pattern Overlay** |

---

## 🛠️ Content Customization Guide

GoDojo is fully data-driven. You can add new chapters, challenges, or pattern cards easily by adding `.yaml` files without recompiling or writing any Go code!

### 1. How to add a Custom Coding Challenge
Create a new file under `content/leetcode/arrays/my-problem.yaml` using this schema:

```yaml
id: leetcode.my_problem.001          # Must be globally unique
title: My Custom DSA Challenge       # Displays as active mission title
chapter: arrays                      # Path/Category matching your Dojo
pattern_id: arrays-and-hashing       # Primary pattern ID from cards
difficulty: 2                        # 1 to 5 stars
type: full_recall                    # full_recall, bug_repair, etc.
role: lesson                         # "lesson" or "boss" (boss locks until lessons pass)
primary_pattern: arrays-and-hashing

prompt: |
  Write a function to search an element in O(1) time.
  (You can use full multiline Markdown here!)

starter: |
  package challenge

  func myCustomSearch(nums []int, target int) bool {
      // godojo:start
      // Write your solution here.
      // godojo:end
      return false
  }

validation:
  compile: true                      # Formats with gofmt and compiles
  gofmt: true
  tests:                             # Lists test function names in test_code
    - TestMyCustomSearch_Visible

hints:
  - "Quote your hints to prevent YAML parsing crashes."
  - "Always store values in map keys for O(1) hashing lookups."

explanation: |
  Describe the optimal syntax or design pattern details here for ADHD learners.

test_code: |
  package challenge

  import "testing"

  func TestMyCustomSearch_Visible(t *testing.T) {
      if !myCustomSearch([]int{1, 2, 3}, 2) {
          t.Error("Expected true, got false")
      }
  }
```

### 2. How to add a Custom Pattern Card
Create a new YAML file under `content/patterns/my-pattern.yaml`:

```yaml
id: my-pattern                       # Must match the pattern_id of challenges
name: Custom Backtracking            # Pattern Title
chapter: backtracking                # Learning track chapter
solves: |
  Enumerate all valid path combinations.
clues:
  - Choosing elements, checking validity, and undoing choices.
invariant: |
  The backtrack recursive branch preserves previous step decisions.
skeleton: |
  // Compilable Go skeleton template
  func backtrackTemplate(candidates []int) {
      var backtrack func(start int, path []int)
      backtrack = func(start int, path []int) {
          if conditionMet() {
              return
          }
          for i := start; i < len(candidates); i++ {
              // Choose element
              backtrack(i + 1, append(path, candidates[i]))
              // Undo choice (implicit on return)
          }
      }
      backtrack(0, []int{})
  }
time_complexity: O(2^n)
space_complexity: O(n)
mistakes:
  - Forgetting base case causing stack overflows.
related:
  - dfs-traversal
```

### 3. Validating your changes
Run the built-in validator command to verify that all directories, challenges, tests, and syntax rules are correct:
```bash
./bin/godojo validate-content
```

---

## 🥷 Developed locally with ❤️
Enjoy training inside Neovim. Keep drilling, and make Go your native muscle memory!
