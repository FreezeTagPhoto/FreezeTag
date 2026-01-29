package plugins

import (
	"fmt"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"
	"log"
)

func ProcessImage(plugin Plugin, hook string, id database.ImageId, repo repositories.ImageRepository) error {
	plugin.IO().In <- PluginMessage{GET, map[string]any{"action": "process-image", "id": id, "hook": hook}}
	webp, err := repo.RetrieveThumbnail(id, 2)
	if err != nil {
		return err
	}
	plugin.IO().In <- PluginMessage{BIN, webp}
	for {
		msg, ok := <-plugin.IO().Out
		if !ok {
			plugin.Shutdown() //nolint:errcheck
			return fmt.Errorf("FATAL: plugin stdout closed during processing")
		}
		switch msg.Type {
		case ERR:
			return fmt.Errorf("%s", string(msg.Contents.([]byte)))
		case LOG:
			log.Printf("[PLUG] %s: %s", plugin.Name(), string(msg.Contents.([]byte)))
		case PUT:
			return handleProcessResponse(plugin, repo, msg.Contents)
		case GET:
			err := handleProcessGet(plugin, repo, msg.Contents)
			if err != nil {
				plugin.IO().In <- PluginMessage{ERR, []byte(err.Error())}
			}
		default:
			plugin.IO().In <- PluginMessage{ERR, []byte("invalid request")}
		}
	}
}

func definitelyStrings(s any) []string {
	sl := s.([]any)
	strings := make([]string, len(sl))
	for i, v := range sl {
		strings[i] = v.(string)
	}
	return strings
}

func definitelyId(i any) database.ImageId {
	return database.ImageId(int64(i.(float64)))
}

func handleProcessResponse(plugin Plugin, repo repositories.ImageRepository, msg any) error {
	if m, ok := msg.(map[string]any); ok {
		switch m["action"] {
		case "skip":
			// skip this image, nothing needs done
		case "tag":
			id := definitelyId(m["id"])
			tags := definitelyStrings(m["tags"])
			res := repo.AddImageTags(database.ImageId(id), tags)
			if res.Err != nil {
				return fmt.Errorf("failed to tag image %d: %s", res.Err.Id, res.Err.Reason)
			}
		default:
			return fmt.Errorf("unsupported process response from plugin: %s", m["action"])
		}
	} else {
		return fmt.Errorf("message was not JSON")
	}
	return nil
}

func handleProcessGet(plugin Plugin, repo repositories.ImageRepository, msg any) error {
	if m, ok := msg.(map[string]any); ok {
		switch m["action"] {
		case "metadata":
			id := definitelyId(m["id"])
			meta, err := repo.GetImageMetadata(id)
			if err != nil {
				return err
			}
			plugin.IO().In <- PluginMessage{PUT, map[string]any{
				"action": "metadata",
				"data":   meta,
			}}
		default:
			return fmt.Errorf("unsupported process request")
		}
	}
	return nil
}
