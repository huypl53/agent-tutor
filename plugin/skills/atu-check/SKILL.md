---
name: atu:check
description: Comprehensive review of your recent coding activity with coaching feedback
---

Review the student's recent work by gathering all available context, then provide coaching feedback.

1. Call `get_student_context` for an overview of recent activity
2. Call `get_recent_file_changes` to see what code was written or modified
3. Call `get_git_activity` to see commits and working tree status
4. Call `get_coaching_config` to check the student's level

Based on all gathered context, provide coaching feedback:
- Point out what the student did well
- Identify one or two areas for improvement (don't overwhelm)
- If there are errors, explain why they happened and how to fix them
- If the code works but could be improved, explain the idiomatic approach
- Tailor your language to the student's level (beginner vs experienced)
