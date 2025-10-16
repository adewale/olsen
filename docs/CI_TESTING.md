# CI Testing Strategy

## Overview

Olsen uses a multi-tier testing strategy to balance CI speed, dependency requirements, and test coverage.

## Test Tiers

### Tier 1: CI Tests (No Dependencies)
**Target**: `make test-ci`
**Environment**: GitHub Actions (Ubuntu)
**CGO**: Disabled (no native dependencies)
**Purpose**: Fast validation of core logic without external dependencies

**What runs:**
- URL parsing tests
- Query string building tests
- WHERE clause generation tests
- Facet URL building logic (no database queries)

**What's excluded:**
- Database tests (require SQLite with CGO)
- RAW processing tests (require LibRaw)
- Diagnostic tests (intentionally fail to document bugs)

**Exit code:** Always 0 (passes in CI)

### Tier 2: Local Development Tests
**Target**: `make test`
**Environment**: Developer machine
**CGO**: Disabled
**Purpose**: Quick local validation without database setup

Similar to CI tests but shows all failures (uses `|| true` to not block development).

### Tier 3: Complete Test Suite
**Target**: `make test-all`
**Environment**: Developer machine with LibRaw installed
**CGO**: Enabled
**Purpose**: Full integration testing before commits

**What runs:**
- All database tests (indexer, query, explorer)
- All RAW processing tests
- All facet tests (including camera facet bug tests)
- All integration tests
- Diagnostic tests (intentionally fail - see below)

**Exit code:** May fail on diagnostic tests (expected)

### Tier 4: Query Package Tests (No Diagnostics)
**Target**: `make test-query-all`
**Environment**: Developer machine with LibRaw
**CGO**: Enabled
**Purpose**: Validate all query functionality without diagnostic noise

**What runs:**
- All functional query tests
- Camera facet tests (multi-word make bug fix validation)
- State machine transition tests
- Facet computation tests

**What's excluded:**
- `TestDiagnostic_*` tests (documentation only)

**Exit code:** 0 (all functional tests pass)

## Diagnostic Tests

Some tests are prefixed with `TestDiagnostic_` and **intentionally fail**. These document bugs that were fixed and demonstrate the debugging methodology.

### Example: Camera Facet Bug

**TestDiagnostic_Layer3_URLBuilding_SplitBug**
- Documents the `strings.SplitN(value, " ", 2)` bug
- Shows how it incorrectly split "Leica Camera AG LEICA M11 Monochrom"
- Demonstrates layer-by-layer debugging approach

**TestDiagnostic_RootCauseDiagnosis**
- Provides comprehensive root cause analysis
- Documents the fix (separate CameraMake/CameraModel fields)
- Shows alternative solutions that were considered

**Why intentionally fail?**
- Makes the bug analysis highly visible
- Prevents accidental regression (tests would pass if bug reintroduced)
- Serves as documentation that's validated by the test runner

### Running Diagnostic Tests

```bash
# Run only diagnostic tests (will fail - that's expected)
make test-camera-facets-diagnostic

# Run all tests INCLUDING diagnostics (some will fail)
CGO_ENABLED=1 CGO_CFLAGS="-w" go test -tags='use_seppedelanghe_libraw' -v ./internal/query/

# Run all tests EXCLUDING diagnostics (all pass)
make test-query-all
```

## CI Configuration

### GitHub Actions Workflow

**File**: `.github/workflows/ci.yml`

**Steps:**
1. Format check (`gofmt`)
2. Static analysis (`go vet`)
3. Unit tests (`make test-ci`) - NO CGO, NO diagnostics
4. Build (`make build`) - NO RAW support
5. Binary validation (`./bin/olsen version`, `./bin/olsen --help`)

**Why no CGO in CI?**
- Installing LibRaw in CI adds complexity
- Installing SQLite driver requires build tools
- Most bugs don't require database for detection
- Fast feedback loop (CI runs in ~2 minutes)

**Future enhancement:** Add optional CGO job for comprehensive testing

## Test Targets Reference

| Target | CGO | LibRaw | Database | Diagnostics | Use Case |
|--------|-----|--------|----------|-------------|----------|
| `make test-ci` | ❌ | ❌ | ❌ | Skip | CI/CD pipeline |
| `make test` | ❌ | ❌ | ❌ | Skip | Quick local check |
| `make test-query-all` | ✅ | ✅ | ✅ | Skip | Pre-commit validation |
| `make test-all` | ✅ | ✅ | ✅ | Include | Full test suite |
| `make test-camera-facets` | ✅ | ✅ | ✅ | Skip | Camera facet validation |
| `make test-camera-facets-diagnostic` | ✅ | ✅ | ✅ | Only | Bug documentation |

## Local Development Workflow

### Before Committing
```bash
# 1. Run fast tests (no CGO)
make test-ci

# 2. Run full query tests (with CGO)
make test-query-all

# 3. If working on indexer/RAW processing
make test-all

# 4. Format and vet
gofmt -w .
go vet ./...
```

### Debugging Failures
```bash
# Run specific test
go test -v ./internal/query/ -run TestCameraFacetWithMultiWordMake

# Run with CGO enabled
CGO_ENABLED=1 CGO_CFLAGS="-w" go test -tags='use_seppedelanghe_libraw' -v ./internal/query/ -run TestCameraFacet

# See diagnostic information
make test-camera-facets-diagnostic
```

## Adding New Tests

### Naming Conventions
- **Functional tests**: `TestFeatureName` (should pass)
- **Diagnostic tests**: `TestDiagnostic_FeatureName` (may intentionally fail)
- **Integration tests**: `TestIntegration_FeatureName` (requires full setup)

### Test Tags
```go
// For tests that need database
// (automatically excluded when CGO_ENABLED=0)

// For diagnostic tests
// Skip with: go test -skip "TestDiagnostic"
func TestDiagnostic_BugName(t *testing.T) {
    t.Error("This test documents a bug that was fixed")
    // ... detailed explanation
}
```

## Troubleshooting

### "database tests skipped"
**Expected** - database tests require `CGO_ENABLED=1`
**Solution**: Run `make test-query-all` or `make test-all`

### "diagnostic tests failing"
**Expected** - they document bugs via intentional failure
**Solution**: Run `make test-query-all` to skip diagnostics

### "LibRaw not found"
**Needed for**: RAW processing tests
**Solution**: `brew install libraw` (macOS) or skip RAW tests

### CI failing
**Check**: Is `make test-ci` passing locally?
**Common causes**:
- Forgot to run `gofmt -w .`
- Added database test without `CGO_ENABLED=1` check
- New test depends on LibRaw

## Best Practices

1. **Write CGO-independent tests when possible** - faster feedback
2. **Use diagnostic tests to document bugs** - they serve as executable documentation
3. **Run `make test-query-all` before pushing** - catches most issues
4. **Keep CI fast** - add expensive tests to `test-all`, not `test-ci`
5. **Name tests clearly** - `TestDiagnostic_` prefix makes intent obvious
