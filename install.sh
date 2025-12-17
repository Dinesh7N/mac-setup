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

echo "Fetching latest release..."
LATEST=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [[ -z "${LATEST}" ]]; then
  echo -e "${RED}Error: Could not fetch latest release${NC}"
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${LATEST}/macsetup-darwin-arm64"
BINARY="/tmp/macsetup"

echo "Downloading macsetup ${LATEST}..."
curl -sL "$URL" -o "$BINARY"
chmod +x "$BINARY"

echo -e "${GREEN}✓${NC} Downloaded successfully"
echo ""

exec "$BINARY"

