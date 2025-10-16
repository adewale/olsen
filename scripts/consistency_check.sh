#!/bin/bash
# consistency_check.sh - Automated consistency checker for Olsen codebase
#
# Purpose: Detect documentation-implementation gaps to prevent stub handler issues
# Usage: ./scripts/consistency_check.sh [--fix]
#
# Exit codes:
#   0 - All checks passed
#   1 - Inconsistencies found
#   2 - Script error

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
PASS=0
FAIL=0
WARN=0

# Check if running in CI
CI_MODE=${CI:-false}

# Parse arguments
FIX_MODE=false
if [ "$1" = "--fix" ]; then
    FIX_MODE=true
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Olsen Consistency Checker"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Helper functions
pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASS++))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    echo "  $2"
    ((FAIL++))
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    echo "  $2"
    ((WARN++))
}

section() {
    echo ""
    echo -e "${BLUE}▸${NC} $1"
}

# ============================================================================
# CHECK 1: Detect stub implementations in CLI
# ============================================================================
section "Checking for stub implementations in CLI..."

STUB_PATTERNS=(
    "not yet implemented"
    "not yet fully implemented"
    "TODO.*stub"
    "Please use .*\.sh"
)

STUBS_FOUND=false
for pattern in "${STUB_PATTERNS[@]}"; do
    # Exclude test files and benchmark files from stub detection
    if grep -rq "$pattern" cmd/olsen/*.go 2>/dev/null; then
        FILES=$(grep -rl "$pattern" cmd/olsen/*.go 2>/dev/null | grep -v "_test.go" | grep -v "benchmark_")
        if [ -n "$FILES" ]; then
            STUBS_FOUND=true
            fail "Found stub pattern: $pattern" "$FILES"
        fi
    fi
done

if [ "$STUBS_FOUND" = false ]; then
    pass "No stub implementations found in CLI"
fi

# ============================================================================
# CHECK 2: Verify documented CLI commands exist and work
# ============================================================================
section "Verifying documented CLI commands..."

# Extract commands from CLAUDE.md
DOCUMENTED_COMMANDS=(
    "index"
    "explore"
    "analyze"
    "stats"
    "show"
    "thumbnail"
    "verify"
)

# Check if binary exists
if [ ! -f "bin/olsen" ]; then
    warn "Binary not found" "Run 'make build' first. Skipping command tests."
else
    for cmd in "${DOCUMENTED_COMMANDS[@]}"; do
        # Check if command appears in main.go switch statement
        if ! grep -q "case \"$cmd\"" cmd/olsen/main.go; then
            fail "Command '$cmd' not found in main.go" "Documented but not implemented"
            continue
        fi

        # Try running command with --help
        OUTPUT=$(./bin/olsen "$cmd" --help 2>&1 || true)

        if echo "$OUTPUT" | grep -q "not yet implemented"; then
            fail "Command '$cmd' is a stub" "Help text contains 'not yet implemented'"
        elif echo "$OUTPUT" | grep -q "Please use.*\.sh"; then
            fail "Command '$cmd' redirects to shell script" "Should be native implementation"
        elif echo "$OUTPUT" | grep -q "Unknown command"; then
            fail "Command '$cmd' not recognized" "May need to add to switch statement"
        else
            pass "Command '$cmd' appears functional"
        fi
    done
fi

# ============================================================================
# CHECK 3: Verify CLI integration tests exist
# ============================================================================
section "Checking CLI integration test coverage..."

if [ ! -f "cmd/olsen/cli_integration_test.go" ]; then
    fail "CLI integration tests missing" "File cmd/olsen/cli_integration_test.go not found"
else
    pass "CLI integration tests exist"

    # Check that each command has a test
    for cmd in "${DOCUMENTED_COMMANDS[@]}"; do
        CMD_TITLE=$(echo "$cmd" | sed 's/\b\(.\)/\u\1/')
        if ! grep -q "Test${CMD_TITLE}Command" cmd/olsen/cli_integration_test.go; then
            warn "No integration test for '$cmd'" "Consider adding Test${CMD_TITLE}Command_NotStub"
        fi
    done
fi

# ============================================================================
# CHECK 4: Verify shell scripts call real commands
# ============================================================================
section "Checking shell scripts..."

SHELL_SCRIPTS=(
    "indexphotos.sh"
    "explorer.sh"
)

for script in "${SHELL_SCRIPTS[@]}"; do
    if [ ! -f "$script" ]; then
        warn "Shell script not found: $script" "Documented but missing"
        continue
    fi

    # Extract olsen commands called by script
    COMMANDS=$(grep -oE '\./bin/olsen [a-z]+' "$script" 2>/dev/null || true)

    if [ -z "$COMMANDS" ]; then
        warn "Script '$script' doesn't call olsen" "May be outdated"
        continue
    fi

    pass "Script '$script' calls olsen commands"
done

# ============================================================================
# CHECK 5: Check for undocumented commands
# ============================================================================
section "Checking for undocumented commands..."

# Extract case statements from main.go
IMPLEMENTED_COMMANDS=$(grep -oE 'case "([a-z]+)"' cmd/olsen/main.go | sed 's/case "\(.*\)"/\1/' | sort -u)

for cmd in $IMPLEMENTED_COMMANDS; do
    # Skip meta commands
    if [[ "$cmd" =~ ^(version|help|-v|--version|-h|--help)$ ]]; then
        continue
    fi

    # Check if documented in CLAUDE.md
    if ! grep -q "\./bin/olsen $cmd" CLAUDE.md 2>/dev/null; then
        warn "Command '$cmd' not documented in CLAUDE.md" "Add usage example"
    fi
done

# ============================================================================
# CHECK 6: Verify referenced documentation files exist
# ============================================================================
section "Verifying referenced documentation files..."

REFERENCED_FILES=(
    "specs/facet_state_machine.spec"
    "specs/dominant_colours.spec"
    "specs/olsen_specs.md"
    "docs/HIERARCHICAL_FACETS.md"
    "docs/LIBRAW_DUAL_LIBRARY_SUPPORT.md"
    "internal/explorer/templates/grid.html"
    "testdata/generate_fixtures.go"
)

for file in "${REFERENCED_FILES[@]}"; do
    if [ ! -f "$file" ]; then
        fail "Referenced file not found: $file" "Remove from documentation or create file"
    else
        pass "File exists: $file"
    fi
done

# ============================================================================
# CHECK 7: Validate claimed component completion percentages
# ============================================================================
section "Analyzing completion percentage claims..."

# This is a heuristic check - not perfect but catches obvious mismatches
check_component() {
    local name=$1
    local claimed_percent=$2
    local files_pattern=$3

    # Count files in component
    FILE_COUNT=$(find $files_pattern -name "*.go" ! -name "*_test.go" 2>/dev/null | wc -l | tr -d ' ')
    TEST_COUNT=$(find $files_pattern -name "*_test.go" 2>/dev/null | wc -l | tr -d ' ')

    if [ "$FILE_COUNT" -eq 0 ]; then
        if [ "$claimed_percent" -gt 10 ]; then
            fail "$name: Claimed ${claimed_percent}% but no files found" "Pattern: $files_pattern"
        else
            pass "$name: Correctly marked as incomplete"
        fi
    else
        # Heuristic: If test coverage is low, claimed percentage is suspicious
        if [ "$TEST_COUNT" -eq 0 ] && [ "$claimed_percent" -gt 50 ]; then
            warn "$name: Claimed ${claimed_percent}% but no tests found" "Add tests to validate completion"
        else
            pass "$name: Has $FILE_COUNT implementation files and $TEST_COUNT test files"
        fi
    fi
}

check_component "Indexer" 100 "internal/indexer"
check_component "Query Engine" 95 "internal/query"
check_component "Explorer" 90 "internal/explorer"
check_component "Database" 100 "internal/database"

# ============================================================================
# CHECK 8: Verify build targets work
# ============================================================================
section "Checking build system..."

# Check if Makefile has expected targets
REQUIRED_TARGETS=(
    "build"
    "build-raw"
    "test"
    "clean"
)

if [ ! -f "Makefile" ]; then
    fail "Makefile not found" "Build system missing"
else
    for target in "${REQUIRED_TARGETS[@]}"; do
        if ! grep -q "^$target:" Makefile; then
            fail "Makefile missing target: $target" "Add to Makefile"
        fi
    done
    pass "Makefile has required targets"
fi

# ============================================================================
# CHECK 9: Verify test data exists
# ============================================================================
section "Checking test data..."

if [ ! -d "testdata/dng" ]; then
    warn "Test data directory not found: testdata/dng" "Integration tests may fail"
else
    DNG_COUNT=$(find testdata/dng -name "*.dng" 2>/dev/null | wc -l | tr -d ' ')
    if [ "$DNG_COUNT" -eq 0 ]; then
        warn "No DNG files found in testdata/dng" "Add test photos"
    else
        pass "Found $DNG_COUNT test DNG files"
    fi
fi

# ============================================================================
# CHECK 10: Scan for TODO comments indicating incomplete features
# ============================================================================
section "Scanning for incomplete features..."

# Count TODOs in main source files (not tests, not vendor)
TODO_COUNT=$(grep -r "TODO" cmd/ internal/ pkg/ 2>/dev/null | grep -v "_test.go" | grep -v "vendor/" | wc -l | tr -d ' ')

if [ "$TODO_COUNT" -gt 50 ]; then
    warn "Found $TODO_COUNT TODO comments in source code" "High number suggests incomplete implementation"
elif [ "$TODO_COUNT" -gt 20 ]; then
    pass "Found $TODO_COUNT TODO comments (moderate)"
else
    pass "Found $TODO_COUNT TODO comments (low)"
fi

# ============================================================================
# Summary
# ============================================================================
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}Passed:${NC} $PASS"
echo -e "${YELLOW}Warnings:${NC} $WARN"
echo -e "${RED}Failed:${NC} $FAIL"
echo ""

# Exit code
if [ $FAIL -gt 0 ]; then
    echo "❌ Consistency check FAILED"
    echo ""
    echo "Action required: Fix the issues above before deploying or merging."
    echo "Run with --fix flag to attempt automatic repairs (not yet implemented)."
    exit 1
elif [ $WARN -gt 0 ]; then
    echo "⚠️  Consistency check passed with warnings"
    echo ""
    echo "Consider addressing warnings to improve code quality."
    exit 0
else
    echo "✅ All consistency checks passed!"
    exit 0
fi
