package challenge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateContentDir_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "godojo-content-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	yaml1 := `
id: json.decode_request.001
title: Decode a JSON request
chapter: json
pattern_id: json.decode_request
difficulty: 3
type: full_recall
prompt: Decode r.Body into input.
starter: |
  package challenge
  import "net/http"
  func decodeRequest(r *http.Request) {}
validation:
  compile: true
hints:
  - hint
explanation: exp
test_code: |
  package challenge
  import "testing"
  func TestFoo(t *testing.T) {}
`
	yaml2 := `
id: json.decode_request.002
title: Decode another JSON request
chapter: json
pattern_id: json.decode_request
difficulty: 3
type: full_recall
prompt: Decode r.Body into input.
starter: |
  package challenge
  import "net/http"
  func decodeRequest(r *http.Request) {}
validation:
  compile: true
hints:
  - hint
explanation: exp
test_code: |
  package challenge
  import "testing"
  func TestFoo(t *testing.T) {}
`

	err = os.WriteFile(filepath.Join(tmpDir, "c1.yaml"), []byte(yaml1), 0644)
	if err != nil {
		t.Fatalf("failed to write c1.yaml: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "c2.yaml"), []byte(yaml2), 0644)
	if err != nil {
		t.Fatalf("failed to write c2.yaml: %v", err)
	}

	err = ValidateContentDir(tmpDir)
	if err != nil {
		t.Errorf("expected validation to pass, got: %v", err)
	}
}

func TestValidateContentDir_DuplicateID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "godojo-content-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	yaml1 := `
id: json.decode_request.001
title: Decode a JSON request
chapter: json
pattern_id: json.decode_request
difficulty: 3
type: full_recall
prompt: Decode r.Body into input.
starter: |
  package challenge
  import "net/http"
  func decodeRequest(r *http.Request) {}
validation:
  compile: true
hints:
  - hint
explanation: exp
test_code: |
  package challenge
  import "testing"
  func TestFoo(t *testing.T) {}
`

	err = os.WriteFile(filepath.Join(tmpDir, "c1.yaml"), []byte(yaml1), 0644)
	if err != nil {
		t.Fatalf("failed to write c1.yaml: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "c2.yaml"), []byte(yaml1), 0644) // Duplicate
	if err != nil {
		t.Fatalf("failed to write c2.yaml: %v", err)
	}

	err = ValidateContentDir(tmpDir)
	if err == nil {
		t.Fatal("expected validation to fail due to duplicate IDs, got nil")
	}

	if !strings.Contains(err.Error(), "duplicate challenge ID") {
		t.Errorf("expected duplicate challenge ID error, got: %v", err)
	}
}
