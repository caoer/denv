# denv - Test-Driven Development Guide

## Phase 0: Project Setup

### Step 0.1: Initialize Go Project
```bash
mkdir denv && cd denv
go mod init github.com/caoer/denv
```

### Step 0.2: Setup Test Framework
```bash
# Create test helpers
mkdir -p internal/testutil

# Install test dependencies
go get github.com/stretchr/testify
```

### Step 0.3: Create Basic Structure
```bash
mkdir -p cmd/denv
mkdir -p internal/{config,project,environment,ports,session,hooks}

# Create main entry point
touch cmd/denv/main.go
echo 'package main; func main() { println("denv") }' > cmd/denv/main.go

# Verify it builds
go build ./cmd/denv
./denv  # Should print "denv"
```

### Step 0.4: Setup Test Runner
```bash
# Create Makefile
cat > Makefile << 'EOF'
test:
	go test -v ./...

test-watch:
	find . -name "*.go" | entr -c go test -v ./...

build:
	go build -o denv ./cmd/denv

.PHONY: test test-watch build
EOF

# Run empty test suite
make test  # Should pass with no tests
```

## Phase 1: Core Utilities

### Step 1.1: Path Management
**Test First:**
```go
// internal/paths/paths_test.go
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
```

**Run test:** `go test ./internal/paths -v` (should fail)

**Implement:**
```go
// internal/paths/paths.go
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
```

**Run test:** `go test ./internal/paths -v` (should pass)

### Step 1.2: Config Loading
**Test First:**
```go
// internal/config/config_test.go
func TestLoadConfig(t *testing.T) {
    // Setup: Create temp config file
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "config.yaml")
    
    yaml := `
projects:
  /path/to/project: custom-name
patterns:
  "*_PORT|PORT":
    action: random_port
    range: [30000, 39999]
`
    os.WriteFile(configPath, []byte(yaml), 0644)
    
    // Test: Load config
    cfg, err := LoadConfig(configPath)
    assert.NoError(t, err)
    assert.Equal(t, "custom-name", cfg.Projects["/path/to/project"])
    assert.Equal(t, "random_port", cfg.Patterns["*_PORT|PORT"].Action)
}

func TestDefaultConfig(t *testing.T) {
    // Test: Returns defaults when no config exists
    cfg, err := LoadConfig("/nonexistent/path")
    assert.NoError(t, err)
    assert.NotNil(t, cfg.Patterns)
    assert.Contains(t, cfg.Patterns, "*_PORT|PORT")
}
```

**Implement config loading...**

## Phase 2: Project Detection

### Step 2.1: Git Detection
**Test First:**
```go
// internal/project/detect_test.go
func TestDetectGitProject(t *testing.T) {
    // Setup: Create temp git repo
    tmpDir := t.TempDir()
    runCmd(tmpDir, "git", "init")
    runCmd(tmpDir, "git", "remote", "add", "origin", "https://github.com/user/myproject.git")
    
    // Test: Should detect project name from git remote
    name, err := DetectProject(tmpDir)
    assert.NoError(t, err)
    assert.Equal(t, "myproject", name)
}

func TestDetectGitWorktree(t *testing.T) {
    // Setup: Create main repo and worktree
    mainDir := t.TempDir() + "/main"
    worktreeDir := t.TempDir() + "/worktree"
    
    runCmd(mainDir, "git", "init")
    runCmd(mainDir, "git", "remote", "add", "origin", "https://github.com/user/myproject.git")
    runCmd(mainDir, "git", "commit", "--allow-empty", "-m", "init")
    runCmd(mainDir, "git", "worktree", "add", worktreeDir)
    
    // Test: Both should detect same project
    mainName, _ := DetectProject(mainDir)
    worktreeName, _ := DetectProject(worktreeDir)
    assert.Equal(t, mainName, worktreeName)
    assert.Equal(t, "myproject", mainName)
}
```

**Implement project detection...**

