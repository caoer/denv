package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDenvHome(t *testing.T) {
	// Test 1: Default to ~/.denv
	os.Unsetenv("DENV_HOME")
	assert.Equal(t, filepath.Join(os.Getenv("HOME"), ".denv"), DenvHome())

	// Test 2: Respect DENV_HOME env var
	os.Setenv("DENV_HOME", "/custom/path")
	assert.Equal(t, "/custom/path", DenvHome())
}

func TestProjectPath(t *testing.T) {
	// Test: Project path construction
	home := DenvHome()
	assert.Equal(t, filepath.Join(home, "myproject"), ProjectPath("myproject"))
}

func TestEnvironmentPath(t *testing.T) {
	// Test: Environment path construction
	home := DenvHome()
	assert.Equal(t, filepath.Join(home, "myproject-default"),
		EnvironmentPath("myproject", "default"))
}

func TestShortenPath(t *testing.T) {
	home := os.Getenv("HOME")
	
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "home directory path",
			path:     home + "/Projects/myproject",
			expected: "~/Projects/myproject",
		},
		{
			name:     "deep nested path in home",
			path:     home + "/Projects/placeholder-soft/lockin-workspace/bot-data-root",
			expected: "~/placeholder-soft/lockin-workspace/bot-data-root",
		},
		{
			name:     "path not in home directory",
			path:     "/var/log/system.log",
			expected: "/var/log/system.log",
		},
		{
			name:     "already shortened path",
			path:     "~/Documents/file.txt",
			expected: "~/Documents/file.txt",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "just home directory",
			path:     home,
			expected: "~",
		},
		{
			name:     "home with trailing slash",
			path:     home + "/",
			expected: "~/",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShortenPath(tt.path, 0)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShortenPathWithMaxSegments(t *testing.T) {
	home := os.Getenv("HOME")
	
	tests := []struct {
		name        string
		path        string
		maxSegments int
		expected    string
	}{
		{
			name:        "limit to 2 segments after home",
			path:        home + "/Projects/placeholder-soft/lockin-workspace/bot-data-root",
			maxSegments: 2,
			expected:    "~/placeholder-soft/.../bot-data-root",
		},
		{
			name:        "limit to 3 segments after home",
			path:        home + "/Projects/placeholder-soft/lockin-workspace/bot-data-root",
			maxSegments: 3,
			expected:    "~/placeholder-soft/lockin-workspace/bot-data-root",
		},
		{
			name:        "path shorter than limit",
			path:        home + "/Projects/myapp",
			maxSegments: 5,
			expected:    "~/Projects/myapp",
		},
		{
			name:        "non-home path with limit",
			path:        "/var/log/nginx/access.log",
			maxSegments: 2,
			expected:    "/var/log/nginx/access.log",
		},
		{
			name:        "limit to 1 segment shows first and last",
			path:        home + "/a/b/c/d/e/f",
			maxSegments: 1,
			expected:    "~/a/.../f",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShortenPath(tt.path, tt.maxSegments)
			assert.Equal(t, tt.expected, result)
		})
	}
}