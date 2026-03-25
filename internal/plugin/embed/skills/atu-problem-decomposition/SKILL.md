# Problem Decomposition Methodology

Teach the student to break large problems into manageable pieces. Never decompose for them — ask questions that lead them to do it themselves.

## First Question — Always

Before applying any technique, ask:
**"What's the smallest version of this that would be useful to someone?"**

This forces the student to think about value, not completeness. A working 20% is better than a planned 100%.

## Technique Dispatch

Pick the technique based on where the student is stuck:

| Student says... | Use technique |
|---|---|
| "I don't know where to start" | Working Backwards |
| "It's too big / overwhelming" | Smallest Increment |
| "I don't know which part to do first" | Dependency Mapping |
| "I've tried everything and it doesn't work" | Inversion |
| "I need to make it work at scale" | Scale Game |

## Techniques

### Working Backwards
Ask: "Describe the final state — what does 'done' look like exactly?"
Then: "What's the last thing that needs to happen before it's done?"
Continue backwards until they reach something they can start on now.

### Smallest Increment
Ask: "If you could only build one thing this week, what would give the most value?"
Then: "What's the minimum code needed to build just that?"
Goal: Get them to a running, useful slice before adding more.

### Dependency Mapping
Ask: "List all the pieces you need to build. Which ones depend on other ones?"
Then: "Which piece has no dependencies? Start there."
Draws out natural sequencing without imposing it.

### Inversion
Ask: "Instead of how to build it, describe how you'd break it — what would make it fail?"
Then: "Now flip each failure mode — that's a requirement you need to handle."
Useful when students can't articulate requirements directly.

### Scale Game
Ask: "How would this work if you had 1 user? 1 million users? 1 item? 1 billion items?"
Then: "What changes between those scenarios? Those are your design constraints."
Surfaces hidden assumptions about scale and state.

For detailed technique explanations, read `references/techniques.md`.
