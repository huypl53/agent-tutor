# Plugin Enhancement Phase 1 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enhance the agent-tutor Claude Code plugin with 4 new teaching commands, structured skill methodologies, and advisory hooks — inspired by Claude's `.claude` patterns.

**Architecture:** Plugin-only changes (no Go backend). New command files activate the agent's teaching mode. Skills live in `.agent-tutor/plugin/skills/` and the agent reads them via the Read tool when commands are invoked. Two JS hooks inject advisory context after Write/Edit/Bash tool uses. `plugin.go` gains settings.json merge logic to install hooks into `.claude/settings.json` without touching user-level `~/.claude/`.

**Tech Stack:** Go (embed, encoding/json, os), Node.js (hook scripts), Markdown (commands, skills)

---

## Context: How the Plugin Works

- `embed/` tree is compiled into the binary via `//go:embed all:embed`
- `installLocal(projectDir)` extracts embed tree to `.agent-tutor/plugin/` and appends to `.claude/CLAUDE.md`
- `restoreColons()` converts `commands/atu-check.md` → `commands/atu:check.md` for Claude Code command naming
- **Bug**: `restoreColons` currently converts ANY path starting with `atu-`, which would corrupt skill directory names like `atu-guided-debugging` → `atu:guided-debugging`. Must fix first.
- Skills are NOT auto-discovered from `.agent-tutor/plugin/skills/`; instead, `CLAUDE.md` injection tells the agent to `Read` them when commands are invoked

---

## Task 1: Fix `restoreColons` Bug

**Files:**
- Modify: `internal/plugin/plugin.go:303-313`
- Modify: `internal/plugin/plugin_test.go:192-211`

**Step 1: Write failing test showing the bug**

Add to `TestRestoreColons` in `plugin_test.go`:
```go
// Skill directories must NOT get colons
{"skills/atu-guided-debugging", "skills/atu-guided-debugging"},
{"skills/atu-guided-debugging/SKILL.md", "skills/atu-guided-debugging/SKILL.md"},
{"hooks/large-file-detect.js", "hooks/large-file-detect.js"},
```

**Step 2: Run to confirm it fails**

```bash
cd /home/huypham/code/spare/agent-tutor && go test ./internal/plugin/ -run TestRestoreColons -v
```

Expected: FAIL — `skills/atu-guided-debugging` would be converted to `skills/atu:guided-debugging`

**Step 3: Fix `restoreColons` — narrow scope to `commands/` only**

Replace the function body in `plugin.go`:
```go
func restoreColons(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	// Only restore colons for command files directly under commands/
	if dir == "commands" && strings.HasPrefix(base, "atu-") && strings.HasSuffix(base, ".md") {
		base = "atu:" + strings.TrimPrefix(base, "atu-")
	}
	if dir == "." {
		return base
	}
	return filepath.Join(dir, base)
}
```

Also update the existing test cases that tested the root-level `atu-check.md` case (no longer applies):
```go
// Remove: {"atu-check.md", "atu:check.md"} — doesn't occur in practice
// Keep: {"commands/atu-check.md", "commands/atu:check.md"}
```

**Step 4: Run tests to verify passing**

```bash
go test ./internal/plugin/ -run TestRestoreColons -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/plugin/plugin.go internal/plugin/plugin_test.go
git commit -m "fix: restoreColons must only apply to commands/ files, not skill dirs"
```

---

## Task 2: Add 4 New Command Files

**Files:**
- Create: `internal/plugin/embed/commands/atu-debug.md`
- Create: `internal/plugin/embed/commands/atu-review.md`
- Create: `internal/plugin/embed/commands/atu-decompose.md`
- Create: `internal/plugin/embed/commands/atu-workflow.md`
- Modify: `internal/plugin/plugin_test.go` (TestInstallLocal)

**Step 1: Write failing test**

Add to `TestInstallLocal` in `plugin_test.go`, extend the `files` slice:
```go
".agent-tutor/plugin/commands/atu:debug.md",
".agent-tutor/plugin/commands/atu:review.md",
".agent-tutor/plugin/commands/atu:decompose.md",
".agent-tutor/plugin/commands/atu:workflow.md",
```

**Step 2: Run to confirm fail**

```bash
go test ./internal/plugin/ -run TestInstallLocal -v
```

Expected: FAIL — files don't exist yet

**Step 3: Create `atu-debug.md`**

```markdown
---
name: atu:debug
description: Guided debugging session — walk through the error step by step instead of fixing it for you
---

You are guiding the student through a debugging session using the atu-guided-debugging methodology.
Load the methodology by reading `.agent-tutor/plugin/skills/atu-guided-debugging/SKILL.md`.

1. Call `get_terminal_activity` to see the current error
2. Call `get_recent_file_changes` to see what code changed
3. Call `get_coaching_config` to check student level
4. Follow the 4-phase methodology from the skill file — guide with questions, not answers
5. Do not fix the bug for the student; help them reason through it
```

**Step 4: Create `atu-review.md`**

```markdown
---
name: atu:review
description: Self-review coaching — learn to critique your own code before asking for help
---

You are teaching the student to self-review their code using the atu-code-review-learning methodology.
Load the methodology by reading `.agent-tutor/plugin/skills/atu-code-review-learning/SKILL.md`.

1. Call `get_git_activity` to see recent commits and changes
2. Call `get_recent_file_changes` to see the code
3. Call `get_coaching_config` to check student level
4. Follow the graduated checklist from the skill file
5. Find one issue yourself, then ask the student to find two more — build their eye for quality
```

**Step 5: Create `atu-decompose.md`**

```markdown
---
name: atu:decompose
description: Problem decomposition coaching — break a big task into manageable pieces using structured thinking
---

You are teaching the student to decompose a problem using the atu-problem-decomposition methodology.
Load the methodology by reading `.agent-tutor/plugin/skills/atu-problem-decomposition/SKILL.md`.

1. Call `get_student_context` to understand what the student is working on
2. Call `get_coaching_config` to check student level
3. If the user provided a problem description as an argument, use that; otherwise infer from context
4. Follow the technique dispatch from the skill file — ask "what's the smallest useful version?" before anything else
5. Guide with questions; never decompose the problem for them entirely
```

**Step 6: Create `atu-workflow.md`**

