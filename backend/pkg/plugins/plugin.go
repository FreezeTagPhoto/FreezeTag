package plugins

import "os/exec"

type pythonPlugin struct {
	process  *exec.Cmd
	io       chan PluginMessage
	ioCloser func()
}

// Initialize a plugin from a command that has not run yet.
// This function will run the command and capture I/O.
func Init(process *exec.Cmd) (Plugin, error) {
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
	return pythonPlugin{process, io, ioCloser}, nil
}
