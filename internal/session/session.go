package session

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/zitao/denv/internal/environment"
)

type SessionHandle struct {
	ID       string
	PID      int
	lock     *FileLock
	envPath  string
}

func CreateSession(envPath, name string) *SessionHandle {
	// Generate session ID
	baseID := generateSessionID()
	if name != "" {
		baseID = name + "-" + baseID
	}

	// Create sessions directory
	sessionsDir := filepath.Join(envPath, "sessions")
	os.MkdirAll(sessionsDir, 0755)

	// Try to acquire lock with the generated ID
	lockPath := filepath.Join(sessionsDir, baseID+".lock")
	lock, err := AcquireLock(lockPath)
	
	// If lock fails, try with different IDs
	id := baseID
	attempts := 0
	for err != nil && attempts < 10 {
		attempts++
		id = generateSessionID()
		if name != "" {
			id = name + "-" + id
		}
		lockPath = filepath.Join(sessionsDir, id+".lock")
		lock, err = AcquireLock(lockPath)
	}

	if err != nil {
		// Failed to acquire any lock, return nil
		return nil
	}

	return &SessionHandle{
		ID:      id,
		PID:     os.Getpid(),
		lock:    lock,
		envPath: envPath,
	}
}

func (s *SessionHandle) Release() {
	if s.lock != nil {
		s.lock.Release()
		s.lock = nil
	}
}

func ListSessions(envPath string) []string {
	sessionsDir := filepath.Join(envPath, "sessions")
	files, err := os.ReadDir(sessionsDir)
	if err != nil {
		return []string{}
	}

	var sessions []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".lock" {
			sessionID := strings.TrimSuffix(file.Name(), ".lock")
			sessions = append(sessions, sessionID)
		}
	}
	return sessions
}

func CleanupOrphaned(envPath string) int {
	runtime, err := environment.LoadRuntime(envPath)
	if err != nil || runtime == nil {
		return 0
	}

	cleaned := 0
	for id, session := range runtime.Sessions {
		if !ProcessExists(session.PID) {
			// Remove from runtime
			delete(runtime.Sessions, id)
			
			// Remove lock file if exists
			lockPath := filepath.Join(envPath, "sessions", id+".lock")
			os.Remove(lockPath)
			
			cleaned++
		}
	}

	if cleaned > 0 {
		environment.SaveRuntime(envPath, runtime)
	}

	return cleaned
}

func ProcessExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	// On Unix, sending signal 0 checks if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func generateSessionID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}