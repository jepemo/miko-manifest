package mikomanifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateKubernetesResourceSchemaValidation(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Test case: Deployment with invalid field name "repicas" instead of "replicas"
	invalidDeployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
  labels:
    app: test-app
spec:
  repicas: 3  # This should be "replicas"
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
        - containerPort: 80
`

	// Write the invalid deployment to a file
	deploymentFile := filepath.Join(tempDir, "invalid-deployment.yaml")
	err := os.WriteFile(deploymentFile, []byte(invalidDeployment), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test that current validation does catch the error (fixed!)
	t.Run("CurrentValidationNowCatchesFieldErrors", func(t *testing.T) {
		// This should now fail because we've improved the validation
		success := validateKubernetesManifests(tempDir, nil)
		if success {
			t.Error("Expected validation to fail for invalid field 'repicas', but it passed")
		}
	})

	// Test case: Valid deployment for comparison
	validDeployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
  labels:
    app: test-app
spec:
  replicas: 3  # Correct field name
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
        - containerPort: 80
`

	// Write the valid deployment to a file
	validDeploymentFile := filepath.Join(tempDir, "valid-deployment.yaml")
	err = os.WriteFile(validDeploymentFile, []byte(validDeployment), 0644)
	if err != nil {
		t.Fatalf("Failed to write valid test file: %v", err)
	}

	t.Run("ValidDeploymentShouldPass", func(t *testing.T) {
		// Test only the directory with the valid deployment
		validTempDir := t.TempDir()
		validDeploymentFile := filepath.Join(validTempDir, "valid-deployment.yaml")
		err = os.WriteFile(validDeploymentFile, []byte(validDeployment), 0644)
		if err != nil {
			t.Fatalf("Failed to write valid test file: %v", err)
		}

		success := validateKubernetesManifests(validTempDir, nil)
		if !success {
			t.Error("Expected valid deployment to pass validation")
		}
	})
}

func TestValidateKubernetesResourceFieldValidation(t *testing.T) {
	// Test various field validation scenarios that should fail but currently don't

	testCases := []struct {
		name       string
		manifest   string
		shouldFail bool
		reason     string
	}{
		{
			name: "DeploymentWithInvalidReplicasField",
			manifest: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test
spec:
  repicas: 3  # Should be "replicas"
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: test
        image: nginx`,
			shouldFail: true,
			reason:     "Invalid field name 'repicas' instead of 'replicas'",
		},
		{
			name: "ServiceWithInvalidSelectorField",
			manifest: `apiVersion: v1
kind: Service
metadata:
  name: test
spec:
  selctor:  # Should be "selector"
    app: test
  ports:
  - port: 80`,
			shouldFail: true,
			reason:     "Invalid field name 'selctor' instead of 'selector'",
		},
		{
			name: "ConfigMapWithInvalidDataField",
			manifest: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test
dta:  # Should be "data"
  key: value`,
			shouldFail: true,
			reason:     "Invalid field name 'dta' instead of 'data'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temp directory and file for this test case
			tempDir := t.TempDir()
			manifestFile := filepath.Join(tempDir, "test.yaml")

			err := os.WriteFile(manifestFile, []byte(tc.manifest), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Current validation - this demonstrates the limitation
			success := validateKubernetesManifests(tempDir, nil)

			if tc.shouldFail && success {
				t.Errorf("Expected validation to fail for %s, but it passed. Reason: %s", tc.name, tc.reason)
			} else if !tc.shouldFail && !success {
				t.Errorf("Expected validation to pass for %s, but it failed", tc.name)
			}
		})
	}
}
