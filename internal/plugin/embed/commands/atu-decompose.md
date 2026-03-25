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
