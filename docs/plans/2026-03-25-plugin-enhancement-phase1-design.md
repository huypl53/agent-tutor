# Plugin Enhancement Phase 1 — Design

**Date:** 2026-03-25
**Status:** Approved
**Approach:** Hybrid (commands flat, skills with references, hooks for detection)
**Scope:** Plugin-only (no Go backend changes in Phase 1)

---

## Goals

Enhance the agent-tutor Claude Code plugin so the coding agent teaches students *how* to think — not just what the answer is. Inspired by Claude's `.claude` skill/workflow patterns.

## Constraints

- Local install only by default. Never touch `~/.claude/` unless `--scope global` explicitly passed.
- All changes are plugin-only (embedded files). No Go recompilation needed for teaching content.
- Hooks are non-blocking (advisory only, always exit 0).
- Existing 4 commands remain unchanged in behavior; enhancements are additive.

---

## New Directory Structure

```
internal/plugin/embed/
├── .claude-plugin/
│   └── plugin.json                        # updated: add skills reference
├── commands/
│   ├── atu-check.md                       # existing (CLAUDE.md enhanced)
│   ├── atu-hint.md                        # existing
│   ├── atu-explain.md                     # existing
│   ├── atu-save.md                        # existing
│   ├── atu-debug.md                       # NEW: guided debugging session
│   ├── atu-review.md                      # NEW: self-review coaching
│   ├── atu-decompose.md                   # NEW: problem decomposition
│   └── atu-workflow.md                    # NEW: development workflow coaching
├── skills/
│   ├── atu-guided-debugging/
│   │   ├── SKILL.md
│   │   └── references/
│   │       ├── phases.md
│   │       └── examples.md
│   ├── atu-problem-decomposition/
│   │   ├── SKILL.md
│   │   └── references/
│   │       └── techniques.md
│   ├── atu-code-review-learning/
│   │   ├── SKILL.md
│   │   └── references/
│   │       └── checklist.md
│   └── atu-dev-workflow/
│       ├── SKILL.md
│       └── references/
│           └── rules.md
└── hooks/
    ├── large-file-detect.js               # PostToolUse Write/Edit: >200 LOC warning
    └── error-pattern-detect.js            # PostToolUse Bash: error pattern detection
```

---

## New Commands

### `/atu:debug`
Guided debugging session. Calls `get_terminal_activity` + `get_recent_file_changes`, then coaches through 4 phases with questions rather than answers. Activates `atu-guided-debugging` skill.

Phases:
1. **Investigate** — Read the error, reproduce, check recent changes
2. **Analyze** — Find working examples, compare with broken code
3. **Hypothesize** — Form a theory, test minimally
4. **Fix** — Write a test first, fix once, verify

### `/atu:review`
Self-review coaching. Calls `get_git_activity` + `get_recent_file_changes`, walks student through a graduated checklist (3 items for beginners, full list for advanced). Agent highlights one missed area and asks student to find two more. Activates `atu-code-review-learning` skill.

### `/atu:decompose`
Problem decomposition coaching. Takes argument or reads current context. Coaches student through smallest useful increment, working backwards from goal, identifying dependencies. Activates `atu-problem-decomposition` skill.

### `/atu:workflow`
Development workflow coaching. Reviews git history + file changes, picks the single highest-impact habit to coach on. Positive framing first (acknowledge good habits), then one improvement. Activates `atu-dev-workflow` skill.

---

## Skills (Teaching Methodologies)

