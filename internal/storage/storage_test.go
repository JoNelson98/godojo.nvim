package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStorage_AttemptTrackingAndStats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "godojo-storage-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test_godojo.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Initially stats should be zeros/empty
	stats, err := db.QueryStats()
	if err != nil {
		t.Fatalf("failed to query stats initially: %v", err)
	}
	if stats.TotalAttempts != 0 {
		t.Errorf("expected 0 total attempts, got %d", stats.TotalAttempts)
	}

	// Record a failing attempt
	err = db.RecordAttempt("json.decode_request.001", "json.decode_request", false, 0, 1500)
	if err != nil {
		t.Fatalf("failed to record first failure attempt: %v", err)
	}

	stats, err = db.QueryStats()
	if err != nil {
		t.Fatalf("failed to query stats after failure: %v", err)
	}
	if stats.TotalAttempts != 1 {
		t.Errorf("expected 1 total attempt, got %d", stats.TotalAttempts)
	}
	if stats.CorrectAttempts != 0 {
		t.Errorf("expected 0 correct attempts, got %d", stats.CorrectAttempts)
	}
	if stats.SuccessRate != 0.0 {
		t.Errorf("expected 0.0 success rate, got %f", stats.SuccessRate)
	}
	if len(stats.WeakPatterns) != 1 || stats.WeakPatterns[0] != "json.decode_request" {
		t.Errorf("expected 'json.decode_request' in weak patterns list, got %+v", stats.WeakPatterns)
	}

	// Record a correct attempt with hint
	err = db.RecordAttempt("json.decode_request.001", "json.decode_request", true, 1, 2200)
	if err != nil {
		t.Fatalf("failed to record correct attempt with hint: %v", err)
	}

	stats, err = db.QueryStats()
	if err != nil {
		t.Fatalf("failed to query stats after correct with hint: %v", err)
	}
	if stats.TotalAttempts != 2 {
		t.Errorf("expected 2 total attempts, got %d", stats.TotalAttempts)
	}
	if stats.CorrectAttempts != 1 {
		t.Errorf("expected 1 correct attempt, got %d", stats.CorrectAttempts)
	}
	if stats.SuccessRate != 0.5 {
		t.Errorf("expected 0.5 success rate, got %f", stats.SuccessRate)
	}

	// Record clean recalls to reach stable mastery
	// 1st clean recall on "pointer.receiver"
	err = db.RecordAttempt("pointer.receiver.001", "pointer.receiver", true, 0, 800)
	if err != nil {
		t.Fatalf("failed to record clean recall: %v", err)
	}
	// 2nd clean recall on "pointer.receiver"
	err = db.RecordAttempt("pointer.receiver.002", "pointer.receiver", true, 0, 600)
	if err != nil {
		t.Fatalf("failed to record clean recall: %v", err)
	}
	// 3rd clean recall on "pointer.receiver" (reaches score >= 0.75)
	err = db.RecordAttempt("pointer.receiver.003", "pointer.receiver", true, 0, 500)
	if err != nil {
		t.Fatalf("failed to record clean recall: %v", err)
	}

	stats, err = db.QueryStats()
	if err != nil {
		t.Fatalf("failed to query stats after clean recalls: %v", err)
	}

	if stats.StableMastery != 0.5 { // 1 out of 2 patterns is stable (pointer.receiver is >= 0.75, json.decode_request is < 0.75)
		t.Errorf("expected 0.5 stable mastery, got %f", stats.StableMastery)
	}
}