### Step 2.2: Config Override
**Test First:**
```go
func TestDetectWithOverride(t *testing.T) {
    // Setup: Create config with override
    cfg := &Config{
        Projects: map[string]string{
            "/my/path": "custom-project",
        },
    }
    
    // Test: Should use override
    name := DetectProjectWithConfig("/my/path", cfg)
    assert.Equal(t, "custom-project", name)
}
```

## Phase 3: Environment State Management

### Step 3.1: Runtime State
**Test First:**
```go
// internal/environment/runtime_test.go
func TestSaveLoadRuntime(t *testing.T) {
    tmpDir := t.TempDir()
    runtime := &Runtime{
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
    }
    
    // Test: Save and load
    err := SaveRuntime(tmpDir, runtime)
    assert.NoError(t, err)
    
    loaded, err := LoadRuntime(tmpDir)
    assert.NoError(t, err)
    assert.Equal(t, runtime.Project, loaded.Project)
    assert.Equal(t, runtime.Ports[3000], loaded.Ports[3000])
}
```

## Phase 4: Port Management

### Step 4.1: Port Assignment
**Test First:**
```go
// internal/ports/ports_test.go
func TestFindFreePort(t *testing.T) {
    // Test: Should find a free port
    port := FindFreePort(30000, 40000)
    assert.Greater(t, port, 30000)
    assert.Less(t, port, 40000)
    
    // Test: Port should actually be free
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    assert.NoError(t, err)
    ln.Close()
}

func TestPortPersistence(t *testing.T) {
    tmpDir := t.TempDir()
    pm := NewPortManager(tmpDir)
    
    // Test: Assign and persist
    port1 := pm.GetPort(3000)
    assert.Greater(t, port1, 30000)
    
    // Test: Same port on reload
    pm2 := NewPortManager(tmpDir)
    port2 := pm2.GetPort(3000)
    assert.Equal(t, port1, port2)
}
```

### Step 4.2: Port Conflict Detection
**Test First:**
```go
func TestPortConflict(t *testing.T) {
    // Start a server on a port
    ln, _ := net.Listen("tcp", ":31234")
    defer ln.Close()
    
    // Test: Should detect port is in use
    assert.False(t, IsPortAvailable(31234))
    
    // Test: Should skip busy port
    pm := NewPortManager(t.TempDir())
    pm.SetRange(31234, 31235)  // Very narrow range
    port := pm.GetPort(3000)
    assert.NotEqual(t, 31234, port)
}
```

## Phase 5: Variable Override System

### Step 5.1: Pattern Matching
**Test First:**
```go
// internal/override/pattern_test.go
func TestPatternMatch(t *testing.T) {
    tests := []struct {
        pattern string
        key     string
        match   bool
    }{
        {"*_PORT", "DB_PORT", true},
        {"*_PORT", "PORT_DB", false},
        {"*_PORT|PORT", "PORT", true},
        {"DATABASE_URL", "DATABASE_URL", true},
        {"*_URL", "DATABASE_URL", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.key, func(t *testing.T) {
            assert.Equal(t, tt.match, MatchesPattern(tt.pattern, tt.key))
        })
    }
}
```

### Step 5.2: URL Rewriting
**Test First:**
```go
func TestRewriteURLPorts(t *testing.T) {
    ports := map[int]int{
        5432: 35432,
        3000: 33000,
    }
    
    tests := []struct {
        input    string
        expected string
    }{
        {
            "postgres://localhost:5432/db",
            "postgres://localhost:35432/db",
        },
        {
            "http://127.0.0.1:3000/api",
            "http://127.0.0.1:33000/api",
        },
        {
            "redis://external.com:6379",
            "redis://external.com:6379", // Don't change external
        },
    }
    
    for _, tt := range tests {
        result := RewriteURL(tt.input, ports)
        assert.Equal(t, tt.expected, result)
    }
}
```