```markdown
---
name: atu:workflow
description: Development workflow coaching — build good habits around commits, testing, and file organization
---

You are coaching the student on development workflow using the atu-dev-workflow methodology.
Load the methodology by reading `.agent-tutor/plugin/skills/atu-dev-workflow/SKILL.md`.

1. Call `get_git_activity` to review recent commit history
2. Call `get_recent_file_changes` to observe file organization patterns
3. Call `get_coaching_config` to check student level
4. Follow the skill file — pick the single highest-impact habit to address
5. Acknowledge one good habit first, then suggest one improvement — never lecture
```

**Step 7: Run tests to verify passing**

```bash
go test ./internal/plugin/ -run TestInstallLocal -v
```

Expected: PASS

**Step 8: Commit**

```bash
git add internal/plugin/embed/commands/ internal/plugin/plugin_test.go
git commit -m "feat: add /atu:debug, /atu:review, /atu:decompose, /atu:workflow commands"
```

---

## Task 3: Add Teaching Skill Directories

**Files:**
- Create: `internal/plugin/embed/skills/atu-guided-debugging/SKILL.md`
- Create: `internal/plugin/embed/skills/atu-guided-debugging/references/phases.md`
- Create: `internal/plugin/embed/skills/atu-guided-debugging/references/examples.md`
- Create: `internal/plugin/embed/skills/atu-problem-decomposition/SKILL.md`
- Create: `internal/plugin/embed/skills/atu-problem-decomposition/references/techniques.md`
- Create: `internal/plugin/embed/skills/atu-code-review-learning/SKILL.md`
- Create: `internal/plugin/embed/skills/atu-code-review-learning/references/checklist.md`
- Create: `internal/plugin/embed/skills/atu-dev-workflow/SKILL.md`
- Create: `internal/plugin/embed/skills/atu-dev-workflow/references/rules.md`
- Modify: `internal/plugin/plugin_test.go`

**Step 1: Write failing test**

Add new test to `plugin_test.go`:
```go
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
```

**Step 2: Run to confirm fail**

```bash
go test ./internal/plugin/ -run TestInstallLocalIncludesSkills -v
```

Expected: FAIL

**Step 3: Create `atu-guided-debugging/SKILL.md`**

```markdown
# Guided Debugging Methodology

Teach the student to debug systematically. Never jump to a fix — walk them through 4 phases.

## Stuck-Type Dispatch

Before starting, identify what kind of problem this is:
- **Error message** (crash, exception, panic) → Start at Phase 1
- **Wrong output** (code runs but result is incorrect) → Start at Phase 2
- **Performance** (too slow) → Start at Phase 3
- **Build failure** (won't compile/lint) → Phase 1, focus on the compiler message

## Phase Gates

Do NOT advance to the next phase until the student has answered the phase's key question.
If student says "I don't know", give a hint — but not the answer.

## Phase 1: Investigate

**Goal:** Understand the error before touching code.
**Key question to ask:** "What does this error message tell you?"
**Teaching prompt:** "Read the error from bottom to top — the last line usually has the most specific clue."
**Do:** Ask them to paste the full error if they haven't. Ask where in the code the error points.
**Don't:** Explain the error yourself.

## Phase 2: Analyze

**Goal:** Find a working example to compare against.
**Key question:** "Can you find a place in the code where something similar works?"
**Teaching prompt:** "What's different between the working version and the broken one?"
**Do:** Ask them to diff the working and broken code in their head.
**Don't:** Find the difference for them.

## Phase 3: Hypothesize

**Goal:** Form one testable theory.
**Key question:** "What do you think is causing this?"
**Teaching prompt:** "A hypothesis should be falsifiable — how would you test yours?"
**Do:** Help them narrow to one theory. Encourage them to add a print/log to test it.
**Don't:** Validate or reject their hypothesis — let the code do that.

## Phase 4: Fix

**Gate:** Only reach here after student has articulated a hypothesis AND tested it.
**Key question:** "Before you change the code, can you write a test that would catch this bug?"
**Teaching prompt:** "A test that would have caught this bug is worth more than the fix itself."
**Do:** Guide them toward the minimal fix. Ask "does this fix just this bug, or does it hide it?"
**Don't:** Write the fix for them.

For detailed worked examples, read `references/examples.md`.
For phase descriptions with common pitfalls, read `references/phases.md`.
```

**Step 4: Create `atu-guided-debugging/references/phases.md`**

```markdown
# Debugging Phases — Detailed Reference

## Phase 1: Investigate — Common Pitfalls

- Students often read only the last line of an error, missing the root cause in the middle
- Stack traces go from most-recent to least-recent; the root cause is usually at the bottom
- "File not found" errors are almost never about the file — they're about the path. Ask "what is the current working directory?"
- Type errors in dynamic languages (Python, JS) usually mean a variable is None/undefined upstream

**Good coaching questions:**
- "Which line number does the error point to?"
- "Is this error coming from your code or a library?"
- "Have you seen this error before? What caused it then?"

## Phase 2: Analyze — Common Pitfalls

- Students search for "the bug" instead of comparing working vs broken
- A working example doesn't have to be in their code — documentation examples count
- Often the difference is one character (missing `await`, wrong variable name, off-by-one index)

**Good coaching questions:**
- "When did this last work?"
- "What changed since it last worked? (git diff is your friend)"
- "Is there any test or example in the codebase that does something similar?"

## Phase 3: Hypothesize — Common Pitfalls

- Students form vague hypotheses ("something is wrong with the data") — push for specificity
- The first hypothesis is usually wrong; that's fine, that's how debugging works
- Students often skip testing the hypothesis and go straight to a fix

**Good coaching questions:**
- "Can you make that more specific? Which data, and what is wrong with it?"
- "How could you add a print statement to test that theory?"
- "What would you expect to see if your hypothesis is correct?"

## Phase 4: Fix — Common Pitfalls

- Fixing without a test means the bug will likely return
- Students often over-fix — making 5 changes at once, making it impossible to know which worked
- "It works now" is not a verification — ask them to explain why it works

**Good coaching questions:**
- "Can you write a test that would fail before your fix and pass after?"
- "Is this the minimal change that fixes the bug?"
- "Why does this fix work? What was the root cause?"
```

**Step 5: Create `atu-guided-debugging/references/examples.md`**

