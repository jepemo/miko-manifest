package mikomanifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitProject(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test initialization
	options := InitOptions{
		ProjectDir: tempDir,
	}

	err := InitProject(options)
	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// Check that templates directory was created
	templatesDir := filepath.Join(tempDir, "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		t.Errorf("Templates directory was not created: %s", templatesDir)
	}

	// Check that config directory was created
	configDir := filepath.Join(tempDir, "config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created: %s", configDir)
	}

	// Check that template files were created
	expectedTemplates := []string{
		"deployment.yaml",
		"configmap.yaml",
		"service.yaml",
	}

	for _, template := range expectedTemplates {
		templatePath := filepath.Join(templatesDir, template)
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			t.Errorf("Template file was not created: %s", templatePath)
		}
	}

	// Check that config file was created
	configPath := filepath.Join(configDir, "dev.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created: %s", configPath)
	}
}

func TestInitProjectWithExistingDirectories(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Pre-create directories
	templatesDir := filepath.Join(tempDir, "templates")
	err := os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	configDir := filepath.Join(tempDir, "config")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Test initialization with existing directories
	options := InitOptions{
		ProjectDir: tempDir,
	}

	err = InitProject(options)
	if err != nil {
		t.Fatalf("InitProject failed with existing directories: %v", err)
	}

	// Check that files were still created
	templatePath := filepath.Join(templatesDir, "deployment.yaml")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Errorf("Template file was not created in existing directory: %s", templatePath)
	}

	configPath := filepath.Join(configDir, "dev.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created in existing directory: %s", configPath)
	}
}

func TestCreateTemplateFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	err := createTemplateFiles(tempDir)
	if err != nil {
		t.Fatalf("createTemplateFiles failed: %v", err)
	}

	// Check deployment.yaml
	deploymentPath := filepath.Join(tempDir, "deployment.yaml")
	deploymentContent, err := os.ReadFile(deploymentPath)
	if err != nil {
		t.Fatalf("Failed to read deployment.yaml: %v", err)
	}

	deploymentStr := string(deploymentContent)
	if !strings.Contains(deploymentStr, "apiVersion: apps/v1") {
		t.Errorf("deployment.yaml should contain 'apiVersion: apps/v1'")
	}
	if !strings.Contains(deploymentStr, "kind: Deployment") {
		t.Errorf("deployment.yaml should contain 'kind: Deployment'")
	}
	if !strings.Contains(deploymentStr, "{{.app_name}}") {
		t.Errorf("deployment.yaml should contain Go template variable '{{.app_name}}'")
	}

	// Check configmap.yaml
	configmapPath := filepath.Join(tempDir, "configmap.yaml")
	configmapContent, err := os.ReadFile(configmapPath)
	if err != nil {
		t.Fatalf("Failed to read configmap.yaml: %v", err)
	}

	configmapStr := string(configmapContent)
	if !strings.Contains(configmapStr, "kind: ConfigMap") {
		t.Errorf("configmap.yaml should contain 'kind: ConfigMap'")
	}
	if !strings.Contains(configmapStr, "{{.config_name}}") {
		t.Errorf("configmap.yaml should contain Go template variable '{{.config_name}}'")
	}

	// Check service.yaml
	servicePath := filepath.Join(tempDir, "service.yaml")
	serviceContent, err := os.ReadFile(servicePath)
	if err != nil {
		t.Fatalf("Failed to read service.yaml: %v", err)
	}

	serviceStr := string(serviceContent)
	if !strings.Contains(serviceStr, "kind: Service") {
		t.Errorf("service.yaml should contain 'kind: Service'")
	}
	if !strings.Contains(serviceStr, "{{.service_name}}") {
		t.Errorf("service.yaml should contain Go template variable '{{.service_name}}'")
	}
}

func TestCreateConfigFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	err := createConfigFile(tempDir)
	if err != nil {
		t.Fatalf("createConfigFile failed: %v", err)
	}

	// Check that config file was created
	configPath := filepath.Join(tempDir, "dev.yaml")
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read dev.yaml: %v", err)
	}

	configStr := string(configContent)

	// Check that config contains expected sections
	if !strings.Contains(configStr, "variables:") {
		t.Errorf("Config should contain 'variables:' section")
	}
	if !strings.Contains(configStr, "include:") {
		t.Errorf("Config should contain 'include:' section")
	}

	// Check that config contains expected variables
	if !strings.Contains(configStr, "app_name") {
		t.Errorf("Config should contain 'app_name' variable")
	}
	if !strings.Contains(configStr, "namespace") {
		t.Errorf("Config should contain 'namespace' variable")
	}
	if !strings.Contains(configStr, "replicas") {
		t.Errorf("Config should contain 'replicas' variable")
	}

	// Check that config contains expected file includes
	if !strings.Contains(configStr, "deployment.yaml") {
		t.Errorf("Config should include 'deployment.yaml'")
	}
	if !strings.Contains(configStr, "configmap.yaml") {
		t.Errorf("Config should include 'configmap.yaml'")
	}
	if !strings.Contains(configStr, "service.yaml") {
		t.Errorf("Config should include 'service.yaml'")
	}

	// Check that config contains different repeat patterns
	if !strings.Contains(configStr, "repeat: same-file") {
		t.Errorf("Config should contain 'repeat: same-file' pattern")
	}
	if !strings.Contains(configStr, "repeat: multiple-files") {
		t.Errorf("Config should contain 'repeat: multiple-files' pattern")
	}
}

func TestInitProjectWithInvalidPath(t *testing.T) {
	// Test with invalid path (read-only filesystem simulation)
	// This test might be skipped on some systems
	if os.Getuid() == 0 {
		t.Skip("Skipping test when running as root")
	}

	invalidPath := "/invalid/path/that/does/not/exist"
	options := InitOptions{
		ProjectDir: invalidPath,
	}

	err := InitProject(options)
	if err == nil {
		t.Errorf("Expected InitProject to fail with invalid path, but it succeeded")
	}
}

func TestInitProjectFileContents(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test initialization
	options := InitOptions{
		ProjectDir: tempDir,
	}

	err := InitProject(options)
	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// Test that the created config file is valid YAML and contains expected structure
	configPath := filepath.Join(tempDir, "config", "dev.yaml")
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	configStr := string(configContent)

	// Check for proper YAML structure
	if !strings.Contains(configStr, "---") {
		t.Errorf("Config file should start with YAML document separator '---'")
	}

	// Check for complete example configuration
	expectedPatterns := []string{
		"variables:",
		"include:",
		"- file: deployment.yaml",
		"- file: configmap.yaml",
		"- file: service.yaml",
		"repeat: same-file",
		"repeat: multiple-files",
		"list:",
		"key:",
		"values:",
		"- name:",
		"value:",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(configStr, pattern) {
			t.Errorf("Config file should contain pattern '%s'", pattern)
		}
	}

	// Test that deployment template contains valid Kubernetes YAML
	deploymentPath := filepath.Join(tempDir, "templates", "deployment.yaml")
	deploymentContent, err := os.ReadFile(deploymentPath)
	if err != nil {
		t.Fatalf("Failed to read deployment template: %v", err)
	}

	deploymentStr := string(deploymentContent)
	k8sPatterns := []string{
		"apiVersion: apps/v1",
		"kind: Deployment",
		"metadata:",
		"spec:",
		"selector:",
		"template:",
		"containers:",
	}

	for _, pattern := range k8sPatterns {
		if !strings.Contains(deploymentStr, pattern) {
			t.Errorf("Deployment template should contain Kubernetes pattern '%s'", pattern)
		}
	}
}
