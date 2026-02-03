//go:build !test

package plugins

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

//go:embed scripts/create_venv.sh
var create_venv string

func createVenv(absPath string, requirements *string, pyVersion *string) error {
	info, err := os.Stat(path.Join(absPath, ".venv"))
	if os.IsNotExist(err) {
		log.Printf("[INFO] creating venv from scratch for '%s', expect it to take a while", absPath)
	} else {
		if !info.IsDir() {
			return fmt.Errorf("some weird file is in the way of .venv")
		}
		log.Printf("[INFO] refreshing venv for '%s'", absPath)
	}
	uvArgs := ""
	if pyVersion != nil {
		uvArgs = "--python " + *pyVersion
	}
	coreLocation, err := filepath.Abs("./plugins/freezetag-core")
	if err != nil {
		return fmt.Errorf("failed to initialize venv: %w", err)
	}
	reqs := ""
	if requirements != nil {
		reqs = *requirements
	}
	venv := exec.Command("sh", "-s", absPath, uvArgs, coreLocation, reqs)
	venv.Stdin = strings.NewReader(create_venv)
	if _, err := venv.Output(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("failed to initialize venv: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("failed to initialize venv: %w", err)
	}
	return nil
}
