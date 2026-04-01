---
name: atu:save
description: Save a lesson from the current session to ./lessons/ for later review
---

Save a structured lesson based on the current coaching context.

1. Call `get_student_context` to see what the student has been working on
2. Call `get_coaching_config` to check the student's level
3. Call `restore_session` to know the active topic

Write a markdown file to `./lessons/YYYY-MM-DD-<topic-slug>.md` where:
- Create the `./lessons/` directory if it does not exist.
- YYYY-MM-DD is today's date
- topic-slug is a short kebab-case summary of the topic (e.g. "understanding-goroutines")

Use this exact template:

    # <Topic Title>

    **Date:** YYYY-MM-DD
    **Topic:** <category>
    **Trigger:** manual

    ## What I Learned
    <Clear explanation of the concept, tailored to student level>

    ## Code Example
    <Relevant code from the session, with annotations>

    ## Key Takeaway
    <One sentence the student should remember>

    ## Common Mistakes
    <Pitfalls to avoid>

If ARGUMENTS are provided, use them as the topic. Otherwise, infer the most significant topic from recent activity.

Do not overwrite an existing lesson file — if the same topic-slug exists today, append a numeric suffix (e.g. `-2`).

**After saving the lesson file, ALWAYS:**
- Call `update_topic` on the active topic with `lessonFile` pointing to the saved file path

ARGUMENTS: $ARGUMENTS
