package plugins

import (
	"context"
	"io"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createTestProtocol(t *testing.T) (*io.PipeReader, *io.PipeWriter, PluginIo, func()) {
	t.Helper()
	input, stdin := io.Pipe()
	stdout, output := io.Pipe()
	proto, closer := protocolFromPipes(stdin, stdout)
	return input, output, proto, closer
}

func TestProtocolPipes(t *testing.T) {
	in, out, proto, closer := createTestProtocol(t)
	defer closer()
	proto.In <- PluginMessage{READY, nil}
	inBuf := make([]byte, 1)
	_, err := io.ReadFull(in, inBuf)
	assert.NoError(t, err)
	assert.Equal(t, []byte{byte(READY)}, inBuf)
	_, err = out.Write([]byte{byte(READY)})
	assert.NoError(t, err)
	select {
	case msg := <-proto.Out:
		assert.Equal(t, READY, msg.Type)
	case <-time.After(time.Second):
		assert.Fail(t, "Capturing output didn't work")
	}
}

func TestProtocolStringMessage(t *testing.T) {
	in, out, proto, closer := createTestProtocol(t)
	defer closer()
	proto.In <- PluginMessage{LOG, []byte("Hello, World!")}
	inBuf := make([]byte, 22)
	_, err := io.ReadFull(in, inBuf)
	assert.NoError(t, err)
	_, err = out.Write(inBuf)
	assert.NoError(t, err)
	select {
	case msg := <-proto.Out:
		assert.Equal(t, PluginMessage{LOG, []byte("Hello, World!")}, msg)
	case <-time.After(time.Second):
		assert.Fail(t, "Capturing output didn't work")
	}
}

func TestProtocolJSONMessage(t *testing.T) {
	in, out, proto, closer := createTestProtocol(t)
	defer closer()
	test := PluginMessage{GET, map[string]any{"a": 2., "b": []any{"c", "d"}}}
	proto.In <- test
	inBuf := make([]byte, 30)
	_, err := io.ReadFull(in, inBuf)
	assert.NoError(t, err)
	_, err = out.Write(inBuf)
	assert.NoError(t, err)
	select {
	case msg := <-proto.Out:
		assert.Equal(t, test, msg)
	case <-time.After(time.Second):
		assert.Fail(t, "Capturing output didn't work")
	}
}

func TestEchoedPluginInitializes(t *testing.T) {
	// this works because cat echoes stdin
	// and the simplest protocol that succeeds is READY -> READY -> SHUTDOWN -> SHUTDOWN
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "cat")
	plugin, err := InitPlugin("cat", cmd, cancel)
	assert.NoError(t, err)
	err = plugin.Shutdown()
	assert.NoError(t, err)
}

func TestEchoedPluginEchoes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "cat")
	plugin, err := InitPlugin("cat", cmd, cancel)
	assert.Equal(t, "cat", plugin.Name())
	assert.NoError(t, err)
	plugin.IO().In <- PluginMessage{BIN, []byte{1, 2, 3}}
	select {
	case msg := <-plugin.IO().Out:
		assert.Equal(t, PluginMessage{BIN, []byte{1, 2, 3}}, msg)
	case <-time.After(time.Second):
		assert.Fail(t, "capturing output didn't work")
	}
	err = plugin.Shutdown()
	assert.NoError(t, err)
}
