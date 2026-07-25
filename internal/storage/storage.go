package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/godojo/godojo/internal/challenge"
)

type DB struct {
	conn *sql.DB
}

type Stats struct {
	TotalAttempts   int      `json:"total_attempts"`
	CorrectAttempts int      `json:"correct_attempts"`
	SuccessRate     float64  `json:"success_rate"`
	ReviewsDue      int      `json:"reviews_due"`
	StableMastery   float64  `json:"stable_mastery"`
	WeakPatterns    []string `json:"weak_patterns"`
}

type ChapterProgress struct {
	Chapter   string  `json:"chapter"`
	Title     string  `json:"title"`
	Total     int     `json:"total"`
	Completed int     `json:"completed"`
	Percent   float64 `json:"percent"`
}

// Open opens a connection to the SQLite database file, automatically creating it
// and running migrations if they are not already present.
func Open(filePath string) (*DB, error) {
	// Create directory if missing
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// RawConn returns the raw database connection.
func (db *DB) RawConn() *sql.DB {
	return db.conn
}

func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS patterns (
			id TEXT PRIMARY KEY,
			chapter TEXT NOT NULL,
			title TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS challenges (
			id TEXT PRIMARY KEY,
			pattern_id TEXT NOT NULL,
			difficulty INTEGER NOT NULL,
			challenge_type TEXT NOT NULL,
			FOREIGN KEY (pattern_id) REFERENCES patterns(id)
		);`,
		`CREATE TABLE IF NOT EXISTS attempts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			challenge_id TEXT NOT NULL,
			pattern_id TEXT NOT NULL,
			correct BOOLEAN NOT NULL,
			first_attempt BOOLEAN NOT NULL,
			hints_used INTEGER NOT NULL DEFAULT 0,
			duration_ms INTEGER NOT NULL,
			error_category TEXT,
			created_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS progress (
			pattern_id TEXT PRIMARY KEY,
			mastery_score REAL NOT NULL DEFAULT 0,
			difficulty INTEGER NOT NULL DEFAULT 1,
			next_review_at DATETIME,
			last_attempt_at DATETIME,
			clean_recall_count INTEGER NOT NULL DEFAULT 0,
			transfer_complete BOOLEAN NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			mode TEXT NOT NULL,
			started_at DATETIME NOT NULL,
			ended_at DATETIME,
			completed_challenges INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS pattern_reveals (
			challenge_id TEXT PRIMARY KEY
		);`,
	}

	for _, q := range queries {
		if _, err := db.conn.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// RecordReveal logs that a pattern was revealed for a specific challenge.
func (db *DB) RecordReveal(challengeID string) error {
	_, err := db.conn.Exec("INSERT OR IGNORE INTO pattern_reveals (challenge_id) VALUES (?)", challengeID)
	return err
}

// IsRevealed checks if a pattern was revealed for a specific challenge.
func (db *DB) IsRevealed(challengeID string) (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM pattern_reveals WHERE challenge_id = ?", challengeID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RecordAttempt persists a training attempt and updates spaced repetition progress.
func (db *DB) RecordAttempt(challengeID, patternID string, correct bool, hintsUsed int, durationMs int64) error {
	now := time.Now()

	// 1. Is this the first attempt for this challenge_id/pattern_id today?
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM attempts WHERE challenge_id = ?", challengeID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to scan attempts count: %w", err)
	}

	firstAttempt := (count == 0)

	// 2. Insert the attempt record
	_, err = db.conn.Exec(`
		INSERT INTO attempts (challenge_id, pattern_id, correct, first_attempt, hints_used, duration_ms, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, challengeID, patternID, correct, firstAttempt, hintsUsed, durationMs, now)
	if err != nil {
		return fmt.Errorf("failed to insert attempt: %w", err)
	}

	// 3. Update the dynamic pattern mastery score
	return db.updateProgress(patternID, correct, hintsUsed, firstAttempt)
}

func (db *DB) updateProgress(patternID string, correct bool, hintsUsed int, firstAttempt bool) error {
	now := time.Now()

	// 1. Find or create progress record
	var score float64
	var difficulty int
	var recallCount int
	var exists bool

	err := db.conn.QueryRow(`
		SELECT mastery_score, difficulty, clean_recall_count 
		FROM progress WHERE pattern_id = ?
	`, patternID).Scan(&score, &difficulty, &recallCount)

	if err == sql.ErrNoRows {
		score = 0.0
		difficulty = 1
		recallCount = 0
		exists = false
	} else if err != nil {
		return fmt.Errorf("failed to query progress: %w", err)
	} else {
		exists = true
	}

	// 2. Recalculate Spaced Repetition Mastery Score
	var interval float64
	if correct {
		if hintsUsed == 0 && firstAttempt {
			recallCount++
			score += 0.25
			if score > 1.0 {
				score = 1.0
			}
			// Exponential Spacing Interval: 1 day -> 3 days -> 7 days -> 14 days -> 30 days
			switch recallCount {
			case 1:
				interval = 24.0 * 60.0 // 1 day
			case 2:
				interval = 3.0 * 24.0 * 60.0 // 3 days
			case 3:
				interval = 7.0 * 24.0 * 60.0 // 7 days
			case 4:
				interval = 14.0 * 24.0 * 60.0 // 14 days
			default:
				interval = 30.0 * 24.0 * 60.0 // 30 days
			}
		} else {
			// Solved but with assistance (hints used): review again tomorrow
			score += 0.10
			if score > 1.0 {
				score = 1.0
			}
			interval = 24.0 * 60.0 // 1 day
		}
	} else {
		// Failed recall: reset cleanliness parameters and cut mastery score in half
		recallCount = 0
		score = score * 0.50
		interval = 1.0 // Review again in 1 minute
	}

	nextReview := now.Add(time.Duration(interval) * time.Minute)

	if !exists {
		_, err = db.conn.Exec(`
			INSERT INTO progress (pattern_id, mastery_score, difficulty, next_review_at, last_attempt_at, clean_recall_count)
			VALUES (?, ?, ?, ?, ?, ?)
		`, patternID, score, difficulty, nextReview, now, recallCount)
	} else {
		_, err = db.conn.Exec(`
			UPDATE progress 
			SET mastery_score = ?, next_review_at = ?, last_attempt_at = ?, clean_recall_count = ?
			WHERE pattern_id = ?
		`, score, nextReview, now, recallCount, patternID)
	}

	if err != nil {
		return fmt.Errorf("failed to update progress database: %w", err)
	}

	return nil
}

// QueryStats aggregates learner metrics to display on the home screen.
func (db *DB) QueryStats() (*Stats, error) {
	stats := &Stats{}

	// 1. Total & correct attempts counts
	err := db.conn.QueryRow("SELECT COUNT(*) FROM attempts").Scan(&stats.TotalAttempts)
	if err != nil {
		return nil, fmt.Errorf("failed to count attempts: %w", err)
	}

	err = db.conn.QueryRow("SELECT COUNT(*) FROM attempts WHERE correct = 1").Scan(&stats.CorrectAttempts)
	if err != nil {
		return nil, fmt.Errorf("failed to count correct: %w", err)
	}

	if stats.TotalAttempts > 0 {
		stats.SuccessRate = float64(stats.CorrectAttempts) / float64(stats.TotalAttempts)
	}

	// 2. Reviews due count
	now := time.Now()
	err = db.conn.QueryRow("SELECT COUNT(*) FROM progress WHERE next_review_at <= ?", now).Scan(&stats.ReviewsDue)
	if err != nil {
		stats.ReviewsDue = 0
	}

	// 3. Stable mastery average (percentage of patterns with score >= 0.75)
	err = db.conn.QueryRow("SELECT COALESCE(SUM(CASE WHEN mastery_score >= 0.75 THEN 1.0 ELSE 0.0 END) / COUNT(*), 0.0) FROM progress").Scan(&stats.StableMastery)
	if err != nil {
		stats.StableMastery = 0.0
	}

	// 4. Weak Patterns list
	rows, err := db.conn.Query("SELECT pattern_id FROM progress WHERE mastery_score < 0.50 LIMIT 5")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var pid string
			if err := rows.Scan(&pid); err == nil {
				stats.WeakPatterns = append(stats.WeakPatterns, pid)
			}
		}
	}

	return stats, nil
}

