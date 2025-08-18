package cmd

import (
	"fmt"
	"os"

	"github.com/jepemo/miko-manifest/pkg/mikomanifest"
	"github.com/spf13/cobra"
)

var validateDir string
var validateEnvironment string
var validateConfigDir string
var validateSkipSchemaValidation bool

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate YAML files in the specified directory",
	Long:  `Validate YAML files in the specified directory using native Go YAML parser and validate Kubernetes manifests. Supports custom resource validation using schemas from environment configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no directory specified but environment is provided, try to detect from environment info
		if validateDir == "" && validateEnvironment != "" {
			return
		}
		
		// If directory is provided as positional argument
		if len(args) > 0 && validateDir == "" {
			validateDir = args[0]
		}
		
		if validateDir == "" {
			fmt.Println("Error: directory is required (use --dir or provide as argument)")
			os.Exit(1)
		}
		
		options := mikomanifest.LintOptions{
			Directory:            validateDir,
			Environment:          validateEnvironment,
			ConfigDir:            validateConfigDir,
			SkipSchemaValidation: validateSkipSchemaValidation,
		}
		
		if err := mikomanifest.LintDirectory(options); err != nil {
			fmt.Printf("Error validating directory: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	validateCmd.Flags().StringVarP(&validateDir, "dir", "d", "", "Directory to validate for YAML files")
	validateCmd.Flags().StringVarP(&validateEnvironment, "env", "e", "", "Environment configuration to use for schema loading")
	validateCmd.Flags().StringVarP(&validateConfigDir, "config", "c", "config", "Configuration directory path (used with --env)")
	validateCmd.Flags().BoolVar(&validateSkipSchemaValidation, "skip-schema-validation", false, "Skip custom resource schema validation")
}