### Step 5.3: Apply Override Rules
**Test First:**
```go
func TestApplyRules(t *testing.T) {
    cfg := &Config{
        Patterns: map[string]Rule{
            "*_PORT": {Action: "random_port"},
            "*_URL": {Action: "rewrite_ports"},
            "*_KEY": {Action: "keep"},
        },
    }
    
    env := map[string]string{
        "DB_PORT": "5432",
        "DATABASE_URL": "postgres://localhost:5432/db",
        "API_KEY": "secret123",
    }
    
    ports := map[int]int{5432: 35432}
    result := ApplyRules(env, cfg, ports)
    
    assert.Equal(t, "35432", result["DB_PORT"])
    assert.Contains(t, result["DATABASE_URL"], "35432")
    assert.Equal(t, "secret123", result["API_KEY"])  // Unchanged
}
```

## Phase 6: Session Management

### Step 6.1: Session Creation
**Test First:**
```go
// internal/session/session_test.go
func TestCreateSession(t *testing.T) {
    tmpDir := t.TempDir()
    
    session := CreateSession(tmpDir, "test-session")
    assert.NotEmpty(t, session.ID)
    assert.Equal(t, os.Getpid(), session.PID)
    
    // Test: Lock file should exist
    lockPath := filepath.Join(tmpDir, "sessions", session.ID+".lock")
    assert.FileExists(t, lockPath)
}
```

