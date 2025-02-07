#!/bin/bash

# Set up colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Testing den's 'go to' functionality...${NC}"

# Create a test project
TEST_DIR=~/projects/den-test-goto
echo -e "${GREEN}Creating test project in ${TEST_DIR}...${NC}"

# Clean up existing test directory if it exists
rm -rf "${TEST_DIR}"

# Create test project
mkdir -p "${TEST_DIR}"
cd "${TEST_DIR}"
echo "# Test Project" > README.md
git init
git add README.md
git commit -m "Initial commit"

# Create den config pointing to test directory
mkdir -p ~/.config/den
cat > ~/.config/den/config.json << EOL
{
    "projectDirs": ["${TEST_DIR}"],
    "preferences": {
        "defaultEditor": "code",
        "editorList": ["code", "vim", "nano"],
        "showHiddenFiles": false,
        "showGitStatus": true,
        "theme": "default",
        "projectListTitle": "Test Projects"
    }
}
EOL

# Build den if needed
cd -
if [ ! -f "./den" ]; then
    echo -e "${GREEN}Building den...${NC}"
    go build
fi

# Ensure shell integration is installed
./den --install

# Test the go to functionality
echo -e "\n${GREEN}Testing 'go to' functionality...${NC}"
echo "Current directory: $(pwd)"
echo -e "Running den with simulated context menu selection...\n"

# The command below simulates:
# 1. Opening den
# 2. Selecting the first project
# 3. Opening context menu (c)
# 4. Selecting "Change Working Directory" (first option)
# 5. Pressing enter
expect << 'EOF'
spawn ./den
sleep 1
send "c"
sleep 0.5
send "\r"
expect eof
EOF

# Check if directory changed
echo -e "\n${GREEN}Test Results:${NC}"
echo "Previous directory: $(pwd)"
echo "Target directory: ${TEST_DIR}"
echo -e "\nTo manually test:"
echo "1. Run: ./den"
echo "2. Press 'c' to open context menu"
echo "3. Press enter to select 'Change Working Directory'"
echo "4. Check if your directory changed to: ${TEST_DIR}"

# Cleanup
echo -e "\n${GREEN}Cleaning up...${NC}"
rm -rf "${TEST_DIR}"
rm -rf ~/.config/den/config.json 