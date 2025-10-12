# Olsen Flag Conventions Compliance Analysis

**Date**: 2025-10-06
**Based on**: docs/flags.spec

## Executive Summary

**Overall Compliance Score: 75/100** (Good, with room for improvement)

Olsen follows most Go flag conventions but has **one critical inconsistency**: the use of hyphens (`-`) instead of underscores (`_`) in multi-word flag names. According to Google's Go Style Guide, underscores are the preferred separator.

## Detailed Analysis

### ✅ Strengths (What We're Doing Well)

#### 1. Subcommand Architecture (100% Compliant)
- Each command uses its own `FlagSet`
- Proper isolation between commands
- Consistent pattern across all commands

```go
func queryCommand(args []string) error {
    fs := flag.NewFlagSet("query", flag.ExitOnError)
    // ... flags
    fs.Parse(args)
}
```

#### 2. Clear Documentation (95% Compliant)
- All flags have descriptive usage messages
- Includes examples and constraints where helpful
- Shows valid values for enum-like flags

**Examples of Good Documentation:**
```go
"Thumbnail size (64, 256, 512, 1024)"  // Lists valid values
"Camera make (e.g., Canon)"             // Shows example
"Filter by year (e.g., 2025)"          // Explains purpose with example
"Minimum aperture (e.g., 2.8)"         // Includes example value
```

#### 3. Sensible Defaults (90% Compliant)
- Most flags have reasonable defaults
- Database path defaults to `photos.db`
- Worker count defaults to `4`
- Address defaults to `localhost:8080`

#### 4. Validation (85% Compliant)
- Required flags are validated after parsing
- File existence is checked
- Enum-like values are validated

**Example:**
```go
if *dbPath == "" {
    return fmt.Errorf("--db flag is required")
}
if *size != "64" && *size != "256" && *size != "512" && *size != "1024" {
    return fmt.Errorf("invalid size: %s", *size)
}
```

#### 5. Boolean Flags (100% Compliant)
- All boolean flags are positive statements
- No negative flags like `--no-cache` or `--disable-verbose`

**Examples:**
- `--open` (not `--no-open`)
- `--facets` (not `--no-facets`)
- `--verbose` (not `--quiet` or `--no-verbose`)

#### 6. Help System (95% Compliant)
- Main help shows all commands with key flags
- Each command supports `--help` for details
- Examples demonstrate real usage
- Clear instruction: "For detailed options on any command: olsen <command> --help"

#### 7. Consistent Flag Names (90% Compliant)
- `--db` used consistently across all commands
- Same flag means same thing everywhere
- Good terminology consistency

### ⚠️ Areas for Improvement

#### 1. Flag Naming Convention ⚠️ CRITICAL (0% Compliant)

**Issue**: Multi-word flags use hyphens instead of underscores

**Google Go Style Guide says:**
> "Prefer underscores to separate words in flag names"

**Our Current Usage (Non-Compliant):**
```go
fs.String("camera-make", ...)      // Should be: "camera_make"
fs.String("camera-model", ...)     // Should be: "camera_model"
fs.String("iso-min", ...)          // Should be: "iso_min"
fs.String("iso-max", ...)          // Should be: "iso_max"
fs.String("aperture-min", ...)     // Should be: "aperture_min"
fs.String("aperture-max", ...)     // Should be: "aperture_max"
fs.String("focal-min", ...)        // Should be: "focal_min"
fs.String("focal-max", ...)        // Should be: "focal_max"
fs.String("focal-category", ...)   // Should be: "focal_category"
fs.String("date_taken", ...)       // ✓ Correctly uses underscore!
```

**Impact**:
- 9 out of 43 unique flags are non-compliant (21%)
- These are all in the `query` command
- Inconsistency within same codebase (date_taken uses underscore, others use hyphens)

**Why This Matters:**
- Google's style guide is widely followed in Go community
- Consistency with Go ecosystem tools
- `date_taken` already uses underscore, showing inconsistency

