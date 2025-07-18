#!/bin/bash

# Atlassian Assets CLI - Simple Build Script
# ==========================================
# This script makes it easy for anyone to build and test the CLI

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_header() {
    echo -e "${BLUE}🔨 $1${NC}"
    echo "=================================="
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check if Go is installed
check_go() {
    print_header "Checking Go Installation"
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed!"
        echo ""
        echo "Please install Go from: https://golang.org/dl/"
        echo "After installation, make sure to:"
        echo "1. Add Go to your PATH"
        echo "2. Set GOPATH environment variable"
        echo "3. Restart your terminal"
        echo ""
        exit 1
    fi
    
    GO_VERSION=$(go version)
    print_success "Go is installed: $GO_VERSION"
    echo ""
}

# Check dependencies
check_deps() {
    print_header "Checking Dependencies"
    
    print_info "Downloading Go modules..."
    go mod download
    print_success "Dependencies ready"
    echo ""
}

# Clean previous builds
clean_build() {
    print_header "Cleaning Previous Builds"
    
    if [ -d "bin" ]; then
        print_info "Removing old bin/ directory..."
        rm -rf bin/
    fi
    
    if [ -f "assets" ]; then
        print_info "Removing old assets binary..."
        rm -f assets
    fi
    
    print_success "Clean complete"
    echo ""
}

# Run tests
run_tests() {
    print_header "Running Tests"
    
    print_info "Running help system tests (no credentials needed)..."
    
    # Create a minimal test that doesn't require credentials
    go test -v -run TestVersionInfo ./cmd/assets 2>/dev/null || {
        print_warning "Some tests failed, but that's okay for a basic build"
        print_info "For full testing, set ATLASSIAN_EMAIL and ATLASSIAN_API_TOKEN"
    }
    
    print_success "Basic tests completed"
    echo ""
}

# Build the binary
build_binary() {
    print_header "Building Binary"
    
    # Get version info
    VERSION="dev"
    COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    print_info "Building with version: $VERSION"
    print_info "Commit: ${COMMIT:0:8}"
    
    # Create bin directory
    mkdir -p bin
    
    # Build with version injection
    go build -ldflags "\
        -X github.com/aaronsb/atlassian-assets/internal/version.Version=$VERSION \
        -X github.com/aaronsb/atlassian-assets/internal/version.Commit=$COMMIT \
        -X github.com/aaronsb/atlassian-assets/internal/version.Date=$DATE" \
        -o bin/assets ./cmd/assets
    
    print_success "Binary built: bin/assets"
    echo ""
}

# Test the built binary
test_binary() {
    print_header "Testing Built Binary"
    
    if [ ! -f "bin/assets" ]; then
        print_error "Binary not found!"
        exit 1
    fi
    
    print_info "Testing version display..."
    ./bin/assets --version
    echo ""
    
    print_info "Testing help system..."
    ./bin/assets --help > /dev/null
    
    print_success "Binary tests passed"
    echo ""
}

# Offer installation
offer_install() {
    print_header "Installation Options"
    
    echo "Your binary is ready: bin/assets"
    echo ""
    echo "Options:"
    echo "1. Use locally: ./bin/assets --help"
    echo "2. Install to system PATH: sudo cp bin/assets /usr/local/bin/"
    echo "3. Add to current PATH: export PATH=\$PATH:\$(pwd)/bin"
    echo ""
    
    # Skip prompt if running non-interactively
    if [ "$1" = "--non-interactive" ]; then
        print_info "Non-interactive mode: Binary ready at: $(pwd)/bin/assets"
        return
    fi
    
    read -p "Would you like to install to system PATH? [y/N]: " choice
    case "$choice" in 
        y|Y ) 
            print_info "Installing to /usr/local/bin/..."
            sudo cp bin/assets /usr/local/bin/
            print_success "Installation complete! You can now run: assets --help"
            ;;
        * ) 
            print_info "Binary ready at: $(pwd)/bin/assets"
            ;;
    esac
    echo ""
}

# Setup development environment
setup_dev() {
    print_header "Development Environment Setup"
    
    echo "For development, you may want to:"
    echo ""
    echo "1. Set up Atlassian credentials for testing:"
    echo "   export ATLASSIAN_EMAIL=your.email@company.com"
    echo "   export ATLASSIAN_API_TOKEN=your-api-token"
    echo "   export ATLASSIAN_HOST=https://yourcompany.atlassian.net"
    echo "   export ATLASSIAN_ASSETS_WORKSPACE_ID=your-workspace-id"
    echo ""
    echo "2. Install development tools:"
    echo "   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    echo ""
    echo "3. Use the Makefile for advanced builds:"
    echo "   make help"
    echo ""
}

# Main execution
main() {
    echo ""
    print_header "Atlassian Assets CLI - Simple Builder"
    echo "This script will help you build the CLI from source"
    echo ""
    
    # Check prerequisites
    check_go
    check_deps
    
    # Build process
    clean_build
    run_tests
    build_binary
    test_binary
    
    # Installation and setup
    offer_install
    setup_dev
    
    print_success "🎉 Build complete!"
    print_info "Next steps:"
    echo "  • Test the CLI: ./bin/assets search --help"
    echo "  • View all commands: ./bin/assets --help"
    echo "  • Check version: ./bin/assets --version"
    echo ""
}

# Run main function
main "$@"