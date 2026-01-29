//go:build !test

package plugins

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"path"
)

func createVenv(absPath string, requirements *string, pyVersion *string) error {
	log.Printf("[INFO] creating venv from scratch for '%s', expect it to take a while", absPath)
	args := []string{"venv", "--managed-python", "--seed", path.Join(absPath, ".venv")}
	if pyVersion != nil {
		args = append(args, "--python", *pyVersion)
	}
	if _, err := exec.Command("uv", args...).Output(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("failed to initialize venv: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("failed to initialize venv: %w", err)
	}
	if _, err := exec.Command(path.Join(absPath, ".venv", "bin", "pip"), "install", "./plugins/freezetag-core").Output(); err != nil {
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
