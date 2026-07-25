package pattern

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPattern_LoadValid(t *testing.T) {
	yamlStr := `
id: variable-sliding-window
name: Variable Sliding Window
chapter: sliding-window
solves: Longest or shortest valid ranges.
clues:
  - contiguous array region
  - longest or shortest valid range
invariant: Window remains valid.
skeleton: |
  for r < len(values) {
      add(values[r])
      for invalid() {
          remove(values[l])
          l++
      }
  }
time_complexity: O(n)
space_complexity: O(1)
mistakes:
  - wrong bounds
`

	p, err := Load(strings.NewReader(yamlStr))
	if err != nil {
		t.Fatalf("failed to load pattern card: %v", err)
	}

	if p.ID != "variable-sliding-window" {
		t.Errorf("expected ID 'variable-sliding-window', got %q", p.ID)
	}
	if p.Name != "Variable Sliding Window" {
		t.Errorf("expected Name 'Variable Sliding Window', got %q", p.Name)
	}
	if len(p.Clues) != 2 {
		t.Errorf("expected 2 clues, got %d", len(p.Clues))
	}
	if p.TimeComplexity != "O(n)" {
		t.Errorf("expected O(n), got %q", p.TimeComplexity)
	}
}

func TestPattern_FindByID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "godojo-patterns-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	yamlStr := `
id: prefix-sum
name: Prefix Sum
chapter: arrays
solves: Range queries
clues:
  - ranges
invariant: sum is cached
skeleton: skeleton
time_complexity: O(1)
space_complexity: O(n)
mistakes: []
`
	err = os.WriteFile(filepath.Join(tmpDir, "prefix-sum.yaml"), []byte(yamlStr), 0644)
	if err != nil {
		t.Fatalf("failed to write pattern card: %v", err)
	}

	p, err := FindByID(tmpDir, "prefix-sum")
	if err != nil {
		t.Fatalf("failed to find by ID: %v", err)
	}

	if p.Name != "Prefix Sum" {
		t.Errorf("expected Name 'Prefix Sum', got %q", p.Name)
	}
}
