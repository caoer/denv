# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`denv` is a zero-configuration development environment manager written in Go that provides automatic port isolation and environment variable management for development projects. It prevents port conflicts and environment variable collisions when working on multiple projects.

## Build and Test Commands

```bash
# Run all tests
make test

# Run tests in watch mode  
make test-watch

# Build the denv binary
make build

# Run a specific test
go test -v ./internal/commands -run TestEnterCommand

# Run integration tests
go test -v ./cmd/denv -run TestIntegration
```

## Architecture

The codebase follows a clean modular architecture organized under `/internal`:

### Core Packages

- **`cmd/denv/main.go`**: Entry point and command routing
- **`internal/commands/`**: CLI command implementations
  - `enter.go`: Core environment entry logic - orchestrates all subsystems
  - `list.go`, `clean.go`, `sessions.go`, `export.go`, `project.go`: Other commands
- **`internal/environment/`**: Runtime state management (Runtime struct in runtime.go)
- **`internal/ports/`**: Port allocation and mapping (30000-39999 range)
- **`internal/override/`**: Pattern-based environment variable transformations
- **`internal/project/`**: Git-based project detection
- **`internal/session/`**: Session management with file locking
- **`internal/shell/`**: Multi-shell support (bash/zsh/fish/sh)
- **`internal/config/`**: YAML configuration loading
- **`internal/paths/`**: Path utilities for `.denv` directory structure

### Key Data Flow

1. **Project Detection**: Extract name from git remote or folder
2. **Environment Creation**: Allocate ports, apply variable rules
3. **Session Management**: Create locked session with unique ID
4. **Shell Integration**: Generate wrapper script for chosen shell
5. **Cleanup**: Remove locks and orphaned sessions on exit

### File System Layout

```
~/.denv/
├── config.yaml                    # Global configuration
├── <project>-<environment>/       # Environment directory
│   ├── runtime.json              # Runtime state
│   └── sessions/                 # Session lock files
└── <project>/                    # Shared project directory
    └── hooks/                    # Entry/exit hooks
```

Project root gets:
```
.denv/
├── current -> ~/.denv/<project>-<environment>  # Current environment symlink
└── project -> ~/.denv/<project>               # Shared directory symlink
```

## Testing Approach

- Unit tests alongside each package (`*_test.go`)
- Integration tests in `cmd/denv/integration_test.go`
- Test utilities in `internal/testutil/helpers.go`
- Use `testify` for assertions

## Key Implementation Details

- **Port Management**: Persists mappings in `ports.json`, verifies availability before assignment
- **Session Locking**: File-based locks prevent concurrent modifications
- **Pattern Matching**: Glob patterns for environment variable rules (see `internal/override/`)
- **Shell Wrappers**: Generated dynamically based on detected shell type
- **Error Handling**: Commands return structured errors with user-friendly messages