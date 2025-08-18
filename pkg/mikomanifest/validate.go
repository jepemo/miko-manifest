package mikomanifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

// LintOptions contains options for linting
type LintOptions struct {
	Directory            string
	Environment          string
	ConfigDir            string
	SkipSchemaValidation bool
}

// CheckOptions contains options for checking
type CheckOptions struct {
	ConfigDir            string
	Environment          string
	SkipSchemaValidation bool
}

// LintDirectory runs native Go YAML linting and kubernetes validation on a directory
func LintDirectory(options LintOptions) error {
	// Auto-detect environment if not provided
	if options.Environment == "" {
		if env, configDir, err := loadEnvironmentInfo(options.Directory); err == nil {
			options.Environment = env
			options.ConfigDir = configDir
			fmt.Printf("🔍 Auto-detected environment: %s\n", env)
		}
	}
	
	fmt.Printf("Linting YAML files in directory: %s\n", options.Directory)
	
	// Check if directory exists
	if stat, err := os.Stat(options.Directory); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory %s not found", options.Directory)
		}
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", options.Directory)
	}

	// Load custom schemas if not skipped
	var schemaRegistry *SchemaRegistry
	if !options.SkipSchemaValidation {
		if options.Environment != "" {
			// Load schemas from environment configuration
			var err error
			schemaRegistry, err = loadSchemasFromEnvironment(options.Environment, options.ConfigDir)
			if err != nil {
				fmt.Printf("Warning: Failed to load schemas from environment config: %v\n", err)
			} else if schemaRegistry != nil {
				fmt.Printf("✓ Loaded schemas from environment: %s\n", options.Environment)
			}
		}
	}
	
	// Run YAML linting
	yamlLintSuccess := lintYAMLFiles(options.Directory)
	
	// Run Kubernetes validation
	k8sSuccess := validateKubernetesManifests(options.Directory, schemaRegistry)
	
	// Final result
	if yamlLintSuccess && k8sSuccess {
		fmt.Println("🎉 All validations passed!")
	} else {
		errorParts := []string{}
		if !yamlLintSuccess {
			errorParts = append(errorParts, "YAML linting")
		}
		if !k8sSuccess {
			errorParts = append(errorParts, "Kubernetes validation")
		}
		
		fmt.Printf("FAILED: %s\n", strings.Join(errorParts, " and "))
		return fmt.Errorf("validation failed")
	}
	
	return nil
}

// CheckConfigDirectory runs native Go YAML linting only on config directory
func CheckConfigDirectory(options CheckOptions) error {
	fmt.Printf("✓ Using config directory: %s\n", options.ConfigDir)
	fmt.Printf("Linting YAML files in directory: %s\n", options.ConfigDir)
	
	// Check if directory exists
	if stat, err := os.Stat(options.ConfigDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory %s not found", options.ConfigDir)
		}
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", options.ConfigDir)
	}

	// Load custom schemas if provided (for config validation)
	var schemaRegistry *SchemaRegistry
	if !options.SkipSchemaValidation {
		if options.Environment != "" {
			// Load schemas from environment configuration
			var err error
			schemaRegistry, err = loadSchemasFromEnvironment(options.Environment, options.ConfigDir)
			if err != nil {
				fmt.Printf("Warning: Failed to load schemas from environment config: %v\n", err)
			} else if schemaRegistry != nil {
				fmt.Printf("✓ Loaded schemas from environment: %s\n", options.Environment)
			}
		}
	}
	
	success := lintYAMLFiles(options.ConfigDir)
	
	if success {
		fmt.Println("SUCCESS: All YAML files passed linting!")
	} else {
		fmt.Println("FAILED: YAML linting failed!")
		return fmt.Errorf("yaml linting failed")
	}
	
	return nil
}

// lintYAMLFiles performs YAML linting using native Go libraries
func lintYAMLFiles(directory string) bool {
	fmt.Printf("Linting YAML files in %s using native Go YAML parser...\n", directory)
	
	// Check if directory exists first
	if stat, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("ERROR: Directory %s not found\n", directory)
			return false
		}
		fmt.Printf("ERROR: Error accessing directory %s: %v\n", directory, err)
		return false
	} else if !stat.IsDir() {
		fmt.Printf("ERROR: %s is not a directory\n", directory)
		return false
	}
	
	// Find YAML files
	yamlFiles, err := filepath.Glob(filepath.Join(directory, "*.yaml"))
	if err != nil {
		fmt.Printf("ERROR: Error finding YAML files: %v\n", err)
		return false
	}
	
	ymlFiles, err := filepath.Glob(filepath.Join(directory, "*.yml"))
	if err != nil {
		fmt.Printf("ERROR: Error finding YML files: %v\n", err)
		return false
	}
	
	allFiles := append(yamlFiles, ymlFiles...)
	
	if len(allFiles) == 0 {
		fmt.Printf("ℹ No YAML files found in %s\n", directory)
		return true
	}
	
	yamlErrors := 0
	yamlValidated := 0
	
	for _, yamlFile := range allFiles {
		if lintSingleYAMLFile(yamlFile) {
			yamlValidated++
		} else {
			yamlErrors++
		}
	}
	
	fmt.Printf("YAML Linting Results: %d file(s) validated successfully, %d error(s)\n", yamlValidated, yamlErrors)
	
	return yamlErrors == 0
}

