package mikomanifest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jepemo/miko-manifest/pkg/output"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	k8syaml "sigs.k8s.io/yaml"
)

// cleanValidationError extracts the useful part from verbose JSON unmarshaling errors
func cleanValidationError(err error) string {
	errMsg := err.Error()

	// Extract just the unknown field from verbose JSON unmarshaling errors
	// From: "error unmarshaling JSON: while decoding JSON: json: unknown field 'repicas'"
	// To: "unknown field 'repicas'"
	if strings.Contains(errMsg, "unknown field") {
		// Find the part after "json: "
		if idx := strings.LastIndex(errMsg, "json: "); idx != -1 {
			return strings.TrimSpace(errMsg[idx+5:]) // Skip "json: " and trim spaces
		}
	}

	// Extract other common validation errors more cleanly
	if strings.Contains(errMsg, "cannot unmarshal") {
		// From: "json: cannot unmarshal string into Go struct field DeploymentSpec.spec.replicas of type int32"
		// To: "invalid type for field 'replicas' (expected number, got string)"
		if strings.Contains(errMsg, "Go struct field") && strings.Contains(errMsg, "of type") {
			// Extract field name from "Go struct field Something.field"
			parts := strings.Split(errMsg, "Go struct field ")
			if len(parts) > 1 {
				fieldPart := strings.Split(parts[1], " of type")
				if len(fieldPart) > 1 {
					fieldName := fieldPart[0]
					if dotIdx := strings.LastIndex(fieldName, "."); dotIdx != -1 {
						fieldName = fieldName[dotIdx+1:]
					}

					// Determine expected type
					if strings.Contains(errMsg, "of type int") {
						return fmt.Sprintf("invalid type for field '%s' (expected number, got text)", fieldName)
					} else if strings.Contains(errMsg, "of type bool") {
						return fmt.Sprintf("invalid type for field '%s' (expected boolean, got text)", fieldName)
					}
				}
			}
		}
	}

	// For other errors, return as-is but try to make them cleaner
	return strings.TrimSpace(errMsg)
}

// LintOptions contains options for linting
type LintOptions struct {
	Directory            string
	Environment          string
	ConfigDir            string
	SkipSchemaValidation bool
	OutputOpts           *output.OutputOptions
}

// CheckOptions contains options for checking
type CheckOptions struct {
	ConfigDir  string
	OutputOpts *output.OutputOptions
}

// LintDirectory runs native Go YAML linting and kubernetes validation on a directory
func LintDirectory(options LintOptions) error {
	// Create a default output options if not provided
	var outputOpts *output.OutputOptions
	if options.OutputOpts != nil {
		outputOpts = options.OutputOpts
	} else {
		outputOpts = &output.OutputOptions{Verbose: false}
	}

	// Auto-detect environment if not provided
	if options.Environment == "" {
		if env, configDir, err := loadEnvironmentInfo(options.Directory); err == nil {
			options.Environment = env
			options.ConfigDir = configDir
			outputOpts.PrintInfo(fmt.Sprintf("Auto-detected environment: %s", env))
		}
	}

	outputOpts.PrintStep(fmt.Sprintf("Linting YAML files in directory: %s", options.Directory))

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
				outputOpts.PrintWarning("Schema loading", fmt.Sprintf("Failed to load schemas from environment config: %v", err))
			} else if schemaRegistry != nil {
				outputOpts.PrintInfo(fmt.Sprintf("Loaded schemas from environment: %s", options.Environment))
			}
		}
	}

	// Run YAML linting
	yamlLintSuccess := lintYAMLFilesWithOutput(options.Directory, outputOpts)

	// Run Kubernetes validation
	k8sSuccess := validateKubernetesManifestsWithOutput(options.Directory, schemaRegistry, outputOpts)

	// Final comprehensive summary
	if yamlLintSuccess && k8sSuccess {
		// Get file count for final summary
		yamlFiles, _ := filepath.Glob(filepath.Join(options.Directory, "*.yaml"))
		ymlFiles, _ := filepath.Glob(filepath.Join(options.Directory, "*.yml"))
		totalFiles := len(yamlFiles) + len(ymlFiles)

		outputOpts.PrintSummary(fmt.Sprintf("All validations passed - %d file(s) validated successfully", totalFiles))
	} else {
		errorParts := []string{}
		if !yamlLintSuccess {
			errorParts = append(errorParts, "YAML linting")
		}
		if !k8sSuccess {
			errorParts = append(errorParts, "Kubernetes validation")
		}

		outputOpts.PrintError("Validation", fmt.Sprintf("FAILED: %s", strings.Join(errorParts, " and ")))
		return fmt.Errorf("validation failed")
	}

	return nil
}

