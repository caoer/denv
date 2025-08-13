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