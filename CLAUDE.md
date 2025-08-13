# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`denv` is a zero-configuration development environment manager written in Go that provides automatic port isolation and environment variable management for development projects. It prevents port conflicts and environment variable collisions when working on multiple projects.

## 🚨 CRITICAL: Test-Driven Development (TDD) REQUIRED 🚨

**ALL code work in this project MUST follow Test-Driven Development methodology. This is non-negotiable.**

### TDD Process (MANDATORY FOR EVERY FEATURE/FIX):

1. **Write failing test first** - Shows what we want to achieve
2. **Run test to see it fail** - Confirms test is actually testing something  
3. **Implement minimal code** - Just enough to make test pass
4. **Refactor if needed** - Clean up while tests protect us

⚠️ **DO NOT write implementation code before writing tests!**  
⚠️ **DO NOT skip the "see it fail" step - it validates your test!**  
⚠️ **DO NOT write more code than needed to pass the test!**

This approach ensures:
- We build exactly what's needed (no over-engineering)
- Tests actually test the right thing (validated by initial failure)
- Code is always covered by tests
- Refactoring is safe with test protection

**Remember: Red → Green → Refactor**

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

# Test final build in isolated directory (use random number to avoid conflicts)
# Example: builds to tmp/bin-12345/denv
go build -o tmp/bin-$(date +%s)/denv ./cmd/denv
# Then test the built binary:
./tmp/bin-*/denv enter
./tmp/bin-*/denv list
# etc.
```

### 🔨 Testing Built Binary

**IMPORTANT**: After making changes, always test the final built CLI binary to ensure it works as expected:

1. Build to an isolated temporary directory with a unique name to avoid conflicts:
   ```bash
   # Create unique directories for both binary and DENV_HOME
   TIMESTAMP="$(date +%s)"
   BUILD_DIR="tmp/bin-${TIMESTAMP}"
   TEST_DENV_HOME="tmp/denv-home-${TIMESTAMP}"
   
   mkdir -p "$BUILD_DIR"
   mkdir -p "$TEST_DENV_HOME"
   go build -o "$BUILD_DIR/denv" ./cmd/denv
   ```

2. Test the built binary with isolated DENV_HOME to avoid conflicts:
   ```bash
   # Export unique DENV_HOME for this test session
   export DENV_HOME="$TEST_DENV_HOME"
   
   # Test basic functionality with isolated environment
   "$BUILD_DIR/denv" --help
   "$BUILD_DIR/denv" enter
   "$BUILD_DIR/denv" list
   "$BUILD_DIR/denv" ps
   "$BUILD_DIR/denv" rm --all
   
   # Clean up when done (optional)
   unset DENV_HOME
   ```

3. Verify the binary works correctly in different scenarios before considering the work complete.

This ensures:
- The final deliverable actually works
- Testing doesn't interfere with other denv instances or user's ~/.denv
- Each test run is completely isolated with its own DENV_HOME
- Catches any build-time or runtime issues that unit tests might miss

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

### Testing Environment

For testing, you can override the default denv home directory using the `DENV_HOME` environment variable:

```bash
# Set DENV_HOME to tmp folder in project root for testing
export DENV_HOME="$(pwd)/tmp"

# Run tests with isolated environment
make test
```

This allows testing in an isolated environment without affecting the user's actual ~/.denv directory.

## Key Implementation Details

- **Port Management**: Persists mappings in `ports.json`, verifies availability before assignment
- **Session Locking**: File-based locks prevent concurrent modifications
- **Pattern Matching**: Glob patterns for environment variable rules (see `internal/override/`)
- **Shell Wrappers**: Generated dynamically based on detected shell type
- **Error Handling**: Commands return structured errors with user-friendly messages