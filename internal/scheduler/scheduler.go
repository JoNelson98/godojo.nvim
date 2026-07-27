package scheduler

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/godojo/godojo/internal/challenge"
	"github.com/godojo/godojo/internal/storage"
)

// Scheduler builds a session queue of challenges based on progress and spaced repetition.
type Scheduler struct {
	db         *storage.DB
	contentDir string
}

func New(db *storage.DB, contentDir string) *Scheduler {
	return &Scheduler{
		db:         db,
		contentDir: contentDir,
	}
}

// SortChallenges sorts a slice of challenges in chronological curriculum order:
// 1. Chapter: "http" -> "json" -> "testing" -> "boss"
// 2. Difficulty: 1 -> 5
// 3. ID alphabetically
func SortChallenges(list []*challenge.Challenge) {
	chapterOrder := map[string]int{
		"http":    1,
		"json":    2,
		"testing": 3,
		"boss":    4,
	}

	sort.Slice(list, func(i, j int) bool {
		c1, c2 := list[i], list[j]
		
		ch1, ch2 := chapterOrder[c1.Chapter], chapterOrder[c2.Chapter]
		if ch1 != ch2 {
			return ch1 < ch2
		}

		if c1.Difficulty != c2.Difficulty {
			return c1.Difficulty < c2.Difficulty
		}

		return c1.ID < c2.ID
	})
}

