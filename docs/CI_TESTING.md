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

### Tier 4: Query Package Tests
**Target**: `make test-query-all`
**Environment**: Developer machine with LibRaw
**CGO**: Enabled
**Purpose**: Validate all query functionality

**What runs:**
- All functional query tests
- Camera facet tests (multi-word make bug fix validation)
- State machine transition tests
- Facet computation tests

**Exit code:** 0 (all functional tests pass)

## CI Configuration

### GitHub Actions Workflow

**File**: `.github/workflows/ci.yml`

**Steps:**
1. Format check (`gofmt`)
2. Static analysis (`go vet`)
3. Unit tests (`make test-ci`) - NO CGO
4. Build (`make build`) - NO RAW support
5. Binary validation (`./bin/olsen version`, `./bin/olsen --help`)

**Why no CGO in CI?**
- Installing LibRaw in CI adds complexity
- Installing SQLite driver requires build tools
- Most bugs don't require database for detection
- Fast feedback loop (CI runs in ~2 minutes)

**Future enhancement:** Add optional CGO job for comprehensive testing

## Test Targets Reference

| Target | CGO | LibRaw | Database | Use Case |
|--------|-----|--------|----------|----------|
| `make test-ci` | ❌ | ❌ | ❌ | CI/CD pipeline |
| `make test` | ❌ | ❌ | ❌ | Quick local check |
| `make test-query-all` | ✅ | ✅ | ✅ | Pre-commit validation |
| `make test-all` | ✅ | ✅ | ✅ | Full test suite |
| `make test-camera-facets` | ✅ | ✅ | ✅ | Camera facet validation |

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
```

## Adding New Tests

### Naming Conventions
- **Functional tests**: `TestFeatureName` (should pass)
- **Integration tests**: `TestIntegration_FeatureName` (requires full setup)

### Test Tags
```go
// For tests that need database
// (automatically excluded when CGO_ENABLED=0)
```

## Troubleshooting

### "database tests skipped"
**Expected** - database tests require `CGO_ENABLED=1`
**Solution**: Run `make test-query-all` or `make test-all`

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
2. **Document a fixed bug with a regression test that fails if the bug returns** - not a skipped or always-failing test
3. **Run `make test-query-all` before pushing** - catches most issues
4. **Keep CI fast** - add expensive tests to `test-all`, not `test-ci`
5. **Name tests clearly** - name the behavior the test protects
