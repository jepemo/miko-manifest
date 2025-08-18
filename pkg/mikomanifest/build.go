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
	Environment string     `yaml:"-"`           // Not serialized, set programmatically
	ConfigDir   string     `yaml:"-"`           // Not serialized, set programmatically
	Resources   []string   `yaml:"resources,omitempty"`
	Schemas     []string   `yaml:"schemas,omitempty"`
	Variables   []Variable `yaml:"variables"`
	Include     []Include  `yaml:"include"`
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
	Environment     string
	OutputDir       string
	ConfigDir       string
	TemplatesDir    string
	Variables       map[string]string
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

// LoadConfig loads configuration from ENV.yaml file with hierarchical resource support
// This is a public function that can be used from external packages
func LoadConfig(configDir, env string, showTree bool) (*Config, error) {
	options := BuildOptions{
		ConfigDir: configDir,
	}
	m := New(options)
	config, err := m.loadConfigInternal(env, showTree)
	if err != nil {
		return nil, err
	}
	config.Environment = env
	config.ConfigDir = configDir
	return config, nil
}

// loadConfigInternal loads configuration from ENV.yaml file with hierarchical resource support
func (m *MikoManifest) LoadConfig(env string) (*Config, error) {
	return m.loadConfigInternal(env, false)
}

// loadConfigInternal is the internal implementation
func (m *MikoManifest) loadConfigInternal(env string, showTree bool) (*Config, error) {
	configPath := filepath.Join(m.options.ConfigDir, fmt.Sprintf("%s.yaml", env))
	return m.LoadConfigWithResources(configPath, make([]string, 0), 0, showTree)
}

// LoadConfigWithResources loads configuration with resource inclusion and circular dependency detection
func (m *MikoManifest) LoadConfigWithResources(configPath string, loadChain []string, depth int, showTree bool) (*Config, error) {
	// Check for maximum recursion depth
	const maxDepth = 5
	if depth > maxDepth {
		return nil, fmt.Errorf("maximum recursion depth (%d) exceeded when loading %s", maxDepth, configPath)
	}
	
	// Check for circular dependencies
	for _, loaded := range loadChain {
		if loaded == configPath {
			return nil, fmt.Errorf("circular dependency detected: %s is already in load chain %v", configPath, loadChain)
		}
	}
	
	// Add current config to the load chain
	currentChain := append(loadChain, configPath)
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file %s not found", configPath)
	}
	
	if showTree {
		indent := strings.Repeat("  ", depth)
		fmt.Printf("%s📄 Loading: %s\n", indent, configPath)
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", configPath, err)
	}
	
	// Process resources if they exist
	if len(config.Resources) > 0 {
		if showTree {
			indent := strings.Repeat("  ", depth)
			fmt.Printf("%s📁 Processing %d resource(s):\n", indent, len(config.Resources))
		}
		
		baseConfig := &Config{
			Variables: []Variable{},
			Include:   []Include{},
		}
		
		// Load and merge all resources
		for _, resource := range config.Resources {
			resourcePath := m.resolveResourcePath(configPath, resource)
			
			if showTree {
				indent := strings.Repeat("  ", depth+1)
				if isDirectory(resourcePath) {
					fmt.Printf("%s📂 %s (directory)\n", indent, resource)
				} else {
					fmt.Printf("%s📄 %s\n", indent, resource)
				}
			}
			
			if isDirectory(resourcePath) {
				// Load all YAML files from directory in alphabetical order
				if err := m.loadConfigFromDirectory(resourcePath, baseConfig, currentChain, depth+1, showTree); err != nil {
					return nil, fmt.Errorf("failed to load from directory %s: %w", resourcePath, err)
				}
			} else {
				// Load single file
				resourceConfig, err := m.LoadConfigWithResources(resourcePath, currentChain, depth+1, showTree)
				if err != nil {
					return nil, fmt.Errorf("failed to load resource %s: %w", resourcePath, err)
				}
				*baseConfig = *m.mergeConfigs(baseConfig, resourceConfig)
			}
		}
		
		// Merge current config with the base config (current config has higher priority)
		config = *m.mergeConfigs(baseConfig, &config)
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
	
	// Save environment info for auto-detection during lint
	if err := m.saveEnvironmentInfo(); err != nil {
		fmt.Printf("Warning: Failed to save environment info: %v\n", err)
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

// resolveResourcePath resolves a resource path relative to the config file location
func (m *MikoManifest) resolveResourcePath(configPath, resourcePath string) string {
	if filepath.IsAbs(resourcePath) {
		return resourcePath
	}
	
	configDir := filepath.Dir(configPath)
	return filepath.Join(configDir, resourcePath)
}

// isDirectory checks if the given path is a directory
func isDirectory(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

// loadConfigFromDirectory loads all YAML files from a directory and merges them
func (m *MikoManifest) loadConfigFromDirectory(dirPath string, baseConfig *Config, loadChain []string, depth int, showTree bool) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}
	
	// Filter and sort YAML files
	var yamlFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			yamlFiles = append(yamlFiles, filepath.Join(dirPath, name))
		}
	}
	
	// Process files in alphabetical order (already sorted by ReadDir)
	for _, yamlFile := range yamlFiles {
		resourceConfig, err := m.LoadConfigWithResources(yamlFile, loadChain, depth, showTree)
		if err != nil {
			return fmt.Errorf("failed to load file %s: %w", yamlFile, err)
		}
		*baseConfig = *m.mergeConfigs(baseConfig, resourceConfig)
	}
	
	return nil
}

