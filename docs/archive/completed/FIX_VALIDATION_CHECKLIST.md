# Fix Validation Checklist

Use this checklist before declaring any bug fix "complete".

## Level 1: Unit Testing ✅

- [ ] Unit tests pass for the modified component
- [ ] Test covers the specific bug being fixed
- [ ] Test covers edge cases (e.g., different bit depths, formats)
- [ ] No panics or crashes in test output

## Level 2: Integration Testing ✅

- [ ] Full pipeline test created (if touching multiple components)
- [ ] Test uses realistic data (not just synthetic test cases)
- [ ] Integration test passes
- [ ] All existing integration tests still pass

## Level 3: End-to-End Testing ✅

- [ ] **Run the actual CLI command users would run**
- [ ] Test with real-world files (not just test fixtures)
- [ ] Check for errors in console output
- [ ] Verify expected output (files created, database updated, etc.)

Example:
```bash
# Don't just test the function in isolation
# Test the actual user workflow
./bin/olsen index path/to/test/files --db test.db

# Check for error messages like:
# "FAILED FILES"
# "failed to decode"
# "unknown format"
```

## Level 4: Impact Analysis ✅

- [ ] **Grep for all consumers of changed types/functions**
  ```bash
  # If you changed a return type from RGB to Gray:
  grep -r "image.Image" internal/
  grep -r "jpeg.Encode" internal/

  # Check each location for compatibility
  ```

- [ ] **Draw the data flow diagram**
  ```
  Input → Component A → Modified Type → Component B → Output
                            ↓
                    Does B handle the new type?
  ```

- [ ] List all components that consume the modified output
- [ ] Verify each component handles the new behavior

## Level 5: Downstream Effects ✅

For type changes (e.g., now returning `image.Gray` instead of `image.RGBA`):

- [ ] Image encoding (JPEG, PNG, etc.) - does it support the new type?
- [ ] Image resizing libraries - do they handle the new type?
- [ ] Color extraction - does it work with grayscale?
- [ ] Database storage - any assumptions about color channels?
- [ ] UI display - can it render the new type?

## Level 6: Error Message Review ✅

- [ ] Check recent logs/output for new error patterns
- [ ] Search codebase for error messages that might be triggered
- [ ] Add logging to track the fix in production

## Level 7: Documentation ✅

- [ ] Update function comments with new behavior
- [ ] Document breaking changes (if any)
- [ ] Update user-facing docs
- [ ] Add example usage

## Common Pitfalls to Check

### For Image Processing Changes
- [ ] Does `jpeg.Encode` support the image type?
- [ ] Does `png.Encode` support the image type?
- [ ] Do resizing libraries handle the type?
- [ ] Are color channels assumed to be 3 or 4?

### For RAW Processing Changes
- [ ] Does thumbnail generation work?
- [ ] Does color extraction work?
- [ ] Are bit depths handled correctly (8-bit vs 16-bit)?
- [ ] Are monochrome images supported everywhere?

### For Buffer/Memory Changes
- [ ] Are all buffer size calculations using the correct variables?
- [ ] Are loop bounds checked against actual buffer size?
- [ ] Is memory allocated based on actual data, not assumptions?

## Final Sign-Off

Before marking as "complete":

- [ ] All above sections reviewed
- [ ] At least one end-to-end test run successfully
- [ ] No new error messages in output
- [ ] Commit message explains what was fixed AND tested
- [ ] Integration test added (if applicable)

## Example of Good vs Bad

### ❌ Bad Process
1. Fix buffer overflow in go-libraw
2. Run unit test showing `ProcessRaw()` succeeds
3. Declare victory ✅
4. Miss downstream JPEG encoding failure ❌

### ✅ Good Process
1. Fix buffer overflow in go-libraw
2. Run unit test showing `ProcessRaw()` succeeds
3. **Run full CLI command: `./bin/olsen index test/files`**
4. **Observe "unknown format" error**
5. Fix JPEG encoding for Gray images
6. **Run CLI command again - success**
7. Add integration test
8. Declare victory ✅

## Template for Commit Message

```
Fix [bug description]

Root cause: [what was wrong]
Impact: [what failed]
Fix: [what changed]

Testing:
- Unit test: [test name] - passes
- Integration test: [test name] - passes
- End-to-end: [CLI command] - succeeds with [N] files

Verified:
- [Downstream component A] works
- [Downstream component B] works
- No new errors in output
```

---

**Remember**: "Tests pass" ≠ "Feature works"

The real test is: **Can a user successfully do what they're trying to do?**
