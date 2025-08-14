# GitHub CI/CD Setup Guide

## Quick Start

Follow these steps to enable the CI/CD pipeline for your denv repository:

## 1. Update Repository References

Replace `caoer` with your GitHub username in these files:
- `.goreleaser.yml`
- `install.sh`
- `docs/badges.md`
- `.github/workflows/*.yml` (if needed)

## 2. Enable GitHub Actions

1. Go to your repository Settings → Actions → General
2. Enable "Allow all actions and reusable workflows"
3. Under "Workflow permissions", select "Read and write permissions"

## 3. Configure Codecov (Optional)

1. Visit [codecov.io](https://codecov.io)
2. Add your repository
3. Copy the CODECOV_TOKEN
4. Add to GitHub Secrets: Settings → Secrets → Actions → New repository secret
   - Name: `CODECOV_TOKEN`
   - Value: Your token from Codecov

## 4. Set Up Homebrew Tap (Optional)

If you want to distribute via Homebrew:

1. Create a new repository named `homebrew-tap`
2. Generate a Personal Access Token with `repo` scope
3. Add to GitHub Secrets:
   - Name: `HOMEBREW_TAP_TOKEN`
   - Value: Your personal access token

## 5. Initial Release

### Option A: Manual Tag

```bash
# Create your first release
git tag v0.1.0
git push origin v0.1.0
```

### Option B: Let Semantic Release Handle It

Simply push commits with conventional commit messages:
```bash
git add .
git commit -m "feat: initial CI/CD setup"
git push origin main
```

## 6. Verify Setup

After pushing:
1. Check Actions tab - CI workflow should run
2. Check Releases page - Release should be created (if tagged)
3. Verify badges work in README

## Commit Message Guidelines

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature (minor version)
- `fix:` - Bug fix (patch version)  
- `feat!:` or `BREAKING CHANGE:` - Breaking change (major version)
- `docs:` - Documentation only
- `chore:` - Maintenance tasks
- `ci:` - CI/CD changes
- `test:` - Test changes

## Triggering Releases

### Automatic (Recommended)
Push commits with proper conventional commit messages. Semantic release will:
1. Analyze commits
2. Determine version bump
3. Create tag and release
4. Update changelog

### Manual
```bash
# Create and push a tag
git tag v1.0.0
git push origin v1.0.0
```

## Workflow Descriptions

| Workflow         | Trigger      | Purpose                            |
| ---------------- | ------------ | ---------------------------------- |
| CI               | Push, PR     | Run tests, linting, security scans |
| Release          | Git tags     | Build and publish releases         |
| Semantic Release | Push to main | Auto-version and changelog         |
| Changelog        | Push to main | Update CHANGELOG.md                |
| Release Drafter  | Push, PR     | Draft release notes                |
| Dependabot       | Schedule     | Update dependencies                |

## Monitoring

- **CI Status**: Check Actions tab
- **Coverage**: View at codecov.io
- **Security**: GitHub Security tab
- **Dependencies**: Dependabot alerts
- **Releases**: GitHub Releases page

## Troubleshooting

### CI Failing
- Check test output in Actions tab
- Ensure all tests pass locally: `make test`
- Verify linting: `golangci-lint run`

### Release Not Created
- Verify tag format: `v*` (e.g., v1.0.0)
- Check Actions tab for errors
- Ensure GoReleaser config is valid

### Coverage Not Updating
- Verify CODECOV_TOKEN is set
- Check Codecov dashboard for issues

## Support

For issues with:
- CI/CD setup: Check `.github/workflows/` files
- Release process: Check `.goreleaser.yml`
- Changelog: Check `.github/cliff.toml`

## Next Steps

1. ✅ Push changes to trigger CI
2. ✅ Create first release
3. ✅ Add badges to README
4. ✅ Configure branch protection rules
5. ✅ Enable GitHub Pages for documentation (optional)