// CheckConfigDirectory runs native Go YAML linting only on config directory
func CheckConfigDirectory(options CheckOptions) error {
	// Create a default output options if not provided
	var outputOpts *output.OutputOptions
	if options.OutputOpts != nil {
		outputOpts = options.OutputOpts
	} else {
		outputOpts = &output.OutputOptions{Verbose: false}
	}

	outputOpts.PrintInfo(fmt.Sprintf("Using config directory: %s", options.ConfigDir))
	outputOpts.PrintStep(fmt.Sprintf("Checking YAML files in directory: %s", options.ConfigDir))

	// Check if directory exists
	if stat, err := os.Stat(options.ConfigDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory %s not found", options.ConfigDir)
		}
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", options.ConfigDir)
	}

	success := lintYAMLFilesWithOutput(options.ConfigDir, outputOpts)

	if success {
		outputOpts.PrintSummary("All configuration files validated successfully")
	} else {
		outputOpts.PrintSummary("Configuration validation failed")
		return fmt.Errorf("yaml configuration validation failed")
	}

	return nil
}

// lintYAMLFilesWithOutput lints YAML files using the new output system
func lintYAMLFilesWithOutput(directory string, outputOpts *output.OutputOptions) bool {
	outputOpts.PrintStep(fmt.Sprintf("Linting YAML files in %s using native Go YAML parser", directory))

	// Check if directory exists first
	if stat, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			outputOpts.PrintError(directory, "Directory not found")
			return false
		}
		outputOpts.PrintError(directory, fmt.Sprintf("Error accessing directory: %v", err))
		return false
	} else if !stat.IsDir() {
		outputOpts.PrintError(directory, "Not a directory")
		return false
	}

	// Find YAML files
	yamlFiles, err := filepath.Glob(filepath.Join(directory, "*.yaml"))
	if err != nil {
		outputOpts.PrintError(directory, fmt.Sprintf("Error finding YAML files: %v", err))
		return false
	}

	ymlFiles, err := filepath.Glob(filepath.Join(directory, "*.yml"))
	if err != nil {
		outputOpts.PrintError(directory, fmt.Sprintf("Error finding YML files: %v", err))
		return false
	}

	allFiles := append(yamlFiles, ymlFiles...)

	if len(allFiles) == 0 {
		outputOpts.PrintInfo(fmt.Sprintf("No YAML files found in %s", directory))
		return true
	}

	yamlErrors := 0
	yamlValidated := 0

	for _, yamlFile := range allFiles {
		if lintSingleYAMLFileWithOutput(yamlFile, outputOpts) {
			yamlValidated++
		} else {
			yamlErrors++
		}
	}

	// Print result using the new output system
	if yamlErrors > 0 {
		outputOpts.PrintResult(fmt.Sprintf("YAML syntax validation - %d file(s) validated successfully, %d error(s)", yamlValidated, yamlErrors))
		return false
	} else {
		outputOpts.PrintResult(fmt.Sprintf("YAML syntax validation - %d file(s) validated successfully, %d error(s)", yamlValidated, yamlErrors))
		return true
	}
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
// lintSingleYAMLFileWithOutput lints a single YAML file using the new output system
func validateYAMLStructureWithOutput(parsed interface{}, fileName string, docIndex int, isMultiDoc bool, outputOpts *output.OutputOptions) bool {
	docInfo := ""
	if isMultiDoc {
		docInfo = fmt.Sprintf(" [doc %d]", docIndex)
	}

	switch v := parsed.(type) {
	case map[string]interface{}:
		// Check for empty maps
		if len(v) == 0 {
			outputOpts.PrintWarning(fileName+docInfo, "Empty YAML document")
			return true
		}

		// Check for common YAML issues like duplicate keys (already handled by yaml.v3)
		// Check for null values in critical fields
		for key, value := range v {
			if value == nil {
				outputOpts.PrintWarning(fileName+docInfo, fmt.Sprintf("Null value for key '%s'", key))
			}
		}

	case []interface{}:
		// Check for empty arrays
		if len(v) == 0 {
			outputOpts.PrintWarning(fileName+docInfo, "Empty YAML array")
			return true
		}

	case nil:
		outputOpts.PrintWarning(fileName+docInfo, "Null YAML document")
		return true

	default:
		// Scalar values are generally fine
	}

	return true
}

