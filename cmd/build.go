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
	buildValidate      bool
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Generate Kubernetes manifests from templates and configuration",
	Long: `Generate Kubernetes manifests by processing templates with environment-specific configurations.

This command combines templates with configuration to produce ready-to-deploy Kubernetes manifests.
Use --validate flag to automatically validate generated manifests after build.

Typical workflow:
  1. miko-manifest config --env <environment>     # Inspect configuration
  2. miko-manifest check                          # Validate configuration  
  3. miko-manifest build --env <environment>      # Generate manifests
  4. miko-manifest validate --dir <output-dir>    # Validate generated manifests

Related commands:
  - Use 'check' to validate configuration before building
  - Use 'validate' to check generated manifests after building`,
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
		}
		
		mikoManifest := mikomanifest.New(options)
		if err := mikoManifest.Build(); err != nil {
			fmt.Printf("Error building project: %v\n", err)
			os.Exit(1)
		}
		
		// Run validation if requested
		if buildValidate {
			fmt.Println("\n🔍 Running validation...")
			lintOptions := mikomanifest.LintOptions{
				Directory:   buildOutputDir,
				Environment: buildEnv,
				ConfigDir:   buildConfigDir,
			}
			
			if err := mikomanifest.LintDirectory(lintOptions); err != nil {
				fmt.Printf("Error during validation: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	buildCmd.Flags().StringVarP(&buildEnv, "env", "e", "", "Environment configuration to use (required)")
	buildCmd.Flags().StringVarP(&buildOutputDir, "output-dir", "o", "", "Output directory for generated files (required)")
	buildCmd.Flags().StringVarP(&buildConfigDir, "config", "c", "config", "Configuration directory path")
	buildCmd.Flags().StringVarP(&buildTemplatesDir, "templates", "t", "templates", "Templates directory path")
	buildCmd.Flags().StringSliceVarP(&buildVariables, "var", "", []string{}, "Override variables in format: --var VAR_NAME=VALUE")
	buildCmd.Flags().BoolVar(&buildValidate, "validate", false, "Run validation after build using schemas from environment config")
	
	// Mark required flags - ignore errors as they're only for documentation purposes
	_ = buildCmd.MarkFlagRequired("env")
	_ = buildCmd.MarkFlagRequired("output-dir")
}
