package mikomanifest

import (
	"fmt"
	"os"
	"path/filepath"
)

// InitOptions contains options for initializing a project
type InitOptions struct {
	ProjectDir string
}

// InitProject initializes a new miko-manifest project
func InitProject(options InitOptions) error {
	fmt.Println("Initializing miko-manifest project...")
	
	// Create templates directory
	templatesDir := filepath.Join(options.ProjectDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}
	fmt.Printf("✓ Created directory: %s\n", templatesDir)
	
	// Create config directory
	configDir := filepath.Join(options.ProjectDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	fmt.Printf("✓ Created directory: %s\n", configDir)
	
	// Create template files
	if err := createTemplateFiles(templatesDir); err != nil {
		return err
	}
	
	// Create configuration file
	if err := createConfigFile(configDir); err != nil {
		return err
	}
	
	fmt.Println("SUCCESS: miko-manifest project initialized successfully!")
	fmt.Println("")
	fmt.Println("Example templates created:")
	fmt.Println("   • deployment.yaml - Simple file processing")
	fmt.Println("   • configmap.yaml - Same-file repeat (multiple sections in one file)")
	fmt.Println("   • service.yaml - Multiple-files repeat (separate files per key)")
	fmt.Println("")
	fmt.Println("To build the project with the example configuration, run:")
	fmt.Println("   miko-manifest build --env dev --output-dir output")
	fmt.Println("")
	fmt.Println("This will generate:")
	fmt.Println("   • output/deployment.yaml")
	fmt.Println("   • output/configmap.yaml (with 2 sections)")
	fmt.Println("   • output/service-frontend.yaml")
	fmt.Println("   • output/service-backend.yaml")
	
	return nil
}

func createTemplateFiles(templatesDir string) error {
	// Create deployment.yaml template
	deploymentContent := `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: "{{.app_name}}"
  namespace: "{{.namespace}}"
spec:
  replicas: {{.replicas}}
  selector:
    matchLabels:
      app: "{{.app_name}}"
  template:
    metadata:
      labels:
        app: "{{.app_name}}"
    spec:
      containers:
        - name: "{{.app_name}}"
          image: "{{.image}}:{{.tag}}"
          ports:
            - containerPort: {{.port}}
          env:
            - name: ENVIRONMENT
              value: "{{.environment}}"
`
	
	deploymentPath := filepath.Join(templatesDir, "deployment.yaml")
	if err := os.WriteFile(deploymentPath, []byte(deploymentContent), 0644); err != nil {
		return fmt.Errorf("failed to create deployment.yaml: %w", err)
	}
	fmt.Printf("✓ Created template file: %s\n", deploymentPath)
	
	// Create configmap.yaml template
	configmapContent := `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: "{{.config_name}}"
  namespace: "{{.namespace}}"
data:
  database_url: "{{.database_url}}"
  redis_url: "{{.redis_url}}"
`
	
	configmapPath := filepath.Join(templatesDir, "configmap.yaml")
	if err := os.WriteFile(configmapPath, []byte(configmapContent), 0644); err != nil {
		return fmt.Errorf("failed to create configmap.yaml: %w", err)
	}
	fmt.Printf("✓ Created template file: %s\n", configmapPath)
	
	// Create service.yaml template
	serviceContent := `---
apiVersion: v1
kind: Service
metadata:
  name: "{{.service_name}}"
  namespace: "{{.namespace}}"
spec:
  selector:
    app: "{{.app_name}}"
  ports:
    - port: {{.service_port}}
      targetPort: {{.target_port}}
`
	
	servicePath := filepath.Join(templatesDir, "service.yaml")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to create service.yaml: %w", err)
	}
	fmt.Printf("✓ Created template file: %s\n", servicePath)
	
	return nil
}

func createConfigFile(configDir string) error {
	devConfigContent := `---
# Example configuration file for miko-manifest
# This demonstrates all three types of file processing

# Variables section - define variables that can be used throughout
# the configuration
variables:
  # Global variables used across all templates
  - name: app_name
    value: my-app
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
  - name: environment
    value: development

# Include section - specify files to include in the configuration
include:
  # Simple file processing - deployment.yaml will be processed once
  # with global variables
  - file: deployment.yaml

  # Same-file repeat pattern - configmap.yaml will generate multiple
  # sections in one file
  - file: configmap.yaml
    repeat: same-file
    list:
      - key: database-config
        values:
          - name: config_name
            value: database-config
          - name: database_url
            value: postgresql://localhost:5432/mydb
          - name: redis_url
            value: redis://localhost:6379
      - key: cache-config
        values:
          - name: config_name
            value: cache-config
          - name: database_url
            value: postgresql://cache-db:5432/cache
          - name: redis_url
            value: redis://cache-redis:6379

  # Multiple-files repeat pattern - service.yaml will generate
  # separate files for each key
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
	
	devConfigPath := filepath.Join(configDir, "dev.yaml")
	if err := os.WriteFile(devConfigPath, []byte(devConfigContent), 0644); err != nil {
		return fmt.Errorf("failed to create dev.yaml: %w", err)
	}
	fmt.Printf("✓ Created configuration file: %s\n", devConfigPath)
	
	// Create example schema configuration file
	schemaConfigContent := `schemas:
  # Example CRD from a URL (Crossplane Composition)
  # - https://raw.githubusercontent.com/crossplane/crossplane/master/cluster/crds/apiextensions.crossplane.io_compositions.yaml
  
  # Example local file
  # - ./schemas/my-custom-crd.yaml
  
  # Example directory with multiple CRDs
  # - ./schemas/operators/
`
	
	schemaConfigPath := filepath.Join(configDir, "schemas.yaml")
	if err := os.WriteFile(schemaConfigPath, []byte(schemaConfigContent), 0644); err != nil {
		return fmt.Errorf("failed to create schemas.yaml: %w", err)
	}
	fmt.Printf("✓ Created schema configuration file: %s\n", schemaConfigPath)
	
	return nil
}
