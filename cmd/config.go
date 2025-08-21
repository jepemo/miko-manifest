package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/jepemo/miko-manifest/pkg/mikomanifest"
	"github.com/jepemo/miko-manifest/pkg/output"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display configuration information for an environment",
	Long: `Display configuration information for a specific environment.

Shows the unified configuration including variables, included files, and schemas.
Use this command to inspect and understand your configuration structure before building.

Options:
  --variables: Show only variables in key=value format
  --schemas: Show list of schema definitions
  --tree: Show configuration tree structure`,
	RunE: runConfig,
}

type ConfigOptions struct {
	Environment string
	ConfigDir   string
	ShowTree    bool
	Variables   bool
	Schemas     bool
	Verbose     bool
}

var configOptions ConfigOptions

func init() {
	configCmd.Flags().StringVarP(&configOptions.Environment, "env", "e", "", "Environment configuration to use (required)")
	configCmd.Flags().StringVarP(&configOptions.ConfigDir, "config", "c", "config", "Configuration directory path")
	configCmd.Flags().BoolVar(&configOptions.ShowTree, "tree", false, "Show the hierarchy of included resources")
	configCmd.Flags().BoolVar(&configOptions.Variables, "variables", false, "Show only variables in format: var=value")
	configCmd.Flags().BoolVar(&configOptions.Schemas, "schemas", false, "Show list of all schemas")
	configCmd.Flags().BoolVarP(&configOptions.Verbose, "verbose", "v", false, "Enable verbose output")

	// Mark required flag - ignore error as it's only for documentation purposes
	_ = configCmd.MarkFlagRequired("env")
}

func runConfig(cmd *cobra.Command, args []string) error {
	if err := configOptions.Validate(); err != nil {
		return fmt.Errorf("validation error: %v", err)
	}

	// Create output options
	outputOpts := &output.OutputOptions{Verbose: configOptions.Verbose}

	// Display based on requested format
	if configOptions.ShowTree {
		return displayConfigTreeWithLoading(configOptions.ConfigDir, configOptions.Environment, outputOpts)
	} else if configOptions.Variables {
		// Load the configuration without verbose tree display
		config, err := mikomanifest.LoadConfig(configOptions.ConfigDir, configOptions.Environment, false)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %v", err)
		}
		return displayVariables(config, outputOpts)
	} else if configOptions.Schemas {
		// Load the configuration without verbose tree display
		config, err := mikomanifest.LoadConfig(configOptions.ConfigDir, configOptions.Environment, false)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %v", err)
		}
		return displaySchemas(config, outputOpts)
	} else {
		// Load the configuration without verbose tree display
		config, err := mikomanifest.LoadConfig(configOptions.ConfigDir, configOptions.Environment, false)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %v", err)
		}
		return displayFullConfig(config, outputOpts)
	}
}

func (opts *ConfigOptions) Validate() error {
	if opts.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	return nil
}

func displayFullConfig(config *mikomanifest.Config, outputOpts *output.OutputOptions) error {
	if outputOpts.Verbose {
		outputOpts.PrintStep(fmt.Sprintf("Displaying full configuration for environment: %s", config.Environment))
		outputOpts.PrintInfo(fmt.Sprintf("Config directory: %s", config.ConfigDir))
	}

	fmt.Printf("# Configuration for environment: %s\n", config.Environment)
	fmt.Printf("# Config directory: %s\n\n", config.ConfigDir)

	// Show resources
	if len(config.Resources) > 0 {
		fmt.Println("resources:")
		for _, resource := range config.Resources {
			fmt.Printf("  - %s\n", resource)
		}
		fmt.Println()
	}

	// Show variables
	if len(config.Variables) > 0 {
		fmt.Println("variables:")
		for _, variable := range config.Variables {
			fmt.Printf("  - name: %s\n", variable.Name)
			fmt.Printf("    value: %s\n", variable.Value)
		}
		fmt.Println()
	}

	// Show includes
	if len(config.Include) > 0 {
		fmt.Println("include:")
		for _, include := range config.Include {
			fmt.Printf("  - file: %s\n", include.File)
			if include.Repeat != "" {
				fmt.Printf("    repeat: %s\n", include.Repeat)
				if len(include.List) > 0 {
					fmt.Printf("    list:\n")
					for _, item := range include.List {
						fmt.Printf("      - key: %s\n", item.Key)
						if len(item.Values) > 0 {
							fmt.Printf("        values:\n")
							for _, value := range item.Values {
								fmt.Printf("          - name: %s\n", value.Name)
								fmt.Printf("            value: %s\n", value.Value)
							}
						}
					}
				}
			}
		}
	}

	if outputOpts.Verbose {
		resourceCount := len(config.Resources)
		variableCount := len(config.Variables)
		includeCount := len(config.Include)
		outputOpts.PrintSummary(fmt.Sprintf("Configuration summary: %d resource(s), %d variable(s), %d include(s)", 
			resourceCount, variableCount, includeCount))
	}

	return nil
}

