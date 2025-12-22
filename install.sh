#!/bin/bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}macsetup${NC}"
echo "Team Mac Onboarding Tool"
echo ""

if [[ "$(uname)" != "Darwin" ]]; then
  echo -e "${RED}Error: This script only works on macOS${NC}"
  exit 1
fi

if [[ "$(uname -m)" != "arm64" ]]; then
  echo -e "${RED}Error: This script only supports Apple Silicon Macs (arm64)${NC}"
  exit 1
fi

echo -e "${GREEN}✓${NC} macOS on Apple Silicon detected"

REPO="Dinesh7N/mac-setup"

if [[ -z "${VERSION:-}" ]]; then
  echo "Fetching latest release..."
  VERSION=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4 || true)
  if [[ -z "${VERSION}" ]]; then
    echo -e "${RED}Error: Could not fetch latest release (Repository might be private or have no releases)${NC}"
    exit 1
  fi
else
  echo "Using specified version: ${VERSION}"
fi

URL="https://github.com/${REPO}/releases/download/${VERSION}/macsetup-darwin-arm64"
BINARY="/tmp/macsetup"

echo "Downloading macsetup ${VERSION}..."
if ! curl -sLf "$URL" -o "$BINARY"; then
    echo -e "${RED}Error: Failed to download version ${VERSION}. Please ensure the version exists.${NC}"
    exit 1
fi
chmod +x "$BINARY"

# Remove quarantine attribute to prevent Gatekeeper from blocking execution
xattr -d com.apple.quarantine "$BINARY" 2>/dev/null || true

echo -e "${GREEN}✓${NC} Downloaded successfully"
echo ""

LOG_FILE="/tmp/macsetup.log"
echo "Starting setup... Logs will be available at $LOG_FILE"
echo ""

# Execute with logging
if ! "$BINARY" --log-file "$LOG_FILE"; then
    echo ""
    echo -e "${RED}Installation failed.${NC}"
    echo "Please check the log file for details: $LOG_FILE"
    echo "Report issues at: https://github.com/Dinesh7N/mac-setup/issues"
    exit 1
fi

