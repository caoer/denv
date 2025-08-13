package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zitao/denv/internal/environment"
)

func TestCreateSession(t *testing.T) {
	tmpDir := t.TempDir()

	session := CreateSession(tmpDir, "test-session")
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, os.Getpid(), session.PID)

	// Test: Lock file should exist
	lockPath := filepath.Join(tmpDir, "sessions", session.ID+".lock")
	assert.FileExists(t, lockPath)
	
	// Clean up
	session.Release()
}

func TestListSessions(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create multiple sessions
	s1 := CreateSession(tmpDir, "session1")
	s2 := CreateSession(tmpDir, "session2")
	
	// Test: Should list both sessions
	sessions := ListSessions(tmpDir)
	assert.Len(t, sessions, 2)
	
	// Clean up
	s1.Release()
	s2.Release()
}

func TestCleanupOrphanedSessions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create fake orphaned session in runtime
	runtime := &environment.Runtime{
		Sessions: map[string]environment.Session{
			"dead-session": {
				PID: 99999, // Non-existent PID
				ID:  "dead-session",
			},
			"alive-session": {
				PID: os.Getpid(), // Current process
				ID:  "alive-session",
			},
		},
	}
	environment.SaveRuntime(tmpDir, runtime)

	// Test: Should detect and clean only orphaned
	cleaned := CleanupOrphaned(tmpDir)
	assert.Equal(t, 1, cleaned)

	// Test: Dead session should be removed, alive should remain
	runtime2, _ := environment.LoadRuntime(tmpDir)
	assert.Len(t, runtime2.Sessions, 1)
	assert.Contains(t, runtime2.Sessions, "alive-session")
}

func TestProcessExists(t *testing.T) {
	// Test: Current process exists
	assert.True(t, ProcessExists(os.Getpid()))
	
	// Test: Non-existent process
	assert.False(t, ProcessExists(99999))
}