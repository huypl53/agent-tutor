---
name: atu:onboard
description: Analyze the current project — detect stack, architecture, patterns — and create a learning path
---

Analyze the student's project to help them understand the codebase. Uses a fast MCP scan followed by sub-agent deep analysis.

## Step 1: Quick Scan

1. Call `scan_project` to get the project skeleton (type, stack, structure, entry points)
2. Call `get_project_profile` to check if analysis docs already exist

**If docs already exist**, ask the student:
- "I've already analyzed this project. Would you like to: (a) rescan, (b) deep-dive into a specific area with /atu:deep-dive, or (c) see the summary?"
- If (a), proceed. If (b), suggest `/atu:deep-dive`. If (c), read and present `.agent-tutor/docs/index.md`.

## Step 2: Present Initial Findings

Show the student what you found:
- Project type and tech stack
- Directory structure overview
- Entry points
- Which analysis domains will be covered

Ask: "I'll now do a deep analysis of [N] areas: [domains]. This will take a minute. Ready?"

## Step 3: Spawn Analysis Sub-Agents

For each domain in the project's `analysisDomains`, spawn a sub-agent using the Agent tool. Run them **in parallel** where possible.

Each sub-agent gets this prompt template (fill in `{domain}`, `{projectType}`, `{skeleton}`):

```
You are analyzing a {projectType} project to help a student understand it.

**Your domain:** {domain}
**Project skeleton:** {skeleton}

Read the relevant source files in the project, then write a clear markdown analysis covering:

**For architecture:** Overall pattern (MVC, microservices, etc.), data flow diagram, key abstractions, design decisions
**For api:** All routes/endpoints, request/response shapes, middleware chain, auth flow
**For data:** Database schema, models, migrations, ORM patterns, validation rules
**For components:** UI component tree, state management, routing, styling approach
**For infra:** CI/CD pipeline, Docker setup, deployment config, environment management
**For testing:** Test patterns, frameworks, coverage approach, fixtures
**For dev-setup:** Prerequisites, install steps, build/run/test commands, environment setup

Write your findings as a complete markdown document. Focus on what a student needs to understand, not exhaustive documentation. Explain WHY things are structured this way, not just WHAT they are.

When done, call the `save_project_doc` MCP tool with name="{domain}" and your markdown content.
```

## Step 4: Collect Results

After all sub-agents complete:

1. Read each generated doc from `.agent-tutor/docs/`
2. Write an `index.md` via `save_project_doc` that links all docs with one-line summaries
3. Present a project overview to the student:
   - What the project does
   - Key architectural decisions
   - How the parts fit together
   - Where to start reading code

## Step 5: Gap Analysis and Learning Plan

1. Call `list_topics` to see what the student already knows
2. Compare against what the project uses (from the analysis)
3. Identify gaps — technologies/patterns the project uses that the student hasn't learned
4. Offer to create a learning plan: "Based on this project, you'd benefit from learning: [gaps]. Want me to create a plan with `/atu:plan`?"

ARGUMENTS: $ARGUMENTS
