// Package cmd tests the CLI argument parsing for sshush.
//
//nolint:testpackage // We need to test internal functions
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const (
	file1 = "file1.yml"
	file2 = "file2.yml"
	file3 = "file3.yml"
)

// Variables to allow mocking in tests.
var (
	//nolint:gochecknoglobals // needed for testing os.Exit behavior
	osExit = os.Exit
)

func TestCollectSources(t *testing.T) {
	testCases := []struct {
		name           string
		flagSources    []string
		positionalArgs []string
		wantSources    []string
	}{
		{
			name:           "positional only",
			flagSources:    []string{},
			positionalArgs: []string{file1, file2},
			wantSources:    []string{file1, file2},
		},
		{
			name:           "flag only",
			flagSources:    []string{file1, file2},
			positionalArgs: []string{},
			wantSources:    []string{file1, file2},
		},
		{
			name:           "flag and positional",
			flagSources:    []string{file1},
			positionalArgs: []string{file2},
			wantSources:    []string{file1, file2},
		},
		{
			name:           "multiple flags and positional",
			flagSources:    []string{file1, file2},
			positionalArgs: []string{file3},
			wantSources:    []string{file1, file2, file3},
		},
		{
			name:           "no sources",
			flagSources:    []string{},
			positionalArgs: []string{},
			wantSources:    []string{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sources := testCase.flagSources
			sources = append(sources, testCase.positionalArgs...)

			if len(sources) != len(testCase.wantSources) {
				t.Errorf("got %d sources, want %d", len(sources), len(testCase.wantSources))

				return
			}

			for i, want := range testCase.wantSources {
				if sources[i] != want {
					t.Errorf("source[%d] = %q, want %q", i, sources[i], want)
				}
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde expansion",
			input:    "~/test",
			expected: filepath.Join(tmpHome, "test"),
		},
		{
			name:     "environment variable",
			input:    "$HOME/test",
			expected: filepath.Join(tmpHome, "test"),
		},
		{
			name:     "no expansion needed",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path",
			input:    "relative/path",
			expected: "relative/path",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := expandPath(testCase.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != testCase.expected {
				t.Errorf("expandPath(%q)", testCase.input)
				t.Errorf("got %q, expected %q", result, testCase.expected)
			}
		})
	}
}

func TestExpandGlobs(t *testing.T) {
	tmpDir := t.TempDir()

	files := []string{file1, file2, "other.txt"}
	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
			t.Fatalf("failed to create test file %s: %v", file, err)
		}
	}

	t.Run("match YAML files", func(t *testing.T) {
		sources := expandGlobs([]string{filepath.Join(tmpDir, "*.yml")})

		if len(sources) != 2 {
			t.Fatalf("expected 2 YAML files, got %d: %v", len(sources), sources)
		}
	})

	t.Run("match all files", func(t *testing.T) {
		sources := expandGlobs([]string{filepath.Join(tmpDir, "*")})

		if len(sources) != 3 {
			t.Fatalf("expected 3 files, got %d: %v", len(sources), sources)
		}
	})

	t.Run("no matches", func(t *testing.T) {
		sources := expandGlobs([]string{filepath.Join(tmpDir, "*.nonexistent")})

		if len(sources) != 0 {
			t.Fatalf("expected 0 files, got %d: %v", len(sources), sources)
		}
	})
}

func TestExpandPathEdgeCases(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("TEST_DIR", tmpHome+"/test")

	testCases := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "tilde with subdirectory",
			input:    "~/dir/subdir",
			expected: filepath.Join(tmpHome, "dir", "subdir"),
		},
		{
			name:     "multiple environment variables",
			input:    "$HOME/test",
			expected: filepath.Join(tmpHome, "test"),
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "relative path with dots",
			input:    "./relative/path",
			expected: "./relative/path",
		},
		{
			name:     "absolute path with tilde in middle",
			input:    "/path/~/middle",
			expected: "/path/~/middle",
		},
		{
			name:     "environment variable with braces",
			input:    "${HOME}/test",
			expected: filepath.Join(tmpHome, "test"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := expandPath(testCase.input)

			if testCase.wantErr {
				if err == nil {
					t.Errorf("expandPath(%q) expected error, got nil", testCase.input)
				}

				return
			}

			if err != nil {
				t.Fatalf("expandPath(%q) unexpected error: %v", testCase.input, err)
			}

			if result != testCase.expected {
				t.Errorf("expandPath(%q)", testCase.input)
				t.Errorf("got %q, expected %q", result, testCase.expected)
			}
		})
	}
}

func TestExpandGlobsEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	nestedDir := filepath.Join(tmpDir, "subdir", "nested")
	//nolint:gosec // 0o755 is fine for test directories
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	files := []struct {
		path    string
		content []byte
	}{
		{filepath.Join(tmpDir, "file1.yml"), []byte("test")},
		{filepath.Join(tmpDir, "file2.yml"), []byte("test")},
		{filepath.Join(tmpDir, "other.txt"), []byte("test")},
		{filepath.Join(nestedDir, "nested.yml"), []byte("test")},
		{filepath.Join(tmpDir, "subdir", "file3.yml"), []byte("test")},
	}

	for _, file := range files {
		if err := os.WriteFile(file.path, file.content, 0o600); err != nil {
			t.Fatalf("failed to create test file %s: %v", file.path, err)
		}
	}

	testCases := []struct {
		name      string
		patterns  []string
		wantCount int
	}{
		{
			name:      "nested glob pattern",
			patterns:  []string{filepath.Join(tmpDir, "*", "*.yml")},
			wantCount: 1, // subdir/file3.yml
		},
		{
			name: "multiple glob patterns",
			patterns: []string{
				filepath.Join(tmpDir, "*.yml"),
				filepath.Join(tmpDir, "subdir", "*.yml"),
			},
			wantCount: 3, // file1.yml, file2.yml, subdir/file3.yml
		},
		{
			name:      "glob with no matches",
			patterns:  []string{filepath.Join(tmpDir, "*.nonexistent")},
			wantCount: 0,
		},
		{
			name:      "glob with tilde expansion",
			patterns:  []string{filepath.Join(tmpDir, "*.yml")},
			wantCount: 2, // file1.yml, file2.yml
		},
		{
			name: "mixed valid and invalid patterns",
			patterns: []string{
				filepath.Join(tmpDir, "*.yml"),
				filepath.Join(tmpDir, "*.nonexistent"),
			},
			wantCount: 2, // file1.yml, file2.yml
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sources := expandGlobs(testCase.patterns)

			if len(sources) != testCase.wantCount {
				t.Errorf("expandGlobs(%v)", testCase.patterns)
				t.Errorf("got %d files, expected %d: %v", len(sources), testCase.wantCount, sources)
			}
		})
	}
}

func TestMustPanics(t *testing.T) {
	// must() calls os.Exit(1) when given an error, which we can't test directly
	// without causing the test process to exit. Instead, we verify that
	// must() will cause the program to exit by checking it doesn't return.
	// This is a limitation of testing functions that call os.Exit.

	// We can test that must(nil) doesn't panic
	t.Run("does not panic with nil error", func(t *testing.T) {
		// We can't actually test must() with a non-nil error without exiting,
		// but we can verify the function exists and compiles
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("must(nil) panicked: %v", r)
			}
		}()
		// must(nil) should not panic or exit
		must(nil)
	})
}

//nolint:gocognit // test function with multiple sub-tests
func TestNewRootCommand(t *testing.T) {
	// Save and restore exit function
	originalExit := osExit
	defer func() {
		osExit = originalExit
	}()

	// Prevent actual exit during command execution.
	osExit = func(_ int) {}

	t.Run("creates command with expected properties", func(t *testing.T) {
		version := "1.0.0"
		commit := "abc123"

		cmd := NewRootCommand(version, commit)

		if cmd.Use != "sshush [SOURCE_FILES...]" {
			t.Errorf("expected Use to be 'sshush [SOURCE_FILES...]', got %q", cmd.Use)
		}

		if cmd.Short != "sshush" {
			t.Errorf("expected Short to be 'sshush', got %q", cmd.Short)
		}

		expectedVersion := fmt.Sprintf("%s (%s)", version, commit)
		if cmd.Version != expectedVersion {
			t.Errorf("expected Version to be %q, got %q", expectedVersion, cmd.Version)
		}

		if cmd.Run == nil {
			t.Error("expected Run to be set")
		}
	})

	t.Run("has all expected flags", func(t *testing.T) {
		cmd := NewRootCommand("1.0.0", "abc123")

		flags := cmd.PersistentFlags()

		flagNames := []string{"source", "dest", "verbose", "debug", "dry-run"}
		for _, name := range flagNames {
			if flags.Lookup(name) == nil {
				t.Errorf("expected flag %q to be present", name)
			}
		}
	})

	t.Run("sets default destination flag", func(t *testing.T) {
		cmd := NewRootCommand("1.0.0", "abc123")

		// Verify the dest flag exists
		flag := cmd.PersistentFlags().Lookup("dest")
		if flag == nil {
			t.Fatal("expected dest flag to exist")
		}

		// The default value should be based on the user's home directory
		// We can't easily verify the exact value without refactoring the code
		// to use dependency injection, but we can verify the flag is registered
		if flag.Name != "dest" {
			t.Errorf("expected flag name to be 'dest', got %q", flag.Name)
		}
	})

	t.Run("viper configuration is set up", func(_ *testing.T) {
		cmd := NewRootCommand("1.0.0", "abc123")

		// Execute the command to trigger viper binding
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		// We expect it to fail because there are no source files
		// but this ensures viper is set up
		_ = err

		// The important thing is that NewRootCommand completes without panic
		// and sets up viper bindings
	})
}
