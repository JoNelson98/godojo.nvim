package challenge

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Validation defines requirements for grading the challenge.
type Validation struct {
	Compile  bool     `yaml:"compile" json:"compile"`
	Gofmt    bool     `yaml:"gofmt" json:"gofmt"`
	Tests    []string `yaml:"tests" json:"tests"`
	ASTRules []string `yaml:"ast_rules" json:"ast_rules"`
}

// Variant defines custom overrides for generating challenge variants.
type Variant map[string]interface{}

// Challenge represents the schema of a data-driven training challenge.
type Challenge struct {
	ID                string            `yaml:"id" json:"id"`
	Title             string            `yaml:"title" json:"title"`
	Chapter           string            `yaml:"chapter" json:"chapter"`
	PatternID         string            `yaml:"pattern_id" json:"pattern_id"`
	Difficulty        int               `yaml:"difficulty" json:"difficulty"`
	Type              string            `yaml:"type" json:"type"`
	Prompt            string            `yaml:"prompt" json:"prompt"`
	Starter           string            `yaml:"starter" json:"starter"`
	Validation        Validation        `yaml:"validation" json:"validation"`
	Feedback          map[string]string `yaml:"feedback" json:"feedback"`
	Hints             []string          `yaml:"hints" json:"hints"`
	Explanation       string            `yaml:"explanation" json:"explanation"`
	TestCode          string            `yaml:"test_code" json:"test_code"`
	Role              string            `yaml:"role,omitempty" json:"role,omitempty"` // "lesson" or "boss"
	PrimaryPattern    string            `yaml:"primary_pattern,omitempty" json:"primary_pattern,omitempty"`
	SecondaryPatterns []string          `yaml:"secondary_patterns,omitempty" json:"secondary_patterns,omitempty"`
	Tags              []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	Prerequisites     []string          `yaml:"prerequisites,omitempty" json:"prerequisites,omitempty"`
	Variants          []Variant         `yaml:"variants,omitempty" json:"variants,omitempty"`
}

// Load decodes a Challenge struct from a YAML reader.
func Load(r io.Reader) (*Challenge, error) {
	var c Challenge
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("failed to decode YAML challenge: %w", err)
	}
	return &c, nil
}

// LoadFile decodes a Challenge from a file path.
func LoadFile(filePath string) (*Challenge, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open challenge file: %w", err)
	}
	defer f.Close()
	return Load(f)
}

// FindByID recursively searches the specified directory for a challenge with the matching ID.
func FindByID(dirPath, id string) (*Challenge, error) {
	var found *Challenge
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".yaml") && !strings.HasSuffix(strings.ToLower(info.Name()), ".yml") {
			return nil
		}
		c, err := LoadFile(path)
		if err != nil {
			return nil // Skip files with loading errors
		}
		if c.ID == id {
			found = c
			return filepath.SkipAll // Stop scanning immediately
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, fmt.Errorf("challenge with ID %q not found in %s", id, dirPath)
	}
	return found, nil
}

// Validate checks that the challenge meets all quality guidelines, required fields,
// unique IDs, and correct data-types.
func (c *Challenge) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return fmt.Errorf("id is required")
	}
	if strings.TrimSpace(c.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if strings.TrimSpace(c.Chapter) == "" {
		return fmt.Errorf("chapter is required")
	}
	if strings.TrimSpace(c.PatternID) == "" {
		return fmt.Errorf("pattern_id is required")
	}
	if c.Difficulty < 0 || c.Difficulty > 5 {
		return fmt.Errorf("difficulty must be between 0 and 5, got %d", c.Difficulty)
	}
	validTypes := map[string]bool{
		"token_completion": true,
		"line_completion":  true,
		"bug_repair":       true,
		"full_recall":      true,
		"test_repair":      true,
		"test_writing":     true,
	}
	if !validTypes[c.Type] {
		return fmt.Errorf("invalid type %q (must be token_completion, line_completion, bug_repair, full_recall, test_repair, or test_writing)", c.Type)
	}
	if strings.TrimSpace(c.Prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	if strings.TrimSpace(c.Starter) == "" {
		return fmt.Errorf("starter code is required")
	}
	if len(c.Hints) == 0 {
		return fmt.Errorf("at least one hint is required for ADHD accessibility")
	}
	for i, hint := range c.Hints {
		if strings.TrimSpace(hint) == "" {
			return fmt.Errorf("hint %d is empty", i+1)
		}
	}
	if strings.TrimSpace(c.Explanation) == "" {
		return fmt.Errorf("explanation is required")
	}
	if strings.TrimSpace(c.TestCode) == "" {
		return fmt.Errorf("test_code is required")
	}
	return nil
}
