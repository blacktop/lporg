package main

import (
	"os"
	"testing"
)

func TestAbsPath(t *testing.T) {
	// get current user
	homeDir, err := os.UserHomeDir()

	if err != nil {
		t.Fatalf("Failed get user homedir: %v", err)
	}

	observed := []string{
		"~",
		"/~",
		"/~/../test",
		"/test/remove/remove/remove/../../..",
		"/test/remove/../.",
		"/test/...",
	}

	expected := []string{
		homeDir,
		"/~",
		"/test",
		"/test",
		"/test",
		"/test/...",
	}

	for i, s := range observed {
		path, err := absPath(s)

		if err != nil {
			t.Fatalf("Error absPath(%q): %v", s, err)
		}

		if path != expected[i] {
			t.Fatalf("Invalid path: %q, excepted: %q", path, expected[i])
		}
	}
}
