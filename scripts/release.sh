#!/bin/bash
set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check if version is provided
if [ $# -eq 0 ]; then
    echo -e "${RED}Error: Version tag required${NC}"
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.0.2"
    exit 1
fi

VERSION=$1

# Validate version format (should start with 'v')
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${YELLOW}Warning: Version should follow format vX.Y.Z (e.g., v1.0.0)${NC}"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo -e "${BLUE}Creating release ${VERSION}${NC}"
echo ""

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo -e "${RED}Error: Tag ${VERSION} already exists${NC}"
    echo "To re-release, delete the tag first:"
    echo "  git tag -d ${VERSION}"
    echo "  git push origin :refs/tags/${VERSION}"
    exit 1
fi

# Check for uncommitted changes
if [[ -n $(git status -s) ]]; then
    echo -e "${YELLOW}Warning: You have uncommitted changes${NC}"
    git status -s
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Build the binary
echo -e "${BLUE}Step 1/4: Building binary${NC}"
BINARY_NAME="macsetup-darwin-arm64"
go build -ldflags "-s -w -X main.version=${VERSION}" -o "bin/${BINARY_NAME}" .
echo -e "${GREEN}✓${NC} Binary built: bin/${BINARY_NAME}"
echo ""

# Create git tag
echo -e "${BLUE}Step 2/4: Creating git tag${NC}"
git tag -a "${VERSION}" -m "Release ${VERSION}"
echo -e "${GREEN}✓${NC} Tag created: ${VERSION}"
echo ""

# Push tag to remote
echo -e "${BLUE}Step 3/4: Pushing tag to remote${NC}"
git push origin "${VERSION}"
echo -e "${GREEN}✓${NC} Tag pushed to origin"
echo ""

# Upload to GitHub releases
echo -e "${BLUE}Step 4/4: Uploading to GitHub releases${NC}"
gh release upload "${VERSION}" "bin/${BINARY_NAME}" --clobber
echo -e "${GREEN}✓${NC} Binary uploaded to GitHub release"
echo ""

echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ Release ${VERSION} completed successfully!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "View the release at:"
echo "https://github.com/Dinesh7N/mac-setup/releases/tag/${VERSION}"
echo ""
echo "Users can now install with:"
echo -e "${BLUE}/bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Dinesh7N/mac-setup/main/install.sh)\"${NC}"
