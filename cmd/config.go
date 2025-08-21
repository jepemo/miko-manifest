package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/jepemo/miko-manifest/pkg/mikomanifest"
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
  --tree: Show configuration tree structure

Typical workflow:
  1. miko-manifest config --env <environment>    # Inspect configuration
  2. miko-manifest check                         # Validate configuration
  3. miko-manifest build --env <environment>     # Generate manifests
  4. miko-manifest validate --dir <output-dir>   # Validate generated manifests`,
	RunE: runConfig,
}

type ConfigOptions struct {
	Environment string
	ConfigDir   string
	ShowTree    bool
	Variables   bool
	Schemas     bool
}

var configOptions ConfigOptions

func init() {
	configCmd.Flags().StringVarP(&configOptions.Environment, "env", "e", "", "Environment configuration to use (required)")
	configCmd.Flags().StringVarP(&configOptions.ConfigDir, "config", "c", "config", "Configuration directory path")
	configCmd.Flags().BoolVar(&configOptions.ShowTree, "tree", false, "Show the hierarchy of included resources")
	configCmd.Flags().BoolVar(&configOptions.Variables, "variables", false, "Show only variables in format: var=value")
	configCmd.Flags().BoolVar(&configOptions.Schemas, "schemas", false, "Show list of all schemas")

	// Mark required flag - ignore error as it's only for documentation purposes
	_ = configCmd.MarkFlagRequired("env")
}

func runConfig(cmd *cobra.Command, args []string) error {
	if err := configOptions.Validate(); err != nil {
		return fmt.Errorf("validation error: %v", err)
	}

	// Load the configuration (always silent unless --tree is specified for detailed loading)
	config, err := mikomanifest.LoadConfig(configOptions.ConfigDir, configOptions.Environment, false)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Display based on requested format
	if configOptions.ShowTree {
		return displayConfigTreeWithLoading(config)
	} else if configOptions.Variables {
		return displayVariables(config)
	} else if configOptions.Schemas {
		return displaySchemas(config)
	} else {
		return displayFullConfig(config)
	}
}

func (opts *ConfigOptions) Validate() error {
	if opts.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	return nil
}

func displayFullConfig(config *mikomanifest.Config) error {
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

	return nil
}

func displayConfigTreeWithLoading(config *mikomanifest.Config) error {
	// First show the loading process with tree details
	fmt.Printf("Loading configuration hierarchy for environment: %s\n", config.Environment)
	fmt.Printf("Config directory: %s\n\n", config.ConfigDir)
	
	// Reload with tree display enabled to show the loading process
	_, err := mikomanifest.LoadConfig(config.ConfigDir, config.Environment, true)
	if err != nil {
		return fmt.Errorf("failed to reload configuration: %v", err)
	}
	
	fmt.Println() // Add separator
	
	// Then show the tree structure
	return displayConfigTree(config)
}

func displayConfigTree(config *mikomanifest.Config) error {
	fmt.Printf("Configuration hierarchy for environment: %s\n", config.Environment)
	fmt.Printf("Config directory: %s\n\n", config.ConfigDir)

	// Show the main configuration file
	fmt.Printf("📄 %s.yaml\n", config.Environment)
	
	// Show resources if they exist  
	if len(config.Resources) > 0 {
		fmt.Println("├── resources:")
		for i, resource := range config.Resources {
			if i == len(config.Resources)-1 {
				fmt.Printf("    └── 📄 %s\n", resource)
			} else {
				fmt.Printf("    ├── 📄 %s\n", resource)
			}
		}
	}
	
	// Show variables
	if len(config.Variables) > 0 {
		fmt.Println("├── variables:")
		for i, variable := range config.Variables {
			if i == len(config.Variables)-1 && len(config.Include) == 0 {
				fmt.Printf("    └── %s=%s\n", variable.Name, variable.Value)
			} else {
				fmt.Printf("    ├── %s=%s\n", variable.Name, variable.Value)
			}
		}
	}
	
	// Show includes
	if len(config.Include) > 0 {
		fmt.Println("└── templates:")
		for i, include := range config.Include {
			if i == len(config.Include)-1 {
				fmt.Printf("    └── 📄 %s", include.File)
			} else {
				fmt.Printf("    ├── 📄 %s", include.File)
			}
			if include.Repeat != "" {
				fmt.Printf(" (repeat: %s)", include.Repeat)
			}
			fmt.Println()
		}
	}

	return nil
}

func displayVariables(config *mikomanifest.Config) error {
	if len(config.Variables) == 0 {
		fmt.Println("No variables defined")
		return nil
	}

	for _, variable := range config.Variables {
		fmt.Printf("%s=%s\n", variable.Name, variable.Value)
	}

	return nil
}

func displaySchemas(config *mikomanifest.Config) error {
	// Load schema configuration if it exists
	schemaPath := filepath.Join(config.ConfigDir, "schemas.yaml")
	schemaConfig, err := mikomanifest.LoadSchemaConfig(schemaPath)
	if err != nil {
		fmt.Printf("No schemas configured or error loading schemas: %v\n", err)
		return nil
	}

	if len(schemaConfig.Schemas) == 0 {
		fmt.Println("No schemas defined")
		return nil
	}

	fmt.Printf("Schemas for environment: %s\n\n", config.Environment)
	
	for _, schema := range schemaConfig.Schemas {
		fmt.Printf("Schema: %s\n", schema)
	}

	return nil
}
