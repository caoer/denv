package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DenvHome() string {
	if home := os.Getenv("DENV_HOME"); home != "" {
		return home
	}
	return filepath.Join(os.Getenv("HOME"), ".denv")
}

func ProjectPath(project string) string {
	return filepath.Join(DenvHome(), project)
}

func EnvironmentPath(project, env string) string {
	return filepath.Join(DenvHome(), fmt.Sprintf("%s-%s", project, env))
}

// ShortenPath shortens a path by replacing the home directory with ~ and optionally limiting segments
// maxSegments controls how many path segments to show after ~/ (0 means no limit)
// For paths with more segments than the limit, it shows first segment, ..., and last segment
func ShortenPath(path string, maxSegments int) string {
	if path == "" {
		return ""
	}
	
	// If already shortened, return as is
	if strings.HasPrefix(path, "~/") || path == "~" {
		return path
	}
	
	home := os.Getenv("HOME")
	if home == "" {
		return path
	}
	
	// Replace home directory with ~
	if path == home {
		return "~"
	}
	
	if strings.HasPrefix(path, home+"/") {
		shortened := "~" + path[len(home):]
		
		// If no segment limit, return the shortened path
		if maxSegments <= 0 {
			// For deep paths in Projects folder, skip "Projects" to make it shorter
			// But only if the path has more than 3 segments after home
			segments := strings.Split(path[len(home)+1:], "/")
			if len(segments) > 3 && segments[0] == "Projects" {
				return "~" + shortened[10:] // Skip "/Projects"
			}
			return shortened
		}
		
		// Apply segment limit
		relativePath := path[len(home)+1:] // Path after home without leading /
		segments := strings.Split(relativePath, "/")
		
		// Skip "Projects" prefix if present and we have many segments
		skipProjects := false
		if len(segments) > 3 && segments[0] == "Projects" {
			segments = segments[1:]
			skipProjects = true
		}
		
		if len(segments) <= maxSegments {
			if skipProjects {
				return "~/" + strings.Join(segments, "/")
			}
			return "~/" + strings.Join(strings.Split(relativePath, "/"), "/")
		}
		
		// Show first segment, ellipsis, and last segment
		if maxSegments == 1 && len(segments) > 1 {
			return "~/" + segments[0] + "/.../" + segments[len(segments)-1]
		}
		
		// Show first n-1 segments and last segment with ellipsis
		result := "~/" + segments[0]
		for i := 1; i < maxSegments-1 && i < len(segments)-1; i++ {
			result += "/" + segments[i]
		}
		result += "/.../" + segments[len(segments)-1]
		return result
	}
	
	return path
}