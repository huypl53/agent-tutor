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
