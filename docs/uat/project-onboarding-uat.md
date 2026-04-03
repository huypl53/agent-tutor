# UAT: Project Onboarding Feature

> How to verify the project onboarding feature works end-to-end via Claude Code.

## Prerequisites

- The `feature/project-onboarding` branch is checked out
- Node.js installed, `npm install` completed
- `npm test` passes (77 tests, 0 failures)

## Test Method: tmux + Claude Code print mode

Use `claude -p` (print/one-shot mode) with `--plugin-dir` to exercise MCP tools non-interactively.
This avoids TUI parsing issues while still testing the full MCP tool chain.

```bash
# Base command pattern (run from a test project directory):
claude -p "<prompt>" --plugin-dir <path-to-plugin>
```

## Setup: Create a Test Project

Create a minimal Node.js backend project to scan:

```bash
TEST_DIR=$(mktemp -d -t uat-onboard-XXXX)
cd "$TEST_DIR"

# Create a recognizable backend-api project
cat > package.json << 'JSON'
{
  "name": "uat-test-api",
  "version": "1.0.0",
  "dependencies": { "express": "^4.18.0", "pg": "^8.0.0" },
  "devDependencies": { "jest": "^29.0.0" }
}
JSON

cat > server.js << 'JS'
const express = require('express');
const app = express();
app.get('/health', (req, res) => res.json({ ok: true }));
module.exports = app;
JS

mkdir -p src/routes src/models test
echo 'module.exports = (app) => { app.get("/api/users", (req, res) => res.json([])); };' > src/routes/users.js
echo 'class User { constructor(name) { this.name = name; } }; module.exports = User;' > src/models/user.js
echo 'const assert = require("assert"); assert.ok(true);' > test/users.test.js
```

## Test Cases

### TC1: scan_project MCP tool

**Goal:** Verify ProjectScanner detects project type, parses manifest, scans structure.

**Command:**
```bash
claude -p "Call the scan_project tool and show me the raw JSON result" \
  --plugin-dir <plugin-dir> \
  --allowedTools "mcp__agent-tutor__scan_project"
```

**Expected output contains:**
- `"projectType": "backend-api"`
- `"displayName": "Backend API"`
- `stack` object with `framework: "Express"` and `language: "JavaScript/TypeScript"`
- `entryPoints` array containing `server.js`
- `analysisDomains` array with `["architecture","api","data","testing","dev-setup"]`
- `manifest.dependencies` with `express` and `pg`
- `scannedAt` timestamp

**Pass criteria:** All fields present and correct project type detected.

### TC2: get_project_profile MCP tool

**Goal:** Verify saved profile can be retrieved.

**Command:**
```bash
claude -p "Call get_project_profile and show me the result" \
  --plugin-dir <plugin-dir> \
  --allowedTools "mcp__agent-tutor__get_project_profile"
```

**Expected output contains:**
- Same profile data from TC1
- `availableDocs` array (may be empty initially)

**Pass criteria:** Profile matches what was saved in TC1. If TC1 hasn't run, returns "No project profile. Run scan_project first."

### TC3: save_project_doc MCP tool

**Goal:** Verify docs are saved to `.agent-tutor/docs/` and tracked in state.

**Command:**
```bash
claude -p 'Call save_project_doc with name="test-architecture" and content="# Architecture\n\nThis is a test doc.\n\n## Patterns\n\nMVC pattern with Express routes."' \
  --plugin-dir <plugin-dir> \
  --allowedTools "mcp__agent-tutor__save_project_doc"
```

**Verify file exists:**
```bash
cat "$TEST_DIR/.agent-tutor/docs/test-architecture.md"
# Should contain the markdown content
```

**Verify state updated:**
```bash
cat "$TEST_DIR/.agent-tutor/state.json" | python3 -m json.tool
# project.docs should include "test-architecture"
```

**Pass criteria:** File exists with correct content, state.json tracks doc name.

### TC4: State migration v1 to v2

**Goal:** Verify v1 state auto-migrates to v2 with `project: null` field.

