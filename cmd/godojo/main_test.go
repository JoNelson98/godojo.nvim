package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Redirect database to a temp file during tests to prevent pollution
	f, err := os.CreateTemp("", "godojo-test-*.db")
	if err == nil {
		os.Setenv("GODOJO_DB", f.Name())
		f.Close()
		defer os.Remove(f.Name())
	}
	os.Exit(m.Run())
}

func TestRun_Ping(t *testing.T) {
	input := `{"action": "ping"}`
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = io.WriteString(w, input)
	}()

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout: %v", err)
	}
	os.Stdout = stdoutW

	err = run()
	stdoutW.Close()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, stdoutR)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal output: %v, raw output: %s", err, buf.String())
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if resp.Message != "pong" {
		t.Errorf("expected message 'pong', got %q", resp.Message)
	}
}

func TestRun_NextChallenge(t *testing.T) {
	if wd, err := os.Getwd(); err == nil && strings.HasSuffix(wd, "cmd/godojo") {
		_ = os.Chdir("../../")
		defer func() { _ = os.Chdir(wd) }()
	}

	input := `{"action": "next_challenge"}`
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = io.WriteString(w, input)
	}()

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout: %v", err)
	}
	os.Stdout = stdoutW

	err = run()
	stdoutW.Close()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, stdoutR)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal output: %v, raw output: %s", err, buf.String())
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q. Error: %s", resp.Status, resp.Error)
	}
	if resp.Challenge == nil {
		t.Fatalf("expected challenge to be loaded and returned, got nil")
	}
	if resp.Challenge.ID != "json.decode_request.001" {
		t.Errorf("expected challenge ID 'json.decode_request.001', got %q", resp.Challenge.ID)
	}
}

func TestRun_GradeAction(t *testing.T) {
	if wd, err := os.Getwd(); err == nil && strings.HasSuffix(wd, "cmd/godojo") {
		_ = os.Chdir("../../")
		defer func() { _ = os.Chdir(wd) }()
	}

	// We pass a syntax error to verify that grading executes and parses properly
	input := `{"action": "grade", "challenge_id": "json.decode_request.001", "submission": "package challenge\ninvalid syntax"}`
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = io.WriteString(w, input)
	}()

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout: %v", err)
	}
	os.Stdout = stdoutW

	err = run()
	stdoutW.Close()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, stdoutR)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal output: %v, raw output: %s", err, buf.String())
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q. Error: %s", resp.Status, resp.Error)
	}
	if resp.Grade == nil {
		t.Fatalf("expected grade report to be populated")
	}
	if resp.Grade.CompileOK {
		t.Errorf("expected compile_ok = false for invalid syntax, got true")
	}
}

func TestRun_StatsAction(t *testing.T) {
	input := `{"action": "stats"}`
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = io.WriteString(w, input)
	}()

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout: %v", err)
	}
	os.Stdout = stdoutW

	err = run()
	stdoutW.Close()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, stdoutR)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal output: %v, raw output: %s", err, buf.String())
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q. Error: %s", resp.Status, resp.Error)
	}
	if resp.Stats == nil {
		t.Fatalf("expected stats block to be populated in response")
	}
}

func TestRun_StartSessionAction(t *testing.T) {
	if wd, err := os.Getwd(); err == nil && strings.HasSuffix(wd, "cmd/godojo") {
		_ = os.Chdir("../../")
		defer func() { _ = os.Chdir(wd) }()
	}

	input := `{"action": "start_session", "mode": "standard"}`
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = io.WriteString(w, input)
	}()

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout: %v", err)
	}
	os.Stdout = stdoutW

	err = run()
	stdoutW.Close()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, stdoutR)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal output: %v, raw output: %s", err, buf.String())
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q. Error: %s", resp.Status, resp.Error)
	}
	if len(resp.Queue) == 0 {
		t.Errorf("expected session queue to contain scheduled challenges, got 0")
	}
}

func TestRun_CurriculumAction(t *testing.T) {
	if wd, err := os.Getwd(); err == nil && strings.HasSuffix(wd, "cmd/godojo") {
		_ = os.Chdir("../../")
		defer func() { _ = os.Chdir(wd) }()
	}

	input := `{"action": "curriculum", "chapter": "http"}`
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = io.WriteString(w, input)
	}()

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout: %v", err)
	}
	os.Stdout = stdoutW

	err = run()
	stdoutW.Close()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, stdoutR)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal output: %v, raw output: %s", err, buf.String())
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q. Error: %s", resp.Status, resp.Error)
	}
	if len(resp.Curriculum) == 0 {
		t.Errorf("expected curriculum progress tracking to load data, got 0 items")
	}
}

func TestRun_PatternsAction(t *testing.T) {
	if wd, err := os.Getwd(); err == nil && strings.HasSuffix(wd, "cmd/godojo") {
		_ = os.Chdir("../../")
		defer func() { _ = os.Chdir(wd) }()
	}

	input := `{"action": "patterns"}`
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = io.WriteString(w, input)
	}()

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout: %v", err)
	}
	os.Stdout = stdoutW

	err = run()
	stdoutW.Close()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, stdoutR)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal output: %v, raw output: %s", err, buf.String())
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q. Error: %s", resp.Status, resp.Error)
	}
	if len(resp.Patterns) == 0 {
		t.Errorf("expected patterns list to load data, got 0 cards")
	}
}

func TestRun_RevealPatternAction(t *testing.T) {
	if wd, err := os.Getwd(); err == nil && strings.HasSuffix(wd, "cmd/godojo") {
		_ = os.Chdir("../../")
		defer func() { _ = os.Chdir(wd) }()
	}

	input := `{"action": "reveal_pattern", "challenge_id": "test.reveal.111"}`
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = io.WriteString(w, input)
	}()

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout: %v", err)
	}
	os.Stdout = stdoutW

	err = run()
	stdoutW.Close()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, stdoutR)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal output: %v, raw output: %s", err, buf.String())
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q. Error: %s", resp.Status, resp.Error)
	}
}

func TestRun_ListChallengesAction(t *testing.T) {
	if wd, err := os.Getwd(); err == nil && strings.HasSuffix(wd, "cmd/godojo") {
		_ = os.Chdir("../../")
		defer func() { _ = os.Chdir(wd) }()
	}

	input := `{"action": "list_challenges", "chapter": "http"}`
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = io.WriteString(w, input)
	}()

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdout: %v", err)
	}
	os.Stdout = stdoutW

	err = run()
	stdoutW.Close()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, stdoutR)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal output: %v, raw output: %s", err, buf.String())
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q. Error: %s", resp.Status, resp.Error)
	}
	if len(resp.Queue) == 0 {
		t.Errorf("expected list of challenges to be returned, got 0")
	}
}

func TestRun_UnknownAction(t *testing.T) {
	input := `{"action": "foo"}`
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	os.Stdin = r
	go func() {
		defer w.Close()
		_, _ = io.WriteString(w, input)
	}()

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	os.Stdout = stdoutW

	err = run()
	stdoutW.Close()
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, stdoutR)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if resp.Status != "error" {
		t.Errorf("expected status 'error', got %q", resp.Status)
	}
	if !strings.Contains(resp.Error, "unknown action") {
		t.Errorf("expected error to contain 'unknown action', got %q", resp.Error)
	}
}
