package challenge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateContentDir scans the specified directory recursively, loading and validating
// all challenge YAML files. It checks for duplicate IDs and enforces schema rules.
func ValidateContentDir(dirPath string) error {
	ids := make(map[string]string) // id -> file path

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only check .yaml files
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".yaml") && !strings.HasSuffix(strings.ToLower(info.Name()), ".yml") {
			return nil
		}

		// Skip pattern cards
		if strings.Contains(path, "/patterns/") || strings.Contains(path, "\\patterns\\") {
			return nil
		}

		c, err := LoadFile(path)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", path, err)
		}

		if err := c.Validate(); err != nil {
			return fmt.Errorf("validation failed in %s: %w", path, err)
		}

		// Ensure unique IDs
		if existingPath, exists := ids[c.ID]; exists {
			return fmt.Errorf("duplicate challenge ID %q found in %s (already defined in %s)", c.ID, path, existingPath)
		}
		ids[c.ID] = path

		fmt.Printf("✓ Validated: %s (ID: %s, Title: %s)\n", path, c.ID, c.Title)
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("\nSuccess: Fully validated %d challenges in %s.\n", len(ids), dirPath)
	return nil
}
