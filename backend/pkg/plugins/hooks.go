package plugins

import (
	"encoding/base64"
	"fmt"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/queries"
	"freezetag/backend/pkg/repositories"
	"log"
	"reflect"
)

func (h *HookedPlugin) RunHook(t HookType, s HookSignature, in any, repo repositories.ImageRepository) (any, error) {
	switch t {
	case PostUpload:
		switch s {
		case ProcessOneImage:
			resolved, ok := in.(repositories.ImageUploadSuccess)
			if !ok {
				return nil, fmt.Errorf("invalid input type for PostUpload,ProcessOneImage")
			}
			return h.processOneImage(resolved, repo)
		case ProcessImageBatch:
			break
		}
	}
	return nil, fmt.Errorf("unknown hook type: %v,%v", t, s)
}

func (h *HookedPlugin) processOneImage(in repositories.ImageUploadSuccess, repo repositories.ImageRepository) (any, error) {
	webp, err := repo.RetrieveThumbnail(in.Id, 2)
	if err != nil {
		return nil, fmt.Errorf("error retrieving image for plugin: %w", err)
	}
	h.IO().In <- PluginMessage{
		GET,
		map[string]any{"action": "single_image", "id": in.Id},
	}
	h.IO().In <- PluginMessage{BIN, webp}
	resp, err := h.handleRequests(repo)
	if err != nil {
		return nil, err
	}
	res, err := h.handlePost(repo, resp)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// loop and handle plugin GETs until the next plugin PUT, then return it
// errors marked with FATAL should stop further processing (the plugin is in an unworkable state)
func (h *HookedPlugin) handleRequests(repo repositories.ImageRepository) (PluginMessage, error) {
	for {
		msg, ok := <-h.IO().Out
		if !ok {
			h.Shutdown() //nolint:errcheck
			return PluginMessage{}, fmt.Errorf("FATAL: plugin stdout closed while processing")
		}
		switch msg.Type {
		case ERR:
			return PluginMessage{}, fmt.Errorf("%s", string(msg.Contents.([]byte)))
		case LOG:
			log.Printf("[PLUG] %s: %s", h.Name(), string(msg.Contents.([]byte)))
		case PUT:
			return msg, nil
		case GET:
			err := h.handleGet(repo, msg)
			if err != nil {
				h.IO().In <- PluginMessage{ERR, []byte(err.Error())}
			}
		default:
			h.IO().In <- PluginMessage{ERR, []byte("invalid message type")}
		}
	}
}

func (h *HookedPlugin) handlePost(repo repositories.ImageRepository, m PluginMessage) (any, error) {
	msg, ok := m.Contents.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("failed to deserialize request contents")
	}
	switch msg["action"] {
	case "add_tags":
		id, err := intoId(msg["id"])
		if err != nil {
			return nil, err
		}
		tags, err := intoList[string](msg["tags"])
		if err != nil {
			return nil, err
		}
		res := repo.AddImageTags(id, tags)
		return res, nil
	case "remove_tags":
		id, err := intoId(msg["id"])
		if err != nil {
			return nil, err
		}
		tags, err := intoList[string](msg["tags"])
		if err != nil {
			return nil, err
		}
		res := repo.RemoveImageTags(id, tags)
		return res, nil
	case "delete_tags":
		tags, err := intoList[string](msg["tags"])
		if err != nil {
			return nil, err
		}
		count, err := repo.DeleteTags(tags)
		if err != nil {
			return nil, err
		}
		return map[string]any{"deleted": count}, nil
	case "delete_image":
		id, err := intoId(msg["id"])
		if err != nil {
			return nil, err
		}
		name, err := repo.DeleteImage(id)
		if err != nil {
			return nil, err
		}
		return map[string]any{"deleted": name}, nil
	case "add_image":
		name, err := into[string](msg["name"])
		if err != nil {
			return nil, err
		}
		data64, err := into[string](msg["data"])
		if err != nil {
			return nil, err
		}
		data, err := base64.StdEncoding.DecodeString(data64)
		if err != nil {
			return nil, err
		}
		id, err := repo.StoreImageBytes(data, name)
		if err != nil {
			return nil, err
		}
		return map[string]any{"name": name, "id": id}, nil
	case "multipart":
		parts, err := intoList[any](msg["parts"])
		if err != nil {
			return nil, err
		}
		results := make([]any, 0, len(parts))
		for _, part := range parts {

		}
	}
	return nil, fmt.Errorf("unrecognized message action")
}

