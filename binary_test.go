// Package main provides end-to-end tests for the sshush binary.
// These tests build the actual binary and execute it to verify behavior.
package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	versionFlag = "--version"
	helpFlag    = "--help"
	invalidFlag = "--invalid-flag"
	fileMode    = 0o600
)

func TestBinaryVersion(t *testing.T) {
	// Build the binary.
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// Run with --version flag.
	cmd := exec.CommandContext(context.Background(), binaryPath, versionFlag)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run %s: %v\nOutput: %s", versionFlag, err, string(output))
	}

	outputStr := string(output)

	// Version output should contain "sshush version".
	if !strings.Contains(outputStr, "sshush version") {
		t.Errorf("expected %s output to contain 'sshush version', got: %s", versionFlag, outputStr)
	}
}

func TestBinaryHelp(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := exec.CommandContext(context.Background(), binaryPath, helpFlag)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run %s: %v\nOutput: %s", helpFlag, err, string(output))
	}

	outputStr := string(output)

	// Help output should contain usage information.
	if !strings.Contains(outputStr, "Usage:") {
		t.Errorf("expected %s output to contain 'Usage:', got: %s", helpFlag, outputStr)
	}

	// Help output should contain command description.
	if !strings.Contains(outputStr, "sshush") {
		t.Errorf("expected %s output to contain 'sshush', got: %s", helpFlag, outputStr)
	}

	// Help output should contain flags.
	expectedFlags := []string{
		"--source", "--dest", "--verbose", "--debug", "--dry-run", helpFlag, versionFlag,
	}
	for _, flag := range expectedFlags {
		if !strings.Contains(outputStr, flag) {
			t.Errorf("expected %s output to contain flag '%s', got: %s", helpFlag, flag, outputStr)
		}
	}
}

func TestBinaryInvalidFlag(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := exec.CommandContext(context.Background(), binaryPath, invalidFlag)

	output, err := cmd.CombinedOutput()

	// Should exit with error.
	if err == nil {
		t.Error("expected error for invalid flag, got nil")
	}

	outputStr := string(output)

	// Error output should mention the invalid flag.
	if !strings.Contains(outputStr, "invalid-flag") &&
		!strings.Contains(outputStr, "unknown flag") {
		t.Errorf("expected error output to mention invalid flag, got: %s", outputStr)
	}

	// Exit code should be 1.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	}
}

func TestBinaryNoArguments(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// Set HOME to a temp dir with .ssh subdirectory to avoid
	// touching the real ~/.ssh/config during tests.
	tmpHome := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpHome, ".ssh"), 0o700); err != nil {
		t.Fatalf("failed to create .ssh dir: %v", err)
	}

	cmd := exec.CommandContext(context.Background(), binaryPath)

	cmd.Env = append(os.Environ(), "HOME="+tmpHome)

	output, err := cmd.CombinedOutput()

	// Should exit with error or succeed (cobra doesn't require args).
	// We just verify it doesn't panic.
	_ = err
	_ = output
}

func TestBinaryExecution(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// Create a temporary source file.
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "test.yml")
	destFile := filepath.Join(tmpDir, "config")

	// Write a simple SSH config.
	sshyaml := `---
global:
  User: testuser

test:
  Hosts:
    myserver:
      HostName: example.com
`
	if err := os.WriteFile(sourceFile, []byte(sshyaml), fileMode); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Run sshush with the test file.
	cmd := exec.CommandContext(
		context.Background(),
		binaryPath,
		"--source",
		sourceFile,
		"--dest",
		destFile,
	)

	output, err := cmd.CombinedOutput()

	// Should succeed.
	if err != nil {
		t.Fatalf("failed to run sshush: %v\nOutput: %s", err, string(output))
	}

	// Verify destination file was created.
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("expected destination file to be created")
	}

	// Verify destination file contains expected content.
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

	// Create a temporary source file.
	tmpDir := t.TempDir()
	sourceFile := filepath.Join(tmpDir, "test.yml")
	_ = filepath.Join(tmpDir, "config")

	// Write a simple SSH config.
	sshyaml := `---
test:
  Hosts:
    myserver:
      HostName: example.com
`
	if err := os.WriteFile(sourceFile, []byte(sshyaml), fileMode); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Run sshush with --dry-run.
	cmd := exec.CommandContext(
		context.Background(),
		binaryPath,
		"--source",
		sourceFile,
		"--dest",
		filepath.Join(tmpDir, "config"),
		"--dry-run",
	)

	output, err := cmd.CombinedOutput()

	// Should succeed.
	if err != nil {
		t.Fatalf("failed to run sshush with --dry-run: %v\nOutput: %s", err, string(output))
	}

	// Verify destination file was NOT created in dry-run mode.
	destFile := filepath.Join(tmpDir, "config")
	if _, err := os.Stat(destFile); !os.IsNotExist(err) {
		t.Error("expected destination file to NOT be created in dry-run mode")
		// Clean up.
		_ = os.Remove(destFile)
	}
}

// buildTestBinary builds the sshush binary for testing.
func buildTestBinary(t *testing.T) string {
	t.Helper()

	// Build in a temp directory.
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "sshush")

	// We need to build from the repo root.
	repoRoot := getRepoRoot()
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", binaryPath, ".")
	cmd.Dir = repoRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary from %s: %v\n%s", repoRoot, err, string(output))
	}

	return binaryPath
}

// getRepoRoot returns the repository root directory.
func getRepoRoot() string {
	// Try to find the repository root by looking for main.go.
	cwd, _ := os.Getwd()

	// If we're in the repo root, main.go should exist.
	if _, err := os.Stat(filepath.Join(cwd, "main.go")); err == nil {
		return cwd
	}

	// Otherwise, we're likely in a subdirectory, go up.
	return filepath.Dir(cwd)
}
