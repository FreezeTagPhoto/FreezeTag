package plugins

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
)

type MessageType uint8

const (
	GET MessageType = iota
	PUT
	BIN
	LOG
	ERR
	READY
	SHUTDOWN
	INVALID // anything >= INVALID is invalid
)

func (t MessageType) IsValid() bool {
	return t < INVALID
}

type PluginMessage struct {
	Type     MessageType
	Contents any
}

type Plugin interface {
	Name() string
	IO() PluginIo
	Shutdown() error
}

type PluginIo struct {
	In  chan<- PluginMessage
	Out <-chan PluginMessage
}

// Create a protocol channel from the stdin and stdout of a process.
// Returns a channel of plugin messages along with a function that can be used to close the channel.
func protocolFromPipes(stdin io.WriteCloser, stdout io.ReadCloser) (PluginIo, func()) {
	in := make(chan PluginMessage)
	out := make(chan PluginMessage)
	cancel := make(chan struct{}, 1)
	cancelFunc := func() {
		cancel <- struct{}{}
		close(in)
		close(out)
		err := stdout.Close()
		if err != nil {
			log.Fatalf("Failed to close plugin stdout: %v", err)
		}
	}
	go func() { // read from stdout
		outBuf := bufio.NewReader(stdout)
	ReadOut:
		for {
			typeBuf := make([]byte, 1)
			_, err := io.ReadFull(outBuf, typeBuf)
			// check for cancellation here just in case output closed because it's done
			select {
			case <-cancel:
				break ReadOut
			default:
				// continue on, no cancel yet
			}
			if err != nil {
				log.Fatalf("Failed to read from plugin stdout: %v", err)
			}
			typeByte := typeBuf[0]
			msgType := MessageType(typeByte)
			if !msgType.IsValid() {
				log.Fatalf("Message from plugin has invalid type")
			}
			switch msgType {
			case READY, SHUTDOWN:
				// special case with nil contents
				out <- PluginMessage{
					Type:     msgType,
					Contents: nil,
				}
			case BIN, LOG, ERR:
				// special case with plain binary/text contents
				sizeBuf := make([]byte, 8)
				_, err := io.ReadFull(outBuf, sizeBuf)
				if err != nil {
					log.Fatalf("Failed to read message size from plugin: %v", err)
				}
				size := binary.LittleEndian.Uint64(sizeBuf)
				contentBuf := make([]byte, size)
				_, err = io.ReadFull(outBuf, contentBuf)
				if err != nil {
					log.Fatalf("Failed to read message contents from plugin: %v", err)
				}
				out <- PluginMessage{
					Type:     msgType,
					Contents: contentBuf,
				}
			default:
				// regular case with JSON contents
				sizeBuf := make([]byte, 8)
				_, err := io.ReadFull(outBuf, sizeBuf)
				if err != nil {
					log.Fatalf("Failed to read message size from plugin: %v", err)
				}
				size := binary.LittleEndian.Uint64(sizeBuf)
				contentBuf := make([]byte, size)
				_, err = io.ReadFull(outBuf, contentBuf)
				if err != nil {
					log.Fatalf("Failed to read message contents from plugin: %v", err)
				}
				var contents any
				err = json.Unmarshal(contentBuf, &contents)
				if err != nil {
					log.Fatalf("Failed to deserialize plugin message contents: %v", err)
				}
				out <- PluginMessage{
					Type:     msgType,
					Contents: contents,
				}
			}
		}
	}()
	go func() { // write to stdin
		for {
			msg, ok := <-in
			if !ok {
				err := stdin.Close()
				if err != nil {
					log.Fatalf("failed to close stdin on plugin: %v", err)
				}
				break
			}
			switch msg.Type {
			case READY, SHUTDOWN:
				// special case messages with nil contents
				_, err := stdin.Write([]byte{byte(msg.Type)})
				if err != nil {
					log.Fatalf("Failed to send plugin message: %v", err)
				}
			case BIN, LOG, ERR:
				// special case with plain binary/text contents
				packet := append(binary.LittleEndian.AppendUint64([]byte{byte(msg.Type)}, uint64(len(msg.Contents.([]byte)))), msg.Contents.([]byte)...)
				_, err := stdin.Write(packet)
				if err != nil {
					log.Fatalf("Failed to send plugin message: %v", err)
				}
			default:
				// regular case messages with JSON contents
				contents, err := json.Marshal(msg.Contents)
				if err != nil {
					log.Fatalf("Failed to serialize plugin message contents: %v", err)
				}
				// type, length, contents
				packet := append(binary.LittleEndian.AppendUint64([]byte{byte(msg.Type)}, uint64(len(contents))), contents...)
				_, err = stdin.Write(packet)
				if err != nil {
					log.Fatalf("Failed to send plugin message: %v", err)
				}
			}
		}
	}()
	return PluginIo{in, out}, cancelFunc
}

func (pp pythonPlugin) IO() PluginIo {
	return pp.io
}

func (pp pythonPlugin) Name() string {
	return pp.name
}
