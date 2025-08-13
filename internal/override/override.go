package override

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/caoer/denv/internal/config"
	"github.com/caoer/denv/internal/environment"
)

func MatchesPattern(pattern, key string) bool {
	// Handle OR patterns (e.g., "*_PORT | PORT")
	patterns := strings.Split(pattern, "|")
	for i := range patterns {
		patterns[i] = strings.TrimSpace(patterns[i])
	}
	for _, p := range patterns {
		if matchSinglePattern(p, key) {
			return true
		}
	}
	return false
}

func matchSinglePattern(pattern, key string) bool {
	// Convert pattern to regex
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = "^" + pattern + "$"
	matched, _ := regexp.MatchString(pattern, key)
	return matched
}

func RewriteURL(url string, ports map[int]int) string {
	// Only rewrite localhost URLs
	if !isLocalURL(url) {
		return url
	}

	// Try to rewrite each port
	result := url
	for oldPort, newPort := range ports {
		// Look for :oldPort in the URL
		oldPattern := fmt.Sprintf(":%d", oldPort)
		newPattern := fmt.Sprintf(":%d", newPort)
		result = strings.ReplaceAll(result, oldPattern, newPattern)
	}
	
	return result
}

func isLocalURL(url string) bool {
	localHosts := []string{"localhost", "127.0.0.1", "0.0.0.0"}
	for _, host := range localHosts {
		if strings.Contains(url, host) {
			return true
		}
	}
	return false
}

func ApplyRules(env map[string]string, cfg *config.Config, ports map[int]int, envPath string) (map[string]string, map[string]environment.Override) {
	result := make(map[string]string)
	overrides := make(map[string]environment.Override)

	for key, value := range env {
		newValue := value
		var rule string

		// Find matching pattern (patterns are now a slice of PatternRule)
		for _, pr := range cfg.Patterns {
			if MatchesPattern(pr.Pattern, key) {
				r := pr.Rule
				switch r.Action {
				case "random_port":
					// Try to parse as port number
					if port, err := strconv.Atoi(value); err == nil {
						if mappedPort, ok := ports[port]; ok {
							newValue = strconv.Itoa(mappedPort)
							rule = "random_port"
						}
					}
				case "rewrite_ports":
					newValue = RewriteURL(value, ports)
					rule = "rewrite_ports"
				case "keep":
					// Do nothing
					rule = "keep"
				case "isolate":
					// Replace with isolated path
					base := r.Base
					if base == "" {
						base = envPath
					} else if base == "${DENV_ENV}" {
						// Expand the DENV_ENV variable
						base = envPath
					}
					// Extract last part of path
					parts := strings.Split(value, "/")
					lastPart := parts[len(parts)-1]
					newValue = filepath.Join(base, lastPart)
					rule = "isolate"
				}
				break // Use first matching pattern
			}
		}

		result[key] = newValue
		if newValue != value {
			overrides[key] = environment.Override{
				Original: value,
				Current:  newValue,
				Rule:     rule,
			}
		}
	}

	return result, overrides
}