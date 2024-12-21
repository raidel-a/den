#!/bin/bash

# Set up colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Setting up Den test environment...${NC}"

# Clear existing config and cache
echo -e "${GREEN}Cleaning up existing configuration...${NC}"
rm -rf ~/.config/den
rm -rf ~/.cache/den

# Create test projects structure with various types of projects
TEST_ROOT=~/projects/den-test
echo -e "${GREEN}Creating test project structure in ${TEST_ROOT}...${NC}"

# Clean up existing test directory if it exists
rm -rf "${TEST_ROOT}"

# Create main test directories
mkdir -p "${TEST_ROOT}"/{go,node,python,rust,misc}

# Go project
echo -e "${GREEN}Setting up Go project...${NC}"
mkdir -p "${TEST_ROOT}/go/awesome-api"
cd "${TEST_ROOT}/go/awesome-api"
go mod init awesome-api
cat > main.go << EOL
package main

func main() {
    println("Hello from Go!")
}
EOL
git init
git add .
git commit -m "Initial Go project commit"

# Node.js project
echo -e "${GREEN}Setting up Node.js project...${NC}"
mkdir -p "${TEST_ROOT}/node/web-app"
cd "${TEST_ROOT}/node/web-app"
cat > package.json << EOL
{
  "name": "web-app",
  "version": "1.0.0",
  "description": "Test web application",
  "main": "index.js",
  "scripts": {
    "test": "echo \"Error: no test specified\" && exit 1"
  }
}
EOL
git init
git add .
git commit -m "Initial Node.js project commit"
echo "node_modules/" > .gitignore

# Python project
echo -e "${GREEN}Setting up Python project...${NC}"
mkdir -p "${TEST_ROOT}/python/data-analyzer"
cd "${TEST_ROOT}/python/data-analyzer"
cat > requirements.txt << EOL
pandas==2.0.0
numpy==1.24.0
EOL
cat > main.py << EOL
def main():
    print("Hello from Python!")

if __name__ == "__main__":
    main()
EOL
git init
git add .
git commit -m "Initial Python project commit"
echo "__pycache__/" > .gitignore

# Rust project
echo -e "${GREEN}Setting up Rust project...${NC}"
mkdir -p "${TEST_ROOT}/rust/cli-tool"
cd "${TEST_ROOT}/rust/cli-tool"
cat > Cargo.toml << EOL
[package]
name = "cli-tool"
version = "0.1.0"
edition = "2021"

[dependencies]
EOL
mkdir src
cat > src/main.rs << EOL
fn main() {
    println!("Hello from Rust!");
}
EOL
git init
git add .
git commit -m "Initial Rust project commit"
echo "target/" > .gitignore

# Create a project with modified files (dirty git state)
echo -e "${GREEN}Setting up project with modified files...${NC}"
mkdir -p "${TEST_ROOT}/misc/modified-project"
cd "${TEST_ROOT}/misc/modified-project"
echo "# Modified Project" > README.md
git init
git add .
git commit -m "Initial commit"
echo "Some changes" >> README.md

# Create a project without git
echo -e "${GREEN}Setting up non-git project...${NC}"
mkdir -p "${TEST_ROOT}/misc/no-git-project"
echo "# Project without Git" > "${TEST_ROOT}/misc/no-git-project/README.md"

# Create initial Den configuration
echo -e "${GREEN}Creating initial Den configuration...${NC}"
mkdir -p ~/.config/den
cat > ~/.config/den/config.json << EOL
{
    "projectDirs": ["${TEST_ROOT}"],
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

echo -e "${BLUE}Test environment setup complete!${NC}"
echo -e "${GREEN}Created test projects:${NC}"
echo "  - Go API (clean git)"
echo "  - Node.js Web App (clean git)"
echo "  - Python Data Analyzer (clean git)"
echo "  - Rust CLI Tool (clean git)"
echo "  - Modified Project (dirty git)"
echo "  - No Git Project"
echo
echo -e "${BLUE}You can now run Den to explore these test projects.${NC}"
echo "Run: go run main.go"