func lintSingleYAMLFileWithOutput(fileName string, outputOpts *output.OutputOptions) bool {
	content, err := os.ReadFile(fileName)
	if err != nil {
		outputOpts.PrintError(fileName, fmt.Sprintf("Error reading file: %v", err))
		return false
	}

	decoder := yaml.NewDecoder(strings.NewReader(string(content)))
	valid := true
	docIndex := 0
	documentCount := 0

	// First pass: count documents
	tempDecoder := yaml.NewDecoder(strings.NewReader(string(content)))
	for {
		var temp interface{}
		if err := tempDecoder.Decode(&temp); err != nil {
			if err == io.EOF {
				break
			}
			break
		}
		documentCount++
	}

	isMultiDoc := documentCount > 1

	// Second pass: validate documents
	for {
		var parsed interface{}
		if err := decoder.Decode(&parsed); err != nil {
			if err == io.EOF {
				break
			}
			outputOpts.PrintError(fileName, fmt.Sprintf("Error parsing YAML: %v", err))
			valid = false
			break
		}

		if !validateYAMLStructureWithOutput(parsed, fileName, docIndex, isMultiDoc, outputOpts) {
			valid = false
		}
		docIndex++
	}

	return valid
}

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
// validateKubernetesManifestsWithOutput validates Kubernetes manifests using the output system
func validateKubernetesManifestsWithOutput(directory string, schemaRegistry *SchemaRegistry, outputOpts *output.OutputOptions) bool {
	outputOpts.PrintStep(fmt.Sprintf("Validating Kubernetes manifests in %s", directory))

	// Find YAML files
	yamlFiles, err := filepath.Glob(filepath.Join(directory, "*.yaml"))
	if err != nil {
		outputOpts.PrintError("File search", fmt.Sprintf("Error finding YAML files: %v", err))
		return false
	}

	ymlFiles, err := filepath.Glob(filepath.Join(directory, "*.yml"))
	if err != nil {
		outputOpts.PrintError("File search", fmt.Sprintf("Error finding YML files: %v", err))
		return false
	}

	allFiles := append(yamlFiles, ymlFiles...)

	if len(allFiles) == 0 {
		outputOpts.PrintInfo(fmt.Sprintf("No YAML files found in %s for Kubernetes validation", directory))
		return true
	}

	k8sErrors := 0
	k8sValidated := 0
	customResourcesValidated := 0

	for _, yamlFile := range allFiles {
		content, err := os.ReadFile(yamlFile)
		if err != nil {
			outputOpts.PrintError(filepath.Base(yamlFile), fmt.Sprintf("Error reading file: %v", err))
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
				outputOpts.PrintError(filepath.Base(yamlFile), fmt.Sprintf("YAML parsing error: %v", err))
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
				outputOpts.PrintInfo(fmt.Sprintf("%s - Not a Kubernetes manifest (missing apiVersion/kind)", filepath.Base(yamlFile)))
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
						outputOpts.PrintError(filepath.Base(yamlFile)+docInfo, fmt.Sprintf("Custom resource validation error: %v", err))
					} else {
						customResourcesValidated++
						k8sValidated++
						outputOpts.PrintValid(filepath.Base(yamlFile)+docInfo, fmt.Sprintf("Valid custom resource %s", kind))
					}
					continue
				}
			}

			// Basic validation - check if it's a valid Kubernetes resource
			if err := validateKubernetesResource(manifest); err != nil {
				k8sErrors++
				outputOpts.PrintError(filepath.Base(yamlFile)+docInfo, fmt.Sprintf("Kubernetes validation error: %v", err))
			} else {
				k8sValidated++
				outputOpts.PrintValid(filepath.Base(yamlFile)+docInfo, fmt.Sprintf("Valid %s manifest", kind))
			}
		}
	}

	if k8sValidated > 0 {
		summary := fmt.Sprintf("Kubernetes schema validation - %d manifest(s) validated successfully", k8sValidated)
		if customResourcesValidated > 0 {
			summary += fmt.Sprintf(" (%d custom resource(s))", customResourcesValidated)
		}
		outputOpts.PrintResult(summary)
	}

	return k8sErrors == 0
}

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

