package cmd

import (
	"fmt"
	"os"

	"github.com/jepemo/miko-manifest/pkg/mikomanifest"
	"github.com/spf13/cobra"
)

var checkConfigDir string

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check YAML files in the specified config directory",
	Long:  `Check YAML files in the specified config directory using yamllint only.`,
	Run: func(cmd *cobra.Command, args []string) {
		options := mikomanifest.CheckOptions{
			ConfigDir: checkConfigDir,
		}
		
		if err := mikomanifest.CheckConfigDirectory(options); err != nil {
			fmt.Printf("Error checking config directory: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	checkCmd.Flags().StringVarP(&checkConfigDir, "config", "c", "config", "Configuration directory path")
}
