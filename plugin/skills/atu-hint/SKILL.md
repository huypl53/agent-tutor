---
name: atu:hint
description: Quick nudge — one teaching point based on what you're currently doing
---

Give the student a brief, focused hint about their current work.

1. Call `get_student_context` to get a quick overview of recent activity
2. Call `get_coaching_config` to check the student's level
3. Call `restore_session` to know the active topic

Based on the context, give exactly ONE teaching point:
- Keep it to 2-3 sentences
- Focus on the most impactful thing they could improve right now
- If they're doing well, say so briefly and don't force a teaching moment
- Frame it as a suggestion, not a correction

**After giving the hint, ALWAYS:**
- Call `update_topic` on the active topic with `moment: { type: "hint", detail: "<your hint>" }`
