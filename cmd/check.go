package cmd

import (
	"fmt"
	"os"

	"github.com/jepemo/miko-manifest/pkg/mikomanifest"
	"github.com/spf13/cobra"
)

var checkConfigDir string
var checkEnvironment string
var checkSchemaConfig string
var checkSkipSchemaValidation bool

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check YAML files in the specified config directory",
	Long:  `Check YAML files in the specified config directory using native Go YAML parser.`,
	Run: func(cmd *cobra.Command, args []string) {
		options := mikomanifest.CheckOptions{
			ConfigDir:            checkConfigDir,
			Environment:          checkEnvironment,
			SchemaConfig:         checkSchemaConfig,
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
	checkCmd.Flags().StringVarP(&checkSchemaConfig, "schema-config", "s", "", "Path to explicit schema configuration file")
	checkCmd.Flags().BoolVar(&checkSkipSchemaValidation, "skip-schema-validation", false, "Skip custom resource schema validation")
}
