---
name: atu-dev-workflow
description: Development workflow habit coaching — commit often, test first, branch well
---

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
