package sshush_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/bencromwell/sshush/sshush"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"
)

// TestFunctional tests the overall functionality. We provide various example
// sources and their expected configuration output.
func TestFunctional(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sources     []string
		destination string
		goldenFile  string
	}{
		{
			name:        "Example",
			sources:     []string{"testdata/example.yml"},
			destination: "example.out.test",
			goldenFile:  "example.golden",
		},
		{
			name:        "Examlpe 2",
			sources:     []string{"testdata/example2.yml"},
			destination: "example2.out.test",
			goldenFile:  "example2.golden",
		},
		{
			name:        "Ciscos",
			sources:     []string{"testdata/ciscos.yml"},
			destination: "ciscos.out.test",
			goldenFile:  "ciscos.golden",
		},
		{
			name:        "Ciscos 2",
			sources:     []string{"testdata/ciscos2.yml"},
			destination: "ciscos2.out.test",
			goldenFile:  "ciscos2.golden",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			sshushRunner := &sshush.Runner{
				Sources:     testCase.sources,
				Destination: filepath.Join("testdata", testCase.destination),
				Out:         &buf,
			}

			err := sshushRunner.Run(true, true, false, "0.0.0-dev")
			require.NoError(t, err)

			generatedContents := string(golden.Get(t, testCase.destination))
			golden.Assert(t, generatedContents, testCase.goldenFile)
		})
	}
}

func TestNoSourceFile(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	sshushRunner := &sshush.Runner{
		Sources:     []string{filepath.Join("testdata", "does_not_exist.yml")},
		Destination: filepath.Join("testdata", "does_not_exist.out"),
		Out:         &buf,
	}

	err := sshushRunner.Run(true, true, false, "0.0.0-dev")
	require.ErrorIs(t, err, sshush.ErrLoadingSources)
}

func TestBadConfig(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	source := filepath.Join("testdata", "bad-config.yml")

	sshushRunner := &sshush.Runner{
		Sources:     []string{source},
		Destination: filepath.Join("testdata", "irrelevant"),
		Out:         &buf,
	}

	err := sshushRunner.Run(true, true, false, "0.0.0-dev")
	require.ErrorIs(t, err, sshush.ErrProducingConfig)
}

func TestDryRun(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	source := filepath.Join("testdata", "aws.yml")

	sshushRunner := &sshush.Runner{
		Sources:     []string{source},
		Destination: filepath.Join("testdata", "dryrun_nofile.golden"),
		Out:         &buf,
	}

	err := sshushRunner.Run(false, false, true, "0.0.0-dev")
	require.NoError(t, err)

	generatedContents := buf.String()
	golden.Assert(t, generatedContents, "dryrun.golden")
}
