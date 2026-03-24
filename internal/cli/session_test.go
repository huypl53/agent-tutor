package cli

import (
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
