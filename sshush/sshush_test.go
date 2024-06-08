package sshush_test

import (
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sshushRunner := &sshush.Runner{
				Sources:     tc.sources,
				Destination: filepath.Join("testdata", tc.destination),
			}

			err := sshushRunner.Run(true, true, "0.0.0-dev")
			require.NoError(t, err)

			generatedContents := string(golden.Get(t, tc.destination))
			golden.Assert(t, generatedContents, tc.goldenFile)
		})
	}
}
