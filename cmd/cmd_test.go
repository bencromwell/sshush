// Package cmd tests the CLI argument parsing for sshush.
//
//nolint:testpackage // We need to test internal functions
package cmd

import (
	"os"
	"path/filepath"
	"testing"
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
			positionalArgs: []string{"file1.yml", "file2.yml"},
			wantSources:    []string{"file1.yml", "file2.yml"},
		},
		{
			name:           "flag only",
			flagSources:    []string{"file1.yml", "file2.yml"},
			positionalArgs: []string{},
			wantSources:    []string{"file1.yml", "file2.yml"},
		},
		{
			name:           "flag and positional",
			flagSources:    []string{"file1.yml"},
			positionalArgs: []string{"file2.yml"},
			wantSources:    []string{"file1.yml", "file2.yml"},
		},
		{
			name:           "multiple flags and positional",
			flagSources:    []string{"file1.yml", "file2.yml"},
			positionalArgs: []string{"file3.yml"},
			wantSources:    []string{"file1.yml", "file2.yml", "file3.yml"},
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

	files := []string{"file1.yml", "file2.yml", "other.txt"}
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
