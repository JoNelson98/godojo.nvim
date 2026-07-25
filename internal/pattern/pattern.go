package pattern

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PatternCard represents the structured data card for a DSA problem-solving pattern.
type PatternCard struct {
	ID              string   `yaml:"id" json:"id"`
	Name            string   `yaml:"name" json:"name"`
	Chapter         string   `yaml:"chapter" json:"chapter"`
	Solves          string   `yaml:"solves" json:"solves"`
	Clues           []string `yaml:"clues" json:"clues"`
	Invariant       string   `yaml:"invariant" json:"invariant"`
	Skeleton        string   `yaml:"skeleton" json:"skeleton"`
	TimeComplexity  string   `yaml:"time_complexity" json:"time_complexity"`
	SpaceComplexity string   `yaml:"space_complexity" json:"space_complexity"`
	Mistakes        []string `yaml:"mistakes" json:"mistakes"`
	Related         []string `yaml:"related,omitempty" json:"related,omitempty"`
}

// Load decodes a PatternCard from an io.Reader.
func Load(r io.Reader) (*PatternCard, error) {
	var p PatternCard
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&p); err != nil {
		return nil, fmt.Errorf("failed to decode YAML pattern card: %w", err)
	}
	return &p, nil
}

// LoadFile decodes a PatternCard from a file path.
func LoadFile(filePath string) (*PatternCard, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open pattern card file: %w", err)
	}
	defer f.Close()
	return Load(f)
}

// LoadAll scans the directory recursively and loads all PatternCards.
func LoadAll(dirPath string) ([]*PatternCard, error) {
	var cards []*PatternCard

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".yaml") && !strings.HasSuffix(strings.ToLower(info.Name()), ".yml") {
			return nil
		}

		c, err := LoadFile(path)
		if err == nil {
			cards = append(cards, c)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan patterns directory: %w", err)
	}

	return cards, nil
}

// FindByID searches the patterns directory for a PatternCard with the matching ID.
func FindByID(dirPath, id string) (*PatternCard, error) {
	var found *PatternCard

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".yaml") && !strings.HasSuffix(strings.ToLower(info.Name()), ".yml") {
			return nil
		}

		c, err := LoadFile(path)
		if err == nil && c.ID == id {
			found = c
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, fmt.Errorf("pattern card %q not found in %s", id, dirPath)
	}

	return found, nil
}
