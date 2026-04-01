# Code Review Checklist — Annotated Reference

## Correctness

**Does the code do what the name/comment says?**
Teaching note: Function names are contracts. `getUserByEmail` that sometimes returns by ID is a bug waiting to happen. Ask student: "Read just the function name — what do you expect it to do? Now read the body. Does it match?"

**What happens on empty/null/zero input?**
Teaching note: This is the most common source of bugs. Ask: "What's the weirdest valid input this could receive? What about invalid input?" Off-by-one errors, nil dereferences, and division by zero live here.

**Are all code paths handled?**
Teaching note: `if/else` chains without a final `else`, switch statements without `default`, promises without `.catch()`. Ask: "Is there any path through this code that doesn't return a value or handle an error?"

## Readability

**Would you understand this in 3 months?**
Teaching note: This is a proxy for "is this well-written?" Ask student: "Without looking at the context, could you explain what this does in one sentence?" If they can't, neither can a future reader.

**Are variable names specific enough?**
Teaching note: `data`, `result`, `temp`, `x` are red flags. The name should encode the type and purpose. `userData` vs `user` — which tells you more?

**Is there duplication that should be extracted?**
Teaching note: The DRY principle isn't about zero duplication — it's about not having two places that need to change in sync. Ask: "If you needed to change how this works, how many places would you need to update?"

## Security (Advanced)

**Is user input validated before use?**
Teaching note: Never trust input from users, APIs, files, or databases. Ask: "Where does this value come from? Could it be malicious?"

**Are there SQL injection, XSS, or path traversal risks?**
Teaching note: String concatenation into queries, HTML templates without escaping, file paths from user input without canonicalization. Show one example and ask them to find others.

## Performance (Advanced)

**Are there N+1 query patterns?**
Teaching note: A loop that makes a database call in each iteration is the most common performance bug. Ask: "How many database queries does this code make for N records?"

**Is there unnecessary work inside loops?**
Teaching note: Function calls that are constant (like `len(slice)` in Go or `array.length` in JS if not changing) moved inside loops.

## Testability (Advanced)

**Is the function testable in isolation?**
Teaching note: Functions that reach out to global state, databases, or the network without dependency injection are hard to test. Ask: "How would you write a unit test for this function?"

**Is there a test for the happy path and at least one edge case?**
Teaching note: A test that only tests the happy path gives false confidence. Ask: "What test would have caught this bug?"