// mergeConfigs merges two configurations, with the override config taking precedence
func (m *MikoManifest) mergeConfigs(base, override *Config) *Config {
	result := &Config{
		Variables: make([]Variable, 0),
		Include:   make([]Include, 0),
		Schemas:   make([]string, 0),
	}
	
	// Create a map for easier variable merging
	variableMap := make(map[string]string)
	
	// Add base variables
	for _, v := range base.Variables {
		variableMap[v.Name] = v.Value
	}
	
	// Override with variables from override config
	for _, v := range override.Variables {
		variableMap[v.Name] = v.Value
	}
	
	// Convert back to slice
	for name, value := range variableMap {
		result.Variables = append(result.Variables, Variable{
			Name:  name,
			Value: value,
		})
	}
	
	// Merge schemas (no duplicates)
	schemaSet := make(map[string]bool)
	
	// Add base schemas
	for _, schema := range base.Schemas {
		if !schemaSet[schema] {
			result.Schemas = append(result.Schemas, schema)
			schemaSet[schema] = true
		}
	}
	
	// Add override schemas
	for _, schema := range override.Schemas {
		if !schemaSet[schema] {
			result.Schemas = append(result.Schemas, schema)
			schemaSet[schema] = true
		}
	}
	
	// Merge includes (no duplicates based on file+key combination)
	includeMap := make(map[string]Include)
	
	// Add base includes
	for _, inc := range base.Include {
		key := m.getIncludeKey(inc)
		includeMap[key] = inc
	}
	
	// Add override includes (may override base includes)
	for _, inc := range override.Include {
		key := m.getIncludeKey(inc)
		includeMap[key] = inc
	}
	
	// Convert back to slice
	for _, inc := range includeMap {
		result.Include = append(result.Include, inc)
	}
	
	return result
}

// getIncludeKey generates a unique key for an include based on file and repeat key
func (m *MikoManifest) getIncludeKey(inc Include) string {
	if inc.Repeat != "" && len(inc.List) > 0 {
		// For repeat includes, create a key based on file and first list key
		return fmt.Sprintf("%s:%s:%s", inc.File, inc.Repeat, inc.List[0].Key)
	}
	return inc.File
}

// saveEnvironmentInfo saves the current environment and config directory for auto-detection
func (m *MikoManifest) saveEnvironmentInfo() error {
	envInfoPath := filepath.Join(m.options.OutputDir, ".miko-manifest-env")
	envInfo := fmt.Sprintf("environment: %s\nconfig_dir: %s\n", m.options.Environment, m.options.ConfigDir)
	return os.WriteFile(envInfoPath, []byte(envInfo), 0644)
}

// loadEnvironmentInfo loads saved environment information
func loadEnvironmentInfo(outputDir string) (string, string, error) {
	envInfoPath := filepath.Join(outputDir, ".miko-manifest-env")
	data, err := os.ReadFile(envInfoPath)
	if err != nil {
		return "", "", err
	}
	
	lines := strings.Split(string(data), "\n")
	var env, configDir string
	
	for _, line := range lines {
		if strings.HasPrefix(line, "environment: ") {
			env = strings.TrimSpace(strings.TrimPrefix(line, "environment: "))
		} else if strings.HasPrefix(line, "config_dir: ") {
			configDir = strings.TrimSpace(strings.TrimPrefix(line, "config_dir: "))
		}
	}
	
	return env, configDir, nil
}
