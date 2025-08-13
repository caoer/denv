package project

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zitao/denv/internal/config"
)

func DetectProject(dir string) (string, error) {
	// Try git remote first
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err == nil {
		// Extract project name from git URL
		url := strings.TrimSpace(string(output))
		return extractProjectName(url), nil
	}

	// Fall back to folder name
	return filepath.Base(dir), nil
}

func DetectProjectWithConfig(dir string, cfg *config.Config) string {
	// Check for override in config
	if cfg != nil && cfg.Projects != nil {
		if name, ok := cfg.Projects[dir]; ok {
			return name
		}
	}

	// Fall back to regular detection
	name, _ := DetectProject(dir)
	return name
}

func extractProjectName(gitURL string) string {
	// Remove .git suffix
	gitURL = strings.TrimSuffix(gitURL, ".git")

	// Handle SSH URLs (git@github.com:user/repo)
	if strings.HasPrefix(gitURL, "git@") {
		parts := strings.Split(gitURL, ":")
		if len(parts) >= 2 {
			gitURL = parts[1]
		}
	}

	// Handle HTTPS URLs (https://github.com/user/repo)
	re := regexp.MustCompile(`^https?://[^/]+/(.+)$`)
	if matches := re.FindStringSubmatch(gitURL); len(matches) > 1 {
		gitURL = matches[1]
	}

	// Extract the repo name (last part of the path)
	parts := strings.Split(gitURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "unknown"
}