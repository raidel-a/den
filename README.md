# Den
A terminal-based project manager with a [charming](https://github.com/charmbracelet) interface.

## Current Status
Den is currently in active development with a functional core feature set.

### Installation
1. Clone the repository
2. Make sure you have Go installed
3. Build and install:
   ```bash
   go build
   ./den --install   # Install shell completions and man pages
   ```

For development:
```bash
# Using air for hot reload
./dev.sh
```

### Currently Implemented Features

#### Core Functionality
- ✅ Interactive TUI using Bubbletea
- ✅ Project scanning and detection
- ✅ Project list navigation
- ✅ Fuzzy search/filtering
- ✅ Git status integration
- ✅ Project caching system
- ✅ Shell completions (Bash, Zsh, Fish)
- ✅ Man page documentation

#### Project Actions
- ✅ Open in editor (configurable)
- ✅ Open in file explorer
- ✅ Change working directory to project
- ✅ Copy project path

#### Configuration
- ✅ Configurable project directories
- ✅ Custom editor preferences
- ✅ Theme support (5 themes available):
  - Default
  - Dracula
  - Nord
  - Gruvbox
  - Solarized
- ✅ Persistent configuration in `~/.config/den/config.yaml`
- ✅ Project cache in `~/.cache/den/projects.json`

#### UI Features
- ✅ Vim-style navigation
- ✅ Context menu
- ✅ Interactive filtering
- ✅ Tab completion for paths
- ✅ Status messages
- ✅ Themed interface

### Command Line Interface
```
den [flags]

Flags:
    -h, --help      Show help information
    -v, --version   Display version information
    --reset         Reset all configuration and start fresh
    --debug         Enable debug logging
    --install       Install shell completions and man pages
```

#### Shell Completion
Den provides shell completion support for:
- Bash: Installed to `~/.bash_completion.d/den`
- Zsh: Installed to `~/.zsh/completion/_den`
- Fish: Installed to `~/.config/fish/completions/den.fish`

To install completions and man pages:
```bash
den --install
```

#### Man Page
After installation, view the man page with:
```bash
man den
```

### Test Environment
The project includes a comprehensive test setup script (`test-setup.sh`) that creates a realistic development environment:

#### Test Projects Created
- Go API project with module initialization
- Node.js web application
- Python data analyzer with dependencies
- Rust CLI tool with Cargo setup
- Project with modified files (for testing git state)
- Non-git project

#### Features Tested
- Git integration (clean/dirty/no-git states)
- Multiple project types detection
- Configuration system
- Project scanning
- Directory structure handling
- Shell completion functionality
- Man page installation

To set up the test environment:
```bash
./test-setup.sh
```
This will create a structured test environment in `~/projects/den-test` with various project types and states.

### Planned Features
- [ ] Project templates
- [ ] Grid/Tree views
- [ ] Project statistics
- [ ] README preview
- [ ] Dependency analysis
- [ ] Project health checks
- [ ] Bulk operations
- [ ] Project tags/categories
- [ ] Favorites system
- [ ] Enhanced Git integration

## Contributing
The project is under active development. Feel free to contribute by:
1. Opening issues for bugs or feature requests
2. Submitting pull requests
3. Improving documentation

## Development Setup
1. Clone the repository
2. Install Air for hot reload (optional):
   ```bash
   go install github.com/cosmtrek/air@latest
   ```
3. Run the development script:
   ```bash
   ./dev.sh
   ```

## Linux specific

### Clipboard
For Linux users, you'll need either xsel or xclip installed for clipboard support:
```bash
# Debian/Ubuntu
sudo apt-get install xsel

# or
sudo apt-get install xclip
```