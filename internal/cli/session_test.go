package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionName_Format(t *testing.T) {
	name := SessionName("/home/user/projects/my-app")
	if !strings.HasPrefix(name, "agent-tutor-my-app-") {
		t.Fatalf("expected prefix 'agent-tutor-my-app-', got %q", name)
	}
	// 8 hex chars after the last dash
	hash := name[len(name)-8:]
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("expected hex hash suffix, got %q in %q", string(c), name)
		}
	}
}

func TestSessionName_Deterministic(t *testing.T) {
	a := SessionName("/home/user/projects/my-app")
	b := SessionName("/home/user/projects/my-app")
	if a != b {
		t.Fatalf("expected deterministic result, got %q != %q", a, b)
	}
}

func TestSessionName_DifferentPaths(t *testing.T) {
	a := SessionName("/home/user/projects/app-a")
	b := SessionName("/home/user/projects/app-b")
	if a == b {
		t.Fatalf("expected different session names for different paths, both got %q", a)
	}
}

func TestSessionName_SameBasename(t *testing.T) {
	// Two dirs with same basename but different parents must differ.
	a := SessionName("/home/alice/code/myapp")
	b := SessionName("/home/bob/code/myapp")
	if a == b {
		t.Fatalf("expected different session names for same basename but different paths, both got %q", a)
	}
	// Both should share the basename prefix.
	if !strings.HasPrefix(a, "agent-tutor-myapp-") || !strings.HasPrefix(b, "agent-tutor-myapp-") {
		t.Fatalf("expected shared prefix, got %q and %q", a, b)
	}
}

func TestSessionName_SanitisesDots(t *testing.T) {
	name := SessionName("/home/user/my.project")
	if strings.Contains(name, ".") {
		t.Fatalf("dots should be replaced, got %q", name)
	}
}

func TestSessionName_SanitisesColons(t *testing.T) {
	name := SessionName("/home/user/my:project")
	if strings.Contains(name, ":") {
		t.Fatalf("colons should be replaced, got %q", name)
	}
}

func TestSessionName_SanitisesSpaces(t *testing.T) {
	name := SessionName("/home/user/my project")
	if strings.Contains(name, " ") {
		t.Fatalf("spaces should be replaced, got %q", name)
	}
}

func TestSessionName_TruncatesLongBasename(t *testing.T) {
	long := "/home/user/" + strings.Repeat("a", 50)
	name := SessionName(long)
	// Format: agent-tutor-<base>-<hash>
	// base should be at most maxBaseLen (30)
	prefix := "agent-tutor-"
	suffix := name[len(name)-9:] // "-" + 8 hex chars
	base := name[len(prefix) : len(name)-len(suffix)]
	if len(base) > maxBaseLen {
		t.Fatalf("basename should be truncated to %d chars, got %d in %q", maxBaseLen, len(base), name)
	}
}

func TestSessionName_RelativeEqualsAbsolute(t *testing.T) {
	// "." should resolve to the same absolute path as os.Getwd
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	a := SessionName(".")
	b := SessionName(cwd)
	if a != b {
		t.Fatalf("relative '.' and absolute %q should produce the same name, got %q vs %q", cwd, a, b)
	}
}

func TestResolveProjectDir_Default(t *testing.T) {
	dir := resolveProjectDir(nil)
	cwd, _ := os.Getwd()
	if dir != cwd {
		t.Fatalf("expected cwd %q, got %q", cwd, dir)
	}
}

func TestResolveProjectDir_WithArg(t *testing.T) {
	tmp := t.TempDir()
	dir := resolveProjectDir([]string{tmp})
	abs, _ := filepath.Abs(tmp)
	if dir != abs {
		t.Fatalf("expected %q, got %q", abs, dir)
	}
}

func TestResolveProjectDir_RelativeArg(t *testing.T) {
	dir := resolveProjectDir([]string{"."})
	cwd, _ := os.Getwd()
	if dir != cwd {
		t.Fatalf("expected cwd %q for relative '.', got %q", cwd, dir)
	}
}