// QueryCurriculum recursively scans the contentDir to determine totals on disk, 
// joins them with SQLite attempts, and returns the live, dynamic path progress checklist.
func (db *DB) QueryCurriculum(contentDir string, activeDojo string) ([]ChapterProgress, error) {
	totals := make(map[string]int)
	chapterChallenges := make(map[string][]string) // chapter -> list of challenge IDs

	// Scan disk content
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".yaml") {
			return nil
		}

		isLeetcode := strings.Contains(path, "/leetcode/") || strings.Contains(path, "\\leetcode\\")
		if activeDojo == "leetcode" && !isLeetcode {
			return nil
		}
		if (activeDojo == "http" || activeDojo == "") && isLeetcode {
			return nil
		}

		// Skip pattern cards
		if strings.Contains(path, "/patterns/") || strings.Contains(path, "\\patterns\\") {
			return nil
		}

		c, err := challenge.LoadFile(path)
		if err == nil {
			totals[c.Chapter]++
			chapterChallenges[c.Chapter] = append(chapterChallenges[c.Chapter], c.ID)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan content for curriculum: %w", err)
	}

	var chapters []string
	var titles map[string]string

	if activeDojo == "leetcode" {
		chapters = []string{"arrays", "pointers", "window"}
		titles = map[string]string{
			"arrays":   "Arrays & Hashing",
			"pointers": "Two Pointers",
			"window":   "Sliding Window",
		}
	} else {
		chapters = []string{"http", "json", "testing", "boss"}
		titles = map[string]string{
			"http":    "net/http fundamentals",
			"json":    "encoding/json API DTOs",
			"testing": "httptest Handler Testing",
			"boss":    "POST /companies Boss Mission",
		}
	}

	var results []ChapterProgress
	for _, ch := range chapters {
		total := totals[ch]
		if total == 0 {
			continue // Skip directories with no yaml files
		}

		// Query database for unique completed challenges in this specific chapter using direct ID lists!
		var completed int
		ids := chapterChallenges[ch]
		if len(ids) > 0 {
			placeholders := make([]string, len(ids))
			args := make([]interface{}, len(ids))
			for i, id := range ids {
				placeholders[i] = "?"
				args[i] = id
			}
			query := fmt.Sprintf("SELECT COUNT(DISTINCT challenge_id) FROM attempts WHERE correct = 1 AND challenge_id IN (%s)", strings.Join(placeholders, ", "))
			_ = db.conn.QueryRow(query, args...).Scan(&completed)
		}

		percent := 0.0
		if total > 0 {
			percent = float64(completed) / float64(total)
		}

		results = append(results, ChapterProgress{
			Chapter:   ch,
			Title:     titles[ch],
			Total:     total,
			Completed: completed,
			Percent:   percent,
		})
	}

	return results, nil
}
