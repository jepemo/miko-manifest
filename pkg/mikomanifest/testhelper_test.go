package mikomanifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelper provides common test utilities
type TestHelper struct {
	t       *testing.T
	tempDir string
}

// NewTestHelper creates a new test helper with a temporary directory
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{
		t:       t,
		tempDir: t.TempDir(),
	}
}

// TempDir returns the temporary directory for this test
func (h *TestHelper) TempDir() string {
	return h.tempDir
}

// CreateFile creates a file with the given content in the temp directory
func (h *TestHelper) CreateFile(filename, content string) string {
	filePath := filepath.Join(h.tempDir, filename)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.t.Fatalf("Failed to create directory %s: %v", dir, err)
	}
	
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		h.t.Fatalf("Failed to create file %s: %v", filePath, err)
	}
	
	return filePath
}

// CreateDir creates a directory in the temp directory
func (h *TestHelper) CreateDir(dirname string) string {
	dirPath := filepath.Join(h.tempDir, dirname)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		h.t.Fatalf("Failed to create directory %s: %v", dirPath, err)
	}
	return dirPath
}

// FileExists checks if a file exists
func (h *TestHelper) FileExists(filename string) bool {
	filePath := filepath.Join(h.tempDir, filename)
	_, err := os.Stat(filePath)
	return err == nil
}

// ReadFile reads a file from the temp directory
func (h *TestHelper) ReadFile(filename string) string {
	filePath := filepath.Join(h.tempDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		h.t.Fatalf("Failed to read file %s: %v", filePath, err)
	}
	return string(content)
}

// AssertFileContains checks if a file contains the expected content
func (h *TestHelper) AssertFileContains(filename, expected string) {
	content := h.ReadFile(filename)
	if !strings.Contains(content, expected) {
		h.t.Errorf("File %s should contain '%s', but got:\n%s", filename, expected, content)
	}
}

// AssertFileNotContains checks if a file does not contain the given content
func (h *TestHelper) AssertFileNotContains(filename, notExpected string) {
	content := h.ReadFile(filename)
	if strings.Contains(content, notExpected) {
		h.t.Errorf("File %s should not contain '%s', but got:\n%s", filename, notExpected, content)
	}
}

// AssertStringContains checks if a string contains the expected substring
func (h *TestHelper) AssertStringContains(actual, expected string) {
	if !strings.Contains(actual, expected) {
		h.t.Errorf("String should contain '%s', but got:\n%s", expected, actual)
	}
}

// AssertNoError checks that an error is nil
func (h *TestHelper) AssertNoError(err error) {
	if err != nil {
		h.t.Fatalf("Expected no error, but got: %v", err)
	}
}

// AssertError checks that an error is not nil
func (h *TestHelper) AssertError(err error) {
	if err == nil {
		h.t.Fatal("Expected an error, but got nil")
	}
}

// AssertErrorContains checks that an error contains the expected message
func (h *TestHelper) AssertErrorContains(err error, expectedMessage string) {
	if err == nil {
		h.t.Fatal("Expected an error, but got nil")
	}
	if !strings.Contains(err.Error(), expectedMessage) {
		h.t.Errorf("Expected error to contain '%s', but got: %v", expectedMessage, err)
	}
}

// Common test data
const (
	ValidDeploymentYAML = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.app_name}}
  namespace: {{.namespace}}
spec:
  replicas: {{.replicas}}
  selector:
    matchLabels:
      app: {{.app_name}}
  template:
    metadata:
      labels:
        app: {{.app_name}}
    spec:
      containers:
        - name: {{.app_name}}
          image: {{.image}}:{{.tag}}
          ports:
            - containerPort: {{.port}}
`

	ValidServiceYAML = `---
apiVersion: v1
kind: Service
metadata:
  name: {{.service_name}}
  namespace: {{.namespace}}
spec:
  selector:
    app: {{.app_name}}
  ports:
    - port: {{.service_port}}
      targetPort: {{.target_port}}
`

	ValidConfigMapYAML = `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.config_name}}
  namespace: {{.namespace}}
data:
  database_url: {{.database_url}}
  redis_url: {{.redis_url}}
`

	ValidConfigYAML = `---