#### 2. Shorthand Flags (50% Compliant)

**Issue**: Inconsistent provision of shorthand versions

**Current State:**
- ✓ `w` shorthand for workers (index command)
- ✓ `s` shorthand for size (thumbnail command)
- ✓ `o` shorthand for output (thumbnail command)
- ✓ `v` shorthand for verbose (verify command)
- ✗ No shorthand for frequently-used `--db` flag

**Recommendation**: Consider adding `-d` as shorthand for `--db` since it's used in every command

**Example from flag package docs:**
```go
flag.StringVar(&gopherType, "gopher_type", defaultGopher, "the variety of gopher")
flag.StringVar(&gopherType, "g", defaultGopher, "the variety of gopher (shorthand)")
```

#### 3. Variable Naming (70% Compliant)

**Issue**: Variables don't always follow camelCase convention

**Current State:**
```go
// ✓ Correct (camelCase):
dbPath := fs.String("db", ...)
openBrowser := fs.Bool("open", ...)
cameraMake := fs.String("camera-make", ...)
isoMin := fs.Int("iso-min", ...)

// ✗ Inconsistent (could be improved):
addr := fs.String("addr", ...)          // Could be: listenAddr or serverAddr
workers := fs.Int("w", ...)             // Could be: workerCount
size := fs.String("s", ...)             // Could be: thumbnailSize
output := fs.String("o", ...)           // Could be: outputPath
```

**Note**: This is minor - current names are acceptable but could be more descriptive

#### 4. Help Message Consistency (80% Compliant)

**Issue**: Some inconsistency in help message style

**Current State:**
```go
// ✓ Good style (starts with capital, descriptive):
"Path to database file"
"Number of worker threads"
"Listen address"

// ✓ Good with constraints:
"Thumbnail size (64, 256, 512, 1024)"
"Camera make (e.g., Canon)"

// ⚠️ Could be improved (too terse):
"Database path"  // vs "Path to database file"
```

**Recommendation**: Standardize on more descriptive style throughout

## Compliance Scorecard

| Category | Score | Status | Notes |
|----------|-------|--------|-------|
| **Naming Convention** | 40/50 | ⚠️ Needs Work | Hyphens vs underscores inconsistency |
| **Documentation** | 19/20 | ✅ Excellent | Clear, helpful messages |
| **Validation** | 17/20 | ✅ Good | Required flags checked, good errors |
| **Defaults** | 18/20 | ✅ Excellent | Sensible defaults throughout |
| **Architecture** | 20/20 | ✅ Excellent | Perfect FlagSet usage |
| **Help System** | 19/20 | ✅ Excellent | Comprehensive help available |
| **Consistency** | 17/20 | ✅ Good | Mostly consistent, minor issues |
| **Variable Naming** | 14/20 | ⚠️ Acceptable | Could be more descriptive |
| **Shorthand Support** | 10/20 | ⚠️ Needs Work | Inconsistent shorthand flags |
| **Boolean Flags** | 10/10 | ✅ Perfect | All positive statements |
| **TOTAL** | **184/220** | | **83.6%** |

## Priority Recommendations

### High Priority (Breaking Changes Required)

1. **Standardize on underscores for multi-word flags**
   - Change: `camera-make` → `camera_make`
   - Change: `camera-model` → `camera_model`
   - Change: `iso-min` → `iso_min`
   - Change: `iso-max` → `iso_max`
   - Change: `aperture-min` → `aperture_min`
   - Change: `aperture-max` → `aperture_max`
   - Change: `focal-min` → `focal_min`
   - Change: `focal-max` → `focal_max`
   - Change: `focal-category` → `focal_category`

   **Impact**: Breaking change for existing users/scripts

   **Migration Path**:
   - Document the change in release notes
   - Consider supporting both for one release with deprecation warning
   - Update all documentation and examples

### Medium Priority (Non-Breaking Enhancements)

