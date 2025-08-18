package mikomanifest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestHierarchicalConfiguration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up directory structure
	configDir := filepath.Join(tempDir, "config")
	templatesDir := filepath.Join(tempDir, "templates")
	componentsDir := filepath.Join(configDir, "components")
	outputDir := filepath.Join(tempDir, "output")
	
	// Create directories
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}
	if err := os.MkdirAll(componentsDir, 0755); err != nil {
		t.Fatalf("Failed to create components dir: %v", err)
	}
	
	// Create base configuration
	baseConfig := `---
schemas:
  - ./schemas/base-crd.yaml

variables:
  - name: app_name
    value: test-app
  - name: version
    value: 1.0.0
  - name: environment
    value: base

include:
  - file: deployment.yaml`
	
	if err := os.WriteFile(filepath.Join(configDir, "base.yaml"), []byte(baseConfig), 0644); err != nil {
		t.Fatalf("Failed to write base.yaml: %v", err)
	}
	
	// Create component configuration
	componentConfig := `---
variables:
  - name: component_enabled
    value: "true"
  - name: component_port
    value: "9090"

include:
  - file: service.yaml`
	
	if err := os.WriteFile(filepath.Join(componentsDir, "monitoring.yaml"), []byte(componentConfig), 0644); err != nil {
		t.Fatalf("Failed to write monitoring.yaml: %v", err)
	}
	
	// Create dev configuration that includes base and components
	devConfig := `---
resources:
  - base.yaml
  - components/

schemas:
  - ./schemas/dev-specific-crd.yaml

variables:
  - name: version
    value: 2.0.0
  - name: environment
    value: development

include:
  - file: configmap.yaml`
	
	if err := os.WriteFile(filepath.Join(configDir, "dev.yaml"), []byte(devConfig), 0644); err != nil {
		t.Fatalf("Failed to write dev.yaml: %v", err)
	}
	
	// Create template files
	deploymentTemplate := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.app_name}}
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: {{.app_name}}
        image: {{.app_name}}:{{.version}}`
	
	if err := os.WriteFile(filepath.Join(templatesDir, "deployment.yaml"), []byte(deploymentTemplate), 0644); err != nil {
		t.Fatalf("Failed to write deployment.yaml template: %v", err)
	}
	
	serviceTemplate := `apiVersion: v1
kind: Service
metadata:
  name: {{.app_name}}-service
spec:
  ports:
  - port: {{.component_port}}`
	
	if err := os.WriteFile(filepath.Join(templatesDir, "service.yaml"), []byte(serviceTemplate), 0644); err != nil {
		t.Fatalf("Failed to write service.yaml template: %v", err)
	}
	
	configmapTemplate := `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.app_name}}-config
