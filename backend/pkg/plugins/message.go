package plugins

import (
	"fmt"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/repositories"
	"log"
)

type imageAction struct {
	action string
	id     database.ImageId
}

func ProcessImage(plugin Plugin, id database.ImageId, repo repositories.ImageRepository) error {
	plugin.IO() <- PluginMessage{GET, imageAction{"process", id}}
	webp, err := repo.RetrieveThumbnail(id, 2)
	if err != nil {
		return err
	}
	plugin.IO() <- PluginMessage{BIN, webp}
	for {
		msg := <-plugin.IO()
		switch msg.Type {
		case ERR:
			return fmt.Errorf("%s", string(msg.Contents.([]byte)))
		case LOG:
			log.Printf("%s: %s", plugin.Name(), string(msg.Contents.([]byte)))
		case PUT:
			return handleProcessResponse(plugin, repo, msg.Contents)
		case GET:
			err := handleProcessGet(plugin, repo, msg.Contents)
			if err != nil {
				plugin.IO() <- PluginMessage{ERR, []byte(err.Error())}
			}
		default:
			plugin.IO() <- PluginMessage{ERR, []byte("invalid request")}
		}
	}
}

func handleProcessResponse(plugin Plugin, repo repositories.ImageRepository, msg any) error {
	if m, ok := msg.(map[string]any); ok {
		switch m["action"] {
		case "tag":
			res := repo.AddImageTags(database.ImageId(m["id"].(int)), m["tags"].([]string))
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
			meta, err := repo.GetImageMetadata(database.ImageId(m["id"].(int)))
			if err != nil {
				return err
			}
			plugin.IO() <- PluginMessage{PUT, map[string]any{
				"action": "metadata",
				"data":   meta,
			}}
		default:
			return fmt.Errorf("unsupported process request")
		}
	}
	return nil
}
