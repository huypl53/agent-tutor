---
name: atu:deep-dive
description: Exhaustive analysis of a specific module or feature in the project
---

Deep-dive into a specific area of the project for focused learning.

## Step 1: Check Prerequisites

1. Call `get_project_profile` — if no profile exists, call `scan_project` first
2. Present the project's directory structure to the student

## Step 2: Pick Target

**If ARGUMENTS specify a path** (e.g., `/atu:deep-dive src/auth`):
- Use that path as the deep-dive target

**If no ARGUMENTS:**
- Show the top-level directories with brief descriptions
- Ask: "Which area would you like to understand deeply?"

## Step 3: Spawn Deep-Dive Sub-Agent

Spawn a single sub-agent with this prompt (fill in `{target}`, `{projectType}`):

```
You are doing an exhaustive analysis of the `{target}` directory in a {projectType} project to help a student understand it deeply.

Read EVERY file in `{target}` (and subdirectories). For each file, document:
- **Purpose:** What this file does (1-2 sentences)
- **Exports:** What it exposes to other files
- **Imports:** What it depends on
- **Key patterns:** Notable code patterns or techniques used

Then synthesize:
- **Dependency graph:** How files in this area relate to each other (which imports which)
- **Data flow:** How data moves through this module (input → processing → output)
- **Integration points:** How this module connects to the rest of the codebase
- **Key abstractions:** The main concepts/classes/functions a student must understand
- **Contributor guidance:** What someone modifying this code needs to know

Write everything as a clear markdown document. Explain patterns and decisions, not just list facts.

When done, call the `save_project_doc` MCP tool with name="deep-dive-{target-slug}" and your markdown content.
```

## Step 4: Present Findings

After the sub-agent completes:

1. Read the deep-dive doc from `.agent-tutor/docs/`
2. Walk through the findings pedagogically:
   - Start with the big picture: "This module does X"
   - Ask comprehension questions: "Looking at the dependency graph, why do you think Y depends on Z?"
   - Highlight non-obvious patterns: "Notice how they use X here — that's because..."
3. Update the project profile's `deepDives` array
4. Offer to create a topic for concepts discovered: "Want me to track your learning on [pattern/concept]?"

ARGUMENTS: $ARGUMENTS
