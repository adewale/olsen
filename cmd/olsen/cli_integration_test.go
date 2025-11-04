package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLICommands_Integration tests that all CLI commands are actually implemented
// and don't just exit with stub messages.
//
// This test ensures that:
// 1. All documented commands exist
// 2. Commands don't exit with "not yet implemented" messages
// 3. Commands actually perform their documented functionality
// 4. Exit codes are correct (0 for success, non-zero for errors)

var olsenBinary = "../../bin/olsen"

// ensureBinary ensures the olsen binary exists
func ensureBinary(t *testing.T) {
	if _, err := os.Stat(olsenBinary); os.IsNotExist(err) {
		t.Fatalf("olsen binary not found at %s. Run 'make build' first.", olsenBinary)
	}
}

// runCommand executes the olsen binary with given args and returns stdout, stderr, and exit code
func runCommand(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(olsenBinary, args...)

	outBytes, err := cmd.CombinedOutput()
	output := string(outBytes)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("Failed to run command: %v", err)
		}
	} else {
		exitCode = 0
	}

	return output, output, exitCode
}

// TestVersionCommand verifies the version command works
func TestVersionCommand(t *testing.T) {
	ensureBinary(t)

	stdout, _, exitCode := runCommand(t, "version")

	if exitCode != 0 {
		t.Errorf("version command should exit with 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "olsen version") {
		t.Errorf("version command should output version info, got: %s", stdout)
	}

	// Should not contain stub messages
	if strings.Contains(stdout, "not yet implemented") {
		t.Error("version command should not be a stub")
	}
}

// TestHelpCommand verifies the help command works
func TestHelpCommand(t *testing.T) {
	ensureBinary(t)

	stdout, _, exitCode := runCommand(t, "help")

	if exitCode != 0 {
		t.Errorf("help command should exit with 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Olsen") {
		t.Errorf("help command should output usage info, got: %s", stdout)
	}
}

// TestIndexCommand_NotStub verifies index command is implemented
func TestIndexCommand_NotStub(t *testing.T) {
	ensureBinary(t)

	// Create temporary test directory
	tmpDir := t.TempDir()
	testPhotoDir := filepath.Join(tmpDir, "photos")
	if err := os.MkdirAll(testPhotoDir, 0755); err != nil {
		t.Fatalf("Failed to create test photo dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")

	stdout, _, exitCode := runCommand(t, "index", testPhotoDir, "--db", dbPath, "--w", "1")

	// Should not contain stub messages
	if strings.Contains(stdout, "not yet fully implemented") {
		t.Error("❌ FAIL: index command is still a stub")
		t.Logf("Output: %s", stdout)
	}

	if strings.Contains(stdout, "Please use ./indexphotos.sh") {
		t.Error("❌ FAIL: index command redirects to shell script instead of working")
	}

	// Should exit with 0 for empty directory (valid case)
	if exitCode != 0 {
		// Allow exit code 0 or success indicators in output
		if !strings.Contains(stdout, "0 photos indexed") && !strings.Contains(stdout, "complete") {
			t.Errorf("index command should succeed for empty directory, got exit code %d", exitCode)
		}
	}

	// Database should be created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("index command should create database file")
	}
}

// TestExploreCommand_NotStub verifies explore command is implemented
func TestExploreCommand_NotStub(t *testing.T) {
	ensureBinary(t)

	// We can't actually start the server in tests, so we'll check for stub messages
	// by running with --help
	stdout, _, _ := runCommand(t, "explore", "--help")

	if strings.Contains(stdout, "not yet fully implemented") {
		t.Error("❌ FAIL: explore command is still a stub")
		t.Logf("Output: %s", stdout)
	}

	if strings.Contains(stdout, "Please use ./explorer.sh") {
		t.Error("❌ FAIL: explore command redirects to shell script instead of working")
	}
}

// TestStatsCommand_NotStub verifies stats command is implemented
func TestStatsCommand_NotStub(t *testing.T) {
	ensureBinary(t)

	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	stdout, _, exitCode := runCommand(t, "stats", "--db", dbPath)

	if strings.Contains(stdout, "not yet implemented") {
		t.Error("❌ FAIL: stats command is still a stub")
		t.Logf("Output: %s", stdout)
	}

	// Should handle missing database gracefully
	if exitCode == 0 {
		t.Error("stats command should exit with non-zero for missing database")
	}
}

// TestAnalyzeCommand_NotStub verifies analyze command is implemented
func TestAnalyzeCommand_NotStub(t *testing.T) {
	ensureBinary(t)

	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	stdout, _, _ := runCommand(t, "analyze", "--db", dbPath)

	if strings.Contains(stdout, "not yet implemented") {
		t.Error("❌ FAIL: analyze command is still a stub")
		t.Logf("Output: %s", stdout)
	}
}

// TestShowCommand_Exists verifies show command exists
func TestShowCommand_Exists(t *testing.T) {
	ensureBinary(t)

	stdout, _, _ := runCommand(t, "show", "--help")

	if strings.Contains(stdout, "Unknown command 'show'") {
		t.Error("❌ FAIL: show command does not exist")
	}

	if strings.Contains(stdout, "not yet implemented") {
		t.Error("❌ FAIL: show command is a stub")
	}
}

// TestThumbnailCommand_Exists verifies thumbnail command exists
func TestThumbnailCommand_Exists(t *testing.T) {
	ensureBinary(t)

	stdout, _, _ := runCommand(t, "thumbnail", "--help")

	if strings.Contains(stdout, "Unknown command 'thumbnail'") {
		t.Error("❌ FAIL: thumbnail command does not exist")
	}

	if strings.Contains(stdout, "not yet implemented") {
		t.Error("❌ FAIL: thumbnail command is a stub")
	}
}

// TestVerifyCommand_Exists verifies verify command exists
func TestVerifyCommand_Exists(t *testing.T) {
	ensureBinary(t)

	stdout, _, _ := runCommand(t, "verify", "--help")

	if strings.Contains(stdout, "Unknown command 'verify'") {
		t.Error("❌ FAIL: verify command does not exist")
	}

	if strings.Contains(stdout, "not yet implemented") {
		t.Error("❌ FAIL: verify command is a stub")
	}
}

// TestIndexCommand_WithRealPhotos tests actual indexing functionality
func TestIndexCommand_WithRealPhotos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ensureBinary(t)

	// Use testdata/dng if it exists
	testPhotoDir := "../../testdata/dng"
	if _, err := os.Stat(testPhotoDir); os.IsNotExist(err) {
		t.Skip("testdata/dng not found, skipping real photo test")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	stdout, _, exitCode := runCommand(t, "index", testPhotoDir, "--db", dbPath, "--w", "2")

	if exitCode != 0 {
		t.Errorf("index command should succeed with test photos, got exit code %d", exitCode)
		t.Logf("Output: %s", stdout)
	}

	// Should show progress or completion message
	if !strings.Contains(stdout, "indexed") && !strings.Contains(stdout, "complete") {
		t.Error("index command should show completion status")
	}

	// Database should exist and have content
	if info, err := os.Stat(dbPath); err != nil {
		t.Error("index command should create database file")
	} else if info.Size() == 0 {
		t.Error("database should not be empty after indexing")
	}
}

// TestStatsCommand_WithDatabase tests stats on actual database
func TestStatsCommand_WithDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ensureBinary(t)

	// First create a database by indexing
	tmpDir := t.TempDir()
	testPhotoDir := "../../testdata/dng"
	if _, err := os.Stat(testPhotoDir); os.IsNotExist(err) {
		t.Skip("testdata/dng not found")
	}

	dbPath := filepath.Join(tmpDir, "test.db")

	// Index first
	_, _, exitCode := runCommand(t, "index", testPhotoDir, "--db", dbPath, "--w", "1")
	if exitCode != 0 {
		t.Skip("Could not create test database")
	}

	// Now test stats
	stdout, _, exitCode := runCommand(t, "stats", "--db", dbPath)

	if exitCode != 0 {
		t.Errorf("stats command should succeed, got exit code %d", exitCode)
		t.Logf("Output: %s", stdout)
	}

	// Should show statistics
	if !strings.Contains(stdout, "photo") && !strings.Contains(stdout, "Photo") {
		t.Error("stats command should display photo statistics")
	}
}

// TestAllCommandsInHelp verifies all documented commands appear in help
func TestAllCommandsInHelp(t *testing.T) {
	ensureBinary(t)

	stdout, _, _ := runCommand(t, "help")

	expectedCommands := []string{
		"index",
		"explore",
		"analyze",
		"stats",
		"show",
		"thumbnail",
		"verify",
		"version",
		"help",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(stdout, cmd) {
			t.Errorf("help output should mention '%s' command", cmd)
		}
	}
}

// TestCommandExitCodes verifies proper exit codes
func TestCommandExitCodes(t *testing.T) {
	ensureBinary(t)

	tests := []struct {
		name     string
		args     []string
		wantZero bool
	}{
		{"version succeeds", []string{"version"}, true},
		{"help succeeds", []string{"help"}, true},
		{"unknown command fails", []string{"notacommand"}, false},
		{"index without args fails", []string{"index"}, false},
		{"stats with missing db fails", []string{"stats", "--db", "/nonexistent/db.db"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, exitCode := runCommand(t, tt.args...)

			if tt.wantZero && exitCode != 0 {
				t.Errorf("Expected exit code 0, got %d", exitCode)
			}
			if !tt.wantZero && exitCode == 0 {
				t.Errorf("Expected non-zero exit code, got 0")
			}
		})
	}
}
