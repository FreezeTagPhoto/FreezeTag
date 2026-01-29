//go:build test

package plugins

import (
	"errors"
	"fmt"
	"os/exec"
	"path"
)

func createVenv(absPath string, requirements *string, _ *string) error {
	if _, err := exec.Command("uv", "venv", "--managed-python", "--seed", path.Join(absPath, ".venv")).Output(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("failed to initialize venv: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("failed to initialize venv: %w", err)
	}
	if _, err := exec.Command(path.Join(absPath, ".venv", "bin", "pip"), "install", "../../plugins/freezetag-core").Output(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("failed to initialize venv: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("failed to install requirements inside venv: %w", err)
	}
	if requirements != nil {
		if _, err := exec.Command(path.Join(absPath, ".venv", "bin", "pip"), "install", "-r", path.Join(absPath, *requirements)).Output(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return fmt.Errorf("failed to initialize venv: %s", string(exitErr.Stderr))
			}
			return fmt.Errorf("failed to install requirements inside venv: %w", err)
		}
	}
	return nil
}
