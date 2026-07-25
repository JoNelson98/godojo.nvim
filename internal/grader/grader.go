package grader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/godojo/godojo/internal/challenge"
)

type TestResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message,omitempty"`
}

type ASTCheck struct {
	Name   string `json:"name" json:"name"`
	Passed bool   `json:"passed" json:"passed"`
}

type FeedbackItem struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
}

type GradeResponse struct {
	Correct    bool           `json:"correct"`
	CompileOK  bool           `json:"compile_ok"`
	Tests      []TestResult   `json:"tests"`
	ASTChecks  []ASTCheck     `json:"ast_checks,omitempty"`
	Feedback   []FeedbackItem `json:"feedback,omitempty"`
	NextAction string         `json:"next_action"` // "retry" or "advance"
}

// GoTestEvent represents a single line of output from "go test -json".
type GoTestEvent struct {
	Action  string  `json:"Action"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Package string  `json:"Package"`
	Elapsed float64 `json:"Elapsed"`
}

// Grade writes the submission and test code into an isolated temporary Go module,
// compiles it, runs "go test -json", and produces a detailed, structured feedback report.
func Grade(c *challenge.Challenge, submission string) (*GradeResponse, error) {
	tmpDir, err := os.MkdirTemp("", "godojo-grade-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp grading dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write go.mod
	goModContent := "module challenge\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write go.mod: %w", err)
	}

	// Write submission
	if err := os.WriteFile(filepath.Join(tmpDir, "challenge.go"), []byte(submission), 0644); err != nil {
		return nil, fmt.Errorf("failed to write challenge.go: %w", err)
	}

	// Write test file
	if err := os.WriteFile(filepath.Join(tmpDir, "challenge_test.go"), []byte(c.TestCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write challenge_test.go: %w", err)
	}

	// Run gofmt on submission
	gofmtCmd := exec.Command("gofmt", "-w", filepath.Join(tmpDir, "challenge.go"))
	var gofmtStderr bytes.Buffer
	gofmtCmd.Stderr = &gofmtStderr
	if err := gofmtCmd.Run(); err != nil {
		// Syntax/formatting error
		return &GradeResponse{
			Correct:   false,
			CompileOK: false,
			Tests:     nil,
			Feedback: []FeedbackItem{
				{
					Line:    1,
					Column:  1,
					Message: fmt.Sprintf("Formatting/Syntax error: %s", strings.TrimSpace(gofmtStderr.String())),
				},
			},
			NextAction: "retry",
		}, nil
	}

	// Execute 'go test -json ./...' inside the temp directory
	cmd := exec.Command("go", "test", "-json", "./...")
	cmd.Dir = tmpDir

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// We run the command and ignore the error, because if tests fail, 'go test' exits with non-zero
	_ = cmd.Run()

	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	// Parse JSON test events
	lines := strings.Split(stdoutStr, "\n")
	testOutputs := make(map[string][]string) // TestName -> []OutputLines
	testPassed := make(map[string]bool)

	// Populate initial test state from challenge expectations
	for _, expectedTest := range c.Validation.Tests {
		testPassed[expectedTest] = false
	}

	hasBuildFailure := false
	var compileErrorMsg strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var ev GoTestEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			// Not standard test event JSON, might be compilation raw output
			continue
		}

		if ev.Test != "" {
			if ev.Action == "output" {
				testOutputs[ev.Test] = append(testOutputs[ev.Test], ev.Output)
			} else if ev.Action == "pass" {
				testPassed[ev.Test] = true
			} else if ev.Action == "fail" {
				testPassed[ev.Test] = false
			}
		} else {
			// Package-level event (e.g. build failure output)
			if ev.Action == "output" {
				// Detect actual compiler errors (not test suite results)
				if strings.Contains(ev.Output, "build failed") || strings.Contains(ev.Output, "# challenge") {
					hasBuildFailure = true
				}
				compileErrorMsg.WriteString(ev.Output)
			}
		}
	}

	// Fallback to stderr for compile errors if stdout was empty
	if compileErrorMsg.Len() == 0 && stderrStr != "" {
		hasBuildFailure = true
		compileErrorMsg.WriteString(stderrStr)
	}

	// If there is any build/compile failure, report it immediately!
	if hasBuildFailure {
		return &GradeResponse{
			Correct:   false,
			CompileOK: false,
			Feedback: []FeedbackItem{
				{
					Line:    1,
					Column:  1,
					Message: fmt.Sprintf("Compilation failure:\n%s", strings.TrimSpace(compileErrorMsg.String())),
				},
			},
			NextAction: "retry",
		}, nil
	}

	// Map results into response
	var results []TestResult
	allPassed := true

	// Ensure we preserve the order of tests defined in the challenge
	for _, expectedTest := range c.Validation.Tests {
		passed := testPassed[expectedTest]
		if !passed {
			allPassed = false
		}

		msg := ""
		if outputs, exists := testOutputs[expectedTest]; exists {
			// Extract fail messages
			var cleanOutput []string
			for _, o := range outputs {
				trimmed := strings.TrimSpace(o)
				if trimmed != "" && !strings.HasPrefix(trimmed, "=== RUN") && !strings.HasPrefix(trimmed, "--- FAIL") {
					cleanOutput = append(cleanOutput, trimmed)
				}
			}
			msg = strings.Join(cleanOutput, "\n")
		}

		results = append(results, TestResult{
			Name:    expectedTest,
			Passed:  passed,
			Message: msg,
		})
	}

	// Run Go AST structural rule checks
	astChecks, astErr := VerifyASTRules(submission, c.Validation.ASTRules)
	if astErr != nil {
		// Gracefully handle parsing failures (e.g. if the code compiled but AST failed, though rare)
		astChecks = []ASTCheck{}
		for _, r := range c.Validation.ASTRules {
			astChecks = append(astChecks, ASTCheck{Name: r, Passed: false})
		}
	}

	// Enforce both behavioral tests and AST rules are correct
	for _, astC := range astChecks {
		if !astC.Passed {
			allPassed = false
		}
	}

	nextAction := "retry"
	if allPassed {
		nextAction = "advance"
	}

	return &GradeResponse{
		Correct:    allPassed,
		CompileOK:  true,
		Tests:      results,
		ASTChecks:  astChecks,
		NextAction: nextAction,
	}, nil
}
