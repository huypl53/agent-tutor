---
name: atu:explain
description: Explain the most recent error or terminal output in detail
---

Explain what just happened in the student's terminal.

1. Call `get_terminal_activity` to see recent terminal output
2. Call `get_coaching_config` to check the student's level

Find the most recent error or notable output and explain it:
- What the error means in plain language
- Why it happened (the root cause, not just the symptom)
- How to fix it, step by step
- For beginners: explain the underlying concept
- For experienced devs: focus on the specific fix and any non-obvious gotchas

If there are no errors in the recent output, explain what the last command did and whether the output looks correct.
