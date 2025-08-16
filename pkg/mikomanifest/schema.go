package mikomanifest

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemaConfig represents the schema configuration
type SchemaConfig struct {
	Schemas []string `yaml:"schemas"`
}

// SchemaRegistry holds registered CRDs for validation
type SchemaRegistry struct {
	crds map[schema.GroupVersionKind]*v1.CustomResourceDefinition
}

// NewSchemaRegistry creates a new schema registry
func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		crds: make(map[schema.GroupVersionKind]*v1.CustomResourceDefinition),
	}
}

// LoadSchemaConfig loads schema configuration from a file
func LoadSchemaConfig(configPath string) (*SchemaConfig, error) {
	if configPath == "" {
		return &SchemaConfig{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &SchemaConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read schema config file %s: %w", configPath, err)
	}

	var config SchemaConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse schema config file %s: %w", configPath, err)
	}

	return &config, nil
}

// LoadSchemas loads CRDs from all sources specified in the configuration
func (sr *SchemaRegistry) LoadSchemas(config *SchemaConfig) error {
	if config == nil || len(config.Schemas) == 0 {
		return nil
	}

	fmt.Printf("Loading custom schemas from %d source(s)...\n", len(config.Schemas))
	loadedCount := 0

	for _, source := range config.Schemas {
		count, err := sr.loadFromSource(source)
		if err != nil {
			fmt.Printf("WARNING: Failed to load schemas from %s: %v\n", source, err)
			continue
		}
		loadedCount += count
	}

	fmt.Printf("✓ Loaded %d custom resource definition(s)\n", loadedCount)
	return nil
}

// loadFromSource loads CRDs from a single source (URL, file, or directory)
func (sr *SchemaRegistry) loadFromSource(source string) (int, error) {
	// Detect source type
	if isURL(source) {
		return sr.loadFromURL(source)
	}

	// Check if it's a file or directory
	info, err := os.Stat(source)
	if err != nil {
		return 0, fmt.Errorf("source not accessible: %w", err)
	}

	if info.IsDir() {
		return sr.loadFromDirectory(source)
	}

	return sr.loadFromFile(source)
}

// isURL checks if a string is a URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// loadFromURL downloads and loads CRDs from a URL
func (sr *SchemaRegistry) loadFromURL(url string) (int, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP %d when downloading from %s", resp.StatusCode, url)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response from %s: %w", url, err)
	}

	return sr.loadFromContent(string(content), url)
}

// loadFromFile loads CRDs from a single file
func (sr *SchemaRegistry) loadFromFile(filePath string) (int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return sr.loadFromContent(string(content), filePath)
}

// loadFromDirectory recursively loads CRDs from all YAML files in a directory
func (sr *SchemaRegistry) loadFromDirectory(dirPath string) (int, error) {
	totalCount := 0

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only process YAML files
		if !strings.HasSuffix(strings.ToLower(path), ".yaml") && !strings.HasSuffix(strings.ToLower(path), ".yml") {
			return nil
		}

		count, err := sr.loadFromFile(path)
		if err != nil {
			fmt.Printf("WARNING: Failed to load CRD from %s: %v\n", path, err)
			return nil // Continue processing other files
		}

		totalCount += count
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to walk directory %s: %w", dirPath, err)
	}

	return totalCount, nil
}

// loadFromContent parses YAML content and extracts CRDs
func (sr *SchemaRegistry) loadFromContent(content, source string) (int, error) {
	documents := strings.Split(content, "\n---\n")
	loadedCount := 0

	for i, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		crd, err := sr.parseCRD(doc)
		if err != nil {
			// Not all documents need to be CRDs, so we'll just skip non-CRD documents
			continue
		}

		if crd != nil {
			sr.registerCRD(crd)
			loadedCount++
			
			docInfo := ""
			if len(documents) > 1 {
				docInfo = fmt.Sprintf(" [doc %d]", i+1)
			}
			fmt.Printf("✓ Registered CRD: %s/%s%s from %s\n", 
				crd.Spec.Group, crd.Spec.Names.Kind, docInfo, source)
		}
	}

	return loadedCount, nil
}

// parseCRD attempts to parse a YAML document as a CRD
func (sr *SchemaRegistry) parseCRD(content string) (*v1.CustomResourceDefinition, error) {
	var crd v1.CustomResourceDefinition

	// First, check if it's a CRD by looking at the kind
	var metadata struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
	}

	if err := yaml.Unmarshal([]byte(content), &metadata); err != nil {
		return nil, err
	}

	// Only process CustomResourceDefinition objects
	if metadata.Kind != "CustomResourceDefinition" {
		return nil, fmt.Errorf("not a CustomResourceDefinition")
	}

	// Parse as CRD
	if err := yaml.Unmarshal([]byte(content), &crd); err != nil {
		return nil, fmt.Errorf("failed to parse CRD: %w", err)
	}

	return &crd, nil
}

// registerCRD registers a CRD in the schema registry
func (sr *SchemaRegistry) registerCRD(crd *v1.CustomResourceDefinition) {
	for _, version := range crd.Spec.Versions {
		gvk := schema.GroupVersionKind{
			Group:   crd.Spec.Group,
			Version: version.Name,
			Kind:    crd.Spec.Names.Kind,
		}
		sr.crds[gvk] = crd
	}
}

// ValidateCustomResource validates a manifest against registered CRDs
func (sr *SchemaRegistry) ValidateCustomResource(manifest map[string]interface{}) (bool, error) {
	// Extract GVK from manifest
	apiVersionStr, ok := manifest["apiVersion"].(string)
	if !ok {
		return false, fmt.Errorf("apiVersion must be a string")
	}

	kind, ok := manifest["kind"].(string)
	if !ok {
		return false, fmt.Errorf("kind must be a string")
	}

	// Parse the apiVersion to get group and version
	gv, err := schema.ParseGroupVersion(apiVersionStr)
	if err != nil {
		return false, fmt.Errorf("invalid apiVersion: %w", err)
	}

	gvk := schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    kind,
	}

	// Check if we have a CRD for this GVK
	crd, exists := sr.crds[gvk]
	if !exists {
		return false, nil // Not a custom resource we know about
	}

	// Basic validation - check required fields
	if err := sr.validateBasicCRStructure(manifest, crd); err != nil {
		return true, err
	}

	return true, nil
}

// validateBasicCRStructure performs basic validation of a custom resource
func (sr *SchemaRegistry) validateBasicCRStructure(manifest map[string]interface{}, crd *v1.CustomResourceDefinition) error {
	// Check for required fields
	if _, ok := manifest["metadata"]; !ok {
		return fmt.Errorf("metadata field is required")
	}

	metadata, ok := manifest["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("metadata must be an object")
	}

	// Check for name in metadata
	if _, ok := metadata["name"]; !ok {
		return fmt.Errorf("metadata.name is required")
	}

	// Check that name is a string
	if _, ok := metadata["name"].(string); !ok {
		return fmt.Errorf("metadata.name must be a string")
	}

	// Check if spec is required (most CRDs require it)
	if _, ok := manifest["spec"]; !ok {
		// Some CRDs might not require spec, so this is just a warning
		fmt.Printf("WARNING: No spec field found in %s\n", crd.Spec.Names.Kind)
	}

	return nil
}

// GetRegisteredCRDs returns information about registered CRDs
func (sr *SchemaRegistry) GetRegisteredCRDs() map[schema.GroupVersionKind]*v1.CustomResourceDefinition {
	return sr.crds
}
