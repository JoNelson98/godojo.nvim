package grader

import (
	"testing"

	"github.com/godojo/godojo/internal/challenge"
)

func getTestChallenge() *challenge.Challenge {
	return &challenge.Challenge{
		ID:         "test.decode_request.001",
		Title:      "Decode a JSON request",
		Chapter:    "json",
		PatternID:  "json.decode_request",
		Difficulty: 3,
		Type:       "full_recall",
		Prompt:     "Decode body.",
		Starter:    "package challenge",
		Hints:      []string{"hint"},
		Explanation: "exp",
		Validation: challenge.Validation{
			Compile: true,
			Gofmt:   true,
			Tests: []string{
				"TestDecodeRequest_Valid",
				"TestDecodeRequest_Invalid",
			},
			ASTRules: []string{
				"uses_json_new_decoder",
				"decoder_reads_request_body",
				"decode_receives_address",
			},
		},
		TestCode: `package challenge

import (
    "net/http"
    "strings"
    "testing"
)

func TestDecodeRequest_Valid(t *testing.T) {
    req, _ := http.NewRequest("POST", "/", strings.NewReader(` + "`" + `{"name":"Acme"}` + "`" + `))
    res, err := decodeRequest(req)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if res.Name != "Acme" {
        t.Errorf("expected Acme, got %q", res.Name)
    }
}

func TestDecodeRequest_Invalid(t *testing.T) {
    req, _ := http.NewRequest("POST", "/", strings.NewReader(` + "`" + `{invalid` + "`" + `))
    _, err := decodeRequest(req)
    if err == nil {
        t.Error("expected error for invalid JSON, got nil")
    }
}
`,
	}
}

func TestGrade_Success(t *testing.T) {
	c := getTestChallenge()

	correctSubmission := `package challenge

import (
    "encoding/json"
    "net/http"
)

type CreateCompanyRequest struct {
    Name string ` + "`" + `json:"name"` + "`" + `
}

func decodeRequest(r *http.Request) (CreateCompanyRequest, error) {
    var input CreateCompanyRequest
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        return input, err
    }
    return input, nil
}
`

	resp, err := Grade(c, correctSubmission)
	if err != nil {
		t.Fatalf("failed to grade: %v", err)
	}

	if !resp.CompileOK {
		t.Errorf("expected compilation to succeed, got compile_ok = false. Message: %v", resp.Feedback)
	}

	if !resp.Correct {
		t.Errorf("expected solution to be correct, got correct = false. Results: %+v", resp.Tests)
	}

	if len(resp.Tests) != 2 {
		t.Errorf("expected 2 tests, got %d", len(resp.Tests))
	}

	for _, ast := range resp.ASTChecks {
		if !ast.Passed {
			t.Errorf("expected AST check %s to pass, but it failed", ast.Name)
		}
	}

	if resp.NextAction != "advance" {
		t.Errorf("expected next action 'advance', got %q", resp.NextAction)
	}
}

func TestGrade_ASTFailure_NoAddress(t *testing.T) {
	c := getTestChallenge()

	// Missing & (address-of) operator: calls Decode(input) instead of Decode(&input)
	submissionWithASTBug := `package challenge

import (
    "encoding/json"
    "net/http"
)

type CreateCompanyRequest struct {
    Name string ` + "`" + `json:"name"` + "`" + `
}

func decodeRequest(r *http.Request) (CreateCompanyRequest, error) {
    var input CreateCompanyRequest
    // Missing & address-of operator! This compiles but doesn't modify input, causing behavior failures.
    if err := json.NewDecoder(r.Body).Decode(input); err != nil {
        return input, err
    }
    return input, nil
}
`

	resp, err := Grade(c, submissionWithASTBug)
	if err != nil {
		t.Fatalf("failed to grade: %v", err)
	}

	if !resp.CompileOK {
		if len(resp.Feedback) > 0 {
			t.Errorf("expected compile to succeed, but failed: %s", resp.Feedback[0].Message)
		} else {
			t.Errorf("expected compile to succeed")
		}
	}

	// Correct should be false because AST check decode_receives_address fails!
	if resp.Correct {
		t.Error("expected solution to be incorrect due to failing AST address-of check, but got correct = true")
	}

	var addressCheckPassed bool
	for _, ast := range resp.ASTChecks {
		if ast.Name == "decode_receives_address" {
			addressCheckPassed = ast.Passed
		}
	}

	if addressCheckPassed {
		t.Error("expected AST check 'decode_receives_address' to fail, but it passed")
	}
}

func TestVerifyASTRules_AllRules(t *testing.T) {
	src := `package main
import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
)

type Server struct {}

func (s *Server) Update(w http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()
    var val string
    _ = json.NewDecoder(r.Body).Decode(&val)
    _ = json.NewEncoder(w).Encode(val)
}

func TestSomething() {
    _ = httptest.NewRequest("GET", "/", nil)
}
`

	rules := []string{
		"uses_json_new_decoder",
		"decoder_reads_request_body",
		"decode_receives_address",
		"uses_json_new_encoder",
		"pointer_receiver",
		"uses_httptest_new_request",
		"uses_defer_body_close",
	}

	checks, err := VerifyASTRules(src, rules)
	if err != nil {
		t.Fatalf("failed to verify AST rules: %v", err)
	}

	if len(checks) != len(rules) {
		t.Fatalf("expected %d AST checks, got %d", len(rules), len(checks))
	}

	for _, check := range checks {
		if !check.Passed {
			t.Errorf("expected rule %q to pass, but it failed", check.Name)
		}
	}
}