// BuildQueue constructs a prioritized list of challenges for a training session.
func (s *Scheduler) BuildQueue(mode string, requestedChapter string) ([]*challenge.Challenge, error) {
	// 1. Load all challenges from disk
	allChallenges, err := s.loadAllChallenges()
	if err != nil {
		return nil, fmt.Errorf("failed to load challenges for scheduler: %w", err)
	}

	if len(allChallenges) == 0 {
		return nil, fmt.Errorf("no challenges found in %s", s.contentDir)
	}

	// Sort curriculum chronologically
	SortChallenges(allChallenges)

	// Filter challenges by requested chapter or Dojo track if specified
	if requestedChapter != "" && requestedChapter != "all" {
		var filtered []*challenge.Challenge
		for _, c := range allChallenges {
			isLeetcode := c.Chapter == "arrays" || c.Chapter == "pointers" || c.Chapter == "window"
			if requestedChapter == "leetcode" {
				if isLeetcode {
					filtered = append(filtered, c)
				}
			} else if requestedChapter == "http" {
				if !isLeetcode {
					filtered = append(filtered, c)
				}
			} else {
				if c.Chapter == requestedChapter {
					filtered = append(filtered, c)
				}
			}
		}
		allChallenges = filtered
	}

	// Query database to find all successfully completed challenge IDs so we can exclude them from new training
	completedChallenges := make(map[string]bool)
	rows, err := s.db.RawConn().Query("SELECT DISTINCT challenge_id FROM attempts WHERE correct = 1")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cid string
			if err := rows.Scan(&cid); err == nil {
				completedChallenges[cid] = true
			}
		}
	}

	// Fetch stats and progress from database
	stats, err := s.db.QueryStats()
	if err != nil {
		return nil, fmt.Errorf("failed to query stats for scheduler: %w", err)
	}

	// We want to construct a queue of about 4 to 8 challenges.
	targetSize := 8
	if mode == "quick" {
		targetSize = 4
	}

	var queue []*challenge.Challenge
	addedIDs := make(map[string]bool)

	// Helper to add a challenge safely without duplicates
	addChallenge := func(c *challenge.Challenge, allowCompleted bool) bool {
		// If building standard or quick session, exclude boss missions from normal flow
		if (mode == "standard" || mode == "quick") && c.Chapter == "boss" && requestedChapter != "boss" {
			return false
		}
		if addedIDs[c.ID] {
			return false
		}
		// If completed check is active, exclude already completed challenges!
		if !allowCompleted && completedChallenges[c.ID] {
			return false
		}

		queue = append(queue, c)
		addedIDs[c.ID] = true
		return true
	}

	// Mode: boss (User explicitly wants the Boss Mission)
	if mode == "boss" || requestedChapter == "boss" {
		for _, c := range allChallenges {
			if c.Chapter == "boss" {
				addChallenge(c, true) // Boss is allowed to repeat
			}
		}
		return queue, nil
	}

	// Mode: list (returns all challenges in the active dojo chronologically)
	if mode == "list" {
		for _, c := range allChallenges {
			addChallenge(c, true) // Allowed to repeat/list
		}
		return queue, nil
	}

	// Mode: reviews (User explicitly wants ONLY due Reviews)
	if mode == "reviews" {
		duePatterns := make(map[string]bool)
		now := time.Now()
		rows, err := s.db.RawConn().Query("SELECT pattern_id FROM progress WHERE next_review_at <= ?", now)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var pid string
				if err := rows.Scan(&pid); err == nil {
					duePatterns[pid] = true
				}
			}
		}

		for _, c := range allChallenges {
			if duePatterns[c.PatternID] {
				addChallenge(c, true) // Spaced reviews are allowed to repeat
			}
		}
		return queue, nil
	}

	// Prioritization Phase 1: Due Reviews (Spaced Reviews ARE allowed to repeat!)
	if stats.ReviewsDue > 0 {
		duePatterns := make(map[string]bool)
		// Query patterns that are due review
		now := time.Now()
		rows, err := s.db.RawConn().Query("SELECT pattern_id FROM progress WHERE next_review_at <= ?", now)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var pid string
				if err := rows.Scan(&pid); err == nil {
					duePatterns[pid] = true
				}
			}
		}

		// Add due review challenges
		for _, c := range allChallenges {
			if duePatterns[c.PatternID] {
				addChallenge(c, true) // Allowed to repeat because it is a due review
				if len(queue) >= targetSize {
					break
				}
			}
		}
	}

	// Prioritization Phase 2: Weak Patterns (Allowed to repeat for active drilling!)
	if len(queue) < targetSize && len(stats.WeakPatterns) > 0 {
		weakMap := make(map[string]bool)
		for _, wp := range stats.WeakPatterns {
			weakMap[wp] = true
		}

		for _, c := range allChallenges {
			if weakMap[c.PatternID] {
				addChallenge(c, true) // Allowed to repeat because it is weak
				if len(queue) >= targetSize {
					break
				}
			}
		}
	}

	// Prioritization Phase 3: Exactly One New Pattern (Excludes completed!)
	if len(queue) < targetSize {
		// Find patterns with NO entries in progress (completely new)
		rows, err := s.db.RawConn().Query("SELECT pattern_id FROM progress")
		attemptedPatterns := make(map[string]bool)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var pid string
				if err := rows.Scan(&pid); err == nil {
					attemptedPatterns[pid] = true
				}
			}
		}

		var newPatternCandidate string
		for _, c := range allChallenges {
			if c.Chapter == "boss" {
				continue
			}
			if !attemptedPatterns[c.PatternID] {
				newPatternCandidate = c.PatternID
				break
			}
		}

		if newPatternCandidate != "" {
			// Add challenges belonging to this single new pattern
			for _, c := range allChallenges {
				if c.PatternID == newPatternCandidate {
					addChallenge(c, false) // Exclude completed!
					if len(queue) >= targetSize {
						break
					}
				}
			}
		}
	}

	// Prioritization Phase 4: Fill Remaining with incremental chronological curriculum items (Excludes completed!)
	if len(queue) < targetSize {
		for _, c := range allChallenges {
			addChallenge(c, false) // Exclude completed!
			if len(queue) >= targetSize {
				break
			}
		}
	}

	// Prioritization Phase 5: Last Resort Fallback (If the user completed everything on disk in this chapter, allow repeats so the session queue can still run)
	if len(queue) == 0 {
		for _, c := range allChallenges {
			addChallenge(c, true) // Allow repeats as fallback
			if len(queue) >= targetSize {
				break
			}
		}
	}

	if len(queue) > targetSize {
		queue = queue[:targetSize]
	}

	return queue, nil
}

func (s *Scheduler) loadAllChallenges() ([]*challenge.Challenge, error) {
	var list []*challenge.Challenge

	err := filepath.Walk(s.contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".yaml") && !strings.HasSuffix(strings.ToLower(info.Name()), ".yml") {
			return nil
		}

		// Skip pattern cards
		if strings.Contains(path, "/patterns/") || strings.Contains(path, "\\patterns\\") {
			return nil
		}

		c, err := challenge.LoadFile(path)
		if err != nil {
			return nil // Skip invalid files during scheduling to avoid crashing sessions
		}
		list = append(list, c)
		return nil
	})

	return list, err
}