### `atu-guided-debugging`
- `SKILL.md`: Decision tree for stuck-type identification, phase gates (don't advance to Fix without a hypothesis), teaching prompts per phase
- `references/phases.md`: Detailed description of each phase with common pitfalls
- `references/examples.md`: 2-3 worked examples (off-by-one, nil pointer, async race condition)

### `atu-problem-decomposition`
- `SKILL.md`: 5-technique dispatch (simplification, inversion, scale game, collision-zone, meta-pattern), student-facing questions per technique
- `references/techniques.md`: Each technique with when-to-use heuristic and student-friendly explanation

### `atu-code-review-learning`
- `SKILL.md`: Evidence-based approach, graduated difficulty by student level, one-issue-then-ask strategy
- `references/checklist.md`: Full checklist (correctness, readability, security, performance, testability) with teaching notes per item

### `atu-dev-workflow`
- `SKILL.md`: Pattern observation from git/file history, single-habit coaching rule, positive framing rule
- `references/rules.md`: Development rules adapted from Claude's development-rules.md with beginner-friendly explanations

---

## Hooks

Both hooks are non-blocking (always exit 0), inject suggestions via `additionalContext`.

### `large-file-detect.js`
- Trigger: PostToolUse on Write/Edit tools
- Logic: Count lines in written/edited file
- Threshold: >200 LOC
- Output: `additionalContext` suggesting the agent use it as a `/atu:decompose` teachable moment
- Never blocks execution

### `error-pattern-detect.js`
- Trigger: PostToolUse on Bash tool
- Logic: Scan stdout/stderr for patterns: `panic`, `FAIL`, `traceback`, `Error:`, `exception`, `segfault`, `fatal`
- Output: `additionalContext` suggesting the agent guide the student through the error using `/atu:debug` or `/atu:explain`
- Never blocks execution

### Installation
- `installLocal()` merges hook entries into project's `.claude/settings.json` (creates if missing)
- Never overwrites existing hooks — appends to PostToolUse array
- Each entry includes a comment identifying it as agent-tutor for clean removal
- `uninstallLocal()` removes only agent-tutor hook entries

---

## Enhanced CLAUDE.md Section

Adds to existing MCP tools table and coaching behavior section:

### Teaching Skills
```
Activate these skills by loading the corresponding SKILL.md when:
- atu-guided-debugging: /atu:debug invoked, or error detected in proactive mode
- atu-problem-decomposition: /atu:decompose invoked, or student appears stuck on large task
- atu-code-review-learning: /atu:review invoked, or student completes a feature
- atu-dev-workflow: /atu:workflow invoked, or student makes 3+ commits in session
```

### Pedagogical Principles
- Ask questions before giving answers
- One teaching point per interaction — never overwhelm
- Praise specific good behavior before suggesting improvements
- Match vocabulary and checklist depth to student level
- Never fix code silently in proactive mode — always explain what and why

### Hook Awareness
- When `additionalContext` mentions a teachable moment, incorporate it naturally into the next response
- Don't parrot hook text verbatim — use it as a trigger for genuine teaching

---

## `plugin.go` Changes

| Function | Change |
|---|---|
| `extractEmbedded()` | No change — already walks full tree, picks up `skills/` and `hooks/` automatically |
| `installLocal()` | After extraction, merge hook entries into `.claude/settings.json` |
| `uninstallLocal()` | Remove agent-tutor hook entries from `.claude/settings.json`; remove `skills/` and `hooks/` dirs |
| `installGlobal()` | Also create skill dirs for 4 teaching methodologies under `~/.claude/skills/` |
| `claudeMDSection` | Updated with teaching skills, pedagogical principles, hook awareness, updated command table |

### Settings Merge Strategy
```
Read .claude/settings.json (create {} if missing)
Append to hooks.PostToolUse[] — never replace existing entries
Tag each entry with agent-tutor identifier for clean removal
Write back atomically
```

---

## What Is NOT in Phase 1

- No Go backend changes (trigger engine rules for complex patterns)
- No stateful detection (repeated edit pattern, no-test detection) — these require Go and are Phase 2
- No new MCP tools — existing 6 tools are sufficient for Phase 1 commands

---

## Success Criteria

- `agent-tutor install-plugin` extracts skills, hooks, and new commands to `.agent-tutor/plugin/`
- `.claude/settings.json` gets hook entries without overwriting existing content
- `/atu:debug`, `/atu:review`, `/atu:decompose`, `/atu:workflow` commands available in Claude Code
- Agent loads appropriate skill when command is invoked
- Hook scripts run after Write/Edit/Bash and inject advisory context when applicable
- `agent-tutor uninstall-plugin` cleanly removes all of the above
- No changes to `~/.claude/` during local install
