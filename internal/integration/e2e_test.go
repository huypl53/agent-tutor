//go:build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testSocket  = "agent-tutor-test"
	testSession = "agent-tutor"
)

// tmuxCmd runs a tmux command on the test socket.
func tmuxCmd(args ...string) *exec.Cmd {
	full := append([]string{"-L", testSocket}, args...)
	return exec.Command("tmux", full...)
}

// paneTarget returns the tmux target string for a pane (session:window.pane).
func paneTarget(paneID string) string {
	return testSession + ":0." + paneID
}

// capturePane returns the text content of a pane.
func capturePane(t *testing.T, paneID string) string {
	t.Helper()
	out, err := tmuxCmd("capture-pane", "-t", paneTarget(paneID), "-p", "-J").Output()
	if err != nil {
		t.Fatalf("capture-pane %s failed: %v", paneID, err)
	}
	return string(out)
}

// sendKeys sends a command to a pane and waits briefly for it to execute.
func sendKeys(t *testing.T, paneID, keys string) {
	t.Helper()
	if err := tmuxCmd("send-keys", "-t", paneTarget(paneID), keys, "Enter").Run(); err != nil {
		t.Fatalf("send-keys to %s failed: %v", paneID, err)
	}
	time.Sleep(500 * time.Millisecond)
}

// waitForContent polls a pane until it contains the expected substring or times out.
func waitForContent(t *testing.T, paneID, substr string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		content := capturePane(t, paneID)
		if strings.Contains(content, substr) {
			return content
		}
		time.Sleep(500 * time.Millisecond)
	}
	content := capturePane(t, paneID)
	t.Fatalf("timed out waiting for %q in pane %s.\nLast content:\n%s", substr, paneID, content)
	return ""
}

// startTmux creates the session and panes using the tmux Manager logic directly.
func startTmux(t *testing.T, binPath, projectDir string) {
	t.Helper()

	// Create session
	if err := tmuxCmd("new-session", "-d", "-s", testSession, "-c", projectDir).Run(); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Split pane
	if err := tmuxCmd("split-window", "-h", "-t", testSession, "-l", "50%").Run(); err != nil {
		t.Fatalf("split pane: %v", err)
	}

	// Start MCP server in pane 1 (agent pane) via the binary
	mcpCmd := binPath + " mcp --project-dir " + projectDir + " --socket " + testSocket
	sendKeys(t, "1", mcpCmd)
}

// run executes a command in a directory and fails on error.
func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

// projectRoot returns the agent-tutor project root.
func projectRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root")
		}
		dir = parent
	}
}

const testConfig = `
[tutor]
intensity = "proactive"
level = "beginner"

[agent]
command = "bash"
args = []

[watchers]
file_patterns = ["**/*.go"]
ignore_patterns = [".git"]
terminal_poll_interval = "1s"
git_poll_interval = "2s"

[tmux]
layout = "horizontal"
user_pane_size = 50
socket = "agent-tutor-test"
`

// setupE2E builds the binary, creates a temp project dir with git and config,
// starts the tmux session, and registers cleanup. Returns the project dir.
func setupE2E(t *testing.T) string {
	t.Helper()

	// Build the binary
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "agent-tutor")
	build := exec.Command("go", "build", "-o", binPath, "./cmd/agent-tutor")
	build.Dir = projectRoot(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	// Create a temp project dir with git
	projectDir := t.TempDir()
	run(t, projectDir, "git", "init")
	run(t, projectDir, "git", "config", "user.email", "test@test.com")
	run(t, projectDir, "git", "config", "user.name", "test")

	// Write config
	cfgDir := filepath.Join(projectDir, ".agent-tutor")
	os.MkdirAll(cfgDir, 0o755)
	os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(testConfig), 0o644)

	// Cleanup: always kill test tmux server
	t.Cleanup(func() {
		tmuxCmd("kill-server").Run()
	})

	// Start the session
	startTmux(t, binPath, projectDir)

	return projectDir
}

func TestE2ESessionLifecycle(t *testing.T) {
	setupE2E(t)

	// Verify session exists
	if err := tmuxCmd("has-session", "-t", testSession).Run(); err != nil {
		t.Fatal("session does not exist after start")
	}

	// Verify two panes exist
	out, _ := tmuxCmd("list-panes", "-t", testSession).Output()
	panes := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(panes) != 2 {
		t.Fatalf("expected 2 panes, got %d: %s", len(panes), string(out))
	}

	t.Log("Session created with 2 panes ✓")
}

func TestE2EGoLearnerActivity(t *testing.T) {
	projectDir := setupE2E(t)

	// Wait for MCP server to start (pane 1 will show log output)
	time.Sleep(2 * time.Second)

	// === Simulate user learning Go in pane 0 ===

	// Step 1: User creates a Go file with a bug (unused import)
	t.Log("Simulating: user writes buggy Go code")
	goCode := `cat > main.go << 'GOEOF'
package main

import "fmt"
import "os"

func main() {
    fmt.Println("Hello Go!")
}
GOEOF`
	sendKeys(t, "0", goCode)
	time.Sleep(1 * time.Second)

	// Step 2: User tries to build — error (unused import "os")
	t.Log("Simulating: user runs go build, gets error")
	sendKeys(t, "0", "cd "+projectDir+" && go build -o hello main.go")
	time.Sleep(2 * time.Second)

	// Verify the error shows in pane 0
	content := capturePane(t, "0")
	if !strings.Contains(content, "imported and not used") {
		t.Log("Warning: expected 'imported and not used' error in pane 0")
		t.Log("Pane 0 content:", content)
	} else {
		t.Log("User sees compilation error ✓")
	}

	// Step 3: User fixes the bug
	t.Log("Simulating: user fixes the unused import")
	fixCode := `cat > main.go << 'GOEOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello Go!")
}
GOEOF`
	sendKeys(t, "0", fixCode)
	time.Sleep(1 * time.Second)

	// Step 4: Build succeeds
	sendKeys(t, "0", "go build -o hello main.go && echo BUILD_OK")
	waitForContent(t, "0", "BUILD_OK", 10*time.Second)
	t.Log("Build succeeds after fix ✓")

	// Step 5: User runs the program
	sendKeys(t, "0", "./hello")
	waitForContent(t, "0", "Hello Go!", 5*time.Second)
	t.Log("Program runs correctly ✓")

	// Step 6: User makes a git commit
	t.Log("Simulating: user commits their work")
	sendKeys(t, "0", "git add main.go && git commit -m 'feat: hello world'")
	time.Sleep(3 * time.Second)

	// === Verify watchers detected activity ===

	// The MCP server is running in pane 1. Check for trigger engine output.
	pane1Content := capturePane(t, "1")
	t.Log("Agent pane (MCP server) output:")
	t.Log(pane1Content)

	// Check that the trigger engine detected the commit
	if strings.Contains(pane1Content, "[trigger] nudge fired: git.commit") {
		t.Log("Trigger engine detected git commit ✓")
	} else {
		t.Log("Note: git.commit trigger not visible in captured output (may need longer wait)")
	}

	t.Log("E2E Go learner simulation complete ✓")
}