// lintSingleYAMLFile lints a single YAML file using Go's yaml.v3 library
func lintSingleYAMLFile(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("ERROR: %s - Error reading file: %v\n", filepath.Base(filePath), err)
		return false
	}
	
	// Check if file is empty
	if len(strings.TrimSpace(string(content))) == 0 {
		fmt.Printf("WARNING: %s - Empty file\n", filepath.Base(filePath))
		return true
	}
	
	// Split by document separator to handle multi-document YAML files
	documents := strings.Split(string(content), "\n---\n")
	
	hasValidDocuments := false
	documentErrors := 0
	
	for i, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}
		
		// Try to parse the YAML document
		var parsed interface{}
		if err := yaml.Unmarshal([]byte(doc), &parsed); err != nil {
			docInfo := ""
			if len(documents) > 1 {
				docInfo = fmt.Sprintf(" [doc %d]", i+1)
			}
			fmt.Printf("ERROR: %s%s - YAML syntax error: %v\n", filepath.Base(filePath), docInfo, err)
			documentErrors++
		} else {
			hasValidDocuments = true
			// Additional validation: check for basic YAML best practices
			if validateYAMLStructure(parsed, filepath.Base(filePath), i+1, len(documents) > 1) {
				docInfo := ""
				if len(documents) > 1 {
					docInfo = fmt.Sprintf(" [doc %d]", i+1)
				}
				fmt.Printf("VALID: %s%s - YAML syntax is correct\n", filepath.Base(filePath), docInfo)
			} else {
				documentErrors++
			}
		}
	}
	
	if !hasValidDocuments && documentErrors == 0 {
		fmt.Printf("WARNING: %s - No valid YAML documents found\n", filepath.Base(filePath))
		return true
	}
	
	return documentErrors == 0
}

// validateYAMLStructure performs additional validation on parsed YAML structure
func validateYAMLStructure(parsed interface{}, fileName string, docIndex int, isMultiDoc bool) bool {
	docInfo := ""
	if isMultiDoc {
		docInfo = fmt.Sprintf(" [doc %d]", docIndex)
	}
	
	switch v := parsed.(type) {
	case map[string]interface{}:
		// Check for empty maps
		if len(v) == 0 {
			fmt.Printf("WARNING: %s%s - Empty YAML document\n", fileName, docInfo)
			return true
		}
		
		// Check for common YAML issues like duplicate keys (already handled by yaml.v3)
		// Check for null values in critical fields
		for key, value := range v {
			if value == nil {
				fmt.Printf("WARNING: %s%s - Null value for key '%s'\n", fileName, docInfo, key)
			}
		}
		
	case []interface{}:
		// Check for empty arrays
		if len(v) == 0 {
			fmt.Printf("WARNING: %s%s - Empty YAML array\n", fileName, docInfo)
			return true
		}
		
	case nil:
		fmt.Printf("WARNING: %s%s - Null YAML document\n", fileName, docInfo)
		return true
		
	default:
		// Scalar values are generally fine
	}
	
	return true
}