```markdown
# Debugging Examples — Worked Walkthroughs

## Example 1: Off-By-One Index Error

**Scenario:** Student has `items[len(items)]` and gets an index out of range panic.

**Phase 1 coaching:** "The panic says index 5 is out of range on a slice of length 5. What's the valid index range for a 5-element slice?"
→ Student realizes indices go 0-4, not 0-5.

**Phase 2 coaching:** "Where else in the code do you access items by index? What pattern do they use?"
→ Student finds `items[i]` in a loop and sees `i < len(items)` not `i <= len(items)`.

**Phase 3 coaching:** "So your hypothesis is that `len(items)` gives the length but not the last valid index. How would you verify that?"
→ Student prints `len(items)` and the index value to confirm.

**Phase 4 coaching:** "Before you fix it, can you write a test with a known slice that triggers this panic?"

---

## Example 2: Nil Pointer / Null Reference

**Scenario:** Student gets a nil pointer dereference. The error points to line 42: `user.Name`.

**Phase 1 coaching:** "The error says you're dereferencing a nil pointer at `user.Name`. What does that mean about `user`?"
→ Student realizes `user` is nil.

**Phase 2 coaching:** "Trace backwards — where is `user` set? Is there a path through the code where it might not get assigned?"
→ Student finds a code path where `getUserByID` returns nil on error but the error isn't checked.

**Phase 3 coaching:** "So your hypothesis is that `getUserByID` returned nil and you used it without checking. How would you add a guard?"

**Phase 4 coaching:** "What should happen when the user isn't found? Error? Default value? Make that explicit before fixing."

---

## Example 3: Async Race Condition

**Scenario (JavaScript):** Student's function sometimes returns undefined. Works in isolation, breaks in production.

**Phase 1 coaching:** "It works sometimes and fails sometimes — what does that tell you about the cause?"
→ Student starts thinking about timing.

**Phase 2 coaching:** "Is this function async? Are there any awaits missing?"
→ Student finds `const data = fetch(url)` without `await`.

**Phase 3 coaching:** "So your hypothesis is that `fetch` returns a Promise and you're treating it as the resolved value. What would you see if you `console.log(data)` right now?"
→ Student logs it and sees `Promise { <pending> }`.

**Phase 4 coaching:** "Now write a test that asserts the return value is a string, not a Promise."
```

**Step 6: Create `atu-problem-decomposition/SKILL.md`**

```markdown
# Problem Decomposition Methodology

Teach the student to break large problems into manageable pieces. Never decompose for them — ask questions that lead them to do it themselves.

## First Question — Always

Before applying any technique, ask:
**"What's the smallest version of this that would be useful to someone?"**

This forces the student to think about value, not completeness. A working 20% is better than a planned 100%.

## Technique Dispatch

Pick the technique based on where the student is stuck:

| Student says... | Use technique |
|---|---|
| "I don't know where to start" | Working Backwards |
| "It's too big / overwhelming" | Smallest Increment |
| "I don't know which part to do first" | Dependency Mapping |
| "I've tried everything and it doesn't work" | Inversion |
| "I need to make it work at scale" | Scale Game |

## Techniques

### Working Backwards
Ask: "Describe the final state — what does 'done' look like exactly?"
Then: "What's the last thing that needs to happen before it's done?"
Continue backwards until they reach something they can start on now.

### Smallest Increment
Ask: "If you could only build one thing this week, what would give the most value?"
Then: "What's the minimum code needed to build just that?"
Goal: Get them to a running, useful slice before adding more.

### Dependency Mapping
Ask: "List all the pieces you need to build. Which ones depend on other ones?"
Then: "Which piece has no dependencies? Start there."
Draws out natural sequencing without imposing it.

### Inversion
Ask: "Instead of how to build it, describe how you'd break it — what would make it fail?"
Then: "Now flip each failure mode — that's a requirement you need to handle."
Useful when students can't articulate requirements directly.

### Scale Game
Ask: "How would this work if you had 1 user? 1 million users? 1 item? 1 billion items?"
Then: "What changes between those scenarios? Those are your design constraints."
Surfaces hidden assumptions about scale and state.

For detailed technique explanations, read `references/techniques.md`.
```

**Step 7: Create `atu-problem-decomposition/references/techniques.md`**

```markdown
# Decomposition Techniques — Detailed Reference

## Working Backwards

**When:** Student has a clear end goal but no idea where to start.
**Depth:** Go at least 4-5 steps back before stopping.
**Watch for:** Students who jump to implementation steps — redirect to outcomes.
**Example:** "Build a REST API" → Last step: "API is deployed and responding" → Before that: "Tests pass" → Before that: "Routes return correct data" → Before that: "Data layer returns correct data" → Before that: "Data model is defined" → Start here.

## Smallest Increment

**When:** Student is overwhelmed by scope or trying to build everything at once.
**Key insight:** Value is delivered when something works, not when it's complete.
**Watch for:** Students who say "but it won't be complete" — completeness is not the goal.
**Good follow-up:** "Who is the user of this smallest version? What problem does it solve for them?"

## Dependency Mapping

**When:** Student has a list of things to build but doesn't know the order.
**How to facilitate:** Ask them to literally list all pieces on paper/in a comment. Then ask about each pair: "Does A depend on B, or B on A?"
**Goal:** A directed graph; the things with no incoming edges are where to start.
**Common mistake students make:** Over-engineering dependencies; often things are more independent than students assume.

## Inversion

**When:** Student is stuck because they can't think positively about the requirements.
**Power:** Flipping to "how would this fail?" is often much easier than "how should this work?"
**Process:** List 5 ways the system could fail → Each failure is a requirement → Prioritize by likelihood and impact.
**Example:** "Users will lose data if the database is down" → Requirement: "Handle database unavailability gracefully."

## Scale Game

**When:** Student is making architecture decisions without understanding constraints.
**Key:** Most systems only need to handle 1-10x their current load; premature scaling is waste.
**Good question:** "At what scale does your current approach break? Is that scale realistic?"
**Outcome:** Student understands the difference between necessary complexity and speculative complexity.
```

**Step 8: Create `atu-code-review-learning/SKILL.md`**