data:
  environment: {{.environment}}`
	
	if err := os.WriteFile(filepath.Join(templatesDir, "configmap.yaml"), []byte(configmapTemplate), 0644); err != nil {
		t.Fatalf("Failed to write configmap.yaml template: %v", err)
	}
	
	// Test loading configuration with resources
	options := BuildOptions{
		Environment:  "dev",
		OutputDir:    outputDir,
		ConfigDir:    configDir,
		TemplatesDir: templatesDir,
		Variables:    make(map[string]string),
	}
	
	mikoManifest := New(options)
	config, err := mikoManifest.LoadConfig("dev")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify that variables from all levels are merged correctly
	variableMap := make(map[string]string)
	for _, v := range config.Variables {
		variableMap[v.Name] = v.Value
	}
	
	// Check that base variables are included
	if variableMap["app_name"] != "test-app" {
		t.Errorf("Expected app_name to be 'test-app', got '%s'", variableMap["app_name"])
	}
	
	// Check that dev config overrides base version
	if variableMap["version"] != "2.0.0" {
		t.Errorf("Expected version to be '2.0.0' (overridden), got '%s'", variableMap["version"])
	}
	
	// Check that environment is set by dev config
	if variableMap["environment"] != "development" {
		t.Errorf("Expected environment to be 'development', got '%s'", variableMap["environment"])
	}
	
	// Check that component variables are included
	if variableMap["component_enabled"] != "true" {
		t.Errorf("Expected component_enabled to be 'true', got '%s'", variableMap["component_enabled"])
	}
	
	// Check that schemas are merged correctly
	expectedSchemas := map[string]bool{
		"./schemas/base-crd.yaml":         false,
		"./schemas/dev-specific-crd.yaml": false,
	}
	
	for _, schema := range config.Schemas {
		if _, exists := expectedSchemas[schema]; exists {
			expectedSchemas[schema] = true
		}
	}
	
	for schema, found := range expectedSchemas {
		if !found {
			t.Errorf("Expected schema '%s' not found in merged config", schema)
		}
	}
	
	// Check that all includes are merged
	expectedFiles := map[string]bool{
		"deployment.yaml": false,
		"service.yaml":    false,
		"configmap.yaml":  false,
	}
	
	for _, inc := range config.Include {
		if _, exists := expectedFiles[inc.File]; exists {
			expectedFiles[inc.File] = true
		}
	}
	
	for file, found := range expectedFiles {
		if !found {
			t.Errorf("Expected include file '%s' not found in merged config", file)
		}
	}
	
	// Test the full build process
	if err := mikoManifest.Build(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}
	
	// Verify that output files were created
	expectedOutputs := []string{
		"deployment.yaml",
		"service.yaml",
		"configmap.yaml",
	}
	
	for _, expectedFile := range expectedOutputs {
		outputPath := filepath.Join(outputDir, expectedFile)
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("Expected output file %s was not created", expectedFile)
		}
	}
}

func TestCircularDependencyDetection(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	
	// Create config A that includes B
	configA := `---
resources:
  - configB.yaml
variables:
  - name: from_a
    value: "value_a"`
	
	if err := os.WriteFile(filepath.Join(configDir, "configA.yaml"), []byte(configA), 0644); err != nil {
		t.Fatalf("Failed to write configA.yaml: %v", err)
	}
	
	// Create config B that includes A (circular dependency)
	configB := `---
resources:
  - configA.yaml
variables:
  - name: from_b
    value: "value_b"`
	
	if err := os.WriteFile(filepath.Join(configDir, "configB.yaml"), []byte(configB), 0644); err != nil {
		t.Fatalf("Failed to write configB.yaml: %v", err)
	}
	
	// Test that circular dependency is detected
	options := BuildOptions{
		Environment:  "configA",
		ConfigDir:    configDir,
		TemplatesDir: tempDir,
		OutputDir:    tempDir,
		Variables:    make(map[string]string),
	}
	
	mikoManifest := New(options)
	_, err := mikoManifest.LoadConfig("configA")
	
	if err == nil {
		t.Fatal("Expected circular dependency error, but got none")
	}
	
	if !contains(err.Error(), "circular dependency") {
		t.Errorf("Expected circular dependency error, got: %v", err)
	}
}

func TestMaxDepthExceeded(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	
	// Create a chain of configs that exceed max depth
	for i := 0; i < 7; i++ {
		var content string
		if i < 6 {
			content = fmt.Sprintf(`---
resources:
  - config%d.yaml
variables:
  - name: level
    value: "%d"`, i+1, i)
		} else {
			content = `---
variables:
  - name: level
    value: "6"`
		}
		
		filename := filepath.Join(configDir, fmt.Sprintf("config%d.yaml", i))
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write config%d.yaml: %v", i, err)
		}
	}
	
	// Test that max depth is enforced
	options := BuildOptions{
		Environment:  "config0",
		ConfigDir:    configDir,
		TemplatesDir: tempDir,
		OutputDir:    tempDir,
		Variables:    make(map[string]string),
	}
	
	mikoManifest := New(options)
	_, err := mikoManifest.LoadConfig("config0")
	
	if err == nil {
		t.Fatal("Expected max depth error, but got none")
	}
	
	if !contains(err.Error(), "maximum recursion depth") {
		t.Errorf("Expected max depth error, got: %v", err)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
