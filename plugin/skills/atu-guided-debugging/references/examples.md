# Debugging Examples — Worked Walkthroughs

## Example 1: Off-By-One Index Error

**Scenario:** Student has `items[len(items)]` and gets an index out of range panic.

**Phase 1 coaching:** "The panic says index 5 is out of range on a slice of length 5. What's the valid index range for a 5-element slice?"
→ Student realizes indices go 0-4, not 0-5.

**Phase 2 coaching:** "Where else in the code do you access items by index? What pattern do they use?"
→ Student finds `items[i]` in a loop and sees `i < len(items)` not `i <= len(items)`.

**Phase 3 coaching:** "So your hypothesis is that `len(items)` gives the length but not the last valid index. How would you verify that?"
→ Student prints `len(items)` and the index value to confirm.

**Phase 4 coaching:** "Before you fix it, can you write a test with a known slice that triggers this panic?"

---

## Example 2: Nil Pointer / Null Reference

**Scenario:** Student gets a nil pointer dereference. The error points to line 42: `user.Name`.

**Phase 1 coaching:** "The error says you're dereferencing a nil pointer at `user.Name`. What does that mean about `user`?"
→ Student realizes `user` is nil.

**Phase 2 coaching:** "Trace backwards — where is `user` set? Is there a path through the code where it might not get assigned?"
→ Student finds a code path where `getUserByID` returns nil on error but the error isn't checked.

**Phase 3 coaching:** "So your hypothesis is that `getUserByID` returned nil and you used it without checking. How would you add a guard?"

**Phase 4 coaching:** "What should happen when the user isn't found? Error? Default value? Make that explicit before fixing."

---

## Example 3: Async Race Condition

**Scenario (JavaScript):** Student's function sometimes returns undefined. Works in isolation, breaks in production.

**Phase 1 coaching:** "It works sometimes and fails sometimes — what does that tell you about the cause?"
→ Student starts thinking about timing.

**Phase 2 coaching:** "Is this function async? Are there any awaits missing?"
→ Student finds `const data = fetch(url)` without `await`.

**Phase 3 coaching:** "So your hypothesis is that `fetch` returns a Promise and you're treating it as the resolved value. What would you see if you `console.log(data)` right now?"
→ Student logs it and sees `Promise { <pending> }`.

**Phase 4 coaching:** "Now write a test that asserts the return value is a string, not a Promise."
