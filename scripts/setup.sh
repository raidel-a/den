#!/bin/bash

# Set up colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up Den dependencies...${NC}"

# Initialize go module if not already initialized
if [ ! -f "go.mod" ]; then
    echo "Initializing go module..."
    go mod init den
fi

# Function to check if a package is already installed
check_package() {
    local pkg=$1
    if go list -m $pkg &>/dev/null; then
        return 0 # Package exists
    else
        return 1 # Package doesn't exist
    fi
}

# Scan for imports in all .go files
echo -e "${BLUE}Scanning for imports...${NC}"
IMPORTS=$(find . -type f -name "*.go" -exec grep -h "^import" {} \; | grep -o '".*"' | tr -d '"' | sort -u)

# Get all found dependencies
if [ ! -z "$IMPORTS" ]; then
    echo -e "${BLUE}Checking dependencies:${NC}"
    while IFS= read -r pkg; do
        if [[ $pkg != "den/"* ]]; then  # Skip internal packages
            if check_package "$pkg"; then
                echo -e "${YELLOW}Already installed: $pkg${NC}"
            else
                echo -e "Installing: $pkg..."
                go get "$pkg"
            fi
        fi
    done <<< "$IMPORTS"
fi

# Check for updates to existing dependencies
echo -e "\n${BLUE}Checking for updates...${NC}"
go list -u -m all | grep '\[' || echo -e "${GREEN}All dependencies are up to date${NC}"

# Tidy up the modules
echo -e "\n${BLUE}Tidying modules...${NC}"
go mod tidy

echo -e "\n${GREEN}Setup complete!${NC}"

# Show final dependencies
echo -e "${BLUE}Current dependencies:${NC}"
go list -m all | grep -v "den$"