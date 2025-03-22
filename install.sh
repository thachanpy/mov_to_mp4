#!/bin/bash

# Script to download and install the latest mov_to_mp4 release

# Define colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Define repository information
REPO="thachanpy/mov_to_mp4"
BINARY_NAME="mov_to_mp4"
INSTALL_DIR="/usr/local/bin"

echo -e "${YELLOW}Fetching the latest release information from GitHub...${NC}"

# Get the latest release URL using GitHub API
LATEST_RELEASE_URL=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" |
                     grep "browser_download_url.*$BINARY_NAME" |
                     cut -d '"' -f 4)

if [ -z "$LATEST_RELEASE_URL" ]; then
    echo -e "${RED}Error: Could not find the latest release download URL.${NC}"
    exit 1
fi

VERSION=$(echo $LATEST_RELEASE_URL | sed -E 's/.*\/v([0-9]+\.[0-9]+\.[0-9]+)\/.*/\1/')
echo -e "${GREEN}Found latest version: ${VERSION}${NC}"

# Create a temporary directory for downloading
TEMP_DIR=$(mktemp -d)
TEMP_FILE="$TEMP_DIR/$BINARY_NAME"

echo -e "${YELLOW}Downloading $BINARY_NAME v${VERSION}...${NC}"
curl -L "$LATEST_RELEASE_URL" -o "$TEMP_FILE"

if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Download failed.${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Make the binary executable
echo -e "${YELLOW}Making the binary executable...${NC}"
chmod +x "$TEMP_FILE"

# Check if file is already in /usr/local/bin and if we need sudo
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo -e "${YELLOW}Removing previous installation...${NC}"
    if [ -w "$INSTALL_DIR" ]; then
        rm "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo rm "$INSTALL_DIR/$BINARY_NAME"
    fi
fi

# Move the binary to /usr/local/bin
echo -e "${YELLOW}Moving binary to $INSTALL_DIR...${NC}"
if [ -w "$INSTALL_DIR" ]; then
    mv "$TEMP_FILE" "$INSTALL_DIR/"
else
    sudo mv "$TEMP_FILE" "$INSTALL_DIR/"
fi

if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to install the binary to $INSTALL_DIR.${NC}"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Clean up temporary directory
rm -rf "$TEMP_DIR"

echo -e "${GREEN}✅ Successfully installed mov_to_mp4 v${VERSION} to $INSTALL_DIR/$BINARY_NAME${NC}"
echo -e "${GREEN}You can now run it by typing 'mov_to_mp4' in your terminal.${NC}"

# Verify the installation
if command -v $BINARY_NAME &>/dev/null; then
    echo -e "${GREEN}Verification: $BINARY_NAME is in your PATH.${NC}"
else
    echo -e "${YELLOW}Note: You may need to restart your terminal or add $INSTALL_DIR to your PATH.${NC}"
fi
