package plugins

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
)

type HookedPlugin struct {
	Plugin
	manifest PluginManifest
	stopped  bool
	subs     []chan<- struct{}
}

// Get all the hooks of the specified type and signature
func (h *HookedPlugin) GetHooks(kind HookType, sig HookSignature) []string {
	var hooks []string
	for hook, info := range h.manifest.Hooks {
		if info.Signature == sig && info.Type == kind {
			hooks = append(hooks, hook)
		}
	}
	return hooks
}

func (h *HookedPlugin) HookDetails(name string) PluginHook {
	return h.manifest.Hooks[name]
}

func (h *HookedPlugin) WaitFinished() <-chan struct{} {
	sub := make(chan struct{}, 1)
	if h.stopped {
		sub <- struct{}{}
		close(sub)
	} else {
		h.subs = append(h.subs, sub)
	}
	return sub
}

func (h *HookedPlugin) Shutdown() error {
	if h.stopped {
		return nil
	}
	err := h.Plugin.Shutdown()
	h.stopped = true
	for _, sub := range h.subs {
		sub <- struct{}{}
		close(sub)
	}
	return err
}

//go:embed scripts/launch_plugin.sh
var launch_plugin string

// Initialize a plugin from a loaded manifest file and a context (for cancel)
// This function will initialize the virtual environment if it doesn't already exist, install requirements, and launch the plugin,
// returning a hook-compatible fully initialized plugin.
func PluginFromManifest(manifest PluginManifest, ctx context.Context) (HookedPlugin, error) {
	// initialize venv (if it doesn't exist)
	if err := createVenv(manifest.AbsPath, manifest.Requirements, manifest.PythonVersion); err != nil {
		return HookedPlugin{}, fmt.Errorf("failed to load plugin: %w", err)
	}
	launchScript, err := os.CreateTemp("", "ftls*.sh")
	if err != nil {
		return HookedPlugin{}, fmt.Errorf("failed to load plugin: %w", err)
	}
	defer os.Remove(launchScript.Name()) //nolint:errcheck
	_, err = launchScript.Write([]byte(launch_plugin))
	if err != nil {
		return HookedPlugin{}, fmt.Errorf("failed to load plugin: %w", err)
	}
	ctx2, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx2, "sh", launchScript.Name(), manifest.AbsPath, manifest.MainFile)
	// cmd := exec.CommandContext(ctx2, path.Join(manifest.AbsPath, ".venv", "bin", "python"), path.Join(manifest.AbsPath, manifest.MainFile))
	plugin, err := PluginFromProcess(manifest.Name, cmd, cancel)
	if err != nil {
		return HookedPlugin{}, err
	}
	return HookedPlugin{plugin, manifest, false, nil}, nil
}

type pythonPlugin struct {
	name          string
	process       *exec.Cmd
	io            PluginIo
	ioCloser      func()
	processCloser context.CancelFunc
}

// Initialize a plugin from a command that has not run yet.
// This function will run the command and capture I/O.
func PluginFromProcess(name string, process *exec.Cmd, cancel context.CancelFunc) (Plugin, error) {
	in, err := process.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := process.StdoutPipe()
	if err != nil {
		return nil, err
	}
	process.Stderr = os.Stderr
	io, ioCloser := protocolFromPipes(in, out)
	err = process.Start()
	if err != nil {
		return nil, err
	}
	io.In <- PluginMessage{READY, nil}
readyLoop:
	for {
		msg, ok := <-io.Out
		if !ok {
			log.Printf("[ERR]  %s: failed to read from stdout during plugin init", name)
			goto initProblem
		}
		switch msg.Type {
		case ERR:
			log.Printf("[ERR]  %s: %s", name, string(msg.Contents.([]byte)))
			goto initProblem
		case LOG:
			log.Printf("[PLUG] %s: %s", name, string(msg.Contents.([]byte)))
		case READY:
			break readyLoop
		default:
			log.Printf("[ERR]  %s: bad init message from plugin", name)
			goto initProblem
		}
	}
	return pythonPlugin{name, process, io, ioCloser, cancel}, nil
initProblem:
	ioCloser()
	cancel()
	return nil, fmt.Errorf("plugin failed to initialize")
}

func (pp pythonPlugin) Shutdown() error {
	pp.io.In <- PluginMessage{SHUTDOWN, nil}
shutdownLoop:
	for {
		msg, ok := <-pp.io.Out
		if !ok {
			log.Printf("[ERR]  %s: plugin closed stdout before shutdown", pp.name)
			pp.ioCloser()
			pp.processCloser()
			return fmt.Errorf("plugin closed stdout before shutdown")
		}
		switch msg.Type {
		case SHUTDOWN:
			break shutdownLoop
		case ERR:
			log.Printf("[ERR]  %s: %s", pp.name, string(msg.Contents.([]byte)))
			pp.ioCloser()
			pp.processCloser()
			return fmt.Errorf("failed to shut down plugin gracefully")
		case LOG:
			log.Printf("[PLUG] %s: %s", pp.name, string(msg.Contents.([]byte)))
		default:
			log.Printf("[ERR]  %s: bad shutdown message from plugin", pp.name)
		}
	}
	pp.ioCloser()
	pp.processCloser()
	return nil
}
