package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/godojo/godojo/internal/challenge"
	"github.com/godojo/godojo/internal/grader"
	"github.com/godojo/godojo/internal/pattern"
	"github.com/godojo/godojo/internal/scheduler"
	"github.com/godojo/godojo/internal/storage"
)

// Request defines the standard schema for incoming Lua commands.
type Request struct {
	Action      string `json:"action"`
	ChallengeID string `json:"challenge_id,omitempty"`
	Submission  string `json:"submission,omitempty"`
	Mode        string `json:"mode,omitempty"`
	Chapter     string `json:"chapter,omitempty"`
	HintsUsed   int    `json:"hints_used,omitempty"`
	DurationMs  int64  `json:"duration_ms,omitempty"`
	ContentDir  string `json:"content_dir,omitempty"` // Absolute path to the content/ folder passed by Lua
}

// Response defines the standard schema for outgoing responses to Lua.
type Response struct {
	Status     string                    `json:"status"`
	Message    string                    `json:"message,omitempty"`
	Error      string                    `json:"error,omitempty"`
	Challenge  *challenge.Challenge      `json:"challenge,omitempty"`
	Grade      *grader.GradeResponse     `json:"grade,omitempty"`
	Stats      *storage.Stats            `json:"stats,omitempty"`
	Queue      []*challenge.Challenge    `json:"queue,omitempty"`
	Curriculum []storage.ChapterProgress `json:"curriculum,omitempty"`
	Patterns   []*pattern.PatternCard    `json:"patterns,omitempty"`
}

func getDBPath() string {
	if env := os.Getenv("GODOJO_DB"); env != "" {
		return env
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".godojo", "godojo.db")
}

func main() {
	// If CLI arguments are provided, handle standard terminal subcommands.
	if len(os.Args) > 1 {
		if os.Args[1] == "validate-content" {
			if err := challenge.ValidateContentDir("./content"); err != nil {
				fmt.Fprintf(os.Stderr, "Content validation failed: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: godojo [validate-content]\n", os.Args[1])
		os.Exit(1)
	}

	// Default mode: process JSON requests from stdin
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Engine error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var req Request
	dec := json.NewDecoder(os.Stdin)
	if err := dec.Decode(&req); err != nil {
		if err == io.EOF {
			return fmt.Errorf("empty input on stdin")
		}
		return fmt.Errorf("failed to decode request JSON: %w", err)
	}

	// Fallback to "./content" if not provided by Lua (for tests / backward compatibility)
	contentDir := req.ContentDir
	if contentDir == "" {
		contentDir = "./content"
	}

	var resp Response

	switch req.Action {
	case "ping":
		resp = Response{
			Status:  "ok",
			Message: "pong",
		}
	case "next_challenge":
		c, err := challenge.LoadFile(filepath.Join(contentDir, "json/decode_request.yaml"))
		if err != nil {
			resp = Response{
				Status: "error",
				Error:  fmt.Sprintf("failed to load challenge: %v", err),
			}
		} else {
			resp = Response{
				Status:    "ok",
				Challenge: c,
			}
		}
	case "grade":
		c, err := challenge.FindByID(contentDir, req.ChallengeID)
		if err != nil {
			resp = Response{
				Status: "error",
				Error:  fmt.Sprintf("failed to find challenge: %v", err),
			}
		} else {
			gradeResp, err := grader.Grade(c, req.Submission)
			if err != nil {
				resp = Response{
					Status: "error",
					Error:  fmt.Sprintf("failed to grade submission: %v", err),
				}
			} else {
				// Record attempt in SQLite
				db, err := storage.Open(getDBPath())
				if err == nil {
					defer db.Close()
					_ = db.RecordAttempt(req.ChallengeID, c.PatternID, gradeResp.Correct, req.HintsUsed, req.DurationMs)
				}

				resp = Response{
					Status: "ok",
					Grade:  gradeResp,
				}
			}
		}
	case "stats":
		db, err := storage.Open(getDBPath())
		if err != nil {
			resp = Response{
				Status: "error",
				Error:  fmt.Sprintf("failed to open database: %v", err),
			}
		} else {
			defer db.Close()
			stats, err := db.QueryStats()
			if err != nil {
				resp = Response{
					Status: "error",
					Error:  fmt.Sprintf("failed to query stats: %v", err),
				}
			} else {
				resp = Response{
					Status: "ok",
					Stats:  stats,
				}
			}
		}
	case "start_session":
		db, err := storage.Open(getDBPath())
		if err != nil {
			resp = Response{
				Status: "error",
				Error:  fmt.Sprintf("failed to open database: %v", err),
			}
		} else {
			defer db.Close()
			sched := scheduler.New(db, contentDir)
			queue, err := sched.BuildQueue(req.Mode, req.Chapter)
			if err != nil {
				resp = Response{
					Status: "error",
					Error:  fmt.Sprintf("failed to build session queue: %v", err),
				}
			} else {
				resp = Response{
					Status: "ok",
					Queue:  queue,
				}
			}
		}
	case "curriculum":
		db, err := storage.Open(getDBPath())
		if err != nil {
			resp = Response{
				Status: "error",
				Error:  fmt.Sprintf("failed to open database: %v", err),
			}
		} else {
			defer db.Close()
			curric, err := db.QueryCurriculum(contentDir, req.Chapter)
			if err != nil {
				resp = Response{
					Status: "error",
					Error:  fmt.Sprintf("failed to query curriculum: %v", err),
				}
			} else {
				resp = Response{
					Status:     "ok",
					Curriculum: curric,
				}
			}
		}
	case "patterns":
		cards, err := pattern.LoadAll(filepath.Join(contentDir, "patterns"))
		if err != nil {
			resp = Response{
				Status: "error",
				Error:  fmt.Sprintf("failed to load pattern cards: %v", err),
			}
		} else {
			resp = Response{
				Status:   "ok",
				Patterns: cards,
			}
		}
	case "reveal_pattern":
		db, err := storage.Open(getDBPath())
		if err != nil {
			resp = Response{
				Status: "error",
				Error:  err.Error(),
			}
		} else {
			defer db.Close()
			_ = db.RecordReveal(req.ChallengeID)
			resp = Response{
				Status: "ok",
			}
		}
	case "list_challenges":
		db, err := storage.Open(getDBPath())
		if err != nil {
			resp = Response{
				Status: "error",
				Error:  err.Error(),
			}
		} else {
			defer db.Close()
			sched := scheduler.New(db, contentDir)
			queue, err := sched.BuildQueue("list", req.Chapter)
			if err != nil {
				resp = Response{
					Status: "error",
					Error:  err.Error(),
				}
			} else {
				resp = Response{
					Status: "ok",
					Queue:  queue,
				}
			}
		}
	default:
		resp = Response{
			Status: "error",
			Error:  fmt.Sprintf("unknown action: %s", req.Action),
		}
	}

	// Write response JSON to stdout
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(resp); err != nil {
		return fmt.Errorf("failed to encode response JSON: %w", err)
	}

	return nil
}
