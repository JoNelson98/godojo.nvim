package challenge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ValidChallenge(t *testing.T) {
	yamlStr := `
id: json.decode_request.001
title: Decode a JSON request
chapter: json
pattern_id: json.decode_request
difficulty: 3
type: full_recall
prompt: Decode r.Body into input.
starter: |
  package challenge
  import "net/http"
  func decodeRequest(r *http.Request) {}
validation:
  compile: true
  gofmt: true
  tests:
    - valid_json_populates_input
  ast_rules:
    - uses_json_new_decoder
feedback:
  missing_error: Please return the error.
hints:
  - The decoder reads from r.Body.
explanation: json.NewDecoder writes decoded fields.
test_code: |
  package challenge
  import "testing"
  func TestFoo(t *testing.T) {}
`

	c, err := Load(strings.NewReader(yamlStr))
	if err != nil {
		t.Fatalf("failed to load valid challenge: %v", err)
	}

	if err := c.Validate(); err != nil {
		t.Fatalf("validation failed for valid challenge: %v", err)
	}

	if c.ID != "json.decode_request.001" {
		t.Errorf("expected ID 'json.decode_request.001', got %q", c.ID)
	}

	if c.Difficulty != 3 {
		t.Errorf("expected difficulty 3, got %d", c.Difficulty)
	}

	if len(c.Validation.Tests) != 1 || c.Validation.Tests[0] != "valid_json_populates_input" {
		t.Errorf("expected test 'valid_json_populates_input'")
	}

	if len(c.Hints) != 1 || c.Hints[0] != "The decoder reads from r.Body." {
		t.Errorf("expected hint 'The decoder reads from r.Body.'")
	}
}

func TestFindByID_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "godojo-find-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	yamlContent := `
id: test.decode.999
title: Decode
chapter: json
pattern_id: json.decode_request
difficulty: 1
type: full_recall
prompt: Decode.
starter: package challenge
hints:
  - hint
explanation: exp
test_code: package challenge
`
	err = os.WriteFile(filepath.Join(tmpDir, "challenge.yaml"), []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	c, err := FindByID(tmpDir, "test.decode.999")
	if err != nil {
		t.Fatalf("failed to find challenge by ID: %v", err)
	}

	if c.Title != "Decode" {
		t.Errorf("expected Title 'Decode', got %q", c.Title)
	}

	// Try finding invalid ID
	_, err = FindByID(tmpDir, "non_existent")
	if err == nil {
		t.Error("expected error when searching non existent challenge ID, got nil")
	}
}

func TestValidate_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Challenge)
		wantErr string
	}{
		{
			name: "missing id",
			mutate: func(c *Challenge) {
				c.ID = ""
			},
			wantErr: "id is required",
		},
		{
			name: "missing title",
			mutate: func(c *Challenge) {
				c.Title = ""
			},
			wantErr: "title is required",
		},
		{
			name: "missing chapter",
			mutate: func(c *Challenge) {
				c.Chapter = ""
			},
			wantErr: "chapter is required",
		},
		{
			name: "missing pattern_id",
			mutate: func(c *Challenge) {
				c.PatternID = ""
			},
			wantErr: "pattern_id is required",
		},
		{
			name: "invalid difficulty",
			mutate: func(c *Challenge) {
				c.Difficulty = 6
			},
			wantErr: "difficulty must be between 0 and 5",
		},
		{
			name: "invalid type",
			mutate: func(c *Challenge) {
				c.Type = "unknown_type"
			},
			wantErr: "invalid type",
		},
		{
			name: "missing prompt",
			mutate: func(c *Challenge) {
				c.Prompt = ""
			},
			wantErr: "prompt is required",
		},
		{
			name: "missing starter",
			mutate: func(c *Challenge) {
				c.Starter = ""
			},
			wantErr: "starter code is required",
		},
		{
			name: "missing hints",
			mutate: func(c *Challenge) {
				c.Hints = nil
			},
			wantErr: "at least one hint is required",
		},
		{
			name: "missing explanation",
			mutate: func(c *Challenge) {
				c.Explanation = ""
			},
			wantErr: "explanation is required",
		},
		{
			name: "missing test_code",
			mutate: func(c *Challenge) {
				c.TestCode = ""
			},
			wantErr: "test_code is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := &Challenge{
				ID:          "test.001",
				Title:       "Test",
				Chapter:     "syntax",
				PatternID:   "test_pattern",
				Difficulty:  1,
				Type:        "token_completion",
				Prompt:      "Do something.",
				Starter:     "package test",
				Hints:       []string{"hint"},
				Explanation: "exp",
				TestCode:    "package test",
			}
			tt.mutate(base)
			err := base.Validate()
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}