### Step 6.2: Lock File Management
**Manual Test (since we can't easily test OS cleanup):**
```go
// cmd/testlock/main.go - Temporary test program
func main() {
    dir := "/tmp/denv-test"
    os.MkdirAll(dir, 0755)
    
    // Create lock
    lockFile := filepath.Join(dir, "test.lock")
    f, _ := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL, 0644)
    
    fmt.Println("Lock created, kill this process...")
    time.Sleep(1 * time.Hour)
}

// Run in one terminal: go run cmd/testlock/main.go
// In another: kill -9 <pid>
// Verify: Lock file should be gone (OS cleanup)
```

### Step 6.3: Session Cleanup
**Test First:**
```go
func TestCleanupOrphanedSessions(t *testing.T) {
    tmpDir := t.TempDir()
    
    // Create fake orphaned session
    runtime := &Runtime{
        Sessions: map[string]Session{
            "dead-session": {
                PID: 99999, // Non-existent PID
                ID:  "dead-session",
            },
        },
    }
    SaveRuntime(tmpDir, runtime)
    
    // Test: Should detect and clean
    cleaned := CleanupOrphaned(tmpDir)
    assert.Equal(t, 1, cleaned)
    
    // Test: Session should be removed
    runtime2, _ := LoadRuntime(tmpDir)
    assert.Empty(t, runtime2.Sessions)
}
```

## Phase 7: Shell Integration

### Step 7.1: Shell Wrapper Script
**Test First:**
```go
// internal/shell/wrapper_test.go
func TestGenerateWrapper(t *testing.T) {
    env := map[string]string{
        "DENV_ENV": "test",
        "DENV_PROJECT": "/home/user/.denv/myproject",
    }
    
    script := GenerateWrapper(env)
    
    // Test: Should contain signal traps
    assert.Contains(t, script, "trap")
    assert.Contains(t, script, "EXIT")
    assert.Contains(t, script, "SIGTERM")
    
    // Test: Should source hooks
    assert.Contains(t, script, "on-enter.sh")
    assert.Contains(t, script, "on-exit.sh")
}
```

### Step 7.2: Signal Handling
**Integration Test:**
```go
func TestSignalHandling(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    tmpDir := t.TempDir()
    
    // Create exit hook that writes a file
    hookDir := filepath.Join(tmpDir, "hooks")
    os.MkdirAll(hookDir, 0755)
    exitHook := filepath.Join(hookDir, "on-exit.sh")
    os.WriteFile(exitHook, []byte(`
        echo "cleaned" > /tmp/denv-test-cleanup
    `), 0755)
    
    // Start denv session in subprocess
    cmd := exec.Command("./denv", "enter", "test")
    cmd.Env = append(os.Environ(), 
        "DENV_HOME="+tmpDir,
        "DENV_TEST_MODE=1",  // Skip interactive shell
    )
    cmd.Start()
    
    time.Sleep(100 * time.Millisecond)
    
    // Send SIGTERM
    cmd.Process.Signal(syscall.SIGTERM)
    cmd.Wait()
    
    // Test: Exit hook should have run
    content, _ := os.ReadFile("/tmp/denv-test-cleanup")
    assert.Equal(t, "cleaned\n", string(content))
}
```

## Phase 8: Commands Implementation

### Step 8.1: Enter Command
**Test First:**
```go
// cmd/denv/enter_test.go
func TestEnterCommand(t *testing.T) {
    tmpDir := t.TempDir()
    os.Setenv("DENV_HOME", tmpDir)
    os.Setenv("DENV_TEST_MODE", "1")  // Don't spawn shell
    
    // Test: First enter creates environment
    err := RunEnter("default")
    assert.NoError(t, err)
    
    // Test: Environment should exist
    envPath := filepath.Join(tmpDir, "testproject-default")
    assert.DirExists(t, envPath)
    
    // Test: Runtime should be saved
    runtime, err := LoadRuntime(envPath)
    assert.NoError(t, err)
    assert.NotEmpty(t, runtime.Ports)
}
```

### Step 8.2: Project Name Prompt
**Test with Mock Input:**
```go
func TestProjectNamePrompt(t *testing.T) {
    // Mock stdin
    oldStdin := os.Stdin
    r, w, _ := os.Pipe()
    os.Stdin = r
    defer func() { os.Stdin = oldStdin }()
    
    // Write mock input
    go func() {
        w.Write([]byte("rename\n"))
        w.Write([]byte("custom-name\n"))
        w.Close()
    }()
    
    // Test: Should prompt and save
    name := PromptProjectName("/path/to/project", "detected-name")
    assert.Equal(t, "custom-name", name)
    
    // Test: Should be saved in config
    cfg, _ := LoadConfig(ConfigPath())
    assert.Equal(t, "custom-name", cfg.Projects["/path/to/project"])
}
```

### Step 8.3: List Command
**Test First:**
```go
func TestListCommand(t *testing.T) {
    tmpDir := t.TempDir()
    os.Setenv("DENV_HOME", tmpDir)
    
    // Create some environments
    os.MkdirAll(filepath.Join(tmpDir, "project-default"), 0755)
    os.MkdirAll(filepath.Join(tmpDir, "project-feature"), 0755)
    
    // Test: Should list environments
    output := CaptureOutput(func() {
        RunList("project")
    })
    
    assert.Contains(t, output, "default")
    assert.Contains(t, output, "feature")
}
```

## Phase 9: Integration Tests

### Step 9.1: Full Workflow Test
```go
// integration_test.go
func TestFullWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    tmpDir := t.TempDir()
    tmpProject := t.TempDir() + "/testproject"
    os.MkdirAll(tmpProject, 0755)
    
    // Initialize git repo
    runCmd(tmpProject, "git", "init")
    runCmd(tmpProject, "git", "remote", "add", "origin", "https://github.com/user/testproject.git")
    
    // Test: Enter environment
    os.Chdir(tmpProject)
    os.Setenv("DENV_HOME", tmpDir)
    os.Setenv("DENV_TEST_MODE", "1")
    
    err := RunEnter("test-env")
    assert.NoError(t, err)
    
    // Test: Ports assigned
    runtime, _ := LoadRuntime(filepath.Join(tmpDir, "testproject-test-env"))
    assert.NotEmpty(t, runtime.Ports)
    
    // Test: Session created
    assert.NotEmpty(t, runtime.Sessions)
    
    // Test: Can list
    output := CaptureOutput(func() { RunList("") })
    assert.Contains(t, output, "test-env")
    
    // Test: Clean removes environment
    RunClean("test-env")
    assert.NoDirExists(t, filepath.Join(tmpDir, "testproject-test-env"))
}
```

### Step 9.2: Multiple Sessions Test
```go
func TestMultipleSessions(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    tmpDir := t.TempDir()
    os.Setenv("DENV_HOME", tmpDir)
    
    // Start two sessions in parallel
    var wg sync.WaitGroup
    for i := 0; i < 2; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // Each "session" creates a lock
            session := CreateSession(filepath.Join(tmpDir, "project-default"), fmt.Sprintf("session%d", id))
            defer session.Release()
            
            time.Sleep(100 * time.Millisecond)
        }(i)
    }
    
    // While sessions are running
    time.Sleep(50 * time.Millisecond)
    
    // Test: Should see 2 active sessions
    sessions := ListSessions(filepath.Join(tmpDir, "project-default"))
    assert.Len(t, sessions, 2)
    
    wg.Wait()
    
    // Test: Sessions should be cleaned up
    sessions = ListSessions(filepath.Join(tmpDir, "project-default"))
    assert.Len(t, sessions, 0)
}
```

## Phase 10: Edge Cases & Error Handling

### Step 10.1: Concurrent Port Assignment
```go
func TestConcurrentPortAssignment(t *testing.T) {
    tmpDir := t.TempDir()
    
    // Multiple goroutines trying to get ports
    var wg sync.WaitGroup
    ports := make([]int, 10)
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            pm := NewPortManager(tmpDir)
            ports[idx] = pm.GetPort(3000)
        }(i)
    }
    
    wg.Wait()
    
    // Test: All should get the same port (properly locked)
    for i := 1; i < 10; i++ {
        assert.Equal(t, ports[0], ports[i])
    }
}
```

### Step 10.2: Crash Recovery
```go
func TestCrashRecovery(t *testing.T) {
    tmpDir := t.TempDir()
    
    // Simulate a crashed session
    runtime := &Runtime{
        Sessions: map[string]Session{
            "crashed": {
                PID: os.Getpid(),  // Current process
                ID:  "crashed",
            },
        },
    }
    SaveRuntime(tmpDir, runtime)
    
    // "Reboot" - clear PID
    runtime.Sessions["crashed"].PID = 99999
    SaveRuntime(tmpDir, runtime)
    
    // Test: Should detect and recover
    RecoverFromCrash(tmpDir)
    
    runtime2, _ := LoadRuntime(tmpDir)
    assert.Empty(t, runtime2.Sessions)
}
```

## Development Workflow

### Running Tests During Development
```bash
# Run all tests
make test

# Run specific package tests
go test -v ./internal/ports

# Run with coverage
go test -v -cover ./...

# Watch mode (requires entr)
make test-watch

# Run only unit tests (skip integration)
go test -short ./...

# Run specific test
go test -v -run TestPortAssignment ./internal/ports
```

### Test Organization
```
denv/
├── cmd/denv/
│   ├── main.go
│   ├── enter_test.go
│   └── integration_test.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   ├── project/
│   │   ├── detect.go
│   │   └── detect_test.go
│   ├── ports/
│   │   ├── manager.go
│   │   └── manager_test.go
│   ├── session/
│   │   ├── session.go
│   │   └── session_test.go
│   └── testutil/
│       └── helpers.go
└── Makefile
```

## Testing Checklist

### Before Each Phase
- [ ] Write failing tests first
- [ ] Run tests to confirm they fail
- [ ] Implement minimal code to pass
- [ ] Refactor if needed
- [ ] All tests still pass

### After Each Phase  
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing works
- [ ] No race conditions (`go test -race`)
- [ ] No goroutine leaks

### Before Release
- [ ] Full test suite passes
- [ ] Coverage > 80% for critical paths
- [ ] Tested on macOS and Linux
- [ ] Tested with bash and zsh
- [ ] Tested with multiple sessions
- [ ] Tested crash recovery
- [ ] Tested signal handling

## Key Testing Principles

1. **Test behavior, not implementation**
   - Test what the user experiences
   - Don't test private functions directly

2. **Use real files when needed**
   - `t.TempDir()` for isolated file tests
   - Real git repos for git detection tests

3. **Mock external commands carefully**
   - Mock stdin for user input
   - Use `DENV_TEST_MODE` to skip shell spawning

4. **Test concurrency explicitly**
   - Port assignment race conditions
   - Session lock conflicts
   - Concurrent file access

5. **Integration tests for workflows**
   - Full enter/exit cycle
   - Multiple sessions
   - Signal handling

6. **Clean up in tests**
   - Use `t.Cleanup()` for deferred cleanup
   - Don't leave test files around