```markdown
# Code Review as Learning Methodology

Teach the student to review their own code — one of the highest-leverage skills a developer can have.

## Core Principle

Never do the full review for the student. Find ONE issue and ask them to find more.
This builds the habit of critical thinking, not dependence on external review.

## Evidence-Based Approach

Avoid vague feedback. Every observation needs:
- **What:** The specific line or pattern
- **Why it matters:** The concrete consequence (bug risk, maintenance burden, performance)
- **What to look for:** How to spot this class of problem in the future

## Graduated Checklist

Match checklist depth to student level from `get_coaching_config`:

**Beginner (3 items):**
1. Does the code do what the comment/function name says?
2. What happens if the input is empty, null, or zero?
3. Would you understand this in 3 months?

**Intermediate (5 items):**
1-3 above, plus:
4. Are there any duplicated patterns that could be extracted?
5. Is every error handled or explicitly ignored?

**Advanced (full checklist):**
1-5 above, plus:
6. Are there any security concerns? (input validation, SQL injection, XSS, path traversal)
7. Are there performance bottlenecks? (N+1 queries, unnecessary allocations, blocking I/O)
8. Does this change preserve backward compatibility?
9. Is this the minimal change that achieves the goal?
10. Is there a test for the edge cases?

## The One-Then-Two Rule

1. Pick the most important issue from the checklist
2. Point to it specifically with the line number or code snippet
3. Ask: "Can you find two more things to improve?"
4. When they answer, validate and add nuance — don't just say "yes" or "no"

For the full annotated checklist, read `references/checklist.md`.
```

**Step 9: Create `atu-code-review-learning/references/checklist.md`**

```markdown
# Code Review Checklist — Annotated Reference

## Correctness

**Does the code do what the name/comment says?**
Teaching note: Function names are contracts. `getUserByEmail` that sometimes returns by ID is a bug waiting to happen. Ask student: "Read just the function name — what do you expect it to do? Now read the body. Does it match?"

**What happens on empty/null/zero input?**
Teaching note: This is the most common source of bugs. Ask: "What's the weirdest valid input this could receive? What about invalid input?" Off-by-one errors, nil dereferences, and division by zero live here.

**Are all code paths handled?**
Teaching note: `if/else` chains without a final `else`, switch statements without `default`, promises without `.catch()`. Ask: "Is there any path through this code that doesn't return a value or handle an error?"

## Readability

**Would you understand this in 3 months?**
Teaching note: This is a proxy for "is this well-written?" Ask student: "Without looking at the context, could you explain what this does in one sentence?" If they can't, neither can a future reader.

**Are variable names specific enough?**
Teaching note: `data`, `result`, `temp`, `x` are red flags. The name should encode the type and purpose. `userData` vs `user` — which tells you more?

**Is there duplication that should be extracted?**
Teaching note: The DRY principle isn't about zero duplication — it's about not having two places that need to change in sync. Ask: "If you needed to change how this works, how many places would you need to update?"

## Security (Advanced)

**Is user input validated before use?**
Teaching note: Never trust input from users, APIs, files, or databases. Ask: "Where does this value come from? Could it be malicious?"

**Are there SQL injection, XSS, or path traversal risks?**
Teaching note: String concatenation into queries, HTML templates without escaping, file paths from user input without canonicalization. Show one example and ask them to find others.

## Performance (Advanced)

**Are there N+1 query patterns?**
Teaching note: A loop that makes a database call in each iteration is the most common performance bug. Ask: "How many database queries does this code make for N records?"

**Is there unnecessary work inside loops?**
Teaching note: Function calls that are constant (like `len(slice)` in Go or `array.length` in JS if not changing) moved inside loops.

## Testability (Advanced)

**Is the function testable in isolation?**
Teaching note: Functions that reach out to global state, databases, or the network without dependency injection are hard to test. Ask: "How would you write a unit test for this function?"

**Is there a test for the happy path and at least one edge case?**
Teaching note: A test that only tests the happy path gives false confidence. Ask: "What test would have caught this bug?"
```

**Step 10: Create `atu-dev-workflow/SKILL.md`**

```markdown
# Development Workflow Methodology

Teach good habits around commits, testing, and code organization — without being preachy.

## Core Rules

1. **One habit per session.** Pick the single most impactful thing to address. Don't lecture on five things.
2. **Positive first.** Always acknowledge something the student did well before suggesting an improvement.
3. **Specific not general.** "Your commit message 'fix stuff' doesn't explain why" beats "write better commit messages."
4. **Show don't tell.** Give a concrete example of the better habit.

## What to Observe

From `get_git_activity`:
- Commit message quality (too vague, too long, missing context)
- Commit frequency (giant commits vs tiny commits)
- Commit scope (one commit for multiple unrelated changes)

From `get_recent_file_changes`:
- File length (>200 lines often means it needs splitting)
- Mixed concerns (database + HTTP + business logic in one file)

## Habit Priority Order

Pick the first habit in this list that the student is NOT doing well:

1. **Commit messages explain the why** — "Add error handling" vs "Handle nil user when auth token is expired"
2. **Commits are focused** — one logical change per commit, not "misc fixes"
3. **Tests accompany code changes** — new functionality has tests; bug fixes have regression tests
4. **Files have single responsibility** — one file, one purpose, reasonable length
5. **No dead code committed** — commented-out blocks, unused variables
6. **Dependencies are intentional** — not adding packages for one small function

## Framing Templates

**Good example:** "I noticed you're committing after each small working piece — that's exactly right. It makes it easy to bisect bugs later."

**Improvement example:** "Your last 3 commits are all called 'wip'. Try: one commit per logical change, with a message that explains *why* you made it, not *what* you changed (the diff already shows what)."

For detailed rules and examples, read `references/rules.md`.
```

**Step 11: Create `atu-dev-workflow/references/rules.md`**

```markdown
# Development Workflow Rules — Reference

## Commit Messages

**The format that works:**
```
<verb> <what changed> [— <why>]

