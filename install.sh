#!/bin/bash

# LAN Relay Installation Script

set -e

# Configuration
BINARY_NAME="lan-relay"
INSTALL_DIR="/usr/local/bin"
TEMP_DIR="/tmp/lan-relay-install"
REPO_URL="https://github.com/yourusername/local_router"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    armv7l)
        ARCH="arm"
        ;;
    *)
        echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

case $OS in
    linux|darwin)
        ;;
    *)
        echo -e "${RED}❌ Unsupported operating system: $OS${NC}"
        exit 1
        ;;
esac

echo -e "${BLUE}🛰️  LAN Relay Installer${NC}"
echo "=================================="
echo -e "Detected OS: ${YELLOW}$OS${NC}"
echo -e "Detected Architecture: ${YELLOW}$ARCH${NC}"
echo ""

# Check if running as root for system-wide installation
if [[ $EUID -eq 0 ]]; then
    INSTALL_DIR="/usr/local/bin"
    echo -e "${YELLOW}⚠️  Running as root - installing system-wide${NC}"
else
    # Try to install in user's local bin
    if [[ ":$PATH:" == *":$HOME/.local/bin:"* ]]; then
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
    elif [[ ":$PATH:" == *":$HOME/bin:"* ]]; then
        INSTALL_DIR="$HOME/bin"
        mkdir -p "$INSTALL_DIR"
    else
        echo -e "${YELLOW}⚠️  Local bin directory not in PATH${NC}"
        echo -e "Would you like to install to ${BLUE}$HOME/.local/bin${NC} and add it to PATH? (y/n)"
        read -r response
        if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            INSTALL_DIR="$HOME/.local/bin"
            mkdir -p "$INSTALL_DIR"
            echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.bashrc"
            echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.zshrc" 2>/dev/null || true
            echo -e "${GREEN}✅ Added $INSTALL_DIR to PATH in shell profiles${NC}"
        else
            echo -e "${RED}❌ Installation cancelled${NC}"
            exit 1
        fi
    fi
fi

echo -e "Installing to: ${BLUE}$INSTALL_DIR${NC}"
echo ""

# Create temporary directory
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"
cd "$TEMP_DIR"

# Function to build from source if needed
build_from_source() {
    echo -e "${YELLOW}📦 Building from source...${NC}"
    
    # Check for required tools
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is required to build from source${NC}"
        echo "Please install Go from https://golang.org/doc/install"
        exit 1
    fi
    
    if ! command -v node &> /dev/null; then
        echo -e "${RED}❌ Node.js is required to build the frontend${NC}"
        echo "Please install Node.js from https://nodejs.org/"
        exit 1
    fi
    
    # Clone the repository
    echo "📥 Cloning repository..."
    git clone "$REPO_URL" .
    
    # Build frontend
    echo "🎨 Building frontend..."
    cd frontend
    npm install
    npm run build
    cd ..
    
    # Copy frontend build to backend
    cp -r frontend/build backend/static
    
    # Build backend
    echo "🚀 Building backend..."
    cd backend
    go mod tidy
    go build -ldflags "-w -s" -o "$BINARY_NAME" .
    
    # Copy binary to install directory
    if [[ $EUID -eq 0 ]] || [[ -w "$INSTALL_DIR" ]]; then
        cp "$BINARY_NAME" "$INSTALL_DIR/"
    else
        sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
    fi
}

# Check if we can download a pre-built binary (when releases are available)
# For now, we'll build from source
echo -e "${YELLOW}ℹ️  Pre-built binaries not yet available, building from source...${NC}"
build_from_source

# Make binary executable
if [[ $EUID -eq 0 ]] || [[ -w "$INSTALL_DIR" ]]; then
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Cleanup
cd /
rm -rf "$TEMP_DIR"

# Verify installation
if command -v "$BINARY_NAME" &> /dev/null; then
    echo ""
    echo -e "${GREEN}✅ LAN Relay installed successfully!${NC}"
    echo ""
    echo "Usage:"
    echo -e "  ${BLUE}$BINARY_NAME start${NC}    - Start the LAN Relay server"
    echo -e "  ${BLUE}$BINARY_NAME status${NC}   - Check if the service is running"
    echo -e "  ${BLUE}$BINARY_NAME --help${NC}   - Show all available commands"
    echo ""
    echo -e "Dashboard will be available at: ${BLUE}http://localhost:8080${NC}"
    echo ""
    echo "Run '$BINARY_NAME start' to get started!"
else
    echo -e "${RED}❌ Installation failed - binary not found in PATH${NC}"
    echo "You may need to restart your shell or run:"
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    exit 1
fi 