#!/usr/bin/env bash
# Installation script for denv bash wrapper

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
WRAPPER_NAME="denv-wrapper.sh"
BINARY_NAME="denv"

echo "Installing denv with bash wrapper..."

# Build the Go binary (from project root)
echo "Building Go binary..."
cd "$PROJECT_ROOT"
if ! make build; then
    echo "Error: Failed to build Go binary" >&2
    exit 1
fi

# Rename binary to denv-core
echo "Installing denv-core..."
sudo cp "$BINARY_NAME" "$INSTALL_DIR/denv-core"
sudo chmod +x "$INSTALL_DIR/denv-core"

# Install the wrapper
echo "Installing bash wrapper..."
sudo cp "$SCRIPT_DIR/$WRAPPER_NAME" "$INSTALL_DIR/$WRAPPER_NAME"
sudo chmod +x "$INSTALL_DIR/$WRAPPER_NAME"

# Create denv command that sources the wrapper
echo "Creating denv command..."
sudo tee "$INSTALL_DIR/denv" > /dev/null <<'EOF'
#!/usr/bin/env bash
# Source the wrapper if not already sourced
if ! type -t denv | grep -q 'function'; then
    source /usr/local/bin/denv-wrapper.sh
fi
# Call the denv function
denv "$@"
EOF
sudo chmod +x "$INSTALL_DIR/denv"

# Detect user's shell
SHELL_NAME=$(basename "$SHELL")
RC_FILE=""

case "$SHELL_NAME" in
    bash)
        RC_FILE="$HOME/.bashrc"
        ;;
    zsh)
        RC_FILE="$HOME/.zshrc"
        ;;
    *)
        echo "Note: Automatic shell integration not available for $SHELL_NAME"
        echo "Please manually add the following to your shell configuration:"
        echo "  source $INSTALL_DIR/$WRAPPER_NAME"
        ;;
esac

if [[ -n "$RC_FILE" ]]; then
    echo ""
    echo "To enable shell integration, add the following to $RC_FILE:"
    echo ""
    echo "# denv shell integration"
    echo "source $INSTALL_DIR/$WRAPPER_NAME"
    echo ""
    read -p "Add automatically? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Check if already added
        if ! grep -q "denv-wrapper.sh" "$RC_FILE" 2>/dev/null; then
            echo "" >> "$RC_FILE"
            echo "# denv shell integration" >> "$RC_FILE"
            echo "source $INSTALL_DIR/$WRAPPER_NAME" >> "$RC_FILE"
            echo "Added to $RC_FILE"
        else
            echo "Already configured in $RC_FILE"
        fi
    fi
fi

echo ""
echo "Installation complete!"
echo ""
echo "Quick start:"
echo "  1. Reload your shell or run: source $INSTALL_DIR/$WRAPPER_NAME"
echo "  2. Navigate to a project directory"
echo "  3. Run: denv enter"
echo ""
echo "Available commands:"
echo "  denv enter [env]  - Enter an environment"
echo "  denv exit        - Exit current environment"
echo "  denv list        - List environments"
echo "  denv clean <env> - Remove an environment"
echo ""
echo "Aliases available after sourcing:"
echo "  de - denv enter"
echo "  dx - denv exit"
echo "  dl - list environments with current marker"
echo "  ds - switch environments"