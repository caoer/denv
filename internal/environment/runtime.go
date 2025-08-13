package environment

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Override struct {
	Original string `json:"original"`
	Current  string `json:"current"`
	Rule     string `json:"rule"`
}

type Session struct {
	ID      string    `json:"id"`
	PID     int       `json:"pid"`
	Started time.Time `json:"started"`
	TTY     string    `json:"tty,omitempty"`
}

type Runtime struct {
	Created     time.Time            `json:"created"`
	Project     string               `json:"project"`
	Environment string               `json:"environment"`
	Ports       map[int]int          `json:"ports"`
	Overrides   map[string]Override  `json:"overrides"`
	Sessions    map[string]Session   `json:"sessions"`
}

func SaveRuntime(envPath string, runtime *Runtime) error {
	data, err := json.MarshalIndent(runtime, "", "  ")
	if err != nil {
		return err
	}

	runtimePath := filepath.Join(envPath, "runtime.json")
	return os.WriteFile(runtimePath, data, 0644)
}

func LoadRuntime(envPath string) (*Runtime, error) {
	runtimePath := filepath.Join(envPath, "runtime.json")
	data, err := os.ReadFile(runtimePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var runtime Runtime
	if err := json.Unmarshal(data, &runtime); err != nil {
		return nil, err
	}

	return &runtime, nil
}

func NewRuntime(project, environment string) *Runtime {
	return &Runtime{
		Created:     time.Now(),
		Project:     project,
		Environment: environment,
		Ports:       make(map[int]int),
		Overrides:   make(map[string]Override),
		Sessions:    make(map[string]Session),
	}
}