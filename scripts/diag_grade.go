package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/godojo/godojo/internal/challenge"
)

func main() {
	c, err := challenge.LoadFile("./content/http/get_header.yaml")
	if err != nil {
		fmt.Printf("Error loading challenge: %v\n", err)
		os.Exit(1)
	}

	submission := `package challenge

import (
	"io"
	"net/http"
)

func headerReaderHandler(w http.ResponseWriter, r *http.Request) {
	// godojo:start
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "no user id found", http.StatusBadRequest)
		return
	}
	io.WriteString(w, "User: "+userID)
	// godojo:end
}
`

	tmpDir, err := os.MkdirTemp("", "godojo-diag-*")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	_ = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module challenge\n\ngo 1.22\n"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "challenge.go"), []byte(submission), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "challenge_test.go"), []byte(c.TestCode), 0644)

	cmd := exec.Command("go", "test", "-v")
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
}
