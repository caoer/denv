# denv - Development Environment Manager

## Core Philosophy

**Zero configuration required.** Just `denv enter` and it works. Focus on environment isolation through automatic port remapping and variable overrides.

## Directory Structure

```
$DENV_HOME (~/.denv)/
├── myapp-default/
│   ├── runtime.json       # Current state and overrides
│   └── sessions/          # Active session locks
│       ├── abc123.lock
│       └── def456.lock
├── myapp-experiment/
│   ├── runtime.json
│   └── sessions/
├── myapp/                 # Shared project directory
│   └── hooks/
│       ├── on-enter.sh   # Per-session setup
│       └── on-exit.sh    # Per-session cleanup
└── config.yaml           # Global override rules
```

## Environment Variables

When inside a denv session:

```bash
# Core paths
DENV_HOME=/home/user/.denv                    # Base directory
DENV_ENV=/home/user/.denv/myapp-default       # Current environment directory
DENV_PROJECT=/home/user/.denv/myapp           # Shared project directory

# Metadata
DENV_ENV_NAME=default                         # Environment name
DENV_PROJECT_NAME=myapp                       # Project name  
DENV_SESSION=abc123                           # Session ID

# Port mappings (automatic)
PORT_3000=33000                               # Remapped ports
PORT_5432=35432
PORT_6379=36379
ORIGINAL_PORT_3000=3000                       # Original values preserved
ORIGINAL_PORT_5432=5432
ORIGINAL_PORT_6379=6379
```

## Project Detection

Automatic project detection (in order):
1. **Check override in config** → Look for path mapping in `$DENV_HOME/config.yaml`
2. **Git remote URL** → Extract repo name (works with git worktrees)
3. **Folder name** → If no git

```bash
# Git worktrees automatically share the same project
~/projects/myapp/          # git remote: github.com/user/myapp
~/projects/myapp-feature/  # Same remote = same project "myapp"

# First time ambiguous detection
$ cd ~/projects/client-project
$ denv enter
Project detected as 'client-project'. Is this correct? (y/n/rename): rename
Enter project name: acme-web
Saved override to ~/.denv/config.yaml
```

## Pattern-Based Configuration

`~/.denv/config.yaml` - Global configuration and overrides:

```yaml
# Project name overrides (set via prompt or denv project command)
projects:
  /Users/me/projects/client-project: acme-web
  /Users/me/projects/another-client: acme-api

# Pattern-based environment variable rules
patterns:
  # Ports - always randomize
  "*_PORT|PORT":
    action: random_port
    range: [30000, 39999]
    
  # Directories - always isolate  
  "*_ROOT|*_DIR|*_PATH|*_HOME":
    action: isolate
    base: "${DENV_ENV}"
    
  # URLs - smart rewrite ports
  "*_URL|*_URI|*_ENDPOINT|DATABASE_URL|REDIS_URL":
    action: rewrite_ports
    
  # Secrets - never change
  "*_KEY|*_TOKEN|*_SECRET|*_PASSWORD|*_CREDENTIAL":
    action: keep
    
  # Hosts - keep if localhost
  "*_HOST|*_HOSTNAME":
    action: keep
    only_if: ["localhost", "127.0.0.1", "0.0.0.0"]
```

## Runtime State

`$DENV_ENV/runtime.json` - Shows what denv did:

```json
{
  "created": "2024-01-10T10:00:00Z",
  "project": "myapp",
  "environment": "default",
  
  "ports": {
    "3000": 33000,
    "5432": 35432,
    "6379": 36379
  },
  
  "overrides": {
    "DATABASE_URL": {
      "original": "postgres://localhost:5432/myapp",
      "current": "postgres://localhost:35432/myapp",
      "rule": "rewrite_ports"
    },
    "REDIS_PORT": {
      "original": "6379",
      "current": "36379",
      "rule": "random_port"
    },
    "DATA_ROOT": {
      "original": "/var/data",
      "current": "/home/user/.denv/myapp-default/data",
      "rule": "isolate"
    }
  },
  
  "sessions": {
    "abc123": {
      "pid": 12345,
      "started": "2024-01-10T10:00:00Z",
      "tty": "/dev/ttys001"
    }
  }
}
```

