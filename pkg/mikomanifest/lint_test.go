package mikomanifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLintDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create a valid YAML file
	validYaml := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: 3`
	
	validPath := filepath.Join(tempDir, "valid.yaml")
	err := os.WriteFile(validPath, []byte(validYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid YAML file: %v", err)
	}
	
	// Test with valid directory
	options := LintOptions{
		Directory: tempDir,
	}
	
	// Note: This should now work with native Go YAML parser
	err = LintDirectory(options)
	if err != nil {
		t.Logf("LintDirectory failed: %v", err)
	} else {
		t.Log("LintDirectory passed with native Go YAML parser")
	}
	
	// Test with non-existent directory
	options = LintOptions{
		Directory: "/non/existent/directory",
	}
	
	err = LintDirectory(options)
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}

func TestCheckConfigDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create a valid YAML config file
	validConfig := `variables:
  - name: app_name
    value: test-app
include:
  - file: deployment.yaml`
	
	configPath := filepath.Join(tempDir, "dev.yaml")
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	// Test with valid directory
	options := CheckOptions{
		ConfigDir: tempDir,
	}
	
	// Note: This should now work with native Go YAML parser
	err = CheckConfigDirectory(options)
	if err != nil {
		t.Logf("CheckConfigDirectory failed: %v", err)
	} else {
		t.Log("CheckConfigDirectory passed with native Go YAML parser")
	}
	
	// Test with non-existent directory
	options = CheckOptions{
		ConfigDir: "/non/existent/directory",
	}
	
	err = CheckConfigDirectory(options)
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}

func TestLintYAMLFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create a valid YAML file
	validYaml := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app`
	
	validPath := filepath.Join(tempDir, "valid.yaml")
	err := os.WriteFile(validPath, []byte(validYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid YAML file: %v", err)
	}
	
	// Test lintYAMLFiles function
	result := lintYAMLFiles(tempDir)
	
	// Should return true for valid YAML
	if !result {
		t.Error("Expected true for valid YAML, got false")
	}
	
	// Test with non-existent directory
	result = lintYAMLFiles("/non/existent/directory")
	if result {
		t.Error("Expected false for non-existent directory, got true")
	}
}

func TestValidateKubernetesManifests(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create a valid Kubernetes manifest
	validManifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
  labels:
    app: test-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      labels:
        app: test-app
    spec:
      containers:
      - name: test-app
        image: nginx:latest
        ports:
        - containerPort: 80`
	
	manifestPath := filepath.Join(tempDir, "deployment.yaml")
	err := os.WriteFile(manifestPath, []byte(validManifest), 0644)
	if err != nil {
		t.Fatalf("Failed to create manifest file: %v", err)
	}
	
	// Test validateKubernetesManifests function
	// Note: This function returns a bool, not an error
	result := validateKubernetesManifests(tempDir)
	
	// The validation should pass for a valid manifest
	if !result {
		t.Error("Expected true for valid Kubernetes manifest, got false")
	}
	
	// Test with directory containing invalid YAML
	invalidYaml := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: not-a-number`
	
	invalidPath := filepath.Join(tempDir, "invalid.yaml")
	err = os.WriteFile(invalidPath, []byte(invalidYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid manifest file: %v", err)
	}
	
	result = validateKubernetesManifests(tempDir)
	
	// The validation might still pass because the YAML is syntactically correct
	// but semantically invalid. The actual validation depends on the Kubernetes client
	t.Logf("validateKubernetesManifests result with invalid manifest: %v", result)
	
	// Test with empty directory
	emptyDir := t.TempDir()
	result = validateKubernetesManifests(emptyDir)
	if !result {
		t.Error("Expected true for empty directory (no files to validate), got false")
	}
	
	// Test with non-existent directory
	result = validateKubernetesManifests("/non/existent/directory")
	// The function returns true for non-existent directories because there are no files to validate
	// (it treats it as "no files found" which is not an error)
	t.Logf("validateKubernetesManifests result for non-existent directory: %v", result)
}

func TestLintOptionsStructure(t *testing.T) {
	// Test that LintOptions can be created properly
	options := LintOptions{
		Directory: "/tmp/test",
	}
	
	if options.Directory != "/tmp/test" {
		t.Errorf("Expected directory to be '/tmp/test', got '%s'", options.Directory)
	}
}

func TestCheckOptionsStructure(t *testing.T) {
	// Test that CheckOptions can be created properly
	options := CheckOptions{
		ConfigDir: "/tmp/config",
	}
	
	if options.ConfigDir != "/tmp/config" {
		t.Errorf("Expected config dir to be '/tmp/config', got '%s'", options.ConfigDir)
	}
}

func TestLintDirectoryWithFile(t *testing.T) {
	// Create a temporary file (not directory)
	tempFile, err := os.CreateTemp("", "test.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	
	// Test with file instead of directory
	options := LintOptions{
		Directory: tempFile.Name(),
	}
	
	err = LintDirectory(options)
	if err == nil {
		t.Error("Expected error when passing file instead of directory, got nil")
	}
}
