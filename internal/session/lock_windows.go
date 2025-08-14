//go:build windows
// +build windows

package session

import (
	"fmt"
	"os"
)

// FileLock represents a file-based lock (Windows version)
type FileLock struct {
	path string
	file *os.File
}

// AcquireLock attempts to acquire an exclusive lock on the given file
// On Windows, we use a simple file-based lock without syscall.Flock
func AcquireLock(path string) (*FileLock, error) {
	// Try to open the file exclusively
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("lock already held")
		}
		return nil, err
	}
	
	return &FileLock{
		path: path,
		file: file,
	}, nil
}

// Release releases the lock
func (l *FileLock) Release() error {
	if l.file == nil {
		return nil
	}

	// Close the file
	err := l.file.Close()
	l.file = nil
	
	// Remove the lock file
	os.Remove(l.path)
	
	return err
}