feat: add retry logic for API calls — transient failures were causing user-visible errors
fix: clamp pagination offset to valid range — negative offsets caused 500 errors
refactor: extract auth middleware — was duplicated across 3 routes
```

**Common anti-patterns:**
- `fix stuff` — what stuff? why?
- `WIP` — never commit work-in-progress to shared branches
- `changes` — this is what git diff is for
- Messages longer than 72 characters in the first line — use a body for context

**Teaching point:** Commit messages are a letter to your future self. "Why did I do this?" is the question they should answer.

## Commit Scope

**One commit = one logical change.**
If you can describe a commit with "and", it should be split:
- ✗ "Fix login bug and add user profile page"
- ✓ Two commits: "fix: redirect loop on login" + "feat: add user profile page"

**Staged commits:** `git add -p` lets you commit only some lines in a file. Teach students this exists.

**When commits are too small:** Committing every line change is noise. A commit should represent a complete thought.

## Testing Discipline

**The workflow that works:**
1. Write the test (it fails — that's good)
2. Write the minimum code to make it pass
3. Commit both together

**Why test first?** Because if you write code first, you'll rationalize away tests. The test is the specification.

**Regression tests:** Every bug fix should come with a test that would have caught the bug. If you can't write the test, you don't understand the bug.

**What to test:** Edge cases, not just happy paths. Empty input, maximum input, concurrent access, network failure.

## File Organization

**Single Responsibility Principle:** One file does one thing. Signs a file should be split:
- >200 lines (rule of thumb, not a hard rule)
- Function at the top has nothing to do with a function at the bottom
- You often import only part of the file

**Naming:** `user-service.go` not `utils.go`. Be specific. `utils.go` eventually becomes a dumping ground.

**Dead code:** Delete it, don't comment it out. Git history preserves it if you ever need it.

## Dependency Management

**The question before adding a dependency:** "Could I write this in under 30 minutes, and would it be more maintainable than the library?"

**Dependencies have costs:** Security vulnerabilities, breaking changes, license incompatibility, abandoned packages.

**When dependencies make sense:** Cryptography (never write your own), database drivers, UI frameworks. When they don't: a function that parses a date, a helper that formats a string.
```

**Step 12: Run tests to verify skill files are picked up**

```bash
go test ./internal/plugin/ -run TestInstallLocalIncludesSkills -v
```

Expected: PASS (embed picks up the new files automatically)

**Step 13: Run all existing tests to check nothing broke**

```bash
go test ./internal/plugin/ -v
```

Expected: all PASS

**Step 14: Commit**

```bash
git add internal/plugin/embed/skills/ internal/plugin/plugin_test.go
git commit -m "feat: add 4 teaching skill directories with methodologies and references"
```

---

## Task 4: Add Hook Scripts

**Files:**
- Create: `internal/plugin/embed/hooks/large-file-detect.js`
- Create: `internal/plugin/embed/hooks/error-pattern-detect.js`
- Modify: `internal/plugin/plugin_test.go`

**Step 1: Write failing test**

Add to `plugin_test.go`:
```go
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
```

**Step 2: Run to confirm fail**

```bash
go test ./internal/plugin/ -run TestInstallLocalIncludesHooks -v
```

Expected: FAIL

**Step 3: Create `large-file-detect.js`**

```javascript
#!/usr/bin/env node
// Agent-tutor hook: warns when a file exceeds 200 LOC — advisory only, never blocks.
'use strict';

let raw = '';
process.stdin.on('data', chunk => { raw += chunk; });
process.stdin.on('end', () => {
  try {
    const input = JSON.parse(raw);
    const toolName = input?.tool_name;
    if (toolName !== 'Write' && toolName !== 'Edit') {
      process.exit(0);
    }

    const filePath = input?.tool_input?.file_path || input?.tool_input?.path;
    if (!filePath) process.exit(0);

    const fs = require('fs');
    let content;
    try {
      content = fs.readFileSync(filePath, 'utf8');
    } catch {
      process.exit(0); // Unreadable — skip silently
    }

    const lines = content.split('\n').length;
    if (lines > 200) {
      const result = {
        continue: true,
        hookSpecificOutput: {
          hookEventName: 'PostToolUse',
          additionalContext: [
            `File ${filePath} is now ${lines} lines — above the 200-line threshold.`,
            'Consider coaching the student on extracting functions or splitting modules.',
            'This is a good moment for /atu:decompose to practice breaking up large files.',
          ],
        },
      };
      process.stdout.write(JSON.stringify(result) + '\n');
    }
  } catch {
    // Malformed input — skip silently
  }
  process.exit(0);
});
```

**Step 4: Create `error-pattern-detect.js`**

```javascript
#!/usr/bin/env node
// Agent-tutor hook: detects error patterns in Bash output — advisory only, never blocks.
'use strict';

const ERROR_PATTERNS = [
  /panic:/i,
  /\bFAIL\b/,
  /Traceback \(most recent call last\)/i,
  /\bError:/,
  /\bexception\b/i,
  /segfault/i,
  /fatal error/i,
  /command not found/i,
  /\bno such file or directory\b/i,
];

let raw = '';
process.stdin.on('data', chunk => { raw += chunk; });
process.stdin.on('end', () => {
  try {
    const input = JSON.parse(raw);
    if (input?.tool_name !== 'Bash') process.exit(0);

    const stdout = input?.tool_response?.stdout || '';
    const stderr = input?.tool_response?.stderr || '';
    const combined = stdout + '\n' + stderr;

    const hasError = ERROR_PATTERNS.some(p => p.test(combined));
    if (hasError) {
      const result = {
        continue: true,
        hookSpecificOutput: {
          hookEventName: 'PostToolUse',
          additionalContext: [
            'An error was detected in the terminal output.',
            'If coaching intensity is not "silent", consider guiding the student through the error.',
            'Use /atu:debug for a guided debugging session, or /atu:explain to explain the specific error.',
          ],
        },
      };
      process.stdout.write(JSON.stringify(result) + '\n');
    }
  } catch {
    // Malformed input — skip silently
  }
  process.exit(0);
});
```

**Step 5: Run tests to verify**

```bash
go test ./internal/plugin/ -run TestInstallLocalIncludesHooks -v
```

Expected: PASS

**Step 6: Commit**

```bash
git add internal/plugin/embed/hooks/ internal/plugin/plugin_test.go
git commit -m "feat: add large-file-detect and error-pattern-detect advisory hooks"
```

---

## Task 5: Hook Settings Merge in `plugin.go`

**Files:**
- Modify: `internal/plugin/plugin.go`
- Modify: `internal/plugin/plugin_test.go`

**Step 1: Write failing tests**

