package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jepemo/miko-manifest/pkg/mikomanifest"
	"github.com/spf13/cobra"
)

var (
	buildEnv           string
	buildOutputDir     string
	buildConfigDir     string
	buildTemplatesDir  string
	buildVariables     []string
	buildDebugConfig   bool
	buildShowConfigTree bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the miko-manifest project",
	Long:  `Build the miko-manifest project by processing templates with environment-specific configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Parse command line variables
		cmdVariables := make(map[string]string)
		for _, varPair := range buildVariables {
			parts := strings.SplitN(varPair, "=", 2)
			if len(parts) != 2 {
				fmt.Printf("Error: Invalid --var format: %s. Expected format: VAR_NAME=VALUE\n", varPair)
				os.Exit(1)
			}
			cmdVariables[parts[0]] = parts[1]
			fmt.Printf("✓ Override variable: %s=%s\n", parts[0], parts[1])
		}
		
		options := mikomanifest.BuildOptions{
			Environment:     buildEnv,
			OutputDir:       buildOutputDir,
			ConfigDir:       buildConfigDir,
			TemplatesDir:    buildTemplatesDir,
			Variables:       cmdVariables,
			DebugConfig:     buildDebugConfig,
			ShowConfigTree:  buildShowConfigTree,
		}
		
		mikoManifest := mikomanifest.New(options)
		if err := mikoManifest.Build(); err != nil {
			fmt.Printf("Error building project: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	buildCmd.Flags().StringVarP(&buildEnv, "env", "e", "", "Environment configuration to use (required)")
	buildCmd.Flags().StringVarP(&buildOutputDir, "output-dir", "o", "", "Output directory for generated files (required)")
	buildCmd.Flags().StringVarP(&buildConfigDir, "config", "c", "config", "Configuration directory path")
	buildCmd.Flags().StringVarP(&buildTemplatesDir, "templates", "t", "templates", "Templates directory path")
	buildCmd.Flags().StringSliceVarP(&buildVariables, "var", "", []string{}, "Override variables in format: --var VAR_NAME=VALUE")
	buildCmd.Flags().BoolVar(&buildDebugConfig, "debug-config", false, "Show the final merged configuration")
	buildCmd.Flags().BoolVar(&buildShowConfigTree, "show-config-tree", false, "Show the hierarchy of included resources")
	
	buildCmd.MarkFlagRequired("env")
	buildCmd.MarkFlagRequired("output-dir")
}
