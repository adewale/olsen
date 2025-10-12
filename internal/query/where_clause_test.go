package query

import (
	"strings"
	"testing"
)

// TestWhereClauseMonthWithoutYear verifies that Month filter is applied even when Year is nil
// This is critical for the state machine model - when computing Year facet, we remove Year
// from params but need to preserve Month filter.
func TestWhereClauseMonthWithoutYear(t *testing.T) {
	engine := &Engine{}

	// Create params with Month but no Year (simulates computing Year facet)
	month := 1
	params := QueryParams{
		Year:  nil,    // Year removed for facet computation
		Month: &month, // Month should still be applied
		Limit: 100,
	}

	where, args := engine.buildWhereClause(params)

	// Should include Month filter
	whereStr := strings.Join(where, " AND ")
	if !strings.Contains(whereStr, "strftime('%m', p.date_taken) = ?") {
		t.Errorf("Expected Month filter to be applied even without Year, but got: %s", whereStr)
	}

	// Should have month value in args
	foundMonth := false
	for _, arg := range args {
		if arg == "01" {
			foundMonth = true
			break
		}
	}
	if !foundMonth {
		t.Errorf("Expected month value '01' in args, but got: %v", args)
	}
}

// TestWhereClauseDayWithoutMonthOrYear verifies that Day filter is applied independently
func TestWhereClauseDayWithoutMonthOrYear(t *testing.T) {
	engine := &Engine{}

	// Create params with Day but no Month or Year (simulates computing Month facet)
	day := 15
	params := QueryParams{
		Year:  nil,
		Month: nil,
		Day:   &day,
		Limit: 100,
	}

	where, args := engine.buildWhereClause(params)

	// Should include Day filter
	whereStr := strings.Join(where, " AND ")
	if !strings.Contains(whereStr, "strftime('%d', p.date_taken) = ?") {
		t.Errorf("Expected Day filter to be applied even without Month/Year, but got: %s", whereStr)
	}

	// Should have day value in args
	foundDay := false
	for _, arg := range args {
		if arg == "15" {
			foundDay = true
			break
		}
	}
	if !foundDay {
		t.Errorf("Expected day value '15' in args, but got: %v", args)
	}
}

// TestWhereClauseAllTemporalFilters verifies all three work together
func TestWhereClauseAllTemporalFilters(t *testing.T) {
	engine := &Engine{}

	year := 2024
	month := 1
	day := 15
	params := QueryParams{
		Year:  &year,
		Month: &month,
		Day:   &day,
		Limit: 100,
	}

	where, args := engine.buildWhereClause(params)
	whereStr := strings.Join(where, " AND ")

	// Should include all three filters
	if !strings.Contains(whereStr, "strftime('%Y', p.date_taken)") {
		t.Errorf("Expected Year filter, but got: %s", whereStr)
	}
	if !strings.Contains(whereStr, "strftime('%m', p.date_taken)") {
		t.Errorf("Expected Month filter, but got: %s", whereStr)
	}
	if !strings.Contains(whereStr, "strftime('%d', p.date_taken)") {
		t.Errorf("Expected Day filter, but got: %s", whereStr)
	}

	// Verify args
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d: %v", len(args), args)
	}
}

// TestWhereClauseMonthOnly verifies Month can be used alone (for "all Januarys")
func TestWhereClauseMonthOnly(t *testing.T) {
	engine := &Engine{}

	month := 1
	params := QueryParams{
		Year:  nil,    // No year filter
		Month: &month, // Just month
		Limit: 100,
	}

	where, args := engine.buildWhereClause(params)
	whereStr := strings.Join(where, " AND ")

	// Should include ONLY Month filter
	if strings.Contains(whereStr, "strftime('%Y', p.date_taken)") {
		t.Errorf("Should not include Year filter, but got: %s", whereStr)
	}
	if !strings.Contains(whereStr, "strftime('%m', p.date_taken)") {
		t.Errorf("Expected Month filter, but got: %s", whereStr)
	}

	// Should have exactly one arg (month)
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d: %v", len(args), args)
	}
	if args[0] != "01" {
		t.Errorf("Expected month '01', got: %s", args[0])
	}
}
