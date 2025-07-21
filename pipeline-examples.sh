#!/bin/bash

# Miko-Manifest Pipeline Examples
# This script demonstrates how to use miko-manifest in CI/CD pipelines

set -e

echo "=== Miko-Manifest Pipeline Examples ==="

# Configuration
IMAGE_NAME="jepemo/miko-manifest:latest"
WORK_DIR="/workspace"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_step() {
    echo -e "${BLUE}==>${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    print_error "Docker is required but not installed"
    exit 1
fi

# Function to run miko-manifest in Docker
run_miko() {
    docker run --rm -v "$(pwd):$WORK_DIR" "$IMAGE_NAME" "$@"
}

# Example 1: Initialize a new project
example_init() {
    print_step "Example 1: Initialize a new project"
    
    # Create temporary directory for demo
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    print_step "Initializing project in $TEMP_DIR"
    docker run --rm -v "$TEMP_DIR:$WORK_DIR" "$IMAGE_NAME" init
    
    print_success "Project initialized successfully"
    
    # Show created structure
    print_step "Created project structure:"
    find . -type f -name "*.yaml" | sort
    
    cd - > /dev/null
    rm -rf "$TEMP_DIR"
}

# Example 2: Validate configuration
example_check() {
    print_step "Example 2: Validate configuration"
    
    if [[ -d "config" ]]; then
        print_step "Checking configuration files"
        run_miko check --config config
        print_success "Configuration validation completed"
    else
        print_warning "No config directory found. Run 'miko-manifest init' first."
    fi
}

# Example 3: Build with different environments
example_build() {
    print_step "Example 3: Build with different environments"
    
    if [[ ! -d "config" || ! -d "templates" ]]; then
        print_warning "Project not initialized. Skipping build example."
        return
    fi
    
    # Build for dev environment
    print_step "Building for dev environment"
    run_miko build --env dev --output-dir output-dev --config config --templates templates
    
    # Build for prod environment (if exists)
    if [[ -f "config/prod.yaml" ]]; then
        print_step "Building for prod environment"
        run_miko build --env prod --output-dir output-prod --config config --templates templates
    fi
    
    print_success "Build completed"
}

# Example 4: Build with variable overrides
example_build_with_vars() {
    print_step "Example 4: Build with variable overrides"
    
    if [[ ! -d "config" || ! -d "templates" ]]; then
        print_warning "Project not initialized. Skipping build with vars example."
        return
    fi
    
    print_step "Building with variable overrides"
    run_miko build \
        --env dev \
        --output-dir output-custom \
        --config config \
        --templates templates \
        --var app_name=custom-app \
        --var replicas=5 \
        --var namespace=custom-namespace
    
    print_success "Build with variables completed"
}

# Example 5: Lint generated files
example_lint() {
    print_step "Example 5: Lint generated files"
    
    if [[ -d "output-dev" ]]; then
        print_step "Linting generated files in output-dev"
        run_miko lint --dir output-dev
        print_success "Lint completed"
    else
        print_warning "No output-dev directory found. Build first."
    fi
}

# Example 6: Complete pipeline simulation
example_complete_pipeline() {
    print_step "Example 6: Complete CI/CD pipeline simulation"
    
    # Create temporary directory for complete pipeline demo
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    print_step "Step 1: Initialize project"
    docker run --rm -v "$TEMP_DIR:$WORK_DIR" "$IMAGE_NAME" init
    
    print_step "Step 2: Validate configuration"
    docker run --rm -v "$TEMP_DIR:$WORK_DIR" "$IMAGE_NAME" check --config config
    
    print_step "Step 3: Build manifests"
    docker run --rm -v "$TEMP_DIR:$WORK_DIR" "$IMAGE_NAME" build \
        --env dev \
        --output-dir output \
        --config config \
        --templates templates
    
    print_step "Step 4: Validate generated manifests"
    docker run --rm -v "$TEMP_DIR:$WORK_DIR" "$IMAGE_NAME" lint --dir output
    
    print_success "Complete pipeline simulation completed"
    
    print_step "Generated files:"
    find output -name "*.yaml" | sort
    
    cd - > /dev/null
    rm -rf "$TEMP_DIR"
}

# Example 7: Custom directories
example_custom_dirs() {
    print_step "Example 7: Using custom directories"
    
    if [[ ! -d "config" || ! -d "templates" ]]; then
        print_warning "Project not initialized. Skipping custom directories example."
        return
    fi
    
    # Create custom directories
    mkdir -p custom-config custom-templates
    
    # Copy existing files (if any)
    if [[ -f "config/dev.yaml" ]]; then
        cp config/dev.yaml custom-config/
    fi
    
    if [[ -f "templates/deployment.yaml" ]]; then
        cp templates/deployment.yaml custom-templates/
    fi
    
    print_step "Building with custom directories"
    run_miko build \
        --env dev \
        --output-dir output-custom-dirs \
        --config custom-config \
        --templates custom-templates
    
    print_success "Custom directories build completed"
    
    # Cleanup
    rm -rf custom-config custom-templates
}

# Main execution
main() {
    print_step "Starting miko-manifest pipeline examples"
    
    # Check if image is available
    if ! docker image inspect "$IMAGE_NAME" &> /dev/null; then
        print_step "Pulling miko-manifest image"
        docker pull "$IMAGE_NAME"
    fi
    
    # Run examples
    example_init
    example_check
    example_build
    example_build_with_vars
    example_lint
    example_custom_dirs
    example_complete_pipeline
    
    print_success "All pipeline examples completed successfully!"
    
    # Cleanup generated directories
    rm -rf output-dev output-prod output-custom output-custom-dirs
}

# Show usage
usage() {
    echo "Usage: $0 [OPTION]"
    echo "Run miko-manifest pipeline examples"
    echo ""
    echo "Options:"
    echo "  init              Run initialization example"
    echo "  check             Run configuration validation example"
    echo "  build             Run build example"
    echo "  build-vars        Run build with variables example"
    echo "  lint              Run lint example"
    echo "  custom-dirs       Run custom directories example"
    echo "  complete          Run complete pipeline example"
    echo "  all               Run all examples (default)"
    echo "  help              Show this help message"
}

# Parse command line arguments
case "${1:-all}" in
    init)
        example_init
        ;;
    check)
        example_check
        ;;
    build)
        example_build
        ;;
    build-vars)
        example_build_with_vars
        ;;
    lint)
        example_lint
        ;;
    custom-dirs)
        example_custom_dirs
        ;;
    complete)
        example_complete_pipeline
        ;;
    all)
        main
        ;;
    help)
        usage
        ;;
    *)
        print_error "Unknown option: $1"
        usage
        exit 1
        ;;
esac
