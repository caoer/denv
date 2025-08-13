package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

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