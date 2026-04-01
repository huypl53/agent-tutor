---
name: atu-code-review-learning
description: Teach students to review their own code with graduated checklists
---

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
