package cmd

import (
	"fmt"
	"os"

	"github.com/jepemo/miko-manifest/pkg/mikomanifest"
	"github.com/spf13/cobra"
)

var lintDir string
var lintEnvironment string
var lintConfigDir string
var lintSchemaConfig string
var lintSkipSchemaValidation bool

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint YAML files in the specified directory",
	Long:  `Lint YAML files in the specified directory using native Go YAML parser and validate Kubernetes manifests. Supports custom resource validation using schemas from environment configuration or explicit schema files.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no directory specified but environment is provided, try to detect from environment info
		if lintDir == "" && lintEnvironment != "" {
			return
		}
		
		// If directory is provided as positional argument
		if len(args) > 0 && lintDir == "" {
			lintDir = args[0]
		}
		
		if lintDir == "" {
			fmt.Println("Error: directory is required (use --dir or provide as argument)")
			os.Exit(1)
		}
		
		options := mikomanifest.LintOptions{
			Directory:            lintDir,
			Environment:          lintEnvironment,
			ConfigDir:            lintConfigDir,
			SchemaConfig:         lintSchemaConfig,
			SkipSchemaValidation: lintSkipSchemaValidation,
		}
		
		if err := mikomanifest.LintDirectory(options); err != nil {
			fmt.Printf("Error linting directory: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	lintCmd.Flags().StringVarP(&lintDir, "dir", "d", "", "Directory to lint for YAML files")
	lintCmd.Flags().StringVarP(&lintEnvironment, "env", "e", "", "Environment configuration to use for schema loading")
	lintCmd.Flags().StringVarP(&lintConfigDir, "config", "c", "config", "Configuration directory path (used with --env)")
	lintCmd.Flags().StringVarP(&lintSchemaConfig, "schema-config", "s", "", "Path to explicit schema configuration file")
	lintCmd.Flags().BoolVar(&lintSkipSchemaValidation, "skip-schema-validation", false, "Skip custom resource schema validation")
}
