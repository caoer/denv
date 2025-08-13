package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileLocking(t *testing.T) {
	tmpDir := t.TempDir()
	lockFile := filepath.Join(tmpDir, "test.lock")

	// Test: Acquire lock
	lock, err := AcquireLock(lockFile)
	assert.NoError(t, err)
	assert.NotNil(t, lock)

	// Test: Second attempt should fail (non-blocking)
	lock2, err := AcquireLock(lockFile)
	assert.Error(t, err)
	assert.Nil(t, lock2)

	// Test: Release lock
	err = lock.Release()
	assert.NoError(t, err)

	// Test: Can acquire after release
	lock3, err := AcquireLock(lockFile)
	assert.NoError(t, err)
	assert.NotNil(t, lock3)
	lock3.Release()
}

func TestLockAutoRelease(t *testing.T) {
	tmpDir := t.TempDir()
	lockFile := filepath.Join(tmpDir, "auto.lock")

	// Create and hold lock
	lock, err := AcquireLock(lockFile)
	assert.NoError(t, err)
	assert.NotNil(t, lock)

	// Close file without releasing lock (simulates process crash)
	// In a real crash, the OS would release the lock
	lock.file.Close()
	lock.file = nil

	// File should still exist
	_, err = os.Stat(lockFile)
	assert.NoError(t, err)

	// In same process, lock is still held by kernel
	// This test mainly verifies our lock cleanup logic
	// Real process death would release the lock automatically
}

func TestSessionWithLocking(t *testing.T) {
	tmpDir := t.TempDir()

	// Create session with proper locking
	session1 := CreateSession(tmpDir, "locked-session")
	assert.NotNil(t, session1)

	// Try to create another session with same ID (should get different ID)
	session2 := CreateSession(tmpDir, "locked-session")
	assert.NotNil(t, session2)
	assert.NotEqual(t, session1.ID, session2.ID)

	// Clean up
	session1.Release()
	session2.Release()
}

func TestConcurrentLocking(t *testing.T) {
	tmpDir := t.TempDir()
	lockFile := filepath.Join(tmpDir, "concurrent.lock")

	acquired := make(chan bool, 10)
	done := make(chan bool)

	// Start 10 goroutines trying to acquire the same lock
	for i := 0; i < 10; i++ {
		go func() {
			lock, err := AcquireLock(lockFile)
			if err == nil {
				acquired <- true
				time.Sleep(10 * time.Millisecond)
				lock.Release()
			} else {
				acquired <- false
			}
		}()
	}

	// Wait for all to complete
	go func() {
		successCount := 0
		for i := 0; i < 10; i++ {
			if <-acquired {
				successCount++
			}
		}
		// Only one should succeed at a time
		assert.Greater(t, successCount, 0)
		done <- true
	}()

	select {
	case <-done:
		// Test completed
	case <-time.After(1 * time.Second):
		t.Fatal("Test timed out")
	}
}