// Package main provides the entry point for sshush.
package main

import (
	"io"
	"log/slog"
	"os"

	"github.com/bencromwell/sshush/cmd"
	"github.com/golang-cz/devslog"
)

//nolint:gochecknoglobals // These variables are set using ldflags.
var (
	version = "0.0.0-dev"
	commit  = ""
)

// Run is the main entry point logic extracted for testing.
// It creates the root command and executes it with the given output writer.
func Run(version, commit string, out io.Writer) error {
	logger := slog.New(devslog.NewHandler(out, nil))
	slog.SetDefault(logger)

	rootCmd := cmd.NewRootCommand(version, commit)

	return rootCmd.Execute()
}

func main() {
	err := Run(version, commit, os.Stdout)
	if err != nil {
		slog.Error("%v\n", "error", err)
		os.Exit(1)
	}
}
