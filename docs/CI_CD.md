# CI/CD Documentation

## Overview

This project uses GitHub Actions for continuous integration and deployment with automated testing, versioning, and multi-platform releases.

## Workflows

### 1. Continuous Integration (`ci.yml`)
- **Triggers**: Push to main/develop, Pull Requests
- **Features**:
  - Multi-OS testing (Linux, macOS, Windows)
  - Multi-version Go testing (1.21, 1.22, 1.23)
  - Code coverage with Codecov
  - Security scanning with gosec and govulncheck
  - Linting with golangci-lint
  - Cross-platform builds (amd64, arm64)

### 2. Release Pipeline (`release.yml`)
- **Triggers**: Git tags (v*), Manual dispatch
- **Features**:
  - Automated testing before release
  - Cross-platform binary generation via GoReleaser
  - Docker image creation and publishing
  - Homebrew formula updates
  - Platform-specific installers (MSI, PKG)
  - Checksums generation

### 3. Semantic Release (`semantic-release.yml`)
- **Triggers**: Push to main
- **Features**:
  - Automatic version bumping based on commit messages
  - Follows Conventional Commits specification
  - Updates CHANGELOG.md automatically
  - Creates GitHub releases with notes

### 4. Changelog Generation (`changelog.yml`)
- **Triggers**: Push to main, Manual dispatch
- **Uses**: git-cliff for changelog generation
- **Features**:
  - Groups changes by type (Features, Bug Fixes, etc.)
  - Links to commits and issues
  - Maintains Keep a Changelog format

### 5. Release Drafter (`release-drafter.yml`)
- **Triggers**: Push to main, Pull Requests
- **Features**:
  - Drafts next release notes automatically
  - Categorizes changes based on labels
  - Includes contributor statistics

## Dependency Management

### Dependabot Configuration
- Automatic dependency updates for:
  - Go modules (weekly)
  - GitHub Actions (weekly)
  - Docker base images (weekly)
- Groups minor and patch updates
- Creates PRs with proper commit prefixes

## Release Process

### Automatic Releases

1. **Commit with Conventional Commits**:
   ```
   feat: add new feature     # Minor version bump
   fix: resolve bug          # Patch version bump
   feat!: breaking change    # Major version bump
   ```

2. **Semantic Release creates version tag**
3. **Release workflow triggers on tag**
4. **GoReleaser builds and publishes**

### Manual Release

1. **Create and push a tag**:
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```

2. **Or trigger workflow manually** from GitHub Actions tab

## Build Artifacts

### Binary Releases
- **Platforms**: Linux, macOS, Windows, FreeBSD
- **Architectures**: amd64, arm64, arm, 386
- **Format**: tar.gz (Unix), zip (Windows)

### Package Formats
- **Linux**: .deb, .rpm, .apk
- **macOS**: .pkg installer, Homebrew formula
- **Windows**: .msi installer
- **Container**: Docker images (multi-arch)

## Code Coverage

- **Service**: Codecov
- **Target**: 80% coverage
- **Reports**: Generated on every push/PR
- **Badge**: Available in README

## Security

- **gosec**: Static security analysis
- **govulncheck**: Vulnerability scanning
- **Dependabot**: Automated security updates
- **SARIF**: Security findings uploaded to GitHub

## Commit Message Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types**:
- `feat`: New feature (MINOR)
- `fix`: Bug fix (PATCH)
- `docs`: Documentation only
- `style`: Code style changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding tests
- `chore`: Maintenance tasks
- `ci`: CI/CD changes
- `build`: Build system changes

**Breaking Changes**: Add `!` after type or `BREAKING CHANGE:` in footer (MAJOR)

## Installation Methods

Users can install denv through:

1. **Direct Download**: From GitHub Releases
2. **Install Script**: `curl -fsSL .../install.sh | bash`
3. **Homebrew**: `brew install denv`
4. **Docker**: `docker pull ghcr.io/.../denv`
5. **Package Managers**: apt, yum, apk

## Environment Variables

Required secrets in GitHub:
- `GITHUB_TOKEN`: Automatically provided
- `CODECOV_TOKEN`: For coverage reports (optional)
- `HOMEBREW_TAP_TOKEN`: For Homebrew updates (optional)

## Monitoring

- **CI Status**: Badge in README
- **Coverage**: Codecov dashboard
- **Dependencies**: Dependabot dashboard
- **Security**: GitHub Security tab
- **Releases**: GitHub Releases page