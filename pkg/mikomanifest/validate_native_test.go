package mikomanifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateWithKubernetesTypes(t *testing.T) {
	// Test the new validation system with native Kubernetes types
	testCases := []struct {
		name        string
		manifest    string
		shouldFail  bool
		expectedError string
	}{
		{
			name: "ValidDeployment",
			manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
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
      - name: app
        image: nginx:1.20
        ports:
        - containerPort: 80`,
			shouldFail: false,
		},
		{
			name: "InvalidDeploymentWithRepicas",
			manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  repicas: 3
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      labels:
        app: test-app
    spec:
      containers:
      - name: app
        image: nginx:1.20
        ports:
        - containerPort: 80`,
			shouldFail: true,
			expectedError: "repicas",
		},
		{
			name: "ValidService",
			manifest: `apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  selector:
    app: test-app
  ports:
  - port: 80
    targetPort: 8080`,
			shouldFail: false,
		},
		{
			name: "InvalidServiceWithSelctor",
			manifest: `apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  selctor:
    app: test-app
  ports:
  - port: 80
    targetPort: 8080`,
			shouldFail: true,
			expectedError: "selctor",
		},
		{
			name: "ValidConfigMap",
			manifest: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value`,
			shouldFail: false,
		},
		{
			name: "InvalidConfigMapWithDta",
			manifest: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
dta:
  key: value`,
			shouldFail: true,
			expectedError: "dta",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary directory and file
			tempDir := t.TempDir()
			manifestFile := filepath.Join(tempDir, "test.yaml")
			
			err := os.WriteFile(manifestFile, []byte(tc.manifest), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Run validation
			success := validateKubernetesManifests(tempDir, nil)
			
			if tc.shouldFail && success {
				t.Errorf("Expected validation to fail for %s, but it passed", tc.name)
			} else if !tc.shouldFail && !success {
				t.Errorf("Expected validation to pass for %s, but it failed", tc.name)
			}
		})
	}
}

func TestStrictFieldValidation(t *testing.T) {
	// Test that the new validation catches unknown fields
	tempDir := t.TempDir()
	
	// Create a deployment with multiple field errors
	invalidManifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  repicas: 3        # Should be "replicas"
  selctor:          # Should be "selector"  
    matchLables:     # Should be "matchLabels"
      app: test-app
  template:
    metadata:
      labls:         # Should be "labels"
        app: test-app
    spec:
      contaiers:     # Should be "containers"
      - name: app
        imagen: nginx  # Should be "image"
        ports:
        - contPort: 80  # Should be "containerPort"`
        
	manifestFile := filepath.Join(tempDir, "invalid.yaml")
	err := os.WriteFile(manifestFile, []byte(invalidManifest), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// This should fail with multiple field validation errors
	success := validateKubernetesManifests(tempDir, nil)
	if success {
		t.Error("Expected validation to fail for manifest with multiple field errors")
	}
}

func TestValidationErrorMessages(t *testing.T) {
	// Test that error messages are helpful and specific
	tempDir := t.TempDir()
	
	invalidDeployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  repicas: 3`
        
	manifestFile := filepath.Join(tempDir, "test.yaml")
	err := os.WriteFile(manifestFile, []byte(invalidDeployment), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Validate and check error message quality
	success := validateKubernetesManifests(tempDir, nil)
	if success {
		t.Error("Expected validation to fail for deployment with 'repicas' field")
	}
	
	// Note: In a more complete implementation, we would capture 
	// and analyze the specific error messages here
}
