package cmd

import (
	"fmt"
	"os"

	"github.com/jepemo/miko-manifest/pkg/mikomanifest"
	"github.com/spf13/cobra"
)

var lintDir string

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint YAML files in the specified directory",
	Long:  `Lint YAML files in the specified directory using yamllint and validate Kubernetes manifests.`,
	Run: func(cmd *cobra.Command, args []string) {
		options := mikomanifest.LintOptions{
			Directory: lintDir,
		}
		
		if err := mikomanifest.LintDirectory(options); err != nil {
			fmt.Printf("Error linting directory: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	lintCmd.Flags().StringVarP(&lintDir, "dir", "d", "", "Directory to lint for YAML files (required)")
	lintCmd.MarkFlagRequired("dir")
}
