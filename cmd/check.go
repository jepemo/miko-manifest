package cmd

import (
	"fmt"
	"os"

	"github.com/jepemo/miko-manifest/pkg/mikomanifest"
	"github.com/spf13/cobra"
)

var checkConfigDir string
var checkEnvironment string
var checkSkipSchemaValidation bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate configuration files before manifest generation",
	Long: `Validate configuration YAML files in the specified directory before generating Kubernetes manifests.

This command checks the syntax and structure of your configuration files (input) to ensure they are
valid before running 'build'. It validates:
  - YAML syntax in configuration files
  - Configuration structure and required fields
  - Schema validation for custom resources (if enabled)

Typical workflow:
  1. miko-manifest check --env <environment>     # Validate configuration
  2. miko-manifest build --env <environment>     # Generate manifests
  3. miko-manifest validate --dir <output-dir>   # Validate generated manifests

Related commands:
  - Use 'validate' to check generated Kubernetes manifests
  - Use 'config' to inspect configuration values`,
	Run: func(cmd *cobra.Command, args []string) {
		options := mikomanifest.CheckOptions{
			ConfigDir:            checkConfigDir,
			Environment:          checkEnvironment,
			SkipSchemaValidation: checkSkipSchemaValidation,
		}
		
		if err := mikomanifest.CheckConfigDirectory(options); err != nil {
			fmt.Printf("Error checking config directory: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	checkCmd.Flags().StringVarP(&checkConfigDir, "config", "c", "config", "Configuration directory path")
	checkCmd.Flags().StringVarP(&checkEnvironment, "env", "e", "", "Environment configuration to use for schema loading")
	checkCmd.Flags().BoolVar(&checkSkipSchemaValidation, "skip-schema-validation", false, "Skip custom resource schema validation")
}
