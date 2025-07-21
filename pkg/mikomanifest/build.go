package mikomanifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration structure
type Config struct {
	Variables []Variable `yaml:"variables"`
	Include   []Include  `yaml:"include"`
}

// Variable represents a configuration variable
type Variable struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// Include represents a file to include in the build
type Include struct {
	File   string      `yaml:"file"`
	Repeat string      `yaml:"repeat,omitempty"`
	List   []ListItem  `yaml:"list,omitempty"`
}

// ListItem represents an item in a repeat list
type ListItem struct {
	Key    string     `yaml:"key"`
	Values []Variable `yaml:"values"`
}

// BuildOptions contains options for building
type BuildOptions struct {
	Environment   string
	OutputDir     string
	ConfigDir     string
	TemplatesDir  string
	Variables     map[string]string
}

// MikoManifest is the main library interface
type MikoManifest struct {
	options BuildOptions
}

// New creates a new MikoManifest instance
func New(options BuildOptions) *MikoManifest {
	return &MikoManifest{
		options: options,
	}
}

// LoadConfig loads configuration from ENV.yaml file
func (m *MikoManifest) LoadConfig(env string) (*Config, error) {
	configPath := filepath.Join(m.options.ConfigDir, fmt.Sprintf("%s.yaml", env))
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file %s not found", configPath)
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", configPath, err)
	}
	
	return &config, nil
}

// ValidateTemplateFiles checks that all template files exist
func (m *MikoManifest) ValidateTemplateFiles(includes []Include) error {
	templatesPath := m.options.TemplatesDir
	
	for _, include := range includes {
		templatePath := filepath.Join(templatesPath, include.File)
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			return fmt.Errorf("template file %s not found", templatePath)
		}
	}
	
	return nil
}

// MergeVariables merges global and local variables
func (m *MikoManifest) MergeVariables(globalVars []Variable, localVars []Variable, cmdVars map[string]string) map[string]string {
	variables := make(map[string]string)
	
	// Add global variables
	for _, v := range globalVars {
		variables[v.Name] = v.Value
	}
	
	// Add local variables (override global if same name)
	for _, v := range localVars {
		variables[v.Name] = v.Value
	}
	
	// Add command line variables (highest priority)
	for k, v := range cmdVars {
		variables[k] = v
	}
	
	return variables
}

// RenderTemplate renders a template with variables
func (m *MikoManifest) RenderTemplate(templateContent string, variables map[string]string, templateName string) (string, error) {
	tmpl, err := template.New(templateName).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}
	
	var result strings.Builder
	if err := tmpl.Execute(&result, variables); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}
	
	return result.String(), nil
}

// ProcessSimpleFile processes a simple file
func (m *MikoManifest) ProcessSimpleFile(templatePath, outputDir string, variables map[string]string) error {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}
	
	rendered, err := m.RenderTemplate(string(content), variables, filepath.Base(templatePath))
	if err != nil {
		return err
	}
	
	// Ensure content ends with newline
	if !strings.HasSuffix(rendered, "\n") {
		rendered += "\n"
	}
	
	outputFile := filepath.Join(outputDir, filepath.Base(templatePath))
	if err := os.WriteFile(outputFile, []byte(rendered), 0644); err != nil {
		return fmt.Errorf("failed to write output file %s: %w", outputFile, err)
	}
	
	fmt.Printf("✓ Processed: %s -> %s\n", filepath.Base(templatePath), outputFile)
	return nil
}

// ProcessSameFileRepeat processes a file with same-file repeat pattern
func (m *MikoManifest) ProcessSameFileRepeat(templatePath, outputDir string, globalVars map[string]string, listItems []ListItem) error {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}
	
	var renderedParts []string
	
	for _, item := range listItems {
		// Merge global variables with item-specific values
		variables := make(map[string]string)
		for k, v := range globalVars {
			variables[k] = v
		}
		for _, v := range item.Values {
			variables[v.Name] = v.Value
		}
		
		rendered, err := m.RenderTemplate(string(content), variables, fmt.Sprintf("%s[%s]", filepath.Base(templatePath), item.Key))
		if err != nil {
			return err
		}
		
		renderedParts = append(renderedParts, rendered)
	}
	
	// Join all parts with separator
	finalContent := strings.Join(renderedParts, "\n---\n")
	
	// Ensure content ends with newline
	if !strings.HasSuffix(finalContent, "\n") {
		finalContent += "\n"
	}
	
	outputFile := filepath.Join(outputDir, filepath.Base(templatePath))
	if err := os.WriteFile(outputFile, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to write output file %s: %w", outputFile, err)
	}
	
	fmt.Printf("✓ Processed (same-file): %s -> %s (%d sections)\n", filepath.Base(templatePath), outputFile, len(listItems))
	return nil
}

