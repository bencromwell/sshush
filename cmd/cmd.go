// Package cmd provides the CLI for sshush.
package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/bencromwell/sshush/sshush"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func must(err error) {
	if err != nil {
		slog.Error("sshush", "error", err)
		os.Exit(1)
	}
}

// expandPath expands environment variables and the tilde (~) to the home directory.
func expandPath(path string) (string, error) {
	// Expand environment variables.
	path = os.ExpandEnv(path)

	// Expand tilde to home directory.
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home dir: %w", err)
		}

		path = filepath.Join(homeDir, path[1:])
	}

	return path, nil
}

// NewRootCommand creates a new root command for sshush.
func NewRootCommand(version, commit string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sshush",
		Short:   "sshush",
		Version: fmt.Sprintf("%s (%s)", version, commit),
		Run: func(cmd *cobra.Command, _ []string) {
			sources := viper.GetStringSlice("source")
			dest := viper.GetString("dest")

			fileSources := expandGlobs(sources)

			runner := &sshush.Runner{
				Sources:     fileSources,
				Destination: dest,
				Out:         os.Stdout,
			}

			verbose, err := cmd.Flags().GetBool("verbose")
			must(err)
			debug, err := cmd.Flags().GetBool("debug")
			must(err)
			dryRun, err := cmd.Flags().GetBool("dry-run")
			must(err)

			err = runner.Run(verbose, debug, dryRun, version)
			must(err)
		},
	}

	homeDir, err := os.UserHomeDir()
	must(err)

	cmd.PersistentFlags().StringSlice("source", []string{}, "the source file(s) to read from")
	cmd.PersistentFlags().String("dest", homeDir+"/.ssh/config", "the destination path to write to")
	cmd.PersistentFlags().BoolP("verbose", "V", false, "verbose output")
	cmd.PersistentFlags().Bool("debug", false, "debug output")
	cmd.PersistentFlags().Bool("dry-run", false, "print diff with current file instead of writing")

	must(viper.BindPFlag("source", cmd.PersistentFlags().Lookup("source")))
	must(viper.BindPFlag("dest", cmd.PersistentFlags().Lookup("dest")))

	viper.SetConfigName("sshush")
	viper.SetConfigType("yaml")

	viper.AddConfigPath(".")
	viper.AddConfigPath(homeDir + "/.ssh/")

	err = viper.ReadInConfig()
	if err != nil {
		// Config file is optional.
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			must(fmt.Errorf("reading config: %w", err))
		}
	}

	return cmd
}

// expandGlobs expands glob patterns and handles tilde and environment variables.
func expandGlobs(sources []string) []string {
	var fileSources []string

	for _, pattern := range sources {
		expandedPattern, err := expandPath(pattern)
		if err != nil {
			must(fmt.Errorf("expanding path: %w", err))
		}

		matches, err := filepath.Glob(expandedPattern)
		if err != nil {
			must(fmt.Errorf("expanding glob pattern: %w", err))
		}

		fileSources = append(fileSources, matches...)
	}

	return fileSources
}