func displayConfigTreeWithLoading(configDir, environment string, outputOpts *output.OutputOptions) error {
	outputOpts.PrintStep(fmt.Sprintf("Loading configuration hierarchy for environment: %s", environment))
	outputOpts.PrintInfo(fmt.Sprintf("Config directory: %s", configDir))
	
	// Load configuration with tree display enabled to show the loading process
	config, err := mikomanifest.LoadConfigWithOutput(configDir, environment, true, outputOpts)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}
	
	outputOpts.PrintInfo(fmt.Sprintf("Configuration hierarchy for environment: %s", environment))
	
	// Then show the tree structure
	return displayConfigTree(config, outputOpts)
}

func displayConfigTree(config *mikomanifest.Config, outputOpts *output.OutputOptions) error {
	outputOpts.PrintInfo("CONFIG TREE:\n")
	fmt.Printf("%s.yaml\n", config.Environment)
	
	// Show resources if they exist  
	if len(config.Resources) > 0 {
		fmt.Println("|-- resources:")
		for i, resource := range config.Resources {
			if i == len(config.Resources)-1 {
				fmt.Printf("    |-- %s\n", resource)
			} else {
				fmt.Printf("    |-- %s\n", resource)
			}
		}
	}
	
	// Show variables
	if len(config.Variables) > 0 {
		fmt.Println("|-- variables:")
		for i, variable := range config.Variables {
			if i == len(config.Variables)-1 && len(config.Include) == 0 {
				fmt.Printf("    |-- %s=%s\n", variable.Name, variable.Value)
			} else {
				fmt.Printf("    |-- %s=%s\n", variable.Name, variable.Value)
			}
		}
	}
	
	// Show includes
	if len(config.Include) > 0 {
		fmt.Println("|-- templates:")
		for i, include := range config.Include {
			if i == len(config.Include)-1 {
				fmt.Printf("    |-- %s", include.File)
			} else {
				fmt.Printf("    |-- %s", include.File)
			}
			if include.Repeat != "" {
				fmt.Printf(" (repeat: %s)", include.Repeat)
			}
			fmt.Println()
		}
	}

	return nil
}

func displayVariables(config *mikomanifest.Config, outputOpts *output.OutputOptions) error {
	if outputOpts.Verbose {
		outputOpts.PrintStep(fmt.Sprintf("Displaying variables for environment: %s", config.Environment))
	}

	if len(config.Variables) == 0 {
		outputOpts.PrintWarning("Variables", "No variables defined")
		return nil
	}

	if outputOpts.Verbose {
		outputOpts.PrintSummary(fmt.Sprintf("Found %d variable(s)", len(config.Variables)))
	}

	for _, variable := range config.Variables {
		fmt.Printf("%s=%s\n", variable.Name, variable.Value)
	}

	return nil
}

func displaySchemas(config *mikomanifest.Config, outputOpts *output.OutputOptions) error {
	if outputOpts.Verbose {
		outputOpts.PrintStep(fmt.Sprintf("Loading schema configuration for environment: %s", config.Environment))
		outputOpts.PrintInfo(fmt.Sprintf("Config directory: %s", config.ConfigDir))
	}

	// Load schema configuration if it exists
	schemaPath := filepath.Join(config.ConfigDir, "schemas.yaml")
	
	if outputOpts.Verbose {
		outputOpts.PrintInfo(fmt.Sprintf("Looking for schemas in: %s", schemaPath))
	}
	
	schemaConfig, err := mikomanifest.LoadSchemaConfig(schemaPath)
	if err != nil {
		outputOpts.PrintWarning("Schema loading", fmt.Sprintf("No schemas configured or error loading schemas: %v", err))
		return nil
	}

	if len(schemaConfig.Schemas) == 0 {
		outputOpts.PrintWarning("Schema validation", "No schemas defined")
		return nil
	}

	if outputOpts.Verbose {
		outputOpts.PrintInfo(fmt.Sprintf("Found %d schema(s) for environment: %s", len(schemaConfig.Schemas), config.Environment))
	}
	
	fmt.Printf("\nSCHEMAS:\n")
	for _, schema := range schemaConfig.Schemas {
		fmt.Printf("|-- %s\n", schema)
	}

	return nil
}
