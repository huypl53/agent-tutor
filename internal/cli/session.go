package cli

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
)

// SessionName derives a unique, deterministic tmux session name from the
// project directory path.  Format: agent-tutor-<basename>-<short-hash>
// where the hash is the first 8 hex chars of SHA-256(abs-path).
func SessionName(projectDir string) string {
	base := filepath.Base(projectDir)
	// Sanitise basename: tmux session names cannot contain dots or colons.
	base = strings.NewReplacer(".", "-", ":", "-").Replace(base)

	h := sha256.Sum256([]byte(projectDir))
	short := fmt.Sprintf("%x", h[:4]) // 8 hex chars

	return fmt.Sprintf("agent-tutor-%s-%s", base, short)
}