// validateKubernetesManifests validates Kubernetes manifests in a directory
func validateKubernetesManifests(directory string, schemaRegistry *SchemaRegistry) bool {
	fmt.Printf("Validating Kubernetes manifests in %s...\n", directory)
	
	// Find YAML files
	yamlFiles, err := filepath.Glob(filepath.Join(directory, "*.yaml"))
	if err != nil {
		fmt.Printf("ERROR: Error finding YAML files: %v\n", err)
		return false
	}
	
	ymlFiles, err := filepath.Glob(filepath.Join(directory, "*.yml"))
	if err != nil {
		fmt.Printf("ERROR: Error finding YML files: %v\n", err)
		return false
	}
	
	allFiles := append(yamlFiles, ymlFiles...)
	
	if len(allFiles) == 0 {
		fmt.Printf("ℹ No YAML files found in %s for Kubernetes validation\n", directory)
		return true
	}
	
	k8sErrors := 0
	k8sValidated := 0
	customResourcesValidated := 0
	
	for _, yamlFile := range allFiles {
		content, err := os.ReadFile(yamlFile)
		if err != nil {
			fmt.Printf("ERROR: %s - Error reading file: %v\n", filepath.Base(yamlFile), err)
			k8sErrors++
			continue
		}
		
		// Split by document separator
		documents := strings.Split(string(content), "\n---\n")
		
		for i, doc := range documents {
			doc = strings.TrimSpace(doc)
			if doc == "" {
				continue
			}
			
			var manifest map[string]interface{}
			if err := yaml.Unmarshal([]byte(doc), &manifest); err != nil {
				fmt.Printf("ERROR: %s - YAML parsing error: %v\n", filepath.Base(yamlFile), err)
				k8sErrors++
				continue
			}
			
			if manifest == nil {
				continue
			}
			
			// Check if it looks like a Kubernetes manifest
			_, hasAPIVersion := manifest["apiVersion"]
			kind, hasKind := manifest["kind"]
			
			if !hasAPIVersion || !hasKind {
				fmt.Printf("ℹ %s - Not a Kubernetes manifest (missing apiVersion/kind)\n", filepath.Base(yamlFile))
				continue
			}
			
			docInfo := ""
			if len(documents) > 1 {
				docInfo = fmt.Sprintf("[doc %d]", i+1)
			}
			
			// Try custom resource validation first if schema registry is available
			if schemaRegistry != nil {
				isCustomResource, err := schemaRegistry.ValidateCustomResource(manifest)
				if isCustomResource {
					if err != nil {
						k8sErrors++
						fmt.Printf("ERROR: %s%s - Custom resource validation error: %v\n", filepath.Base(yamlFile), docInfo, err)
					} else {
						customResourcesValidated++
						k8sValidated++
						fmt.Printf("VALID: %s%s - Valid custom resource %s\n", filepath.Base(yamlFile), docInfo, kind)
					}
					continue
				}
			}
			
			// Basic validation - check if it's a valid Kubernetes resource
			if err := validateKubernetesResource(manifest); err != nil {
				k8sErrors++
				fmt.Printf("ERROR: %s%s - Kubernetes validation error: %v\n", filepath.Base(yamlFile), docInfo, err)
			} else {
				k8sValidated++
				fmt.Printf("VALID: %s%s - Valid %s manifest\n", filepath.Base(yamlFile), docInfo, kind)
			}
		}
	}
	
	if k8sValidated > 0 {
		fmt.Printf("SUCCESS: Kubernetes validation: %d manifest(s) validated successfully", k8sValidated)
		if customResourcesValidated > 0 {
			fmt.Printf(" (%d custom resource(s))", customResourcesValidated)
		}
		fmt.Println()
	}
	
	return k8sErrors == 0
}

// validateKubernetesResource performs basic validation of a Kubernetes resource
func validateKubernetesResource(manifest map[string]interface{}) error {
	// Get apiVersion and kind
	apiVersionStr, ok := manifest["apiVersion"].(string)
	if !ok {
		return fmt.Errorf("apiVersion must be a string")
	}
	
	kind, ok := manifest["kind"].(string)
	if !ok {
		return fmt.Errorf("kind must be a string")
	}
	
	// Parse the apiVersion to get group and version
	gv, err := schema.ParseGroupVersion(apiVersionStr)
	if err != nil {
		return fmt.Errorf("invalid apiVersion: %w", err)
	}
	
	// Check if it's a known Kubernetes resource type
	gvk := schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    kind,
	}
	
	// Try to find the resource in the scheme
	_, err = scheme.Scheme.New(gvk)
	if err != nil {
		// If not found in scheme, it might be a custom resource or newer API
		// We'll just do basic structure validation
		return validateBasicStructure(manifest)
	}
	
	return nil
}

// validateBasicStructure validates basic Kubernetes resource structure
func validateBasicStructure(manifest map[string]interface{}) error {
	// Check for required fields
	if _, ok := manifest["metadata"]; !ok {
		return fmt.Errorf("metadata field is required")
	}
	
	metadata, ok := manifest["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("metadata must be an object")
	}
	
	// Check for name in metadata (most resources require it)
	if _, ok := metadata["name"]; !ok {
		return fmt.Errorf("metadata.name is required")
	}
	
	// Check that name is a string
	if _, ok := metadata["name"].(string); !ok {
		return fmt.Errorf("metadata.name must be a string")
	}
	
	return nil
}

// loadSchemasFromEnvironment loads schemas from the environment configuration
func loadSchemasFromEnvironment(environment, configDir string) (*SchemaRegistry, error) {
	if environment == "" || configDir == "" {
		return nil, fmt.Errorf("environment and config directory are required")
	}
	
	// Create a temporary MikoManifest to load the config
	tempOptions := BuildOptions{
		Environment: environment,
		ConfigDir:   configDir,
	}
	mikoManifest := New(tempOptions)
	
	// Load configuration
	config, err := mikoManifest.LoadConfig(environment)
	if err != nil {
		return nil, fmt.Errorf("failed to load environment config: %w", err)
	}
	
	// If no schemas defined, return nil
	if len(config.Schemas) == 0 {
		return nil, nil
	}
	
	// Create schema registry and load schemas
	schemaRegistry := NewSchemaRegistry()
	
	// Create a temporary SchemaConfig structure for compatibility
	schemaConfig := &SchemaConfig{
		Schemas: config.Schemas,
	}
	
	if err := schemaRegistry.LoadSchemas(schemaConfig); err != nil {
		return nil, fmt.Errorf("failed to load schemas from environment config: %w", err)
	}
	
	return schemaRegistry, nil
}
