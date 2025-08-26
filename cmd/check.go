package cmd

import (
	"fmt"
	"os"

	"github.com/jepemo/miko-manifest/pkg/mikomanifest"
	"github.com/jepemo/miko-manifest/pkg/output"
	"github.com/spf13/cobra"
)

var checkConfigDir string
var checkVerbose bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate configuration files before manifest generation",
	Long: `Validate configuration YAML files in the specified directory before generating Kubernetes manifests.

This command checks the syntax and structure of your configuration files to ensure they are
valid before running 'build'. It validates:
  - YAML syntax in configuration files
  - Configuration structure and required fields
  - Variable definitions and references`,
	Run: func(cmd *cobra.Command, args []string) {
		outputOpts := output.NewOutputOptions(checkVerbose)
		options := mikomanifest.CheckOptions{
			ConfigDir:  checkConfigDir,
			OutputOpts: outputOpts,
		}

		if err := mikomanifest.CheckConfigDirectory(options); err != nil {
			fmt.Printf("Error checking config directory: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	checkCmd.Flags().StringVarP(&checkConfigDir, "config", "c", "config", "Configuration directory path")
	checkCmd.Flags().BoolVarP(&checkVerbose, "verbose", "v", false, "Show detailed processing information")
}
