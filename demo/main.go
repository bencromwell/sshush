package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	// Test StringSlice behavior
	sliceCmd := &cobra.Command{
		Use:   "slice-test",
		Short: "Test StringSlice behavior",
		Run: func(cmd *cobra.Command, args []string) {
			sources, _ := cmd.Flags().GetStringSlice("source")
			fmt.Printf("StringSlice result: %v (length: %d)\n", sources, len(sources))
			for i, s := range sources {
				fmt.Printf("  [%d]: %s\n", i, s)
			}
		},
	}
	sliceCmd.Flags().StringSlice("source", []string{}, "source files")

	// Test StringArray behavior
	arrayCmd := &cobra.Command{
		Use:   "array-test",
		Short: "Test StringArray behavior",
		Run: func(cmd *cobra.Command, args []string) {
			sources, _ := cmd.Flags().GetStringArray("source")
			fmt.Printf("StringArray result: %v (length: %d)\n", sources, len(sources))
			for i, s := range sources {
				fmt.Printf("  [%d]: %s\n", i, s)
			}
		},
	}
	arrayCmd.Flags().StringArrayP("source", "s", []string{}, "source files")

	rootCmd := &cobra.Command{Use: "flag-demo"}
	rootCmd.AddCommand(sliceCmd, arrayCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
