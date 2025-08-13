# GitHub Actions Debugging with Claude Code

## Overview
This document outlines strategies for debugging and fixing GitHub Actions workflow failures using Claude Code.

## Key Debugging Methods

### 1. Common Issues and Solutions

**404 Error with GitHub App Installation**
- Cause: Missing workflow permissions on gh auth token
- Solution: Run `gh auth refresh -h github.com -s workflow` and retry

### 2. Debugging Techniques

**Use Verbose Mode**
- Add `--verbose` flag for detailed output during debugging
- Disable in production for cleaner logs

**Check Workflow Triggers**
- Ensure using GitHub App or custom app (not Actions user)
- Verify workflow triggers include necessary events
- Confirm app permissions include CI triggers

**Verify API Keys**
- Confirm API key validity and permissions
- For Bedrock/Vertex: Check credentials configuration
- Ensure secrets are correctly named in workflows

### 3. Structured Debugging Workflow

1. Read relevant files, images, or URLs first
2. Use subagents for complex problems
3. Create documentation or GitHub issues with plan before implementation
4. Reset if implementation doesn't match expectations

### 4. Monitoring and Logs

- Check Actions tab in GitHub for workflow debugging
- View Claude's logs to understand workflow execution
- Use output from failed steps to identify issues

## Iterative Fix Process

### Critical Loop Pattern
When fixing GitHub Actions issues:

1. **Identify Issues**
   - Check CI workflow failure details
   - Check other workflow failure details
   - Analyze error messages and logs

2. **Fix Issues**
   - Fix linting errors
   - Fix test failures
   - Update dependencies or action versions

3. **Commit and Push**
   ```bash
   git add .
   git commit -m "fix: resolve CI/CD issues"
   git push origin main
   ```

4. **Wait for Results**
   - Monitor GitHub Actions runs
   - Check for new failures

5. **Loop Until Fixed**
   - If failures persist, return to step 1
   - Continue until all workflows pass

### Example Fix Session

```bash
# List recent runs
gh run list --limit 5

# View specific run details
gh run view <RUN_ID> --log-failed

# After fixes, commit and push
git add -A
git commit -m "fix: resolve linting and test issues"
git push

# Monitor results
gh run list --limit 1 --watch
```

## Common CI/CD Issues

### Linting Errors
- Unused imports
- Unchecked error returns
- Ineffectual assignments
- Deprecated functions (e.g., rand.Seed in Go 1.20+)

### Test Failures
- Environment-specific issues
- Missing test data
- Race conditions
- Incorrect assertions

### Action Version Issues
- Deprecated action versions
- Breaking changes in new versions
- Docker image EOL (e.g., Debian Buster)

## Best Practices

1. **Test Locally First**
   - Run linters locally before pushing
   - Execute test suite locally
   - Verify build process

2. **Use Tool Batching**
   - Run multiple checks in parallel
   - Batch related fixes together
   - Use multi-edit tools for efficiency

3. **Document Changes**
   - Clear commit messages
   - Update relevant documentation
   - Note breaking changes

4. **Version Management**
   - Keep actions updated
   - Pin versions for stability
   - Monitor deprecation notices

## Resources

- [Claude Code GitHub Actions Documentation](https://docs.anthropic.com/en/docs/claude-code/github-actions)
- [GitHub Actions Debugging Guide](https://docs.github.com/en/actions/monitoring-and-troubleshooting-workflows)
- [git-cliff Action Documentation](https://git-cliff.org/docs/github-actions/)