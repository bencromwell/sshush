// Package main provides the entry point for sshush.
package main

import (
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

func main() {
	logger := slog.New(devslog.NewHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	rootCmd := cmd.NewRootCommand(version, commit)

	err := rootCmd.Execute()
	if err != nil {
		slog.Error("%v\n", "error", err)
		os.Exit(1)
	}
}
