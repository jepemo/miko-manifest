package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "miko-manifest",
	Short: "A CLI application for Kubernetes manifest configuration management",
	Long: `miko-manifest is a powerful CLI tool for managing Kubernetes manifest configurations.
It provides templating capabilities using Go templates and supports multiple deployment patterns.

Typical workflow:
  1. miko-manifest init                          # Initialize new project
  2. miko-manifest check                         # Validate configuration
  3. miko-manifest build --env <environment>     # Generate manifests
  4. miko-manifest validate --dir <output-dir>   # Validate generated manifests

Use "miko-manifest [command] --help" for detailed information about each command.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add all subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(configCmd)
}