func (h *HookedPlugin) handleGet(repo repositories.ImageRepository, m PluginMessage) error {
	msg, ok := m.Contents.(map[string]any)
	if !ok {
		return fmt.Errorf("failed to deserialize request contents")
	}
	switch msg["action"] {
	case "metadata":
		id, err := intoId(msg["id"])
		if err != nil {
			return err
		}
		meta, err := repo.GetImageMetadata(id)
		if err != nil {
			return err
		}
		h.IO().In <- PluginMessage{
			PUT,
			map[string]any{"action": "metadata", "data": meta},
		}
	case "search":
		res, err := pluginSearchRequest(msg, repo)
		if err != nil {
			return err
		}
		h.IO().In <- PluginMessage{
			PUT,
			map[string]any{"action": "search", "results": res},
		}
	case "image":
		id, err := intoId(msg["id"])
		if err != nil {
			return err
		}
		webp, err := repo.RetrieveThumbnail(id, 2)
		if err != nil {
			return err
		}
		h.IO().In <- PluginMessage{
			PUT,
			map[string]any{"action": "image"},
		}
		h.IO().In <- PluginMessage{BIN, webp}
	case "tags":
		id, err := intoId(msg["id"])
		if err != nil {
			return err
		}
		tags, err := repo.RetrieveImageTags(id)
		if err != nil {
			return err
		}
		h.IO().In <- PluginMessage{
			PUT,
			map[string]any{"action": "tags", "tags": tags},
		}
	case "all_tags":
		tags, err := repo.RetrieveAllTags()
		if err != nil {
			return err
		}
		h.IO().In <- PluginMessage{
			PUT,
			map[string]any{"action": "all_tags", "tags": tags},
		}
	case "tags_query":
		res, err := pluginTagSearchRequest(msg, repo)
		if err != nil {
			return err
		}
		h.IO().In <- PluginMessage{
			PUT,
			map[string]any{"action": "tags_query", "tags": res},
		}
	default:
		return fmt.Errorf("unknown request type")
	}
	return nil
}

func pluginSearchRequest(msg map[string]any, repo repositories.ImageRepository) ([]database.ImageId, error) {
	qParams, err := intoMap[string, any](msg["query"])
	if err != nil {
		return nil, err
	}
	q := queries.ImageQueryParams{
		Make:           intoOrEmpty[string](qParams["make"]),
		Model:          intoOrEmpty[string](qParams["model"]),
		MakeLike:       intoOrEmpty[string](qParams["makeLike"]),
		ModelLike:      intoOrEmpty[string](qParams["modelLike"]),
		TakenBefore:    intoOrEmpty[string](qParams["takenBefore"]),
		TakenAfter:     intoOrEmpty[string](qParams["takenAfter"]),
		UploadedBefore: intoOrEmpty[string](qParams["uploadedBefore"]),
		UploadedAfter:  intoOrEmpty[string](qParams["uploadedAfter"]),
		Near:           intoOrEmpty[string](qParams["near"]),
		Tags:           intoListOrEmpty[string](qParams["tags"]),
		TagsLike:       intoListOrEmpty[string](qParams["tagsLike"]),
	}
	query, err := queries.QueryFromStruct(q)
	if err != nil {
		return nil, err
	}
	return repo.SearchImage(query)
}

func pluginTagSearchRequest(msg map[string]any, repo repositories.ImageRepository) (map[string]int64, error) {
	qParams, err := intoMap[string, any](msg["query"])
	if err != nil {
		return nil, err
	}
	q := queries.ImageQueryParams{
		Make:           intoOrEmpty[string](qParams["make"]),
		Model:          intoOrEmpty[string](qParams["model"]),
		MakeLike:       intoOrEmpty[string](qParams["makeLike"]),
		ModelLike:      intoOrEmpty[string](qParams["modelLike"]),
		TakenBefore:    intoOrEmpty[string](qParams["takenBefore"]),
		TakenAfter:     intoOrEmpty[string](qParams["takenAfter"]),
		UploadedBefore: intoOrEmpty[string](qParams["uploadedBefore"]),
		UploadedAfter:  intoOrEmpty[string](qParams["uploadedAfter"]),
		Near:           intoOrEmpty[string](qParams["near"]),
		Tags:           intoListOrEmpty[string](qParams["tags"]),
		TagsLike:       intoListOrEmpty[string](qParams["tagsLike"]),
	}
	query, err := queries.QueryFromStruct(q)
	if err != nil {
		return nil, err
	}
	return repo.GetQueryTagCounts(query)
}

// special case because JSON deserialized numbers are dumb
func intoId(a any) (database.ImageId, error) {
	i, ok := a.(float64)
	if !ok {
		return 0, fmt.Errorf("failed to extract ID")
	}
	return database.ImageId(int64(i)), nil
}

func into[V any](a any) (V, error) {
	t, ok := a.(V)
	if !ok {
		return t, fmt.Errorf("failed to extract %s", reflect.TypeFor[V]().Name())
	}
	return t, nil
}

func intoOrEmpty[V any](a any) V {
	t, err := into[V](a)
	if err != nil {
		var empty V
		return empty
	}
	return t
}

func intoList[V any](v any) ([]V, error) {
	a, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("failed to extract array")
	}
	list := make([]V, len(a))
	for i, v := range a {
		t, err := into[V](v)
		if err != nil {
			return nil, err
		}
		list[i] = t
	}
	return list, nil
}

func intoListOrEmpty[V any](v any) []V {
	l, err := intoList[V](v)
	if err != nil {
		return []V{}
	}
	return l
}

func intoMap[K comparable, V any](v any) (map[K]V, error) {
	a, ok := v.(map[any]any)
	if !ok {
		return nil, fmt.Errorf("failed to extract map")
	}
	m := make(map[K]V, len(a))
	for k, v := range a {
		kt, err := into[K](k)
		if err != nil {
			return nil, err
		}
		vt, err := into[V](v)
		if err != nil {
			return nil, err
		}
		m[kt] = vt
	}
	return m, nil
}
