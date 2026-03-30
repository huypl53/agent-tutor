---
name: atu:plan
description: Create a structured learning plan or show progress on the current plan
---

Manage the student's learning plan.

**If ARGUMENTS contain a learning goal** (e.g., "Build a REST API in Go"):

1. Call `get_coaching_config` to check the student's level
2. Call `get_student_context` to understand what they've been working on
3. Propose 4-8 learning steps, ordered by dependency, each with:
   - A clear learning objective (bold title)
   - A brief description of what the student will do
   - Appropriate difficulty for their level
4. Write the plan to `.agent-tutor/learning-plan.md` using this format:

        # Learning Plan: <Goal>

        **Goal:** <one-sentence description>
        **Created:** YYYY-MM-DD
        **Progress:** 0/N complete

        ## Steps

        - [ ] 1. **<Step title>** — <what the student will learn/do>
        - [ ] 2. **<Step title>** — <what the student will learn/do>
        ...

        ## Notes
        - <any relevant observations about student level or context>

5. Set step 1 as the current topic in `.agent-tutor/current-topic.md`
6. Announce the plan and ask if the student wants to adjust it before starting

**If ARGUMENTS are "next":**

1. Read `.agent-tutor/learning-plan.md`
2. Mark the current step as done (change `- [ ]` to `- [x]`)
3. Update the progress count
4. Save a lesson for the completed step
5. Set the next uncompleted step as the current topic in `.agent-tutor/current-topic.md`
6. Announce what was completed and what's next

**If no ARGUMENTS (or plan does not exist):**

1. Read `.agent-tutor/learning-plan.md`
2. If it exists, display the plan with current progress and which step is active
3. If it does not exist, tell the student they can create one with `/atu:plan <goal>`

ARGUMENTS: $ARGUMENTS