// validateKubernetesResource performs comprehensive validation using native Kubernetes types
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

	// Create the GroupVersionKind
	gvk := schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    kind,
	}

	// Convert manifest back to YAML for validation
	yamlData, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest to YAML: %w", err)
	}

	// Use strict validation with native Kubernetes types
	return validateWithKubernetesTypes(yamlData, gvk)
}

// validateWithKubernetesTypes validates YAML data against native Kubernetes types with strict decoding
func validateWithKubernetesTypes(yamlData []byte, gvk schema.GroupVersionKind) error {
	// Use specific validation for known types to avoid false positives
	switch gvk.Kind {
	case "Deployment":
		if gvk.Group == "apps" && gvk.Version == "v1" {
			return validateDeploymentWithNativeTypes(yamlData)
		}
	case "Service":
		if gvk.Group == "" && gvk.Version == "v1" {
			return validateServiceWithNativeTypes(yamlData)
		}
	case "ConfigMap":
		if gvk.Group == "" && gvk.Version == "v1" {
			return validateConfigMapWithNativeTypes(yamlData)
		}
	case "Secret":
		if gvk.Group == "" && gvk.Version == "v1" {
			return validateSecretWithNativeTypes(yamlData)
		}
	case "Pod":
		if gvk.Group == "" && gvk.Version == "v1" {
			return validatePodWithNativeTypes(yamlData)
		}
	}

	// For other types, try to create using the scheme and do basic validation
	_, err := scheme.Scheme.New(gvk)
	if err != nil {
		// If not found in scheme, it might be a custom resource
		return validateBasicStructureFromYAML(yamlData)
	}

	// For known types in scheme but not specifically handled, do basic validation
	return validateBasicStructureFromYAML(yamlData)
}

// validateDeploymentWithNativeTypes validates a Deployment using the complete appsv1.Deployment type
func validateDeploymentWithNativeTypes(yamlData []byte) error {
	var deployment appsv1.Deployment
	if err := k8syaml.UnmarshalStrict(yamlData, &deployment); err != nil {
		return fmt.Errorf("invalid Deployment: %s", cleanValidationError(err))
	}

	return nil
}

// validateServiceWithNativeTypes validates a Service using the complete corev1.Service type
func validateServiceWithNativeTypes(yamlData []byte) error {
	var service corev1.Service
	if err := k8syaml.UnmarshalStrict(yamlData, &service); err != nil {
		return fmt.Errorf("invalid Service: %s", cleanValidationError(err))
	}

	return nil
}

// validateConfigMapWithNativeTypes validates a ConfigMap using the complete corev1.ConfigMap type
func validateConfigMapWithNativeTypes(yamlData []byte) error {
	var configMap corev1.ConfigMap
	if err := k8syaml.UnmarshalStrict(yamlData, &configMap); err != nil {
		return fmt.Errorf("invalid ConfigMap: %s", cleanValidationError(err))
	}

	return nil
}

// validateSecretWithNativeTypes validates a Secret using the complete corev1.Secret type
func validateSecretWithNativeTypes(yamlData []byte) error {
	var secret corev1.Secret
	if err := k8syaml.UnmarshalStrict(yamlData, &secret); err != nil {
		return fmt.Errorf("invalid Secret: %s", cleanValidationError(err))
	}

	return nil
}

// validatePodWithNativeTypes validates a Pod using the complete corev1.Pod type
func validatePodWithNativeTypes(yamlData []byte) error {
	var pod corev1.Pod
	if err := k8syaml.UnmarshalStrict(yamlData, &pod); err != nil {
		return fmt.Errorf("invalid Pod: %s", cleanValidationError(err))
	}

	return nil
}

// validateBasicStructureFromYAML validates basic structure for unknown resource types
func validateBasicStructureFromYAML(yamlData []byte) error {
	var manifest map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &manifest); err != nil {
		return fmt.Errorf("invalid YAML structure: %w", err)
	}

	return validateBasicStructure(manifest)
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