**Setup:**
```bash
mkdir -p "$TEST_DIR/.agent-tutor"
cat > "$TEST_DIR/.agent-tutor/state.json" << 'JSON'
{
  "version": 1,
  "topics": { "js-basics": { "id": "js-basics", "title": "JS Basics", "status": "introduced", "complexity": null, "dependencies": [], "started": "2026-01-01", "moments": [], "lessonFile": null } },
  "plan": null,
  "session": null
}
JSON
```

**Command:**
```bash
claude -p "Call get_project_profile" \
  --plugin-dir <plugin-dir> \
  --allowedTools "mcp__agent-tutor__get_project_profile"
```

**Verify migration:**
```bash
cat "$TEST_DIR/.agent-tutor/state.json" | python3 -m json.tool
# version should be 2
# project should be null
# topics.js-basics should be preserved
```

**Pass criteria:** State version is 2, project field exists (null), existing topics preserved.

### TC5: /atu:onboard skill registration

**Goal:** Verify the skill is discoverable by Claude Code.

**Command:**
```bash
claude -p "List your available skills that start with atu" \
  --plugin-dir <plugin-dir>
```

**Expected:** Response mentions `atu:onboard` with description about project analysis.

**Pass criteria:** Skill is listed and description matches SKILL.md frontmatter.

### TC6: /atu:deep-dive skill registration

**Goal:** Verify the deep-dive skill is discoverable.

**Command:**
```bash
claude -p "List your available skills that start with atu" \
  --plugin-dir <plugin-dir>
```

**Expected:** Response mentions `atu:deep-dive` with description about module analysis.

**Pass criteria:** Skill is listed and description matches SKILL.md frontmatter.

### TC7: End-to-end scan + doc + profile workflow

**Goal:** Verify the full workflow chains correctly.

**Commands (sequential):**
```bash
# 1. Scan
claude -p "Scan this project using scan_project" \
  --plugin-dir <plugin-dir> \
  --allowedTools "mcp__agent-tutor__scan_project"

# 2. Save a doc
claude -p 'Save a project doc named "overview" with content "# Project Overview\nA test Express API."' \
  --plugin-dir <plugin-dir> \
  --allowedTools "mcp__agent-tutor__save_project_doc"

# 3. Get profile (should show doc in availableDocs)
claude -p "Get the project profile" \
  --plugin-dir <plugin-dir> \
  --allowedTools "mcp__agent-tutor__get_project_profile"
```

**Pass criteria:**
- Step 3 output includes `"availableDocs": ["overview"]`
- Profile shows `backend-api` type
- `.agent-tutor/docs/overview.md` exists on disk

## Teardown

```bash
rm -rf "$TEST_DIR"
```

## Results — 2026-04-03

| Test | Status | Notes |
|------|--------|-------|
| TC1: scan_project | PASS | Detected `backend-api` type, Express framework, JS/TypeScript language, `server.js` entry point, 5 analysis domains, manifest with express+pg deps, scannedAt timestamp present |
| TC2: get_project_profile | PASS | Returned same profile from TC1, availableDocs empty as expected, all fields matched |
| TC3: save_project_doc | PASS | File created at `.agent-tutor/docs/test-architecture.md` with correct content, state.json `project.docs` contains `["test-architecture"]` |
| TC4: v1→v2 migration | PASS | State version migrated from 1 to 2, `project` field set to `null`, `js-basics` topic preserved intact |
| TC5: /atu:onboard registration | PASS | Skill listed as `agent-tutor:atu-onboard` with description "Analyze the current project — detect stack, architecture, patterns" |
| TC6: /atu:deep-dive registration | PASS | Skill listed as `agent-tutor:atu-deep-dive` with description "Exhaustive analysis of a specific module or feature in the project" |
| TC7: E2E workflow | PASS | scan_project -> save_project_doc -> get_project_profile chain worked; profile shows `backend-api`, `availableDocs: ["overview"]`, `overview.md` exists on disk |
