package plugins

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

type pythonPlugin struct {
	name     string
	process  *exec.Cmd
	io       chan PluginMessage
	ioCloser func()
}

// Initialize a plugin from a command that has not run yet.
// This function will run the command and capture I/O.
func Init(name string, process *exec.Cmd, cancel context.CancelFunc) (Plugin, error) {
	in, err := process.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := process.StdoutPipe()
	if err != nil {
		return nil, err
	}
	io, ioCloser := protocolFromPipes(in, out)
	err = process.Start()
	if err != nil {
		return nil, err
	}
	io <- PluginMessage{READY, nil}
readyLoop:
	for {
		msg, ok := <-io
		if !ok {
			goto initProblem
		}
		switch msg.Type {
		case ERR:
			log.Printf("%s [ERR]: %s", name, string(msg.Contents.([]byte)))
			goto initProblem
		case LOG:
			log.Printf("%s: %s", name, string(msg.Contents.([]byte)))
		case READY:
			break readyLoop
		default:
			log.Printf("%s [ERR]: bad init message from plugin", name)
			goto initProblem
		}
	}
	return pythonPlugin{name, process, io, ioCloser}, nil
initProblem:
	cancel()
	return nil, fmt.Errorf("Plugin failed to initialize")
}
