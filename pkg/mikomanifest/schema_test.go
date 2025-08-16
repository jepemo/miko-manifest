package mikomanifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSchemaConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Test with valid config file
	validConfig := `schemas:
  - https://example.com/crd.yaml
  - ./local/crd.yaml
  - ./schemas/`
	
	configPath := filepath.Join(tempDir, "schemas.yaml")
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	config, err := LoadSchemaConfig(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if len(config.Schemas) != 3 {
		t.Errorf("Expected 3 schemas, got %d", len(config.Schemas))
	}
	
	expectedSchemas := []string{
		"https://example.com/crd.yaml",
		"./local/crd.yaml",
		"./schemas/",
	}
	
	for i, expected := range expectedSchemas {
		if config.Schemas[i] != expected {
			t.Errorf("Expected schema %d to be '%s', got '%s'", i, expected, config.Schemas[i])
		}
	}
}

func TestLoadSchemaConfigEmpty(t *testing.T) {
	// Test with empty config path
	config, err := LoadSchemaConfig("")
	if err != nil {
		t.Fatalf("Expected no error for empty config path, got: %v", err)
	}
	
	if len(config.Schemas) != 0 {
		t.Errorf("Expected 0 schemas for empty config, got %d", len(config.Schemas))
	}
}

func TestLoadSchemaConfigNonExistent(t *testing.T) {
	// Test with non-existent file
	config, err := LoadSchemaConfig("/non/existent/file.yaml")
	if err != nil {
		t.Fatalf("Expected no error for non-existent file, got: %v", err)
	}
	
	if len(config.Schemas) != 0 {
		t.Errorf("Expected 0 schemas for non-existent file, got %d", len(config.Schemas))
	}
}

func TestSchemaRegistry(t *testing.T) {
	registry := NewSchemaRegistry()
	
	if registry == nil {
		t.Fatal("Expected registry to be created")
	}
	
	if len(registry.crds) != 0 {
		t.Errorf("Expected empty registry, got %d CRDs", len(registry.crds))
	}
}

func TestSchemaRegistryLoadFromContent(t *testing.T) {
	registry := NewSchemaRegistry()
	
	// Valid CRD content
	crdContent := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: databases.example.com
spec:
  group: example.com
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
  scope: Namespaced
  names:
    plural: databases
    singular: database
    kind: Database`
	
	count, err := registry.loadFromContent(crdContent, "test-source")
	if err != nil {
		t.Fatalf("Expected no error loading CRD, got: %v", err)
	}
	
	if count != 1 {
		t.Errorf("Expected 1 CRD loaded, got %d", count)
	}
	
	// Check if CRD was registered
	crds := registry.GetRegisteredCRDs()
	if len(crds) != 1 {
		t.Errorf("Expected 1 registered CRD, got %d", len(crds))
	}
}

func TestSchemaRegistryValidateCustomResource(t *testing.T) {
	registry := NewSchemaRegistry()
	
	// Load a CRD first
	crdContent := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: databases.example.com
spec:
  group: example.com
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
  scope: Namespaced
  names:
    plural: databases
    singular: database
    kind: Database`
	
	_, err := registry.loadFromContent(crdContent, "test-source")
	if err != nil {
		t.Fatalf("Failed to load CRD: %v", err)
	}
	
	// Test validation of a custom resource
	manifest := map[string]interface{}{
		"apiVersion": "example.com/v1",
		"kind":       "Database",
		"metadata": map[string]interface{}{
			"name": "test-db",
		},
		"spec": map[string]interface{}{
			"host": "localhost",
		},
	}
	
	isCustomResource, err := registry.ValidateCustomResource(manifest)
	if !isCustomResource {
		t.Error("Expected manifest to be recognized as custom resource")
	}
	
	if err != nil {
		t.Errorf("Expected no validation error, got: %v", err)
	}
}

func TestSchemaRegistryValidateNonCustomResource(t *testing.T) {
	registry := NewSchemaRegistry()
	
	// Test with a standard Kubernetes resource
	manifest := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name": "test-pod",
		},
	}
	
	isCustomResource, err := registry.ValidateCustomResource(manifest)
	if isCustomResource {
		t.Error("Expected manifest not to be recognized as custom resource")
	}
	
	if err != nil {
		t.Errorf("Expected no error for non-custom resource, got: %v", err)
	}
}
