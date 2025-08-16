package cmd

import (
	"fmt"
	"os"

	"github.com/jepemo/miko-manifest/pkg/mikomanifest"
	"github.com/spf13/cobra"
)

var lintDir string
var lintSchemaConfig string

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint YAML files in the specified directory",
	Long:  `Lint YAML files in the specified directory using native Go YAML parser and validate Kubernetes manifests. Optionally validate custom resources using CRD schemas.`,
	Run: func(cmd *cobra.Command, args []string) {
		options := mikomanifest.LintOptions{
			Directory:    lintDir,
			SchemaConfig: lintSchemaConfig,
		}
		
		if err := mikomanifest.LintDirectory(options); err != nil {
			fmt.Printf("Error linting directory: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	lintCmd.Flags().StringVarP(&lintDir, "dir", "d", "", "Directory to lint for YAML files (required)")
	lintCmd.Flags().StringVarP(&lintSchemaConfig, "schema-config", "s", "", "Path to schema configuration file for custom resource validation")
	lintCmd.MarkFlagRequired("dir")
}
