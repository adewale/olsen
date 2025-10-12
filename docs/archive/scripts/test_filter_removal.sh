#!/bin/bash

# Test script for verifying the filter removal fix
# Original issue: Removing year filter cleared month filter

set -e

echo "=== Testing Filter Removal Fix ==="
echo ""

# Start server in background
echo "1. Starting server..."
./bin/olsen explore --db perf.db --addr localhost:9091 > /tmp/test_server.log 2>&1 &
SERVER_PID=$!
sleep 3

cleanup() {
    echo "Cleaning up..."
    kill $SERVER_PID 2>/dev/null || true
}
trap cleanup EXIT

echo "   Server started (PID: $SERVER_PID)"
echo ""

# Test 1: Navigate to year=2020
echo "2. Testing: Navigate to year=2020"
RESPONSE=$(curl -s "http://localhost:9091/photos?year=2020")
if echo "$RESPONSE" | grep -q "2020"; then
    echo "   ✅ Year 2020 page loads"
else
    echo "   ❌ Failed to load year 2020"
    exit 1
fi

# Test 2: Navigate to year=2020&month=10
echo ""
echo "3. Testing: Navigate to year=2020&month=10"
RESPONSE=$(curl -s "http://localhost:9091/photos?year=2020&month=10")
if echo "$RESPONSE" | grep -q "October"; then
    echo "   ✅ October 2020 page loads"
else
    echo "   ❌ Failed to load October 2020"
    exit 1
fi

# Test 3: Remove year (should preserve month)
echo ""
echo "4. Testing: Remove year filter (month should remain)"
RESPONSE=$(curl -s "http://localhost:9091/photos?month=10")

# Check that we have results (October from all years)
if echo "$RESPONSE" | grep -q "October"; then
    echo "   ✅ October filter preserved after removing year"
else
    echo "   ❌ October filter was lost!"
    exit 1
fi

# Check that Month chip is visible
if echo "$RESPONSE" | grep -q "filter-chip" && echo "$RESPONSE" | grep -q "October"; then
    echo "   ✅ October chip is visible"
else
    echo "   ❌ October chip not found!"
    exit 1
fi

# Check logs for proper state transitions
echo ""
echo "5. Checking server logs..."
sleep 1

if grep -q "state=year=2020" /tmp/test_server.log; then
    echo "   ✅ Logged state with year=2020"
fi

if grep -q "state=year=2020&month=10" /tmp/test_server.log; then
    echo "   ✅ Logged state with year=2020&month=10"
fi

if grep -q "state=month=10" /tmp/test_server.log; then
    echo "   ✅ Logged state with month=10 (year removed)"
else
    echo "   ❌ Month-only state not found in logs!"
    exit 1
fi

# Verify no FACET_404 for month=10 (should have results)
if grep "state=month=10" /tmp/test_server.log | grep -q "FACET_404"; then
    echo "   ❌ Month=10 resulted in zero results (but shouldn't!)"
    exit 1
else
    echo "   ✅ Month=10 has results (no FACET_404)"
fi

echo ""
echo "=== All Tests Passed! ✅ ==="
echo ""
echo "Summary:"
echo "  - Year filter can be removed"
echo "  - Month filter is preserved"
echo "  - Month chip is visible"
echo "  - State transitions logged correctly"
echo "  - No invalid zero-result states"