## Session Management

### File Lock Approach

Each session creates a lock file that OS automatically releases on process death:

```bash
$DENV_ENV/sessions/
├── abc123.lock  # Held by PID 12345
└── def456.lock  # Held by PID 12350
```

### Session Commands

```bash
denv sessions              # List active sessions
denv sessions --cleanup    # Remove orphaned locks and run cleanup hooks
denv sessions --kill       # Send SIGTERM to all sessions for graceful exit
```

### Signal Handling and Cleanup

When `denv enter` starts a session, it sets up signal handlers:

```bash
# Trap signals to ensure on-exit.sh runs
trap 'run_exit_hook' EXIT
trap 'run_exit_hook; exit' SIGINT SIGTERM

run_exit_hook() {
    # Load the environment state
    source $DENV_ENV/session_env
    
    # Run user's on-exit.sh hook
    if [ -f "$DENV_PROJECT/hooks/on-exit.sh" ]; then
        source "$DENV_PROJECT/hooks/on-exit.sh"
    fi
    
    # Release lock file
    rm -f "$DENV_ENV/sessions/$DENV_SESSION.lock"
    
    # Update runtime.json
    update_session_state "exited"
}
```

When `denv sessions --kill` is run:
1. Sends SIGTERM to all session processes
2. Each session's trap handler runs `on-exit.sh`
3. Lock files are cleaned up
4. Runtime state is updated

For orphaned sessions (process died without cleanup):
```bash
denv sessions --cleanup
# Detects dead PIDs
# Runs on-exit.sh with last known environment
# Removes stale lock files
```

## Hook System

Hooks are created automatically on first use:

```bash
$DENV_PROJECT/hooks/
├── on-enter.sh    # Runs for every session enter
└── on-exit.sh     # Runs for every session exit
```

### Example Hooks

```bash
# on-enter.sh
echo "Entering $DENV_ENV_NAME environment"

# Create database if needed
createdb ${DENV_PROJECT_NAME}_${DENV_ENV_NAME} 2>/dev/null || true

# Set up additional variables
export CUSTOM_VAR="development"

# on-exit.sh  
echo "Exiting $DENV_ENV_NAME environment"

# Kill processes using our ports
for port in $PORT_3000 $PORT_3001; do
    lsof -ti:$port | xargs kill 2>/dev/null || true
done
```

## Commands

```bash
# Core commands
denv enter [name]      # Enter environment (default: "default")
denv exit             # Exit current environment
denv ls               # List environments for current project
denv clean [name]     # Remove environment

# Session management  
denv sessions         # Show active sessions
denv sessions --cleanup # Clean orphaned sessions
denv sessions --kill  # Gracefully terminate all sessions

# Project management
denv project          # Show current project name
denv project rename   # Rename current project
denv project unset    # Remove project override

# Utility
denv export           # Export variables (for direnv integration)
denv help            # Show help
```

## Project Integration

### Symlinks (auto-created)

The only thing denv creates in your project:

```bash
myproject/
└── .denv/                                    # Only this directory
    ├── current -> ~/.denv/myapp-default     # Points to $DENV_ENV
    └── project -> ~/.denv/myapp             # Points to $DENV_PROJECT
```

### Global gitignore

Add to `~/.gitignore_global`:
```bash
.denv/
```

### direnv Integration (optional)

Add to end of `.envrc`:
```bash
[ -n "$DENV_ENV" ] && eval "$(denv export)"
```

### Zero Project Pollution

denv **never** writes any configuration files to your project:
- No `.denv-name` files
- No `.denv.env` files  
- No `.denv.sh` files
- No `.gitignore` modifications

