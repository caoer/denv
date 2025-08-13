package environment

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSaveLoadRuntime(t *testing.T) {
	tmpDir := t.TempDir()
	runtime := &Runtime{
		Created:     time.Now(),
		Project:     "myproject",
		Environment: "default",
		Ports: map[int]int{
			3000: 33000,
			5432: 35432,
		},
		Overrides: map[string]Override{
			"DATABASE_URL": {
				Original: "postgres://localhost:5432/db",
				Current:  "postgres://localhost:35432/db",
				Rule:     "rewrite_ports",
			},
		},
		Sessions: map[string]Session{},
	}

	// Test: Save and load
	err := SaveRuntime(tmpDir, runtime)
	assert.NoError(t, err)

	loaded, err := LoadRuntime(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, runtime.Project, loaded.Project)
	assert.Equal(t, runtime.Environment, loaded.Environment)
	assert.Equal(t, runtime.Ports[3000], loaded.Ports[3000])
	assert.Equal(t, runtime.Overrides["DATABASE_URL"].Current, loaded.Overrides["DATABASE_URL"].Current)
}

func TestLoadRuntimeNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Test: Should return nil when file doesn't exist
	runtime, err := LoadRuntime(tmpDir)
	assert.NoError(t, err)
	assert.Nil(t, runtime)
}