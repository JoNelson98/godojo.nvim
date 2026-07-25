package scheduler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/godojo/godojo/internal/storage"
)

func TestScheduler_BuildQueue(t *testing.T) {
	// 1. Create temporary directory representing content dir
	tmpDir, err := os.MkdirTemp("", "godojo-sched-content-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy YAML challenge
	yamlContent := `
id: json.decode_request.001
title: Decode a JSON request
chapter: json
pattern_id: json.decode_request
difficulty: 3
type: full_recall
prompt: Decode.
starter: package challenge
hints:
  - hint
explanation: exp
test_code: package challenge
`
	err = os.WriteFile(filepath.Join(tmpDir, "c1.yaml"), []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("failed to write c1.yaml: %v", err)
	}

	// 2. Create temporary database
	dbFile, err := os.CreateTemp("", "godojo-sched-db-*.db")
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	dbFile.Close()
	defer os.Remove(dbFile.Name())

	db, err := storage.Open(dbFile.Name())
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// 3. Initialize Scheduler and build queue
	s := New(db, tmpDir)
	queue, err := s.BuildQueue("standard", "") // Fixed: passed two arguments!
	if err != nil {
		t.Fatalf("failed to build queue: %v", err)
	}

	if len(queue) != 1 {
		t.Errorf("expected queue of length 1, got %d", len(queue))
	}

	if queue[0].ID != "json.decode_request.001" {
		t.Errorf("expected first challenge ID to be 'json.decode_request.001', got %q", queue[0].ID)
	}
}
