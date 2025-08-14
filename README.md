# denv - Development Environment Manager 🚀

> **Zero-configuration environment isolation for developers**  
> Automatically prevent port conflicts and environment variable collisions when working on multiple projects.

[![Go](https://img.shields.io/badge/Go-1.19%2B-blue)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

## 🎯 Why denv?

Ever experienced these frustrations?
- **Port conflicts** when running multiple projects (`Error: port 3000 already in use`)
- **Environment variable collisions** between different projects
- **Database connection strings** pointing to the wrong environment
- **Manual environment setup** every time you switch projects
- **Forgetting which ports** are used by which project

**denv solves all of these with zero configuration!**

## ✨ Quick Demo

```bash
# Terminal 1: Work on project A
$ cd ~/projects/webapp
$ denv enter
[denv:webapp-default] $ npm start  # Runs on port 33000 (auto-mapped from 3000)

# Terminal 2: Work on project B simultaneously  
$ cd ~/projects/api
$ denv enter
[denv:api-default] $ npm start     # Runs on port 33001 (no conflict!)

# Both projects run simultaneously without any port conflicts! 🎉
```

## 🚀 Features

### Core Benefits
- **🔥 Zero Configuration** - Works immediately, no setup required
- **🔒 Automatic Port Isolation** - Each environment gets unique ports (30000-39999 range)
- **🎯 Smart Variable Management** - Pattern-based environment variable overrides
- **🏷️ Project Auto-Detection** - Identifies projects via git remote or folder name
- **🔄 Multi-Session Support** - Multiple terminals work correctly with file locks
- **🌲 Git Worktree Aware** - Different worktrees share the same project environment pool
- **🧹 Non-Invasive** - Never modifies your project files (only creates `.denv/` symlinks)

### Advanced Features
- **📦 Environment Stacking** - Push/pop between environments
- **🎨 Visual Indicators** - Color-coded port mappings and environment status
- **⚡ Shell Integration** - Native aliases and instant environment switching
- **🪝 Hook System** - Run custom scripts on enter/exit
- **🔌 direnv Compatible** - Integrate with existing direnv workflows
- **📊 Session Management** - Track and manage active sessions across terminals

## 📦 Installation

### Option 1: Quick Install (Recommended)

```bash
# Install Go binary
go install github.com/caoer/denv/cmd/denv@latest

# Or build from source
git clone https://github.com/caoer/denv.git
cd denv
make build
sudo mv ./denv /usr/local/bin/
```

### Option 2: With Shell Integration (Enhanced Experience) 

For the best experience with aliases and instant switching:

```bash
# Clone the repository
git clone https://github.com/caoer/denv.git
cd denv

# Run the installer
./shell/install-wrapper.sh

# Or manually:
make build
sudo cp denv /usr/local/bin/denv-core
sudo cp shell/denv-wrapper.sh /usr/local/bin/
echo 'source /usr/local/bin/denv-wrapper.sh' >> ~/.bashrc  # or ~/.zshrc
source ~/.bashrc
```

## 🎮 Usage Guide

### Basic Commands

| Command             | Description                      | Example                              |
| ------------------- | -------------------------------- | ------------------------------------ |
| `denv enter [name]` | Enter an environment             | `denv enter` or `denv enter staging` |
| `denv list`         | List all environments            | `denv list` or `denv ls`             |
| `denv ps [name]`    | Show environment status          | `denv ps`                            |
| `denv rm <name>`    | Remove an environment            | `denv rm feature-x`                  |
| `denv rm --all`     | Remove all inactive environments | `denv rm --all`                      |
| `denv exit`         | Exit current environment         | `denv exit` or `Ctrl+D`              |

### Shell Aliases (when using wrapper)

| Alias      | Command       | Description                   |
| ---------- | ------------- | ----------------------------- |
| `de`       | `denv enter`  | Quick enter                   |
| `dx`       | `denv exit`   | Quick exit                    |
| `dl`       | `denv list`   | List with current highlighted |
| `ds <env>` | `denv switch` | Instant environment switch    |

### Session Management

```bash
# View active sessions
$ denv sessions
Active sessions for myapp:
  Session abc123 (PID 12345) in /dev/ttys001 - default environment
  Session def456 (PID 12350) in /dev/ttys002 - staging environment

# Clean up orphaned sessions
$ denv sessions --cleanup

# Terminate all sessions gracefully
$ denv sessions --kill
```

### Project Management

```bash
# Show current project name
$ denv project
myapp

# Rename project (updates config)
$ denv project rename my-awesome-app

# Remove project name override
$ denv project unset
```

## 🎯 Real-World Examples

### Example 1: Running Multiple Development Servers

```bash
# Project 1: React App
$ cd ~/projects/frontend
$ denv enter
[denv:frontend-default] $ npm start
# ✅ Runs on port 33000 (mapped from 3000)

# Project 2: Node.js API (different terminal)
$ cd ~/projects/backend  
$ denv enter
[denv:backend-default] $ npm run dev
# ✅ Runs on port 33001 (mapped from 3000) - no conflict!

# Project 3: Another React App (different terminal)
$ cd ~/projects/admin
$ denv enter
[denv:admin-default] $ npm start  
# ✅ Runs on port 33002 (mapped from 3000) - still no conflict!
```

### Example 2: Database Connections Auto-Adjusted

```bash
$ cd ~/projects/myapp
$ denv enter

# Original DATABASE_URL: postgresql://localhost:5432/myapp
# After denv: postgresql://localhost:35432/myapp
# ✅ Port automatically remapped!

[denv:myapp-default] $ echo $DATABASE_URL
postgresql://localhost:35432/myapp

# Your app connects to the right port without any code changes!
[denv:myapp-default] $ npm run db:migrate  # Uses port 35432
```

### Example 3: Multiple Environments for Same Project

```bash
# Work on main feature
$ denv enter main
[denv:myapp-main] $ npm run dev  # Port 33000

# Test feature branch (new terminal)
$ denv enter feature-auth  
[denv:myapp-feature-auth] $ npm run dev  # Port 33001

# Test another feature (new terminal)
$ denv enter feature-ui
[denv:myapp-feature-ui] $ npm run dev  # Port 33002

# All three run simultaneously! 🚀
```

### Example 4: Git Worktrees Share Environments

```bash
# Main repository
$ cd ~/projects/myapp
$ denv enter
[denv:myapp-default] $ echo $PORT_5432
35432

# Git worktree for bugfix
$ cd ~/projects/myapp-bugfix  
$ denv enter
[denv:myapp-default] $ echo $PORT_5432
35432  # Same port! Recognized as same project

# Both directories share the same environment pool
```

### Example 5: Quick Environment Switching (with shell wrapper)

```bash
# Start in development
$ de dev
[denv:myapp-dev] $ npm run test

# Quick switch to staging
[denv:myapp-dev] $ ds staging
[denv:myapp-staging] $ npm run test  # Now testing with staging config

# Quick switch to production  
[denv:myapp-staging] $ ds prod
[denv:myapp-prod] $ npm run test  # Now testing with prod config

# Exit when done
[denv:myapp-prod] $ dx
$
```

## ⚙️ Configuration

### Global Configuration

Located at `~/.denv/config.yaml`:

```yaml
# Project name overrides (set via prompts or denv project command)
projects:
  /Users/me/work/client-project: acme-web
  /Users/me/work/another-client: acme-api

# Pattern-based environment variable rules
patterns:
  # Port variables - always randomize
  "*_PORT|PORT":
    action: random_port
    range: [30000, 39999]
  
  # URLs - intelligently rewrite ports
  "*_URL|*_URI|*_ENDPOINT|DATABASE_URL|REDIS_URL":
    action: rewrite_ports
  
  # Directory paths - isolate per environment  
  "*_ROOT|*_DIR|*_PATH|*_HOME":
    action: isolate
    base: "${DENV_ENV}"
  
  # Secrets - never modify
  "*_KEY|*_TOKEN|*_SECRET|*_PASSWORD":
    action: keep
  
  # System paths - preserve
  "PATH|GOPATH|CARGO_HOME|NVM_DIR":
    action: keep
```

### Environment Variables Available

Inside a denv session, these variables are automatically set:

| Variable            | Description              | Example                          |
| ------------------- | ------------------------ | -------------------------------- |
| `DENV_HOME`         | Base denv directory      | `/home/user/.denv`               |
| `DENV_ENV`          | Current environment path | `/home/user/.denv/myapp-staging` |
| `DENV_PROJECT`      | Shared project directory | `/home/user/.denv/myapp`         |
| `DENV_ENV_NAME`     | Environment name         | `staging`                        |
| `DENV_PROJECT_NAME` | Project name             | `myapp`                          |
| `DENV_SESSION`      | Unique session ID        | `abc123def456`                   |
| `PORT_*`            | Remapped ports           | `PORT_3000=33000`                |
| `ORIGINAL_PORT_*`   | Original port values     | `ORIGINAL_PORT_3000=3000`        |

## 🪝 Hooks System

Create custom scripts that run on environment enter/exit:

### Setup Hooks

```bash
# Hooks are stored in the shared project directory
~/.denv/myapp/hooks/
├── on-enter.sh    # Runs when entering any environment
└── on-exit.sh     # Runs when exiting any environment
```

### Example: Auto-start Services

```bash
# ~/.denv/myapp/hooks/on-enter.sh
#!/bin/bash
echo "🚀 Starting services for $DENV_ENV_NAME environment..."

# Start PostgreSQL if not running
if ! pg_isready -p $PORT_5432 > /dev/null 2>&1; then
    postgres -D ~/.denv/$DENV_PROJECT_NAME/pgdata -p $PORT_5432 &
    echo "PostgreSQL started on port $PORT_5432"
fi

# Start Redis if not running
if ! redis-cli -p $PORT_6379 ping > /dev/null 2>&1; then
    redis-server --port $PORT_6379 --daemonize yes
    echo "Redis started on port $PORT_6379"
fi
```

```bash
# ~/.denv/myapp/hooks/on-exit.sh
#!/bin/bash
echo "🛑 Cleaning up $DENV_ENV_NAME environment..."

# Stop services using our ports
for port in $PORT_5432 $PORT_6379; do
    lsof -ti:$port | xargs kill 2>/dev/null || true
done

echo "Services stopped"
```

## 🔧 Integration with Other Tools

### direnv Integration

Add to your project's `.envrc`:

```bash
# .envrc
if command -v denv >/dev/null 2>&1; then
    eval "$(denv export)"
fi
```

### Docker Compose Integration

Use denv's port mappings in your `docker-compose.yml`:

```yaml
version: '3'
services:
  web:
    ports:
      - "${PORT_3000:-3000}:3000"
  
  postgres:
    ports:
      - "${PORT_5432:-5432}:5432"
  
  redis:
    ports:
      - "${PORT_6379:-6379}:6379"
```

### CI/CD Integration

```bash
# In your CI script
eval "$(denv export ci)"
npm test
npm run build
```

## 📁 File System Structure

### Global Structure
```
~/.denv/                           # DENV_HOME
├── config.yaml                    # Global configuration
├── myapp-default/                 # Environment directory
│   ├── runtime.json              # Current state & mappings
│   ├── ports.json                # Port allocations
│   └── sessions/                 # Active session locks
│       └── abc123.lock
├── myapp-staging/                # Another environment
├── myapp/                        # Shared project directory
│   └── hooks/
│       ├── on-enter.sh          # Entry hook
│       └── on-exit.sh           # Exit hook
└── another-project-default/
```

### Project Structure (Auto-created)
```
your-project/
└── .denv/                        # Only directory created
    ├── current -> ~/.denv/myapp-default   # Symlink to active env
    └── project -> ~/.denv/myapp          # Symlink to shared dir
```

**Note:** Add `.denv/` to your global gitignore: `echo ".denv/" >> ~/.gitignore_global`

## 🐛 Troubleshooting

### Common Issues and Solutions

#### Port is still in use after entering environment
```bash
# Check what's using the port
lsof -i :33000

# Clean up sessions
denv sessions --cleanup
```

#### Can't enter environment - "already in an environment"
```bash
# Exit current environment first
denv exit
# Or force exit
exit
```

#### Environment variables not updating
```bash
# Make sure you're in a denv session
echo $DENV_SESSION  # Should show session ID

# Re-enter environment
denv exit
denv enter
```

#### "Project detected as X. Is this correct?"
```bash
# Option 1: Accept the detection
y

# Option 2: Rename the project
n
Enter project name: my-better-name

# Option 3: Set permanently
denv project rename my-project-name
```

#### Permission denied errors
```bash
# Check denv home permissions
ls -la ~/.denv

# Fix permissions
chmod -R 755 ~/.denv
```

## 🧪 Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/caoer/denv.git
cd denv

# Run tests (TDD approach)
make test

# Watch tests during development
make test-watch

# Build binary
make build

# Test the built binary
./denv enter
```

### Testing in Isolated Environment

```bash
# Use temporary DENV_HOME for testing
export DENV_HOME="$(pwd)/tmp"
./denv enter test
# Your tests won't affect ~/.denv
```

### Project Structure

```
denv/
├── cmd/denv/          # CLI entry point
├── internal/
│   ├── commands/      # Command implementations
│   ├── config/        # Configuration management
│   ├── environment/   # Runtime state management
│   ├── ports/         # Port allocation system
│   ├── project/       # Project detection
│   ├── session/       # Session & lock management
│   ├── shell/         # Shell integration
│   └── color/         # Terminal colors
├── shell/             # Bash wrapper & integration
└── docs/              # Additional documentation
```

## 🤝 Contributing

We follow Test-Driven Development (TDD):

1. **Write failing test first**
2. **Run test to see it fail**
3. **Implement minimal code to pass**
4. **Refactor if needed**

See [CLAUDE.md](CLAUDE.md) for detailed development guidelines.

## 📚 Comparison with Other Tools

| Feature                  | denv | direnv   | dotenv | docker-compose |
| ------------------------ | ---- | -------- | ------ | -------------- |
| Zero configuration       | ✅    | ❌        | ❌      | ❌              |
| Automatic port isolation | ✅    | ❌        | ❌      | ⚠️ Manual       |
| Multiple environments    | ✅    | ⚠️ Manual | ❌      | ⚠️ Manual       |
| Git worktree aware       | ✅    | ❌        | ❌      | ❌              |
| Shell integration        | ✅    | ✅        | ❌      | ❌              |
| Project auto-detection   | ✅    | ❌        | ❌      | ❌              |
| Visual indicators        | ✅    | ❌        | ❌      | ❌              |

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Built with ❤️ using:
- [Go](https://golang.org) - For robust core implementation
- [Cobra](https://github.com/spf13/cobra) - CLI framework (if used)
- [testify](https://github.com/stretchr/testify) - Testing assertions

---

**Ready to eliminate environment conflicts forever?** 

```bash
# Get started in 10 seconds
go install github.com/caoer/denv/cmd/denv@latest
denv enter
```

🚀 **Happy coding without conflicts!**