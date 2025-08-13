package testutil

import (
	"os/exec"
	"testing"
)

func RunCmd(t *testing.T, dir string, command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %s %v\nOutput: %s\nError: %v", command, args, output, err)
	}
}