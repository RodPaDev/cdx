#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Installing CDX (Change Directory Xplorer)...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed. Please install Go before continuing.${NC}"
    exit 1
fi

# Create installation directory
INSTALL_DIR="$HOME/.cdx"
BIN_DIR="$INSTALL_DIR/bin"
mkdir -p "$BIN_DIR"

# Clone or update repository
if [ -d "$INSTALL_DIR/src" ]; then
    echo -e "${BLUE}Updating CDX...${NC}"
    cd "$INSTALL_DIR/src"
    git pull
else
    echo -e "${BLUE}Downloading CDX...${NC}"
    mkdir -p "$INSTALL_DIR/src"
    git clone https://github.com/RodPaDev/cdx.git "$INSTALL_DIR/src"
    cd "$INSTALL_DIR/src"
fi

# Build the application
echo -e "${BLUE}Building CDX...${NC}"
go build -o "$BIN_DIR/cdx"

# Make executable
chmod +x "$BIN_DIR/cdx"

# Add to PATH
PROFILE_FILE=""
if [ -f "$HOME/.zshrc" ]; then
    PROFILE_FILE="$HOME/.zshrc"
elif [ -f "$HOME/.bashrc" ]; then
    PROFILE_FILE="$HOME/.bashrc"
elif [ -f "$HOME/.bash_profile" ]; then
    PROFILE_FILE="$HOME/.bash_profile"
else
    PROFILE_FILE="$HOME/.profile"
fi

# Check if already in PATH
if ! grep -q "export PATH=\"\$HOME/.cdx/bin:\$PATH\"" "$PROFILE_FILE"; then
    echo -e "\n# CDX Path" >> "$PROFILE_FILE"
    echo "export PATH=\"\$HOME/.cdx/bin:\$PATH\"" >> "$PROFILE_FILE"
    echo -e "${GREEN}Added CDX to PATH in $PROFILE_FILE${NC}"
    echo -e "${BLUE}Please run: source $PROFILE_FILE${NC}"
else
    echo -e "${GREEN}CDX is already in PATH${NC}"
fi

echo -e "${GREEN}CDX installed successfully!${NC}"
echo -e "${BLUE}Run 'cdx' to start the application${NC}"