Add to `plugin_test.go`:
```go
func TestInstallLocalMergesHookSettings(t *testing.T) {
	dir := t.TempDir()
	if err := Install(dir, ScopeLocal); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("settings.json not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "large-file-detect.js") {
		t.Error("missing large-file-detect.js hook in settings.json")
	}
	if !strings.Contains(content, "error-pattern-detect.js") {
		t.Error("missing error-pattern-detect.js hook in settings.json")
	}
}

func TestInstallLocalHookSettingsIdempotent(t *testing.T) {
	dir := t.TempDir()
	Install(dir, ScopeLocal)
	Install(dir, ScopeLocal)

	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	count := strings.Count(string(data), "large-file-detect.js")
	if count != 1 {
		t.Errorf("expected 1 hook entry, got %d", count)
	}
}

func TestInstallLocalPreservesExistingSettings(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0o755)
	existing := `{"permissions":{"allow":["Bash"]}}`
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(existing), 0o644)

	if err := Install(dir, ScopeLocal); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	content := string(data)
	if !strings.Contains(content, `"allow"`) {
		t.Error("existing permissions were lost")
	}
	if !strings.Contains(content, "large-file-detect.js") {
		t.Error("hook entries not added")
	}
}

func TestUninstallLocalRemovesHookSettings(t *testing.T) {
	dir := t.TempDir()
	Install(dir, ScopeLocal)
	if err := Uninstall(dir, ScopeLocal); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		return // settings.json removed entirely — acceptable
	}
	content := string(data)
	if strings.Contains(content, "large-file-detect.js") {
		t.Error("hook entry should be removed on uninstall")
	}
}
```

**Step 2: Run to confirm fail**

```bash
go test ./internal/plugin/ -run "TestInstallLocalMergesHookSettings|TestInstallLocalHookSettingsIdempotent|TestInstallLocalPreservesExistingSettings|TestUninstallLocalRemovesHookSettings" -v
```

Expected: FAIL

**Step 3: Add imports and types to `plugin.go`**

Add to the import block:
```go
"encoding/json"
```

Add after the `Scope` const block:
```go
// hookGroup matches Claude Code's settings.json PostToolUse hook format.
type hookGroup struct {
	Matcher string    `json:"matcher"`
	Hooks   []hookCmd `json:"hooks"`
}

type hookCmd struct {
	Type    string `json:"command"`
	Command string `json:"command"`
}
```

Wait — `hookCmd` has two fields with tag `json:"command"`. That's a mistake. Correct:
```go
type hookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}
```

**Step 4: Add `mergeHookSettings` to `plugin.go`**

```go
const agentTutorHookMarker = ".agent-tutor/plugin/hooks/"

// mergeHookSettings merges agent-tutor hook entries into .claude/settings.json.
// Preserves all existing settings. Idempotent.
func mergeHookSettings(settingsPath, hooksAbsDir string) error {
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return err
	}

	// Read existing settings as raw JSON map to preserve unknown fields.
	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &raw)
	}

	// Parse existing hooks section.
	hooks := make(map[string]json.RawMessage)
	if h, ok := raw["hooks"]; ok {
		_ = json.Unmarshal(h, &hooks)
	}

	// Parse existing PostToolUse entries.
	var postToolUse []hookGroup
	if p, ok := hooks["PostToolUse"]; ok {
		_ = json.Unmarshal(p, &postToolUse)
	}

	// Remove any existing agent-tutor entries (idempotency).
	postToolUse = removeAgentTutorHookGroups(postToolUse)

	// Add our two hooks.
	postToolUse = append(postToolUse,
		hookGroup{
			Matcher: "Write|Edit",
			Hooks: []hookCmd{{
				Type:    "command",
				Command: "node " + filepath.Join(hooksAbsDir, "large-file-detect.js"),
			}},
		},
		hookGroup{
			Matcher: "Bash",
			Hooks: []hookCmd{{
				Type:    "command",
				Command: "node " + filepath.Join(hooksAbsDir, "error-pattern-detect.js"),
			}},
		},
	)

	// Marshal back, preserving other fields.
	ptu, err := json.Marshal(postToolUse)
	if err != nil {
		return err
	}
	hooks["PostToolUse"] = ptu
	hooksRaw, err := json.Marshal(hooks)
	if err != nil {
		return err
	}
	raw["hooks"] = hooksRaw

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, out, 0o644)
}

// removeAgentTutorHookGroups filters out any hook groups that reference agent-tutor hooks.
func removeAgentTutorHookGroups(groups []hookGroup) []hookGroup {
	var result []hookGroup
	for _, g := range groups {
		isAgentTutor := false
		for _, h := range g.Hooks {
			if strings.Contains(h.Command, agentTutorHookMarker) {
				isAgentTutor = true
				break
			}
		}
		if !isAgentTutor {
			result = append(result, g)
		}
	}
	return result
}

// removeHookSettings removes agent-tutor hook entries from .claude/settings.json.
func removeHookSettings(settingsPath string) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil // File doesn't exist — nothing to do
	}

	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil // Unparseable — leave as-is
	}

	hooks := make(map[string]json.RawMessage)
	if h, ok := raw["hooks"]; ok {
		_ = json.Unmarshal(h, &hooks)
	}

	var postToolUse []hookGroup
	if p, ok := hooks["PostToolUse"]; ok {
		_ = json.Unmarshal(p, &postToolUse)
	}

	postToolUse = removeAgentTutorHookGroups(postToolUse)

	ptu, _ := json.Marshal(postToolUse)
	hooks["PostToolUse"] = ptu
	hooksRaw, _ := json.Marshal(hooks)
	raw["hooks"] = hooksRaw

	out, _ := json.MarshalIndent(raw, "", "  ")
	return os.WriteFile(settingsPath, out, 0o644)
}
```

**Step 5: Update `installLocal` to call `mergeHookSettings`**

Replace `installLocal` function body:
```go
func installLocal(projectDir string) error {
	destDir := PluginDir(projectDir)

	// Extract embedded files to .agent-tutor/plugin/
	if err := extractEmbedded(destDir); err != nil {
		return fmt.Errorf("extracting plugin files: %w", err)
	}

	// Merge hooks into .claude/settings.json (project-level, not user-level)
	settingsPath := filepath.Join(projectDir, ".claude", "settings.json")
	hooksDir := filepath.Join(destDir, "hooks")
	if err := mergeHookSettings(settingsPath, hooksDir); err != nil {
		return fmt.Errorf("updating settings.json: %w", err)
	}

	// Append to .claude/CLAUDE.md
	claudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	if err := appendCLAUDEmd(claudeMD); err != nil {
		return fmt.Errorf("updating CLAUDE.md: %w", err)
	}

	return nil
}
```

