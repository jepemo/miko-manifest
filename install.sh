#!/bin/bash

# Miko-Manifest Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/jepemo/miko-manifest/main/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="jepemo/miko-manifest"
BINARY_NAME="miko-manifest"
INSTALL_DIR="/usr/local/bin"
USER_INSTALL_DIR="$HOME/.local/bin"

# Print functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect platform and architecture
detect_platform() {
    local os arch
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          print_error "Unsupported operating system: $(uname -s)" ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        aarch64|arm64)  arch="arm64" ;;
        *)              print_error "Unsupported architecture: $(uname -m)" ;;
    esac
    
    echo "${os}_${arch}"
}

# Get latest release version from GitHub
get_latest_version() {
    local version
    version=$(curl -sSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        print_error "Failed to get latest version from GitHub"
    fi
    
    echo "$version"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check dependencies
check_dependencies() {
    print_info "Checking dependencies..."
    
    if ! command_exists curl; then
        print_error "curl is required but not installed. Please install curl and try again."
    fi
    
    if ! command_exists tar; then
        print_error "tar is required but not installed. Please install tar and try again."
    fi
    
    print_success "All dependencies are available"
}

# Remove existing installation
remove_existing() {
    local binary_path
    
    # Check common installation locations
    for dir in "$INSTALL_DIR" "$USER_INSTALL_DIR" "$HOME/bin" "/usr/bin"; do
        binary_path="$dir/$BINARY_NAME"
        if [ -f "$binary_path" ]; then
            print_warning "Found existing installation at $binary_path"
            if [ -w "$dir" ] || [ "$dir" = "$USER_INSTALL_DIR" ] || [ "$dir" = "$HOME/bin" ]; then
                rm -f "$binary_path"
                print_success "Removed existing installation from $binary_path"
            else
                print_warning "Cannot remove $binary_path (requires sudo). Please remove manually if needed."
            fi
        fi
    done
}

# Uninstall function
uninstall() {
    echo ""
    echo "🗑️  Miko-Manifest Uninstallation"
    echo "================================="
    echo ""
    
    local found_installations=false
    
    # Check common installation locations
    for dir in "$INSTALL_DIR" "$USER_INSTALL_DIR" "$HOME/bin" "/usr/bin"; do
        local binary_path="$dir/$BINARY_NAME"
        if [ -f "$binary_path" ]; then
            found_installations=true
            print_info "Found installation at $binary_path"
            
            # Try to get version before removing
            if [ -x "$binary_path" ]; then
                local version_output
                version_output=$("$binary_path" --version 2>/dev/null || echo "Version info not available")
                print_info "Current version: $version_output"
            fi
            
            if [ -w "$dir" ] || [ "$dir" = "$USER_INSTALL_DIR" ] || [ "$dir" = "$HOME/bin" ]; then
                rm -f "$binary_path"
                print_success "Removed $BINARY_NAME from $binary_path"
            else
                print_warning "Cannot remove $binary_path (requires sudo)"
                print_info "To remove manually, run: sudo rm -f $binary_path"
            fi
        fi
    done
    
    if [ "$found_installations" = true ]; then
        echo ""
        print_success "🎉 Miko-Manifest uninstallation completed!"
        print_info "Thank you for using Miko-Manifest!"
    else
        echo ""
        print_warning "No Miko-Manifest installations found"
        print_info "Nothing to uninstall"
    fi
    
    echo ""
}

# Determine installation directory
get_install_dir() {
    if [ -w "$INSTALL_DIR" ]; then
        echo "$INSTALL_DIR"
    else
        # Create user local bin if it doesn't exist
        mkdir -p "$USER_INSTALL_DIR"
        echo "$USER_INSTALL_DIR"
    fi
}

# Download and install binary
install_binary() {
    local platform="$1"
    local version="$2"
    local install_dir="$3"
    local temp_dir archive_name download_url
    
    temp_dir=$(mktemp -d)
    archive_name="${BINARY_NAME}_${version#v}_${platform}"
    
    # Determine file extension
    if [[ "$platform" == *"windows"* ]]; then
        archive_name="${archive_name}.zip"
    else
        archive_name="${archive_name}.tar.gz"
    fi
    
    download_url="https://github.com/$REPO/releases/download/$version/$archive_name"
    
    print_info "Downloading $BINARY_NAME $version for $platform..."
    print_info "Download URL: $download_url"
    
    # Download the archive
    if ! curl -sSL -o "$temp_dir/$archive_name" "$download_url"; then
        rm -rf "$temp_dir"
        print_error "Failed to download $archive_name"
    fi
    
    print_success "Downloaded $archive_name"
    
    # Extract the archive
    print_info "Extracting archive..."
    cd "$temp_dir"
    
    if [[ "$archive_name" == *.zip ]]; then
        if command_exists unzip; then
            unzip -q "$archive_name"
        else
            print_error "unzip is required to extract Windows binaries"
        fi
    else
        tar -xzf "$archive_name"
    fi
    
    # Find the binary (it might be in a subdirectory)
    local binary_path
    binary_path=$(find . -name "$BINARY_NAME" -type f | head -1)
    
    if [ -z "$binary_path" ]; then
        rm -rf "$temp_dir"
        print_error "Binary not found in downloaded archive"
    fi
    
    # Make binary executable
    chmod +x "$binary_path"
    
    # Install binary
    print_info "Installing to $install_dir..."
    if ! cp "$binary_path" "$install_dir/$BINARY_NAME"; then
        rm -rf "$temp_dir"
        print_error "Failed to install binary to $install_dir"
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
    
    print_success "Installed $BINARY_NAME to $install_dir"
}

# Verify installation
verify_installation() {
    local install_dir="$1"
    local binary_path="$install_dir/$BINARY_NAME"
    
    if [ ! -f "$binary_path" ]; then
        print_error "Installation verification failed: binary not found at $binary_path"
    fi
    
    if [ ! -x "$binary_path" ]; then
        print_error "Installation verification failed: binary is not executable"
    fi
    
    # Test the binary
    print_info "Testing installation..."
    if "$binary_path" --version >/dev/null 2>&1; then
        print_success "Installation verified successfully"
        
        # Show version
        local version_output
        version_output=$("$binary_path" --version 2>/dev/null || echo "Version info not available")
        print_info "Installed version: $version_output"
    else
        print_warning "Binary installed but version check failed"
    fi
}

# Add to PATH instructions
show_path_instructions() {
    local install_dir="$1"
    
    if [ "$install_dir" = "$USER_INSTALL_DIR" ]; then
        print_warning "Binary installed to $install_dir"
        print_info "Make sure $install_dir is in your PATH."
        print_info "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo ""
        echo "    export PATH=\"$install_dir:\$PATH\""
        echo ""
        print_info "Then restart your shell or run: source ~/.bashrc (or ~/.zshrc)"
    fi
}

# Main installation function
main() {
    echo ""
    echo "🎯 Miko-Manifest Installation Script"
    echo "======================================"
    echo ""
    
    # Check dependencies
    check_dependencies
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    print_info "Detected platform: $platform"
    
    # Get latest version
    local version
    version=$(get_latest_version)
    print_info "Latest version: $version"
    
    # Remove existing installations
    remove_existing
    
    # Determine installation directory
    local install_dir
    install_dir=$(get_install_dir)
    print_info "Installation directory: $install_dir"
    
    # Install binary
    install_binary "$platform" "$version" "$install_dir"
    
    # Verify installation
    verify_installation "$install_dir"
    
    # Show PATH instructions if needed
    show_path_instructions "$install_dir"
    
    echo ""
    print_success "🎉 Miko-Manifest installation completed!"
    echo ""
    print_info "Quick start:"
    echo "  $BINARY_NAME init                    # Initialize a new project"
    echo "  $BINARY_NAME build --env dev        # Build manifests"
    echo "  $BINARY_NAME --help                 # Show help"
    echo ""
    print_info "Documentation: https://github.com/$REPO"
    echo ""
}

# Handle command line arguments
case "${1:-}" in
    "--help"|"-h")
        echo "Miko-Manifest Installation Script"
        echo ""
        echo "Usage:"
        echo "  curl -sSL https://raw.githubusercontent.com/jepemo/miko-manifest/main/install.sh | bash"
        echo "  curl -sSL https://raw.githubusercontent.com/jepemo/miko-manifest/main/install.sh | bash -s -- --uninstall"
        echo ""
        echo "Options:"
        echo "  --help, -h        Show this help message"
        echo "  --uninstall, -u   Uninstall miko-manifest from the system"
        echo ""
        echo "Environment Variables:"
        echo "  INSTALL_DIR       Custom installation directory (default: /usr/local/bin or ~/.local/bin)"
        echo ""
        echo "Examples:"
        echo "  # Install latest version"
        echo "  curl -sSL https://raw.githubusercontent.com/jepemo/miko-manifest/main/install.sh | bash"
        echo ""
        echo "  # Uninstall"
        echo "  curl -sSL https://raw.githubusercontent.com/jepemo/miko-manifest/main/install.sh | bash -s -- --uninstall"
        echo ""
        exit 0
        ;;
    "--uninstall"|"-u")
        uninstall
        ;;
    "")
        main
        ;;
    *)
        print_error "Unknown option: $1. Use --help for usage information."
        ;;
esac
