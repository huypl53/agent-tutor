package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallLocal(t *testing.T) {
	dir := t.TempDir()
	if err := Install(dir, ScopeLocal); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify plugin files exist
	files := []string{
		".agent-tutor/plugin/.claude-plugin/plugin.json",
		".agent-tutor/plugin/commands/atu:check.md",
		".agent-tutor/plugin/commands/atu:hint.md",
		".agent-tutor/plugin/commands/atu:explain.md",
		".agent-tutor/plugin/commands/atu:save.md",
		".agent-tutor/plugin/commands/atu:debug.md",
		".agent-tutor/plugin/commands/atu:review.md",
		".agent-tutor/plugin/commands/atu:decompose.md",
		".agent-tutor/plugin/commands/atu:workflow.md",
	}
	for _, f := range files {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to exist: %v", f, err)
		}
	}

	// Verify CLAUDE.md was created with sentinels
	claudeMD := filepath.Join(dir, ".claude", "CLAUDE.md")
	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatalf("CLAUDE.md not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<!-- BEGIN AGENT-TUTOR -->") {
		t.Error("missing BEGIN sentinel")
	}
	if !strings.Contains(content, "<!-- END AGENT-TUTOR -->") {
		t.Error("missing END sentinel")
	}
	if !strings.Contains(content, "get_student_context") {
		t.Error("missing MCP tool reference")
	}
}

func TestInstallLocalIncludesAutoSaveInstructions(t *testing.T) {
	dir := t.TempDir()
	if err := Install(dir, ScopeLocal); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "Lesson Auto-Save") {
		t.Error("missing Lesson Auto-Save section in CLAUDE.md")
	}
	if !strings.Contains(content, "./lessons/") {
		t.Error("missing ./lessons/ path reference in auto-save instructions")
	}
}

func TestInstallLocalAppendsCLAUDEmd(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0o755)
	os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte("# My Project\n\nExisting content.\n"), 0o644)

	if err := Install(dir, ScopeLocal); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(claudeDir, "CLAUDE.md"))
	content := string(data)
	if !strings.HasPrefix(content, "# My Project") {
		t.Error("existing content was overwritten")
	}
	if !strings.Contains(content, "<!-- BEGIN AGENT-TUTOR -->") {
		t.Error("agent-tutor section not appended")
	}
}

func TestInstallLocalIdempotent(t *testing.T) {
	dir := t.TempDir()
	Install(dir, ScopeLocal)
	Install(dir, ScopeLocal)

	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	count := strings.Count(string(data), "<!-- BEGIN AGENT-TUTOR -->")
	if count != 1 {
		t.Errorf("expected 1 sentinel block, got %d", count)
	}
}

func TestUninstallLocal(t *testing.T) {
	dir := t.TempDir()
	Install(dir, ScopeLocal)
	if err := Uninstall(dir, ScopeLocal); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	// Plugin dir removed
	pluginDir := filepath.Join(dir, ".agent-tutor", "plugin")
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Error("plugin directory should be removed")
	}

	// CLAUDE.md sentinel section removed
	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	if strings.Contains(string(data), "<!-- BEGIN AGENT-TUTOR -->") {
		t.Error("sentinel section should be removed")
	}
}

func TestUninstallPreservesOtherContent(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0o755)
	os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte("# My Project\n\nKeep this.\n"), 0o644)

	Install(dir, ScopeLocal)
	Uninstall(dir, ScopeLocal)

	data, _ := os.ReadFile(filepath.Join(claudeDir, "CLAUDE.md"))
	content := string(data)
	if !strings.Contains(content, "Keep this.") {
		t.Error("non-agent-tutor content was removed")
	}
}

func TestInstallGlobal(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	if err := Install("", ScopeGlobal); err != nil {
		t.Fatalf("Install global failed: %v", err)
	}

	// Verify skills exist
	skills := []string{
		".claude/skills/atu-check/SKILL.md",
		".claude/skills/atu-hint/SKILL.md",
		".claude/skills/atu-explain/SKILL.md",
		".claude/skills/atu-save/SKILL.md",
	}
	for _, s := range skills {
		path := filepath.Join(dir, s)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to exist: %v", s, err)
		}
	}

	// Verify CLAUDE.md
	data, err := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("global CLAUDE.md not created: %v", err)
	}
	if !strings.Contains(string(data), "<!-- BEGIN AGENT-TUTOR -->") {
		t.Error("missing sentinel in global CLAUDE.md")
	}
}

func TestUninstallGlobal(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	Install("", ScopeGlobal)
	if err := Uninstall("", ScopeGlobal); err != nil {
		t.Fatalf("Uninstall global failed: %v", err)
	}

	// Skill directories removed
	for _, name := range []string{"atu-check", "atu-hint", "atu-explain", "atu-save"} {
		path := filepath.Join(dir, ".claude", "skills", name)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("skill directory %s should be removed", name)
		}
	}

	// CLAUDE.md sentinel section removed
	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	if strings.Contains(string(data), "<!-- BEGIN AGENT-TUTOR -->") {
		t.Error("sentinel section should be removed from global CLAUDE.md")
	}
}

func TestRestoreColons(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"commands/atu-check.md", "commands/atu:check.md"},
		{"commands/atu-hint.md", "commands/atu:hint.md"},
		{"commands/atu-explain.md", "commands/atu:explain.md"},
		{"commands/atu-save.md", "commands/atu:save.md"},
		{".claude-plugin/plugin.json", ".claude-plugin/plugin.json"},
		{"atu-check.md", "atu-check.md"},
		{"other-file.md", "other-file.md"},
		{"commands", "commands"},
		// Skill directories must NOT get colons
		{"skills/atu-guided-debugging", "skills/atu-guided-debugging"},
		{"skills/atu-guided-debugging/SKILL.md", "skills/atu-guided-debugging/SKILL.md"},
		{"hooks/large-file-detect.js", "hooks/large-file-detect.js"},
	}
	for _, tt := range tests {
		got := restoreColons(tt.input)
		if got != tt.want {
			t.Errorf("restoreColons(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInstallLocalIncludesSkills(t *testing.T) {
	dir := t.TempDir()
	if err := Install(dir, ScopeLocal); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	skills := []string{
		".agent-tutor/plugin/skills/atu-guided-debugging/SKILL.md",
		".agent-tutor/plugin/skills/atu-guided-debugging/references/phases.md",
		".agent-tutor/plugin/skills/atu-guided-debugging/references/examples.md",
		".agent-tutor/plugin/skills/atu-problem-decomposition/SKILL.md",
		".agent-tutor/plugin/skills/atu-problem-decomposition/references/techniques.md",
		".agent-tutor/plugin/skills/atu-code-review-learning/SKILL.md",
		".agent-tutor/plugin/skills/atu-code-review-learning/references/checklist.md",
		".agent-tutor/plugin/skills/atu-dev-workflow/SKILL.md",
		".agent-tutor/plugin/skills/atu-dev-workflow/references/rules.md",
	}
	for _, s := range skills {
		path := filepath.Join(dir, s)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to exist: %v", s, err)
		}
	}
}

func TestInstallLocalIncludesHooks(t *testing.T) {
	dir := t.TempDir()
	if err := Install(dir, ScopeLocal); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	hooks := []string{
		".agent-tutor/plugin/hooks/large-file-detect.js",
		".agent-tutor/plugin/hooks/error-pattern-detect.js",
	}
	for _, h := range hooks {
		path := filepath.Join(dir, h)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to exist: %v", h, err)
		}
	}
}
