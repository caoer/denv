package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionExists(t *testing.T) {
	t.Run("should detect current process as existing", func(t *testing.T) {
		// Get current process PID which definitely exists
		currentPID := os.Getpid()
		
		// This should return true for the current process
		exists := sessionExists(currentPID)
		assert.True(t, exists, "Current process should be detected as existing")
	})
	
	t.Run("should detect non-existent process as not existing", func(t *testing.T) {
		// Use an impossible PID that won't exist
		nonExistentPID := 999999999
		
		// This should return false for a non-existent process
		exists := sessionExists(nonExistentPID)
		assert.False(t, exists, "Non-existent process should be detected as not existing")
	})
	
	t.Run("should detect parent process as existing", func(t *testing.T) {
		// Get parent process PID which should exist
		parentPID := os.Getppid()
		
		// This should return true for the parent process
		exists := sessionExists(parentPID)
		assert.True(t, exists, "Parent process should be detected as existing")
	})
}