// ProcessMultipleFilesRepeat processes a file with multiple-files repeat pattern
func (m *MikoManifest) ProcessMultipleFilesRepeat(templatePath, outputDir string, globalVars map[string]string, listItems []ListItem) error {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}
	
	filename := filepath.Base(templatePath)
	ext := filepath.Ext(filename)
	stem := strings.TrimSuffix(filename, ext)
	
	for _, item := range listItems {
		// Merge global variables with item-specific values
		variables := make(map[string]string)
		for k, v := range globalVars {
			variables[k] = v
		}
		for _, v := range item.Values {
			variables[v.Name] = v.Value
		}
		
		rendered, err := m.RenderTemplate(string(content), variables, fmt.Sprintf("%s[%s]", filename, item.Key))
		if err != nil {
			return err
		}
		
		// Ensure content ends with newline
		if !strings.HasSuffix(rendered, "\n") {
			rendered += "\n"
		}
		
		outputFilename := fmt.Sprintf("%s-%s%s", stem, item.Key, ext)
		outputFile := filepath.Join(outputDir, outputFilename)
		
		if err := os.WriteFile(outputFile, []byte(rendered), 0644); err != nil {
			return fmt.Errorf("failed to write output file %s: %w", outputFile, err)
		}
		
		fmt.Printf("✓ Processed (multi-file): %s -> %s\n", filename, outputFile)
	}
	
	return nil
}

// Build builds the manifest project
func (m *MikoManifest) Build() error {
	fmt.Printf("Building miko-manifest project with environment: %s\n", m.options.Environment)
	fmt.Printf("✓ Using config directory: %s\n", m.options.ConfigDir)
	fmt.Printf("✓ Using templates directory: %s\n", m.options.TemplatesDir)
	
	// Validate directories
	if err := m.validateDirectories(); err != nil {
		return err
	}
	
	// Load configuration
	config, err := m.LoadConfig(m.options.Environment)
	if err != nil {
		return err
	}
	
	if len(config.Include) == 0 {
		return fmt.Errorf("no 'include' section found in configuration")
	}
	
	// Check that all template files exist
	if err := m.ValidateTemplateFiles(config.Include); err != nil {
		return err
	}
	
	// Create output directory
	if err := os.MkdirAll(m.options.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", m.options.OutputDir, err)
	}
	fmt.Printf("✓ Output directory: %s\n", m.options.OutputDir)
	
	// Get global variables and merge with command line overrides
	globalVariables := m.MergeVariables(config.Variables, nil, m.options.Variables)
	
	// Process each file in include
	for _, include := range config.Include {
		templatePath := filepath.Join(m.options.TemplatesDir, include.File)
		
		switch include.Repeat {
		case "":
			// Simple file include
			if err := m.ProcessSimpleFile(templatePath, m.options.OutputDir, globalVariables); err != nil {
				return err
			}
		case "same-file":
			// Same-file repeat
			if err := m.ProcessSameFileRepeat(templatePath, m.options.OutputDir, globalVariables, include.List); err != nil {
				return err
			}
		case "multiple-files":
			// Multiple-files repeat
			if err := m.ProcessMultipleFilesRepeat(templatePath, m.options.OutputDir, globalVariables, include.List); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown repeat type: %s", include.Repeat)
		}
	}
	
	fmt.Println("SUCCESS: Build completed successfully!")
	return nil
}

// validateDirectories validates that required directories exist
func (m *MikoManifest) validateDirectories() error {
	// Check templates directory
	if stat, err := os.Stat(m.options.TemplatesDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("templates directory %s not found", m.options.TemplatesDir)
		}
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", m.options.TemplatesDir)
	}
	
	// Check config directory
	if stat, err := os.Stat(m.options.ConfigDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config directory %s not found", m.options.ConfigDir)
		}
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", m.options.ConfigDir)
	}
	
	return nil
}
