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
