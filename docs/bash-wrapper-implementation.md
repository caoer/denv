# Bash Wrapper Implementation for denv

## Overview

The Bash wrapper provides seamless shell integration for denv by handling environment manipulation natively in the shell while delegating complex operations to the Go binary. This hybrid approach combines the best of both worlds.

## Architecture

```
User → denv (bash function) → denv-core (Go binary)
         ↓                           ↓
    Shell Integration          Complex Logic
    - source/export            - Port allocation
    - subshell spawn           - Config parsing
    - prompt modification       - State management
    - signal handling          - Session tracking
```

## Key Benefits

### 1. **Native Shell Integration**
- Direct environment variable manipulation without subprocess limitations
- Proper `source` command support for loading `.env` files
- Natural subshell spawning with inherited environment
- Shell-specific prompt modifications

### 2. **Preserved Go Advantages**
- Type-safe port allocation and tracking
- Robust JSON/YAML configuration parsing
- Atomic file operations with proper locking
- Cross-platform compatibility for core logic

### 3. **Enhanced User Experience**
- Instant environment switching without binary restart
- Shell aliases for common operations (`de`, `dx`, `dl`)
- Tab completion support (can be added)
- Nested environment support with stack management

## Implementation Details

### Wrapper Functions

#### `denv enter`
```bash
# Workflow:
1. Detect current shell (bash/zsh/fish/sh)
2. Call Go binary to prepare environment (allocate ports, create session)
3. Get JSON response with environment details
4. Generate shell-specific environment script
5. Spawn subshell with modified environment
6. Setup cleanup traps for proper exit handling
```

#### `denv exit`
```bash
# Workflow:
1. Run exit hooks from project
2. Call Go binary to cleanup session
3. Exit subshell to return to parent
```

### Go Binary Commands

#### `prepare-env`
Returns JSON with:
- Environment and project paths
- Session ID
- Port mappings
- Environment variable overrides

#### `get-env-overrides`
Returns shell export commands for environment-specific variables

#### `cleanup-session`
Removes session locks and updates runtime state

## Shell Detection

The wrapper automatically detects the current shell and adapts its behavior:

```bash
detect_shell() {
    if [[ -n "$BASH_VERSION" ]]; then
        echo "bash"
    elif [[ -n "$ZSH_VERSION" ]]; then
        echo "zsh"
    elif [[ "$SHELL" == *"fish"* ]]; then
        echo "fish"
    else
        echo "sh"
    fi
}
```

## Environment Variables

The wrapper sets these environment variables in the subshell:

```bash
DENV_HOME        # Base directory (~/.denv)
DENV_ENV         # Current environment directory
DENV_PROJECT     # Shared project directory
DENV_ENV_NAME    # Environment name (e.g., "default")
DENV_PROJECT_NAME # Project name
DENV_SESSION     # Unique session ID
PORT_*           # Remapped ports (PORT_3000=30001)
```

## Advanced Features

### 1. **Nested Environments**
```bash
denv push staging    # Push current env to stack, enter staging
denv pop            # Return to previous environment
```

### 2. **Quick Switching**
```bash
ds production       # Switch directly to production env
```

### 3. **Visual Indicators**
```bash
dl                  # List with arrow showing current env
→ default (current)
  staging
  production
```

### 4. **Auto-detection**
```bash
cd /path/to/project
# Automatically detects .denv/config.yaml
# Prompts: "denv environment detected. Run 'denv enter' to activate."
```

## Installation

### Method 1: System-wide Installation
```bash
# Run the installation script
./install-wrapper.sh

# Or manually:
sudo cp denv-wrapper.sh /usr/local/bin/
sudo cp denv /usr/local/bin/denv-core
echo 'source /usr/local/bin/denv-wrapper.sh' >> ~/.bashrc
```

### Method 2: User Installation
```bash
# Copy files to user directory
cp denv-wrapper.sh ~/.local/bin/
cp denv ~/.local/bin/denv-core

# Add to shell configuration
echo 'source ~/.local/bin/denv-wrapper.sh' >> ~/.bashrc
```

## Integration with Existing Tools

### direnv
```bash
# .envrc
eval "$(denv eval)"
```

### tmux
```bash
# Preserve environment in new panes/windows
set-option -ga update-environment ' DENV_*'
```

### VS Code
```json
// settings.json
{
  "terminal.integrated.env.linux": {
    "DENV_AUTO_ENTER": "1"
  }
}
```

## Performance Considerations

### Wrapper Overhead
- **Minimal**: ~10ms for shell function invocation
- **Go binary calls**: Only for complex operations
- **Environment script**: Generated once, cached in temp file

### Optimization Tips
1. **Lazy Loading**: Only source wrapper when needed
2. **Caching**: Port mappings cached in `runtime.json`
3. **Selective Exports**: Only export changed variables

## Troubleshooting

### Issue: "command not found: denv"
**Solution**: Ensure wrapper is sourced in shell configuration
```bash
source /usr/local/bin/denv-wrapper.sh
```

### Issue: Environment variables not persisting
**Solution**: Check you're using the wrapper, not calling binary directly
```bash
type denv  # Should show: "denv is a function"
```

### Issue: Fish shell not working
**Solution**: Fish support requires special syntax, fallback to Go binary
```bash
denv-core enter  # Use binary directly for fish
```

## Migration Path

### Phase 1: Current Go Implementation
- All logic in Go binary
- Shell spawning via exec.Command
- Works but limited shell integration

### Phase 2: Hybrid with Bash Wrapper (Current)
- Bash wrapper for shell operations
- Go binary for complex logic
- Best of both worlds

### Phase 3: Future Enhancements
- Native fish/powershell support
- Tab completion
- Plugin system
- Remote environment support

## Testing Strategy

### Unit Tests
```bash
# Test Go binary commands
go test ./internal/commands -run TestPrepareEnv

# Test bash wrapper functions
bash test-wrapper.sh
```

### Integration Tests
```bash
# Test full workflow
source denv-wrapper.sh
denv enter test
echo $DENV_ENV_NAME  # Should output: test
denv exit
```

### Shell Compatibility Tests
```bash
# Test across shells
for shell in bash zsh sh; do
    $shell -c 'source denv-wrapper.sh && denv enter test'
done
```

## Conclusion

The Bash wrapper implementation provides superior shell integration while maintaining the robustness of the Go implementation. This hybrid approach offers:

1. **Immediate value**: Better shell integration without rewriting core logic
2. **Maintainability**: Clear separation of concerns
3. **Flexibility**: Easy to extend and customize
4. **Performance**: Optimal for both simple and complex operations

The wrapper acts as a thin layer that enhances the user experience while the Go binary continues to handle all complex operations reliably.