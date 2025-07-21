package mikomanifest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

// LintOptions contains options for linting
type LintOptions struct {
	Directory string
}

// CheckOptions contains options for checking
type CheckOptions struct {
	ConfigDir string
}

// LintDirectory runs yamllint and kubernetes validation on a directory
func LintDirectory(options LintOptions) error {
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
	
	// Run yamllint
	yamllintSuccess := runYamllint(options.Directory)
	
	// Run Kubernetes validation
	k8sSuccess := validateKubernetesManifests(options.Directory)
	
	// Final result
	if yamllintSuccess && k8sSuccess {
		fmt.Println("🎉 All validations passed!")
	} else {
		errorParts := []string{}
		if !yamllintSuccess {
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

// CheckConfigDirectory runs yamllint only on config directory
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
	
	success := runYamllint(options.ConfigDir)
	
	if success {
		fmt.Println("SUCCESS: All YAML files passed linting!")
	} else {
		fmt.Println("FAILED: YAML linting failed!")
		return fmt.Errorf("yaml linting failed")
	}
	
	return nil
}

// runYamllint runs yamllint on a directory
func runYamllint(directory string) bool {
	// Check if yamllint is available
	if _, err := exec.LookPath("yamllint"); err != nil {
		fmt.Println("ERROR: yamllint is not installed.")
		fmt.Println("   Install it with: pip install yamllint")
		return false
	}
	
	// Build yamllint command
	args := []string{}
	
	// Check if yamllint.config exists
	if _, err := os.Stat("yamllint.config"); err == nil {
		args = append(args, "-c", "yamllint.config")
		fmt.Println("✓ Using config file: yamllint.config")
	} else {
		fmt.Println("ℹ No yamllint.config found, using default configuration")
	}
	
	// Add the directory to lint
	args = append(args, directory)
	
	fmt.Printf("Running: yamllint %s\n", strings.Join(args, " "))
	
	// Run yamllint
	cmd := exec.Command("yamllint", args...)
	output, err := cmd.CombinedOutput()
	
	if len(output) > 0 {
		fmt.Print(string(output))
	}
	
	return err == nil
}

// validateKubernetesManifests validates Kubernetes manifests in a directory
func validateKubernetesManifests(directory string) bool {
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
			
			// Basic validation - check if it's a valid Kubernetes resource
			if err := validateKubernetesResource(manifest); err != nil {
				k8sErrors++
				docInfo := ""
				if len(documents) > 1 {
					docInfo = fmt.Sprintf("[doc %d]", i+1)
				}
				fmt.Printf("ERROR: %s%s - Kubernetes validation error: %v\n", filepath.Base(yamlFile), docInfo, err)
			} else {
				k8sValidated++
				docInfo := ""
				if len(documents) > 1 {
					docInfo = fmt.Sprintf("[doc %d]", i+1)
				}
				fmt.Printf("VALID: %s%s - Valid %s manifest\n", filepath.Base(yamlFile), docInfo, kind)
			}
		}
	}
	
	if k8sValidated > 0 {
		fmt.Printf("SUCCESS: Kubernetes validation: %d manifest(s) validated successfully\n", k8sValidated)
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