All configuration and customization happens in `$DENV_HOME` (`~/.denv/`).

## Workflow Examples

### Basic Usage

```bash
$ cd ~/projects/myapp
$ denv enter
Entering 'default' environment for myapp

Port mappings:
  PostgreSQL: 5432 → 35432
  Redis: 6379 → 36379
  Port 3000 → 33000

(default) $ echo $DATABASE_URL
postgresql://localhost:35432/myapp    # Automatically rewritten!

(default) $ npm run dev               # Uses PORT_3000=33000

(default) $ denv exit                 # or Ctrl+D
Exiting 'default' environment
```

### First Time Project Detection

```bash
$ cd ~/projects/client-work
$ denv enter
Project detected as 'client-work'. Is this correct? (y/n/rename): rename
Enter project name: acme-api
Saved to ~/.denv/config.yaml

Entering 'default' environment for acme-api
...

# Next time, it remembers
$ cd ~/projects/client-work
$ denv enter
Entering 'default' environment for acme-api
```

### Multiple Environments

```bash
# Terminal 1
$ denv enter feature-x
(feature-x) $ npm run dev    # Port 33000

# Terminal 2  
$ denv enter feature-y
(feature-y) $ npm run dev    # Port 33001

# No conflicts!
```

### Git Worktrees

```bash
$ cd ~/projects/myapp
$ denv enter
(default) $ echo $PORT_5432
35432

$ cd ~/projects/myapp-bugfix    # Different worktree
$ denv enter
(default) $ echo $PORT_5432
35432                            # Same ports! Same project detected
```

### Graceful Session Cleanup

```bash
# See what's running
$ denv sessions
Active sessions for myapp:
  - Session abc123 (PID 12345) in terminal /dev/ttys001
  - Session def456 (PID 12350) in terminal /dev/ttys002

# Gracefully terminate all (runs on-exit.sh hooks)
$ denv sessions --kill
Sending SIGTERM to session abc123...
Sending SIGTERM to session def456...
All sessions terminated.

# Or clean up dead sessions
$ denv sessions --cleanup
Found 1 orphaned session (PID 99999 dead)
Running cleanup hooks...
Removed orphaned locks.
```

## Implementation Priorities

### Phase 1 - Core (MVP)
- [x] Automatic project detection
- [x] Port randomization and remapping
- [x] Environment variable overrides via patterns
- [x] Basic enter/exit commands
- [x] Session tracking with file locks

### Phase 2 - Hooks
- [ ] on-enter.sh / on-exit.sh hooks
- [ ] Project-specific hooks (.denv.sh)
- [ ] Session count tracking

### Phase 3 - Advanced (Future)
- [ ] on-first-enter.sh / on-last-exit.sh
- [ ] Service management
- [ ] Data backup/restore
- [ ] Environment templates

## Design Principles

1. **Zero Configuration** - Works immediately without setup
2. **Project Isolation** - Automatic via git remote or folder name
3. **Simple Flat Structure** - No nested complexity
4. **Pattern-Based Rules** - Flexible without being complicated
5. **Session Safe** - Multiple terminals work correctly
6. **Git Worktree Aware** - Same project = same environment pool
7. **Non-Invasive** - Never modifies project files

## What denv Does NOT Do

- **Service Management** - Use docker-compose, systemd, etc. (maybe v2)
- **Version Management** - Use mise, asdf, nvm, etc.
- **Secrets Management** - Use proper secret stores
- **Complex Orchestration** - Keep it simple

## Summary

denv provides automatic, zero-config environment isolation for development. It focuses solely on preventing port conflicts and isolating environment variables across multiple development environments. 

**Key principle: denv never writes to your project.** All configuration and state lives in `~/.denv/`. The only trace in your project is a `.denv/` directory with symlinks (add `.denv/` to your global gitignore).

Simple, predictable, and completely non-invasive.