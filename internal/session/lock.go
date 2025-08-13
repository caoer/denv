package session

import (
	"fmt"
	"os"
	"syscall"
)

// FileLock represents a file-based lock
type FileLock struct {
	file *os.File
	path string
}

// AcquireLock attempts to acquire an exclusive lock on a file
func AcquireLock(path string) (*FileLock, error) {
	// Open or create the lock file
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	// Try to acquire exclusive lock (non-blocking)
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		file.Close()
		if err == syscall.EWOULDBLOCK {
			return nil, fmt.Errorf("lock is already held")
		}
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return &FileLock{
		file: file,
		path: path,
	}, nil
}

// Release releases the lock and closes the file
func (l *FileLock) Release() error {
	if l.file == nil {
		return nil
	}

	// Release the lock
	syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	
	// Close the file
	err := l.file.Close()
	l.file = nil
	
	// Optionally remove the lock file
	os.Remove(l.path)
	
	return err
}

// TryAcquireLock attempts to acquire a lock with retries
func TryAcquireLock(path string, maxAttempts int) (*FileLock, error) {
	for i := 0; i < maxAttempts; i++ {
		lock, err := AcquireLock(path)
		if err == nil {
			return lock, nil
		}
		if i < maxAttempts-1 {
			// Generate a new path with a suffix
			newPath := fmt.Sprintf("%s.%d", path, i+1)
			lock, err = AcquireLock(newPath)
			if err == nil {
				return lock, nil
			}
		}
	}
	return nil, fmt.Errorf("failed to acquire lock after %d attempts", maxAttempts)
}