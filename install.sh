#!/bin/bash

# LAN Relay Installation Script

set -e

# Configuration
BINARY_NAME="lan-relay"
INSTALL_DIR="/usr/local/bin"
TEMP_DIR="/tmp/lan-relay-install"
REPO_URL="https://github.com/itsniaz/proxino"

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
        echo "Choose installation method:"
        echo -e "  ${BLUE}1)${NC} Install to ~/.local/bin and add to PATH (user-only)"
        echo -e "  ${BLUE}2)${NC} Install to /usr/local/bin (system-wide, requires sudo)"
        echo -e "  ${BLUE}3)${NC} Cancel installation"
        echo ""
        read -p "Enter your choice (1-3): " choice
        
        case $choice in
            1)
                INSTALL_DIR="$HOME/.local/bin"
                mkdir -p "$INSTALL_DIR"
                echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.bashrc"
                echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.zshrc" 2>/dev/null || true
                echo -e "${GREEN}✅ Added $INSTALL_DIR to PATH in shell profiles${NC}"
                echo -e "${YELLOW}ℹ️  You may need to restart your shell or run: source ~/.zshrc${NC}"
                ;;
            2)
                INSTALL_DIR="/usr/local/bin"
                echo -e "${YELLOW}ℹ️  System-wide installation requires sudo privileges${NC}"
                ;;
            3|*)
                echo -e "${RED}❌ Installation cancelled${NC}"
                exit 1
                ;;
        esac
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
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Failed to clone repository. Please check your internet connection.${NC}"
        echo "Repository URL: $REPO_URL"
        exit 1
    fi
    
    # Build frontend
    echo "🎨 Building frontend..."
    if [ ! -d "frontend" ]; then
        echo -e "${RED}❌ Frontend directory not found. Repository may be incomplete.${NC}"
        exit 1
    fi
    cd frontend
    npm install
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Frontend dependency installation failed${NC}"
        exit 1
    fi
    npm run build
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Frontend build failed${NC}"
        exit 1
    fi
    cd ..
    
    # Copy frontend build to backend
    echo "📦 Copying frontend build files..."
    if [ ! -d "frontend/build" ]; then
        echo -e "${RED}❌ Frontend build directory not found${NC}"
        exit 1
    fi
    cp -r frontend/build backend/static
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Failed to copy frontend build files${NC}"
        exit 1
    fi
    
    # Build backend
    echo "🚀 Building backend..."
    if [ ! -d "backend" ]; then
        echo -e "${RED}❌ Backend directory not found. Repository may be incomplete.${NC}"
        exit 1
    fi
    cd backend
    go mod tidy
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Go module tidy failed${NC}"
        exit 1
    fi
    go build -ldflags "-w -s" -o "$BINARY_NAME" .
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Backend build failed${NC}"
        exit 1
    fi
    
    # Copy binary to install directory
    echo "📦 Installing binary to $INSTALL_DIR..."
    if [[ "$INSTALL_DIR" == "$HOME"* ]]; then
        # User directory - no sudo needed
        cp "$BINARY_NAME" "$INSTALL_DIR/"
        if [ $? -ne 0 ]; then
            echo -e "${RED}❌ Failed to copy binary to $INSTALL_DIR${NC}"
            exit 1
        fi
    else
        # System directory - use sudo
        sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
        if [ $? -ne 0 ]; then
            echo -e "${RED}❌ Failed to copy binary to $INSTALL_DIR (sudo required)${NC}"
            exit 1
        fi
    fi
}

# Check if we can download a pre-built binary (when releases are available)
# For now, we'll build from source
echo -e "${YELLOW}ℹ️  Pre-built binaries not yet available, building from source...${NC}"
build_from_source

# Make binary executable
echo "🔧 Setting executable permissions..."
if [[ "$INSTALL_DIR" == "$HOME"* ]]; then
    # User directory - no sudo needed
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Failed to set executable permissions${NC}"
        exit 1
    fi
else
    # System directory - use sudo
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ Failed to set executable permissions (sudo required)${NC}"
        exit 1
    fi
fi

# Cleanup
cd /
rm -rf "$TEMP_DIR"

# Verify installation
echo "🔍 Verifying installation..."

# For user installation, we may need to update PATH for this session
if [[ "$INSTALL_DIR" == "$HOME"* ]]; then
    export PATH="$PATH:$INSTALL_DIR"
fi

if command -v "$BINARY_NAME" &> /dev/null; then
    # Test that the binary actually works
    echo "🧪 Testing installed binary..."
    if "$BINARY_NAME" version &> /dev/null; then
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
        echo -e "${RED}❌ Binary installed but not working correctly${NC}"
        echo "Try running: $INSTALL_DIR/$BINARY_NAME version"
        exit 1
    fi
else
    echo -e "${RED}❌ Installation failed - binary not found in PATH${NC}"
    if [[ "$INSTALL_DIR" == "$HOME"* ]]; then
        echo "The binary was installed to $INSTALL_DIR"
        echo "You may need to restart your shell or run:"
        echo -e "  ${BLUE}source ~/.zshrc${NC}  # or source ~/.bashrc"
        echo -e "  ${BLUE}export PATH=\"\$PATH:$INSTALL_DIR\"${NC}  # for current session"
    else
        echo "The binary was installed to $INSTALL_DIR"
        echo "This directory should already be in your PATH."
        echo "Try running: $INSTALL_DIR/$BINARY_NAME version"
    fi
    exit 1
fi 