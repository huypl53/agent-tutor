# Project Onboarding Feature — Design

## Goal

Help students understand existing codebases through automated analysis and guided learning. The tutor analyzes the project's structure, stack, architecture, and patterns, then tailors teaching around the student's actual codebase.

## Target Users

- Students joining an **unfamiliar codebase** — need to understand someone else's project to contribute
- Students building their **own project** — want the tutor to understand their project so teaching is contextual ("your `auth.js` uses this pattern")

## Architecture

Hybrid MCP extraction + LLM sub-agent reasoning.

```
/atu:onboard
    │
    ├─ 1. Call scan_project (MCP) ──► project-profile.json
    │      Fast mechanical extraction: file tree, manifests,
    │      config detection, entry points. No source reading.
    │
    ├─ 2. Look up project type in CSV ──► which domains to analyze
    │
    ├─ 3. Spawn sub-agents in parallel (one per domain)
    │      Each sub-agent:
    │        - Receives the project skeleton JSON
    │        - Reads source files in its domain
    │        - Writes findings via save_project_doc
    │        - Returns a 2-3 sentence summary
    │
    ├─ 4. Collect summaries, write index.md
    │
    └─ 5. Generate learning plan from gaps
         (what the project uses vs student's known topics)
```

**Trigger modes:**
- **Auto** — quick scan runs on first session when `state.project` is null. Fast (~seconds), no sub-agents.
- **Manual** — `/atu:onboard` runs full exhaustive analysis with sub-agents.
- **Focused** — `/atu:deep-dive <module>` spawns a single sub-agent for one directory/feature.

## New MCP Tools

### `scan_project`

Fast mechanical extraction, no source reading.

- Scans file tree (respecting .gitignore)
- Parses manifests (package.json, requirements.txt, Cargo.toml, go.mod, pyproject.toml, etc.)
- Detects project type from CSV lookup (14 types)
- Identifies entry points, config files, test patterns
- Returns structured JSON:

```json
{
  "projectType": "backend-api",
  "stack": { "language": "Python", "framework": "FastAPI", "database": "PostgreSQL" },
  "structure": "monolith",
  "entryPoints": ["src/main.py"],
  "configFiles": [".env", "alembic.ini"],
  "testPatterns": ["tests/**/*.py"],
  "dependencies": { "fastapi": "^0.100.0", "sqlalchemy": "^2.0.0" }
}
```

### `get_project_profile`

Returns stored project profile and doc index.

- Reads `.agent-tutor/project-profile.json` and `.agent-tutor/docs/index.md`
- Returns what's available so the tutor knows what's been analyzed

### `save_project_doc`

Writes an analysis doc from a sub-agent.

- Input: `{ name, content }` (e.g., `name: "architecture"`, `content: "# Architecture\n..."`)
- Writes to `.agent-tutor/docs/<name>.md`
- Updates `index.md` with the new entry

## Project Type Detection

CSV-driven type system at `plugin/data/project-types.csv`.

| Column | Purpose |
|--------|---------|
| `type_id` | e.g., `web-app`, `backend-api`, `ai-llm` |
| `display_name` | Human-readable name |
| `key_files` | Detection patterns (e.g., `package.json,next.config.*,vite.config.*`) |
| `critical_dirs` | Where to focus deep analysis (e.g., `src/,app/,pages/`) |
| `analysis_domains` | Which sub-agents to spawn (e.g., `api,components,state,routing`) |
| `test_patterns` | Test file patterns (e.g., `*.test.ts,*.spec.ts`) |
| `entry_points` | Typical entry files (e.g., `index.ts,main.ts,app.ts`) |

### The 14 Project Types

1. `web-app` — React, Vue, Angular, Svelte, Next.js, etc.
2. `backend-api` — Express, FastAPI, Django, Spring, Go HTTP, etc.
3. `full-stack` — combined frontend + backend (monorepo or mono-app)
4. `cli-tool` — Commander, Click, Cobra, clap, etc.
5. `library` — npm package, PyPI, crate, Go module
6. `mobile-app` — React Native, Flutter, Swift, Kotlin
7. `desktop-app` — Electron, Tauri, Qt, etc.
8. `game` — Unity, Godot, Bevy, Pygame, etc.
9. `data-pipeline` — Spark, Airflow, dbt, pandas, etc.
10. `extension` — browser extension, VS Code extension, plugin
11. `infra` — Terraform, Pulumi, CloudFormation, Ansible
12. `embedded` — Arduino, ESP32, Zephyr, bare-metal
13. `ai-llm` — LangChain, LlamaIndex, transformers, MCP servers, agents
14. `devops` — CI/CD configs, Docker, K8s, Helm, monitoring

Detection runs top-down — first match wins, with `full-stack` checking for both frontend + backend signals. If no match, defaults to a generic profile.

## Analysis Domains & Sub-Agent Orchestration

| Domain | What it analyzes | Output doc |
|--------|-----------------|------------|
| `architecture` | Overall pattern (MVC, microservices, hexagonal), data flow, key abstractions | `architecture.md` |
| `api` | Routes/endpoints, request/response shapes, middleware chain, auth flow | `api-contracts.md` |
| `data` | Database schema, models, migrations, ORM patterns, validation | `data-models.md` |
| `components` | UI components, state management, routing, styling patterns | `component-inventory.md` |
| `infra` | CI/CD, Docker, deployment config, env management | `infra-guide.md` |
| `testing` | Test patterns, coverage approach, fixtures, mocks | `testing-guide.md` |
| `dev-setup` | How to install, build, run, and contribute | `development-guide.md` |

