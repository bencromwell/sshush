// Package main provides end-to-end tests for the sshush binary.
// These tests build the actual binary and execute it to verify behavior.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBinaryVersion(t *testing.T) {
	// Build the binary
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// Run with --version flag
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run --version: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	
	// Version output should contain "sshush version"
	if !strings.Contains(outputStr, "sshush version") {
		t.Errorf("expected --version output to contain 'sshush version', got: %s", outputStr)
	}
}

func TestBinaryHelp(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run --help: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	
	// Help output should contain usage information
	if !strings.Contains(outputStr, "Usage:") {
		t.Errorf("expected --help output to contain 'Usage:', got: %s", outputStr)
	}

	// Help output should contain command description
	if !strings.Contains(outputStr, "sshush") {
		t.Errorf("expected --help output to contain 'sshush', got: %s", outputStr)
	}

	// Help output should contain flags
	expectedFlags := []string{"--source", "--dest", "--verbose", "--debug", "--dry-run", "--help", "--version"}
	for _, flag := range expectedFlags {
		if !strings.Contains(outputStr, flag) {
			t.Errorf("expected --help output to contain flag '%s', got: %s", flag, outputStr)
		}
	}
}

func TestBinaryInvalidFlag(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := exec.Command(binaryPath, "--invalid-flag")
	output, err := cmd.CombinedOutput()
	
	// Should exit with error
	if err == nil {
		t.Error("expected error for invalid flag, got nil")
	}

	outputStr := string(output)
	
	// Error output should mention the invalid flag
	if !strings.Contains(outputStr, "invalid-flag") && !strings.Contains(outputStr, "unknown flag") {
		t.Errorf("expected error output to mention invalid flag, got: %s", outputStr)
	}

	// Exit code should be 1
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	}
}

func TestBinaryNoArguments(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := exec.Command(binaryPath)
	output, err := cmd.CombinedOutput()
	
	// Should exit with error or succeed (cobra doesn't require args)
	// We just verify it doesn't panic
	_ = err
	_ = output
}

func TestBinaryExecution(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// Create a temporary source file
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "test.yml")
	destFile := filepath.Join(tmpDir, "config")
	
	// Write a simple SSH config
	sshyaml := `---
global:
  User: testuser

test:
  Hosts:
    myserver:
      HostName: example.com
`
	if err := os.WriteFile(sourceFile, []byte(sshyaml), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Run sshush with the test file
	cmd := exec.Command(binaryPath, "--source", sourceFile, "--dest", destFile)
	output, err := cmd.CombinedOutput()
	
	// Should succeed
	if err != nil {
		t.Fatalf("failed to run sshush: %v\nOutput: %s", err, string(output))
	}

	// Verify destination file was created
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("expected destination file to be created")
	}

	// Verify destination file contains expected content
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "example.com") {
		t.Errorf("expected destination file to contain 'example.com', got: %s", contentStr)
	}

	if !strings.Contains(contentStr, "testuser") {
		t.Errorf("expected destination file to contain 'testuser', got: %s", contentStr)
	}
}

func TestBinaryDryRun(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// Create a temporary source file
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "test.yml")
	destFile := filepath.Join(tmpDir, "config")
	
	// Write a simple SSH config
	sshyaml := `---
test:
  Hosts:
    myserver:
      HostName: example.com
`
	if err := os.WriteFile(sourceFile, []byte(sshyaml), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Run sshush with --dry-run
	cmd := exec.Command(binaryPath, "--source", sourceFile, "--dest", destFile, "--dry-run")
	output, err := cmd.CombinedOutput()
	
	// Should succeed
	if err != nil {
		t.Fatalf("failed to run sshush with --dry-run: %v\nOutput: %s", err, string(output))
	}

	// Verify destination file was NOT created in dry-run mode
	if _, err := os.Stat(destFile); !os.IsNotExist(err) {
		t.Error("expected destination file to NOT be created in dry-run mode")
		os.Remove(destFile) // Clean up
	}
}

// buildTestBinary builds the sshush binary for testing
func buildTestBinary(t *testing.T) string {
	t.Helper()
	
	// Build in a temp directory
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "sshush")
	
	// We need to build from the repo root
	repoRoot := getRepoRoot()
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary from %s: %v\n%s", repoRoot, err, string(output))
	}
	
	return binaryPath
}

// getRepoRoot returns the repository root directory
func getRepoRoot() string {
	// Try to find the repository root by looking for main.go
	cwd, _ := os.Getwd()
	
	// If we're in the repo root, main.go should exist
	if _, err := os.Stat(filepath.Join(cwd, "main.go")); err == nil {
		return cwd
	}
	
	// Otherwise, we're likely in a subdirectory, go up
	return filepath.Dir(cwd)
}
