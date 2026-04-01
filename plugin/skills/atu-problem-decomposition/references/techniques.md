# Decomposition Techniques — Detailed Reference

## Working Backwards

**When:** Student has a clear end goal but no idea where to start.
**Depth:** Go at least 4-5 steps back before stopping.
**Watch for:** Students who jump to implementation steps — redirect to outcomes.
**Example:** "Build a REST API" → Last step: "API is deployed and responding" → Before that: "Tests pass" → Before that: "Routes return correct data" → Before that: "Data layer returns correct data" → Before that: "Data model is defined" → Start here.

## Smallest Increment

**When:** Student is overwhelmed by scope or trying to build everything at once.
**Key insight:** Value is delivered when something works, not when it's complete.
**Watch for:** Students who say "but it won't be complete" — completeness is not the goal.
**Good follow-up:** "Who is the user of this smallest version? What problem does it solve for them?"

## Dependency Mapping

**When:** Student has a list of things to build but doesn't know the order.
**How to facilitate:** Ask them to literally list all pieces on paper/in a comment. Then ask about each pair: "Does A depend on B, or B on A?"
**Goal:** A directed graph; the things with no incoming edges are where to start.
**Common mistake students make:** Over-engineering dependencies; often things are more independent than students assume.

## Inversion

**When:** Student is stuck because they can't think positively about the requirements.
**Power:** Flipping to "how would this fail?" is often much easier than "how should this work?"
**Process:** List 5 ways the system could fail → Each failure is a requirement → Prioritize by likelihood and impact.
**Example:** "Users will lose data if the database is down" → Requirement: "Handle database unavailability gracefully."

## Scale Game

**When:** Student is making architecture decisions without understanding constraints.
**Key:** Most systems only need to handle 1-10x their current load; premature scaling is waste.
**Good question:** "At what scale does your current approach break? Is that scale realistic?"
**Outcome:** Student understands the difference between necessary complexity and speculative complexity.
