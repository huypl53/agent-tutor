---
name: atu:plan
description: Create a structured learning plan or show progress on the current plan
---

Manage the student's learning plan using MCP state tools.

**If ARGUMENTS contain a learning goal** (e.g., "Build a REST API in Go"):

1. Call `get_coaching_config` to check the student's level
2. Call `get_student_context` to understand what they've been working on
3. Design 4-8 learning steps, ordered by dependency
4. Call `create_topic` for each step (with id, title, complexity, dependencies)
5. Call `create_plan` with the goal and steps referencing the topic IDs
6. Call `update_topic` on the first topic with `status: "practicing"`
7. Call `save_session` with the first topic as `activeTopicId`
8. Announce the plan and ask if the student wants to adjust it before starting

**If ARGUMENTS are "next":**

1. Call `get_plan` to see current progress
2. Call `update_topic` on the current step's topic with `status: "mastered"`
3. Call `update_plan` to mark the current step as `mastered`
4. Save a lesson file for the completed step, then call `update_topic` with `lessonFile`
5. Call `update_topic` on the next step's topic with `status: "practicing"`
6. Call `save_session` with the new active topic
7. Announce what was completed and what's next

**If no ARGUMENTS (or plan does not exist):**

1. Call `get_plan` to check if a plan exists
2. If it exists, call `get_learning_summary` and display progress with topic statuses
3. If it does not exist, tell the student they can create one with `/atu:plan <goal>`

ARGUMENTS: $ARGUMENTS
