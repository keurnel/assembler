#!/bin/bash

# Set strict mode
#
set -euo pipefail

# Color codes
#
NC='\033[0m'
RED='\033[0;31m'

REQUIRED_GO_VERSION="go1.25.6 linux/amd64"

# Is go installed?
#
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: go is not installed.${NC}"
    exit 1
fi

# Go should equal `go version go1.25.6 linux/amd64`
# in order to build for x86_64
#
if [[ "$(go version)" != "$REQUIRED_GO_VERSION" ]]; then
    echo -e "${RED}Error: go version must be '$REQUIRED_GO_VERSION' to build for x86_64.${NC}"
    echo -e "${RED}Current go version: $(go version)${NC}"
    exit 1
fi

# Build the assembler
#
