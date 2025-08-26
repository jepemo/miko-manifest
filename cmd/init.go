package cmd

import (
	"fmt"
	"os"

	"github.com/jepemo/miko-manifest/pkg/mikomanifest"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new miko-manifest project",
	Long:  `Initialize a new miko-manifest project by creating necessary directories and example files.`,
	Run: func(cmd *cobra.Command, args []string) {
		workingDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting working directory: %v\n", err)
			os.Exit(1)
		}

		options := mikomanifest.InitOptions{
			ProjectDir: workingDir,
		}

		if err := mikomanifest.InitProject(options); err != nil {
			fmt.Printf("Error initializing project: %v\n", err)
			os.Exit(1)
		}
	},
}
