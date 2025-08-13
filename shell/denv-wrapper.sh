#!/usr/bin/env bash
# denv - Bash wrapper for seamless shell integration
# This wrapper handles shell-specific operations while delegating complex logic to denv-core

# Configuration
DENV_CORE="${DENV_CORE:-denv-core}"  # Path to Go binary
DENV_HOME="${DENV_HOME:-$HOME/.denv}"

# Detect current shell
detect_shell() {
    local shell_name
    if [[ -n "$BASH_VERSION" ]]; then
        shell_name="bash"
    elif [[ -n "$ZSH_VERSION" ]]; then
        shell_name="zsh"
    elif [[ "$SHELL" == *"fish"* ]]; then
        shell_name="fish"
    else
        shell_name="sh"
    fi
    echo "$shell_name"
}

# Main denv function
denv() {
    local cmd="${1:-}"
    shift || true

    case "$cmd" in
        enter|e)
            _denv_enter "$@"
            ;;
        
        exit|x)
            _denv_exit
            ;;
        
        eval)
            # For direnv integration
            _denv_eval "$@"
            ;;
        
        shell-init)
            # Initialize shell integration
            _denv_shell_init
            ;;
        
        # Pass through all other commands to Go binary
        *)
            "$DENV_CORE" "$cmd" "$@"
            ;;
    esac
}

# Enter environment with native shell integration
_denv_enter() {
    local env_name="${1:-default}"
    local shell_type=$(detect_shell)
    
    # Get environment setup from Go binary
    local env_json
    env_json=$("$DENV_CORE" prepare-env "$env_name" 2>/dev/null)
    
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to prepare environment" >&2
        return 1
    fi
    
    # Parse JSON response (using jq if available, fallback to grep/sed)
    if command -v jq >/dev/null 2>&1; then
        local env_path=$(echo "$env_json" | jq -r '.env_path')
        local project_path=$(echo "$env_json" | jq -r '.project_path')
        local session_id=$(echo "$env_json" | jq -r '.session_id')
        local project_name=$(echo "$env_json" | jq -r '.project_name')
        local port_mappings=$(echo "$env_json" | jq -r '.ports | to_entries[] | "PORT_\(.key)=\(.value)"')
    else
        # Fallback parsing without jq
        local env_path=$("$DENV_CORE" get-env-path "$env_name" 2>/dev/null)
        local project_path=$("$DENV_CORE" get-project-path 2>/dev/null)
        local session_id=$("$DENV_CORE" create-session "$env_name" 2>/dev/null)
        local project_name=$("$DENV_CORE" get-project-name 2>/dev/null)
    fi
    
    # Create temporary environment file
    local temp_env=$(mktemp /tmp/denv-env.XXXXXX)
    
    # Generate environment script based on shell type
    case "$shell_type" in
        bash)
            _generate_bash_env "$temp_env" "$env_path" "$project_path" "$session_id" "$env_name" "$project_name" "$port_mappings"
            
            # For bash, we can use a subshell with modified environment
            (
                # Source the environment
                source "$temp_env"
                
                # Run enter hooks
                [[ -f "$DENV_PROJECT/hooks/on-enter.sh" ]] && source "$DENV_PROJECT/hooks/on-enter.sh"
                
                # Setup cleanup trap
                trap '_denv_cleanup' EXIT INT TERM
                
                # Start interactive shell with modified prompt
                PS1="[denv:$env_name] $PS1" exec bash --norc -i
            )
            ;;
        
        zsh)
            _generate_zsh_env "$temp_env" "$env_path" "$project_path" "$session_id" "$env_name" "$project_name" "$port_mappings"
            
            # For zsh, similar approach
            (
                source "$temp_env"
                [[ -f "$DENV_PROJECT/hooks/on-enter.sh" ]] && source "$DENV_PROJECT/hooks/on-enter.sh"
                trap '_denv_cleanup' EXIT INT TERM
                
                # Preserve zsh config
                [[ -f ~/.zshrc ]] && source ~/.zshrc
                PS1="[denv:$env_name] $PS1" exec zsh -i
            )
            ;;
        
        fish)
            # Fish requires different syntax, delegate to Go binary for now
            "$DENV_CORE" enter "$env_name"
            ;;
        
        *)
            # Fallback to Go implementation
            "$DENV_CORE" enter "$env_name"
            ;;
    esac
    
    # Cleanup temp file
    rm -f "$temp_env"
}

