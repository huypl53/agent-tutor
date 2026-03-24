package watcher

import (
	"strings"
	"testing"
)

func TestTerminalDiff(t *testing.T) {
	tw := &TerminalWatcher{}

	old := "$ ls\nfile1.go\nfile2.go\n"
	newContent := "$ ls\nfile1.go\nfile2.go\n$ go build\nerror: syntax\n"

	diff := tw.diff(old, newContent)
	if diff == "" {
		t.Error("expected non-empty diff")
	}
	if !strings.Contains(diff, "go build") {
		t.Errorf("expected diff to contain 'go build', got: %s", diff)
	}
}

func TestTerminalDiffCleared(t *testing.T) {
	tw := &TerminalWatcher{}

	old := "$ ls\nfile1.go\nfile2.go\n$ go build\nerror: syntax\n"
	newContent := "$ echo hello\nhello\n"

	diff := tw.diff(old, newContent)
	if diff == "" {
		t.Error("expected non-empty diff when screen cleared")
	}
	if !strings.Contains(diff, "hello") {
		t.Errorf("expected diff to contain 'hello', got: %s", diff)
	}
}

func TestTerminalDiffNoChange(t *testing.T) {
	tw := &TerminalWatcher{}

	content := "$ ls\nfile1.go\n"
	diff := tw.diff(content, content)
	if diff != "" {
		t.Errorf("expected empty diff for identical content, got: %s", diff)
	}
}

func TestDetectErrorPattern(t *testing.T) {
	tw := &TerminalWatcher{}

	tests := []struct {
		content string
		isError bool
	}{
		{"$ go build\nOK", false},
		{"error: something failed", true},
		{"Error: cannot find module", true},
		{"FAIL: TestFoo", true},
		{"panic: runtime error", true},
		{"all tests passed", false},
		{"fatal: not a git repository", true},
		{"Fatal: crash", true},
		{"Traceback (most recent call last):", true},
		{"exception: bad value", true},
	}

	for _, tt := range tests {
		got := tw.hasError(tt.content)
		if got != tt.isError {
			t.Errorf("hasError(%q) = %v, want %v", tt.content, got, tt.isError)
		}
	}
}
