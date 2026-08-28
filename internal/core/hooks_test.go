package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExecuteHook(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zenpomo-hook-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	outFile := filepath.Join(tmpDir, "hook_out.txt")
	hookCmd := "echo $ZENPOMO_SESSION > " + outFile

	ExecuteHook(hookCmd, SessionWork, "Test Task")

	// Wait up to 1 second for async execution
	time.Sleep(200 * time.Millisecond)

	b, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Hook did not execute: %v", err)
	}
	if len(b) == 0 {
		t.Fatalf("Hook output file is empty")
	}
}
