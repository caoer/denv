# denv - Development Environment Manager

A zero-configuration development environment manager that provides automatic port isolation and environment variable management for your projects.

## Features

- **Zero Configuration**: Works immediately without setup
- **Automatic Port Remapping**: Prevents port conflicts between environments
- **Environment Variable Overrides**: Smart pattern-based variable management
- **Project Detection**: Automatic via git remote or folder name
- **Session Management**: Multiple terminals work correctly with file locks
- **Git Worktree Aware**: Same project = same environment pool
- **Non-Invasive**: Never modifies your project files

## Installation

```bash
go install github.com/zitao/denv/cmd/denv@latest
```

Or build from source:

```bash
git clone https://github.com/zitao/denv.git
cd denv
make build
```

## Quick Start

```bash
# Enter default environment
denv enter

# Enter named environment
denv enter feature-x

# List environments
denv ls

# Clean up environment
denv clean feature-x

# Manage sessions
denv sessions           # List active sessions
denv sessions --cleanup # Clean orphaned sessions
denv sessions --kill    # Terminate all sessions
```

## How It Works

When you enter a denv environment:

1. **Project Detection**: Automatically detects your project name from git remote or folder name
2. **Port Assignment**: Assigns unique ports to prevent conflicts (e.g., 3000 → 33000)
3. **Variable Overrides**: Applies pattern-based rules to environment variables
4. **Session Creation**: Creates a new shell session with modified environment

## Environment Variables

Inside a denv session, you have access to:

- `DENV_HOME`: Base directory (~/.denv)
- `DENV_ENV`: Current environment directory
- `DENV_PROJECT`: Shared project directory
- `DENV_ENV_NAME`: Environment name
- `DENV_PROJECT_NAME`: Project name
- `DENV_SESSION`: Session ID
- `PORT_*`: Remapped ports (e.g., PORT_3000=33000)

## Configuration

Global configuration in `~/.denv/config.yaml`:

```yaml
# Project name overrides
projects:
  /path/to/project: custom-name

# Pattern-based rules
patterns:
  "*_PORT|PORT":
    action: random_port
    range: [30000, 39999]
  
  "*_URL|*_URI|DATABASE_URL":
    action: rewrite_ports
  
  "*_KEY|*_SECRET":
    action: keep
```

## Project Structure

```
~/.denv/
├── myproject-default/      # Environment directory
│   ├── runtime.json        # Current state
│   └── sessions/           # Active session locks
├── myproject-feature/      # Another environment
├── myproject/              # Shared project directory
│   └── hooks/
│       ├── on-enter.sh    # Run on session start
│       └── on-exit.sh     # Run on session end
└── config.yaml            # Global configuration
```

## Development

```bash
# Run tests
make test

# Watch tests
make test-watch

# Build
make build
```

## Testing

The project follows Test-Driven Development with comprehensive test coverage:

- Unit tests for each component
- Integration tests for workflows
- Session management tests
- Port allocation tests
- Pattern matching tests

## Architecture

The project is organized into focused packages:

- `internal/paths`: Path management utilities
- `internal/config`: Configuration loading
- `internal/project`: Project detection
- `internal/environment`: Runtime state management
- `internal/ports`: Port allocation and tracking
- `internal/override`: Variable override system
- `internal/session`: Session management
- `internal/shell`: Shell wrapper generation
- `internal/commands`: CLI command implementations

## License

MIT