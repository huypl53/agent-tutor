package cli

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
)

const maxBaseLen = 30

// SessionName derives a unique, deterministic tmux session name from the
// project directory path.  Format: agent-tutor-<basename>-<short-hash>
// where the hash is the first 8 hex chars of SHA-256(abs-path).
// The input is resolved to an absolute path before hashing so that
// relative and absolute forms of the same directory produce the same name.
func SessionName(projectDir string) string {
	abs, err := filepath.Abs(projectDir)
	if err == nil {
		projectDir = abs
	}

	base := filepath.Base(projectDir)
	// Sanitise basename: tmux session names cannot contain dots, colons, or spaces.
	base = strings.NewReplacer(".", "-", ":", "-", " ", "-").Replace(base)
	if len(base) > maxBaseLen {
		base = base[:maxBaseLen]
	}

	h := sha256.Sum256([]byte(projectDir))
	short := fmt.Sprintf("%x", h[:4]) // 8 hex chars

	return fmt.Sprintf("agent-tutor-%s-%s", base, short)
}