variables:
  - name: app_name
    value: test-app
  - name: namespace
    value: default
  - name: replicas
    value: "3"
  - name: image
    value: nginx
  - name: tag
    value: latest
  - name: port
    value: "80"

include:
  - file: deployment.yaml
  - file: service.yaml
    repeat: multiple-files
    list:
      - key: frontend
        values:
          - name: service_name
            value: frontend-service
          - name: service_port
            value: "80"
          - name: target_port
            value: "8080"
      - key: backend
        values:
          - name: service_name
            value: backend-service
          - name: service_port
            value: "3000"
          - name: target_port
            value: "3000"
`

	InvalidYAML = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
  invalid yaml content [[[
spec:
  replicas: 3
`
)

// CreateTestProject creates a complete test project structure
func (h *TestHelper) CreateTestProject() {
	// Create directories
	h.CreateDir("templates")
	h.CreateDir("config")
	h.CreateDir("output")

	// Create template files
	h.CreateFile("templates/deployment.yaml", ValidDeploymentYAML)
	h.CreateFile("templates/service.yaml", ValidServiceYAML)
	h.CreateFile("templates/configmap.yaml", ValidConfigMapYAML)

	// Create config file
	h.CreateFile("config/test.yaml", ValidConfigYAML)
}

// GetBuildOptions returns standard build options for testing
func (h *TestHelper) GetBuildOptions() BuildOptions {
	return BuildOptions{
		Environment:  "test",
		OutputDir:    filepath.Join(h.tempDir, "output"),
		ConfigDir:    filepath.Join(h.tempDir, "config"),
		TemplatesDir: filepath.Join(h.tempDir, "templates"),
		Variables:    map[string]string{},
	}
}

// GetInitOptions returns standard init options for testing
func (h *TestHelper) GetInitOptions() InitOptions {
	return InitOptions{
		ProjectDir: h.tempDir,
	}
}

// GetLintOptions returns standard lint options for testing
func (h *TestHelper) GetLintOptions() LintOptions {
	return LintOptions{
		Directory: filepath.Join(h.tempDir, "output"),
	}
}

// GetCheckOptions returns standard check options for testing
func (h *TestHelper) GetCheckOptions() CheckOptions {
	return CheckOptions{
		ConfigDir: filepath.Join(h.tempDir, "config"),
	}
}

// SkipIfYAMLLintingUnavailable skips the test if YAML linting is not available
func (h *TestHelper) SkipIfYAMLLintingUnavailable() {
	// Since we're using native Go YAML parsing, this is always available
	// This function is kept for compatibility but doesn't skip anymore
}

// Benchmark helper functions for performance testing
func (h *TestHelper) BenchmarkBuild(b *testing.B, options BuildOptions) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manifest := New(options)
		if err := manifest.Build(); err != nil {
			b.Fatalf("Build failed: %v", err)
		}
	}
}

func (h *TestHelper) BenchmarkInit(b *testing.B, options InitOptions) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a new temp dir for each iteration
		tempDir := b.TempDir()
		opts := InitOptions{ProjectDir: tempDir}
		if err := InitProject(opts); err != nil {
			b.Fatalf("Init failed: %v", err)
		}
	}
}

// MockError represents a mock error for testing
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}

// NewMockError creates a new mock error
func NewMockError(message string) error {
	return &MockError{message: message}
}

// TestValidation provides validation helpers for testing
func (h *TestHelper) ValidateKubernetesManifest(content string) error {
	// Basic validation - check for required Kubernetes fields
	if !strings.Contains(content, "apiVersion:") {
		return fmt.Errorf("missing apiVersion field")
	}
	if !strings.Contains(content, "kind:") {
		return fmt.Errorf("missing kind field")
	}
	if !strings.Contains(content, "metadata:") {
		return fmt.Errorf("missing metadata field")
	}
	return nil
}

// CleanupTestFiles removes test files (usually called in defer)
func (h *TestHelper) CleanupTestFiles() {
	// t.TempDir() automatically cleans up, but this can be used for manual cleanup
	if h.tempDir != "" {
		if err := os.RemoveAll(h.tempDir); err != nil {
			// In test cleanup, just log the error - don't fail the test
			fmt.Printf("Warning: failed to remove temp dir %s: %v\n", h.tempDir, err)
		}
	}
}