Not all domains run for every project type — the CSV's `analysis_domains` column controls which ones are relevant. A CLI tool gets `architecture,testing,dev-setup` but not `components` or `api`.

**Deep-dive** (`/atu:deep-dive <module>`) spawns a single focused sub-agent that does exhaustive file-by-file analysis of one directory/feature — imports, exports, dependency graph, data flow, key patterns, contributor guidance. Writes to `.agent-tutor/docs/deep-dive-<name>.md`.

## State Schema Changes

Extend `state.json` with a `project` key (v1 → v2):

```json
{
  "version": 2,
  "topics": {},
  "plan": null,
  "session": null,
  "project": {
    "type": "backend-api",
    "stack": {
      "language": "Python",
      "framework": "FastAPI",
      "database": "PostgreSQL",
      "orm": "SQLAlchemy",
      "testing": "pytest"
    },
    "structure": "monolith",
    "entryPoints": ["src/main.py"],
    "scannedAt": "2026-04-03T10:00:00Z",
    "docs": ["architecture.md", "api-contracts.md", "data-models.md", "development-guide.md", "testing-guide.md"],
    "deepDives": ["auth-middleware"]
  }
}
```

**Version migration:** On load, if `version === 1`, adds `project: null` and bumps to v2. Existing data untouched.

**Auto-scan trigger:** When `get_student_context` or `get_coaching_config` is called and `state.project` is null, `scan_project` runs automatically.

**Staleness:** `scannedAt` lets the tutor decide if a rescan is needed. Suggest rescanning if profile is older than 7 days.

## New Skills

### `/atu:onboard` (`plugin/skills/atu-onboard/SKILL.md`)

Full analysis orchestrator:

1. Calls `scan_project` to get the project skeleton
2. Checks if docs already exist — if so, offers: rescan, deep-dive, or skip
3. Looks up project type in CSV to determine which domains to analyze
4. Spawns sub-agents in parallel (one per domain)
5. Collects summaries, writes `index.md` via `save_project_doc`
6. Presents project overview to the student
7. Compares project stack against student's existing topics — identifies gaps
8. Offers to create a learning plan based on the gaps

### `/atu:deep-dive` (`plugin/skills/atu-deep-dive/SKILL.md`)

Focused module analysis:

1. If no project profile exists, runs quick scan first
2. Presents project structure — student picks a directory or feature
3. Spawns a single sub-agent for exhaustive file-by-file analysis
4. Produces: purpose, file inventory, dependency graph, data flow, key patterns, contributor guidance
5. Writes to `.agent-tutor/docs/deep-dive-<name>.md`
6. Tutor walks through findings pedagogically

### Updates to Tutor Instructions

New "Project Awareness" section — when `.agent-tutor/docs/` exists, reference project-specific code in teaching instead of abstract examples.

## File Layout

```
.agent-tutor/
├── state.json                      # adds project key (v2)
├── config.json                     # unchanged
├── project-profile.json            # quick scan result
└── docs/
    ├── index.md                    # master doc index with summaries
    ├── architecture.md
    ├── api-contracts.md
    ├── data-models.md
    ├── component-inventory.md
    ├── infra-guide.md
    ├── testing-guide.md
    ├── development-guide.md
    └── deep-dive-<name>.md

plugin/
├── data/
│   └── project-types.csv           # 14 project type definitions
├── skills/
│   ├── atu-onboard/SKILL.md
│   └── atu-deep-dive/SKILL.md
└── servers/
    ├── tutoring-mcp.js             # +3 new tools
    ├── state-manager.js            # v2 schema, project methods
    └── project-scanner.js          # scan logic module
```

## Testing Strategy

**Unit tests** (`test/project-scanner.test.js`):
- Project type detection for each of the 14 types (fixture directories with key files)
- Manifest parsing (package.json, requirements.txt, go.mod, pyproject.toml, Cargo.toml)
- CSV loading and lookup
- Fallback to generic profile when no type matches
- Multi-type detection (full-stack = frontend + backend signals)

**Unit tests** (`test/state-manager.test.js` — additions):
- v1 → v2 migration adds `project: null`
- `saveProjectProfile` / `getProjectProfile` CRUD
- `saveProjectDoc` / `getProjectDoc` write and read
- `listProjectDocs` returns doc inventory

**MCP tool tests** (`test/mcp-tools.test.js` — additions):
- `scan_project` returns valid skeleton for a fixture project
- `get_project_profile` returns null when no scan exists, returns profile after scan
- `save_project_doc` writes file and updates index

Sub-agent orchestration is not unit-tested — it's skill-level behavior (markdown instructions) using Claude Code's Agent tool.

## Success Criteria

After running `/atu:onboard`, the student can:

1. Understand what the project does (purpose, architecture pattern)
2. Know the tech stack and why those choices were made
3. Navigate the codebase (where things live, entry points, key files)
4. Understand data flow (how a request/action travels through the system)
5. Know how to build, test, and run the project
6. Get a personalized learning plan based on what the project uses vs what they know