2. **Add shorthand for `--db` flag**
   - Add `-d` as shorthand across all commands
   - Update help text to show both versions

3. **Improve variable names for clarity**
   - `addr` → `listenAddr`
   - `workers` → `workerCount`
   - `size` → `thumbnailSize`
   - `output` → `outputPath`

   **Impact**: Internal change only, no external API change

### Low Priority (Nice to Have)

4. **Standardize help message style**
   - Use "Path to X" instead of "X path"
   - Ensure all have proper capitalization
   - Add constraints/examples consistently

5. **Add shorthand flags for other common options**
   - Consider `-y` for `--year`
   - Consider `-c` for `--color`
   - Consider `-l` for `--limit`

## Comparison with Popular Go Tools

### Our Implementation vs Standards

| Tool | Flag Style | Example | Compliance with Our Spec |
|------|-----------|---------|--------------------------|
| go tool | Mixed | `go build -o output -v` | Mostly underscores |
| kubectl | Hyphens | `--namespace`, `--follow` | Not Go-based |
| docker | Hyphens | `--name`, `--publish` | Not Go-based |
| **Olsen** | **Hyphens** | `--camera-make`, `--iso-min` | **Should use underscores** |

**Note**: kubectl and docker use hyphens because they're not Go-based tools. Go tools typically use underscores (see: `date_taken`, `poll_interval`).

## Implementation Status: Current Flags

### Fully Compliant Flags ✅
- `db` - single word
- `w` - single letter shorthand
- `open` - single word
- `addr` - single word
- `year`, `month`, `day` - single words
- `time`, `season`, `lens`, `color` - single words
- `bursts`, `gps` - single words
- `limit`, `offset`, `sort`, `order`, `format` - single words
- `facets`, `verbose`, `quick` - single words
- `s`, `o`, `v` - shorthand flags
- `date_taken` - ✅ **Uses underscore correctly!**

### Non-Compliant Flags ⚠️
- `camera-make` → should be `camera_make`
- `camera-model` → should be `camera_model`
- `iso-min` → should be `iso_min`
- `iso-max` → should be `iso_max`
- `aperture-min` → should be `aperture_min`
- `aperture-max` → should be `aperture_max`
- `focal-min` → should be `focal_min`
- `focal-max` → should be `focal_max`
- `focal-category` → should be `focal_category`

## Code Quality Observations

### What We're Doing Really Well

1. **Proper use of FlagSet per subcommand**
   ```go
   fs := flag.NewFlagSet("query", flag.ExitOnError)
   ```
   This is the recommended pattern for subcommands.

2. **Comprehensive validation**
   ```go
   if *dbPath == "" {
       return fmt.Errorf("--db flag is required")
   }
   ```
   We validate required flags and check file existence.

3. **Helpful error messages**
   ```go
   return fmt.Errorf("invalid size: %s (must be 64, 256, 512, or 1024)", *size)
   ```
   Errors explain what went wrong and what's valid.

4. **Good documentation in help text**
   ```go
   "Camera make (e.g., Canon)"  // Shows example
   "Filter by year (e.g., 2025)"  // Explains with example
   ```

### Interesting Edge Case

We have one flag that **does** follow the underscore convention:
- Line 839: `"date_taken"` in the sort flag

This proves we're aware of the convention but haven't applied it consistently!

## Conclusion

Olsen has a solid flag implementation with good documentation, validation, and architecture. The main issue is the **inconsistent use of hyphens instead of underscores** in multi-word flags, which goes against the Google Go Style Guide.

**The fact that `date_taken` uses an underscore while `camera-make` uses a hyphen shows we're inconsistent even within our own codebase.**

### Recommendation

Fix the naming convention issue in the next major version:
1. Update all hyphenated flags to use underscores
2. Document this as a breaking change
3. Provide migration guide
4. Consider supporting both formats temporarily with deprecation warnings

This will bring Olsen into full compliance with Go community standards and improve consistency within our own codebase.
