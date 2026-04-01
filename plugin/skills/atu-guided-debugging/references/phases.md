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
