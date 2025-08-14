#!/usr/bin/env bash

set -e

# denv installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/caoer/denv/main/install.sh | bash

REPO="caoer/denv"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
DENV_HOME="${DENV_HOME:-$HOME/.denv}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$OS" in
        linux)
            OS="Linux"
            ;;
        darwin)
            OS="Darwin"
            ;;
        msys*|mingw*|cygwin*)
            OS="Windows"
            ;;
        *)
            log_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)
            ARCH="x86_64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l|armv7)
            ARCH="armv7"
            ;;
        i686|i386)
            ARCH="i386"
            ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    log_info "Detected platform: $OS $ARCH"
}

# Get latest release version
get_latest_version() {
    log_info "Fetching latest version..."
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        log_error "Failed to fetch latest version"
        exit 1
    fi
    
    log_info "Latest version: $VERSION"
}

# Download and install denv
install_denv() {
    local url="https://github.com/$REPO/releases/download/$VERSION/denv_${VERSION#v}_${OS}_${ARCH}.tar.gz"
    local tmp_dir=$(mktemp -d)
    local archive="$tmp_dir/denv.tar.gz"
    
    log_info "Downloading denv from $url..."
    
    if ! curl -fsSL "$url" -o "$archive"; then
        log_error "Failed to download denv"
        rm -rf "$tmp_dir"
        exit 1
    fi
    
    log_info "Extracting archive..."
    tar -xzf "$archive" -C "$tmp_dir"
    
    log_info "Installing denv to $INSTALL_DIR..."
    
    if [ -w "$INSTALL_DIR" ]; then
        mv "$tmp_dir/denv" "$INSTALL_DIR/"
    else
        log_warning "Need sudo permissions to install to $INSTALL_DIR"
        sudo mv "$tmp_dir/denv" "$INSTALL_DIR/"
    fi
    
    chmod +x "$INSTALL_DIR/denv"
    
    rm -rf "$tmp_dir"
    
    log_info "denv installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v denv &> /dev/null; then
        log_info "Verification successful! denv version: $(denv --version)"
    else
        log_warning "denv is installed but not in PATH"
        log_warning "Add $INSTALL_DIR to your PATH:"
        log_warning "  export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
}

# Setup denv home directory
setup_denv_home() {
    if [ ! -d "$DENV_HOME" ]; then
        log_info "Creating denv home directory at $DENV_HOME..."
        mkdir -p "$DENV_HOME"
    fi
}

# Add shell completion
setup_completion() {
    local shell=$(basename "$SHELL")
    
    case "$shell" in
        bash)
            local completion_file="$HOME/.bash_completion.d/denv"
            mkdir -p "$(dirname "$completion_file")"
            denv completion bash > "$completion_file"
            log_info "Bash completion installed to $completion_file"
            ;;
        zsh)
            local completion_file="$HOME/.zsh/completions/_denv"
            mkdir -p "$(dirname "$completion_file")"
            denv completion zsh > "$completion_file"
            log_info "Zsh completion installed to $completion_file"
            ;;
        fish)
            local completion_file="$HOME/.config/fish/completions/denv.fish"
            mkdir -p "$(dirname "$completion_file")"
            denv completion fish > "$completion_file"
            log_info "Fish completion installed to $completion_file"
            ;;
        *)
            log_warning "Shell completion not available for $shell"
            ;;
    esac
}

# Main installation flow
main() {
    echo "======================================"
    echo "        denv Installer"
    echo "======================================"
    echo
    
    detect_platform
    get_latest_version
    install_denv
    setup_denv_home
    setup_completion
    verify_installation
    
    echo
    echo "======================================"
    echo "     Installation Complete!"
    echo "======================================"
    echo
    echo "To get started, run:"
    echo "  denv --help"
    echo
    echo "To enter a development environment:"
    echo "  denv enter"
    echo
}

# Run main function
main "$@"