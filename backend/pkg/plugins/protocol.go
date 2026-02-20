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
				close(out)
				break ReadOut
			default:
				// continue on, no cancel yet
			}
			if err == io.EOF {
				log.Printf("[WARN] Plugin stdout may have closed at an unexpected time.")
				log.Printf("[WARN] Did you remember to call 'freezetag.run()' in the plugin's main file?")
				close(out)
				break ReadOut
			} else if err != nil {
				log.Printf("[ERR]  Failed to read from plugin stdout: %v", err)
				close(out)
				break ReadOut
			}
			typeByte := typeBuf[0]
			msgType := MessageType(typeByte)
			if !msgType.IsValid() {
				log.Printf("[ERR]  Message from plugin has invalid type")
				continue
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
					log.Printf("[ERR]  Failed to read message size from plugin: %v", err)
					continue
				}
				size := binary.LittleEndian.Uint64(sizeBuf)
				contentBuf := make([]byte, size)
				_, err = io.ReadFull(outBuf, contentBuf)
				if err != nil {
					log.Printf("[ERR]  Failed to read message contents from plugin: %v", err)
					continue
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
					log.Printf("[ERR]  Failed to read message size from plugin: %v", err)
					continue
				}
				size := binary.LittleEndian.Uint64(sizeBuf)
				contentBuf := make([]byte, size)
				_, err = io.ReadFull(outBuf, contentBuf)
				if err != nil {
					log.Printf("[ERR]  Failed to read message contents from plugin: %v", err)
					continue
				}
				var contents any
				err = json.Unmarshal(contentBuf, &contents)
				if err != nil {
					log.Printf("[ERR]  Failed to deserialize plugin message contents: %v", err)
					continue
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
					log.Printf("[ERR]  failed to close stdin on plugin: %v", err)
				}
				break
			}
			switch msg.Type {
			case READY, SHUTDOWN:
				// special case messages with nil contents
				_, err := stdin.Write([]byte{byte(msg.Type)})
				if err != nil {
					log.Printf("[ERR]  Failed to send plugin message: %v", err)
					continue
				}
			case BIN, LOG, ERR:
				// special case with plain binary/text contents
				packet := append(binary.LittleEndian.AppendUint64([]byte{byte(msg.Type)}, uint64(len(msg.Contents.([]byte)))), msg.Contents.([]byte)...)
				_, err := stdin.Write(packet)
				if err != nil {
					log.Printf("[ERR]  Failed to send plugin message: %v", err)
					continue
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
					log.Printf("[ERR]  Failed to send plugin message: %v", err)
					continue
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