**Step 6: Update `uninstallLocal` to call `removeHookSettings`**

Replace `uninstallLocal` function body:
```go
func uninstallLocal(projectDir string) error {
	// Remove plugin directory
	pluginDir := PluginDir(projectDir)
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("removing plugin directory: %w", err)
	}

	// Remove hook entries from .claude/settings.json
	settingsPath := filepath.Join(projectDir, ".claude", "settings.json")
	if err := removeHookSettings(settingsPath); err != nil {
		return fmt.Errorf("removing hook settings: %w", err)
	}

	// Remove CLAUDE.md section
	claudeMD := filepath.Join(projectDir, ".claude", "CLAUDE.md")
	return removeCLAUDEmdSection(claudeMD)
}
```

**Step 7: Run failing tests to verify pass**

```bash
go test ./internal/plugin/ -run "TestInstallLocalMergesHookSettings|TestInstallLocalHookSettingsIdempotent|TestInstallLocalPreservesExistingSettings|TestUninstallLocalRemovesHookSettings" -v
```

Expected: PASS

**Step 8: Run full test suite**

```bash
go test ./internal/plugin/ -v
```

Expected: all PASS

**Step 9: Commit**

```bash
git add internal/plugin/plugin.go internal/plugin/plugin_test.go
git commit -m "feat: merge hook entries into .claude/settings.json on install/uninstall"
```

---

## Task 6: Enhance `claudeMDSection`

**Files:**
- Modify: `internal/plugin/plugin.go` (`claudeMDSection` constant)
- Modify: `internal/plugin/plugin_test.go`

**Step 1: Write failing test**

Add to `plugin_test.go`:
```go
func TestInstallLocalCLAUDEmdHasTeachingContent(t *testing.T) {
	dir := t.TempDir()
	if err := Install(dir, ScopeLocal); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
	content := string(data)

	checks := []string{
		"atu-guided-debugging",
		"atu-problem-decomposition",
		"atu-code-review-learning",
		"atu-dev-workflow",
		"Ask questions before giving answers",
		"One teaching point per interaction",
		"additionalContext",
		"/atu:debug",
		"/atu:review",
		"/atu:decompose",
		"/atu:workflow",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("CLAUDE.md missing expected content: %q", want)
		}
	}
}
```

**Step 2: Run to confirm fail**

```bash
go test ./internal/plugin/ -run TestInstallLocalCLAUDEmdHasTeachingContent -v
```

Expected: FAIL

**Step 3: Update `claudeMDSection` constant in `plugin.go`**

Replace the `claudeMDSection` constant (from `const claudeMDSection =` to closing backtick) with:

```go
const claudeMDSection = `<!-- BEGIN AGENT-TUTOR -->
# Agent Tutor

You are a programming tutor. A student is working in a terminal pane next to you.
You have MCP tools to observe their work — use them to provide relevant coaching.

## MCP Tools Reference

| Tool | Purpose | When to use |
|------|---------|-------------|
| ` + "`get_student_context`" + ` | 5-minute activity summary (markdown) | Quick overview of what the student is doing |
| ` + "`get_recent_file_changes`" + ` | File changes with diffs | When reviewing code the student wrote |
| ` + "`get_terminal_activity`" + ` | Recent terminal output | When the student hits errors or runs commands |
| ` + "`get_git_activity`" + ` | Commits and status changes | When the student commits or has uncommitted work |
| ` + "`get_coaching_config`" + ` | Current intensity and level | Check before deciding how proactive to be |
| ` + "`set_coaching_intensity`" + ` | Change coaching mode | When the student asks to adjust coaching |

## Commands Available

| Command | Purpose |
|---------|---------|
| ` + "`/atu:check`" + ` | Comprehensive review of recent activity |
| ` + "`/atu:hint`" + ` | Quick one-point nudge |
| ` + "`/atu:explain`" + ` | Explain the most recent error |
| ` + "`/atu:save`" + ` | Save current session as a lesson |
| ` + "`/atu:debug`" + ` | Guided debugging session (4-phase methodology) |
| ` + "`/atu:review`" + ` | Self-review coaching (graduated checklist) |
| ` + "`/atu:decompose`" + ` | Problem decomposition coaching |
| ` + "`/atu:workflow`" + ` | Development workflow habit coaching |

## Teaching Skills

When these commands are invoked, load the methodology by reading the corresponding skill file:

- ` + "`/atu:debug`" + ` → read ` + "`.agent-tutor/plugin/skills/atu-guided-debugging/SKILL.md`" + `
- ` + "`/atu:decompose`" + ` → read ` + "`.agent-tutor/plugin/skills/atu-problem-decomposition/SKILL.md`" + `
- ` + "`/atu:review`" + ` → read ` + "`.agent-tutor/plugin/skills/atu-code-review-learning/SKILL.md`" + `
- ` + "`/atu:workflow`" + ` → read ` + "`.agent-tutor/plugin/skills/atu-dev-workflow/SKILL.md`" + `

For deeper reference material, read the ` + "`references/`" + ` subdirectory of each skill.

## Coaching Behavior

- **proactive**: After messages, check ` + "`get_student_context`" + ` for teachable moments. On ` + "`tutor_nudge`" + `, offer coaching.
- **on-demand**: Only use tutor tools when the student asks or uses ` + "`/atu:check`" + `.
- **silent**: Never coach unless explicitly asked.

## Pedagogical Principles

- **Ask questions before giving answers.** "What do you think this error means?" before explaining.
- **One teaching point per interaction.** Never overwhelm with five things at once.
- **Praise specific good behavior first.** Acknowledge what worked before suggesting improvements.
- **Match depth to student level.** Vocabulary and checklist depth from ` + "`get_coaching_config`" + `.
- **Never fix code silently in proactive mode.** Always explain what and why.
- **If the student is doing well, say nothing.** Silence is valid coaching.

## Hook Awareness

The project has advisory hooks that inject ` + "`additionalContext`" + ` when:
- A file exceeds 200 lines after a Write/Edit (suggests ` + "`/atu:decompose`" + `)
- An error pattern appears in terminal output after a Bash command (suggests ` + "`/atu:debug`" + ` or ` + "`/atu:explain`" + `)

