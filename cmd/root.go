package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version information - injected at build time via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
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
	// Add version flag
	var versionFlag bool
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Show version information")
	
	// Handle version flag before command execution
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Printf("miko-manifest version %s\n", version)
			fmt.Printf("commit: %s\n", commit)
			fmt.Printf("built: %s\n", date)
			os.Exit(0)
		}
	}

	// Add all subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(configCmd)
}