# Generate bash environment script
_generate_bash_env() {
    local env_file="$1"
    local env_path="$2"
    local project_path="$3"
    local session_id="$4"
    local env_name="$5"
    local project_name="$6"
    local port_mappings="$7"
    
    cat > "$env_file" <<EOF
# denv environment setup for bash
export DENV_HOME="$DENV_HOME"
export DENV_ENV="$env_path"
export DENV_PROJECT="$project_path"
export DENV_ENV_NAME="$env_name"
export DENV_PROJECT_NAME="$project_name"
export DENV_SESSION="$session_id"

# Port mappings
$port_mappings

# Apply environment overrides from Go binary
eval "\$("$DENV_CORE" get-env-overrides "$env_name" 2>/dev/null)"

# Add project bin to PATH if exists
[[ -d "$project_path/bin" ]] && export PATH="$project_path/bin:\$PATH"

# Load project-specific environment
[[ -f "$project_path/.env" ]] && source "$project_path/.env"

# Define cleanup function
_denv_cleanup() {
    [[ -f "$DENV_PROJECT/hooks/on-exit.sh" ]] && source "$DENV_PROJECT/hooks/on-exit.sh"
    "$DENV_CORE" cleanup-session "$session_id" 2>/dev/null
}
EOF
}

# Generate zsh environment script
_generate_zsh_env() {
    local env_file="$1"
    shift
    # Similar to bash but with zsh-specific adjustments
    _generate_bash_env "$env_file" "$@"
    
    # Add zsh-specific configurations
    cat >> "$env_file" <<'EOF'

# Zsh-specific settings
setopt prompt_subst
EOF
}

# Exit current denv environment
_denv_exit() {
    if [[ -z "$DENV_SESSION" ]]; then
        echo "Not in a denv environment" >&2
        return 1
    fi
    
    # Run exit hooks
    [[ -f "$DENV_PROJECT/hooks/on-exit.sh" ]] && source "$DENV_PROJECT/hooks/on-exit.sh"
    
    # Cleanup session
    "$DENV_CORE" cleanup-session "$DENV_SESSION" 2>/dev/null
    
    # Exit the subshell
    exit 0
}

# Eval mode for direnv integration
_denv_eval() {
    local env_name="${1:-default}"
    
    # Get environment variables from Go binary
    "$DENV_CORE" export "$env_name"
}

# Shell initialization (add to .bashrc/.zshrc)
_denv_shell_init() {
    cat <<'EOF'
# denv shell integration
if [[ -f /usr/local/bin/denv-wrapper.sh ]]; then
    source /usr/local/bin/denv-wrapper.sh
fi

# Optional: Auto-enter denv on directory change
_denv_auto_enter() {
    if [[ -f .denv/config.yaml ]] && [[ -z "$DENV_SESSION" ]]; then
        echo "denv environment detected. Run 'denv enter' to activate."
    fi
}

# Hook into cd for auto-detection
if [[ -n "$BASH_VERSION" ]]; then
    PROMPT_COMMAND="_denv_auto_enter; $PROMPT_COMMAND"
elif [[ -n "$ZSH_VERSION" ]]; then
    precmd_functions+=(_denv_auto_enter)
fi
EOF
}

# Advanced features

# Nested environment support
_denv_push() {
    local env_name="${1:-default}"
    export DENV_STACK="${DENV_SESSION}:${DENV_STACK}"
    _denv_enter "$env_name"
}

_denv_pop() {
    if [[ -z "$DENV_STACK" ]]; then
        echo "No environment to pop to" >&2
        return 1
    fi
    
    _denv_exit
    local prev_session="${DENV_STACK%%:*}"
    export DENV_STACK="${DENV_STACK#*:}"
    
    # Restore previous environment
    "$DENV_CORE" restore-session "$prev_session"
}

# Quick environment switching
_denv_switch() {
    local new_env="$1"
    if [[ -n "$DENV_SESSION" ]]; then
        _denv_exit
    fi
    _denv_enter "$new_env"
}

# List available environments with current marker
_denv_list() {
    "$DENV_CORE" list | while IFS= read -r line; do
        if [[ "$line" == *"$DENV_ENV_NAME"* ]] && [[ -n "$DENV_SESSION" ]]; then
            echo "→ $line"
        else
            echo "  $line"
        fi
    done
}

# Aliases for convenience
alias de='denv enter'
alias dx='denv exit'
alias dl='_denv_list'
alias ds='_denv_switch'

# Export the main function
export -f denv

# If sourced directly, provide instructions
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "This script should be sourced, not executed directly."
    echo ""
    echo "Add to your shell configuration:"
    echo "  source $(readlink -f "$0")"
    echo ""
    echo "Or install system-wide:"
    echo "  sudo cp $(readlink -f "$0") /usr/local/bin/denv-wrapper.sh"
    echo "  echo 'source /usr/local/bin/denv-wrapper.sh' >> ~/.bashrc"
fi