When ` + "`additionalContext`" + ` mentions a teachable moment, incorporate it naturally into your next response.
Do not parrot the hook text verbatim — use it as a trigger for genuine teaching.

## Lesson Auto-Save

After giving coaching feedback in these situations, also save a lesson file to ` + "`./lessons/`" + `:
- After responding to ` + "`/atu:check`" + ` — save the coaching feedback as a lesson
- After a ` + "`tutor_nudge`" + ` triggered by a git commit — save what was learned in that commit
- Whenever you explain a non-trivial concept and it would be valuable for review

Write each lesson to ` + "`./lessons/YYYY-MM-DD-<topic-slug>.md`" + ` using this template:
Create the ` + "`./lessons/`" + ` directory if it does not exist.

    # <Topic Title>

    **Date:** YYYY-MM-DD
    **Topic:** <category>
    **Trigger:** <check|commit|nudge|manual>

    ## What I Learned
    <Clear explanation tailored to student level>

    ## Code Example
    <Relevant code with annotations>

    ## Key Takeaway
    <One sentence to remember>

    ## Common Mistakes
    <Pitfalls to avoid>

Do not duplicate — if a lesson file for the same topic already exists today, skip it.
<!-- END AGENT-TUTOR -->`
```

**Step 4: Run failing test to verify pass**

```bash
go test ./internal/plugin/ -run TestInstallLocalCLAUDEmdHasTeachingContent -v
```

Expected: PASS

**Step 5: Run full test suite**

```bash
go test ./internal/plugin/ -v
```

Expected: all PASS

**Step 6: Commit**

```bash
git add internal/plugin/plugin.go internal/plugin/plugin_test.go
git commit -m "feat: enhance CLAUDE.md injection with teaching skills, pedagogy, and hook awareness"
```

---

## Task 7: Update `uninstallGlobal` for New Skill Directories

**Files:**
- Modify: `internal/plugin/plugin.go:198`
- Modify: `internal/plugin/plugin_test.go` (`TestUninstallGlobal`)

**Step 1: Write failing test**

Update `TestUninstallGlobal` — add new skill names to the checked list:
```go
for _, name := range []string{
    "atu-check", "atu-hint", "atu-explain", "atu-save",
    "atu-guided-debugging", "atu-problem-decomposition",
    "atu-code-review-learning", "atu-dev-workflow",
} {
```

Also add them to `TestInstallGlobal` to verify global install creates them:
```go
skills := []string{
    ".claude/skills/atu-check/SKILL.md",
    ".claude/skills/atu-hint/SKILL.md",
    ".claude/skills/atu-explain/SKILL.md",
    ".claude/skills/atu-save/SKILL.md",
    ".claude/skills/atu-guided-debugging/SKILL.md",
    ".claude/skills/atu-problem-decomposition/SKILL.md",
    ".claude/skills/atu-code-review-learning/SKILL.md",
    ".claude/skills/atu-dev-workflow/SKILL.md",
}
```

**Step 2: Run to confirm fail**

```bash
go test ./internal/plugin/ -run "TestInstallGlobal|TestUninstallGlobal" -v
```

Expected: FAIL for new skill directories

**Step 3: Update `installGlobal` to also extract skill directories**

In `installGlobal`, after the commands loop, add a skills extraction block:

```go
// Install teaching skills as global skills
skills, err := fs.ReadDir(pluginFS, "embed/skills")
if err == nil {
    for _, skillEntry := range skills {
        if !skillEntry.IsDir() {
            continue
        }
        skillName := skillEntry.Name()
        skillDestDir := filepath.Join(home, ".claude", "skills", skillName)
        skillSrcDir := "embed/skills/" + skillName
        if err := fs.WalkDir(pluginFS, skillSrcDir, func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return err
            }
            rel, _ := filepath.Rel(skillSrcDir, path)
            if rel == "." {
                return os.MkdirAll(skillDestDir, 0o755)
            }
            dest := filepath.Join(skillDestDir, rel)
            if d.IsDir() {
                return os.MkdirAll(dest, 0o755)
            }
            data, err := pluginFS.ReadFile(path)
            if err != nil {
                return err
            }
            return os.WriteFile(dest, data, 0o644)
        }); err != nil {
            return fmt.Errorf("installing skill %s: %w", skillName, err)
        }
    }
}
```

**Step 4: Update `uninstallGlobal` skill list**

Replace the hardcoded list:
```go
for _, name := range []string{
    "atu-check", "atu-hint", "atu-explain", "atu-save",
    "atu-guided-debugging", "atu-problem-decomposition",
    "atu-code-review-learning", "atu-dev-workflow",
} {
```

**Step 5: Run tests to verify pass**

```bash
go test ./internal/plugin/ -run "TestInstallGlobal|TestUninstallGlobal" -v
```

Expected: PASS

**Step 6: Run full test suite**

```bash
go test ./internal/plugin/ -v
```

Expected: all PASS

**Step 7: Build and smoke test**

```bash
go build ./... && echo "Build OK"
```

Expected: `Build OK`

**Step 8: Commit**

```bash
git add internal/plugin/plugin.go internal/plugin/plugin_test.go
git commit -m "feat: update global install/uninstall for 4 new teaching skill directories"
```

---

## Final Verification

**Step 1: Run all tests**

```bash
go test ./... -v 2>&1 | tail -30
```

Expected: all PASS, no FAIL lines

**Step 2: Build binary**

```bash
go build -o /tmp/agent-tutor-test ./cmd/agent-tutor/
```

Expected: no errors

**Step 3: Smoke test install**

```bash
cd /tmp && mkdir test-project && /tmp/agent-tutor-test install-plugin --project-dir /tmp/test-project && \
  ls /tmp/test-project/.agent-tutor/plugin/commands/ && \
  ls /tmp/test-project/.agent-tutor/plugin/skills/ && \
  ls /tmp/test-project/.agent-tutor/plugin/hooks/ && \
  cat /tmp/test-project/.claude/settings.json && \
  rm -rf /tmp/test-project /tmp/agent-tutor-test
```

Expected: 8 command files, 4 skill dirs, 2 hook files, settings.json with hook entries

**Step 4: Final commit if any cleanup needed, then done**
