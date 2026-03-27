package plugins

import (
	"encoding/base64"
	"fmt"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/queries"
	"freezetag/backend/pkg/images/imagedata"
	"freezetag/backend/pkg/repositories"
	"log"
	"reflect"
	"strings"
)

type PluginResult map[string]any

func (h *HookedPlugin) RunHook(hookName string, in any, repo repositories.ImageRepository) (PluginResult, error) {
	details := h.HookDetails(hookName)
	s := details.Signature
	t := details.Type
	switch t {
	case PostUpload:
		switch s {
		case ProcessOneImage:
			resolved, ok := in.(repositories.ImageUploadSuccess)
			if !ok {
				return nil, fmt.Errorf("invalid input type for post_upload,single_image")
			}
			return h.processOneImage(hookName, resolved.Id, repo)
		case ProcessImageBatch:
			resolved, ok := in.([]repositories.ImageUploadSuccess)
			if !ok {
				return nil, fmt.Errorf("invalid input type for post_upload,image_batch")
			}
			ids := make([]database.ImageId, len(resolved))
			for i, res := range resolved {
				ids[i] = res.Id
			}
			return h.processImageBatch(hookName, ids, repo)
		case ProcessFormData:
			return nil, fmt.Errorf("form_data not legal for post_upload")
		}
	case ManualTrigger, GenerateForm:
		switch s {
		case ProcessOneImage:
			resolved, ok := in.(database.ImageId)
			if !ok {
				return nil, fmt.Errorf("invalid input type for manual_trigger,single_image")
			}
			return h.processOneImage(hookName, resolved, repo)
		case ProcessImageBatch:
			resolved, ok := in.([]database.ImageId)
			if !ok {
				return nil, fmt.Errorf("invalid input type for manual_trigger,image_batch")
			}
			return h.processImageBatch(hookName, resolved, repo)
		case ProcessFormData:
			resolved, ok := in.(string)
			if !ok {
				return nil, fmt.Errorf("invalid input type for manual_trigger,form_data")
			}
			return h.processFormData(hookName, resolved, repo);
		}
	}
	return nil, fmt.Errorf("unknown hook type: %v,%v", t, s)
}

func (h *HookedPlugin) processOneImage(hookName string, id database.ImageId, repo repositories.ImageRepository) (PluginResult, error) {
	webp, err := repo.RetrieveThumbnail(id, 2)
	if webp == nil {
		return map[string]any{"skipped": true}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error retrieving image for plugin: %w", err)
	}
	h.IO().In <- PluginMessage{
		GET,
		map[string]any{"action": "single_image", "id": id, "hook": hookName},
	}
	h.IO().In <- PluginMessage{BIN, webp}
	resp, err := h.handleRequests(repo)
	if err != nil {
		return nil, err
	}
	res, err := h.handlePost(repo, resp.Contents)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *HookedPlugin) processImageBatch(hookName string, ids []database.ImageId, repo repositories.ImageRepository) (PluginResult, error) {
	h.IO().In <- PluginMessage{
		GET,
		map[string]any{"action": "image_batch", "ids": ids, "hook": hookName},
	}
	resp, err := h.handleRequests(repo)
	if err != nil {
		return nil, err
	}
	res, err := h.handlePost(repo, resp.Contents)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *HookedPlugin) processFormData(hookName string, data string, repo repositories.ImageRepository) (PluginResult, error) {
	h.IO().In <- PluginMessage{
		GET,
		map[string]any{"action": "form_data", "json": data, "hook": hookName},
	}
	resp, err := h.handleRequests(repo)
	if err != nil {
		return nil, err
	}
	res, err := h.handlePost(repo, resp.Contents)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// loop and handle plugin GETs until the next plugin PUT, then return it
// errors marked with FATAL should stop further processing (the plugin is in an unworkable state)
func (h *HookedPlugin) handleRequests(repo repositories.ImageRepository) (PluginMessage, error) {
	res, err := handlePluginRequests(h, repo, PUT)
	if err != nil {
		if strings.Contains(err.Error(), "FATAL") {
			h.Shutdown() //nolint:errcheck
		}
		return PluginMessage{}, err
	}
	return res, nil
}

func handlePluginRequests(pl Plugin, repo repositories.ImageRepository, ty MessageType) (PluginMessage, error) {
	for {
		msg, ok := <-pl.IO().Out
		if !ok {
			return PluginMessage{}, fmt.Errorf("FATAL: plugin stdout closed while processing")
		}
		switch msg.Type {
		case ty:
			return msg, nil
		case ERR:
			return PluginMessage{}, fmt.Errorf("%s", string(msg.Contents.([]byte)))
		case LOG:
			log.Printf("[PLUG] %s: %s", pl.Name(), string(msg.Contents.([]byte)))
		case GET:
			err := handlePluginGet(pl, repo, msg)
			if err != nil {
				pl.IO().In <- PluginMessage{ERR, []byte(err.Error())}
			}
		default:
			pl.IO().In <- PluginMessage{ERR, []byte("unexpected message type")}
			return PluginMessage{}, fmt.Errorf("unexpected message type from plugin")
		}
	}
}

func (h *HookedPlugin) handlePost(repo repositories.ImageRepository, m any) (PluginResult, error) {
	msg, ok := m.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("failed to deserialize request contents")
	}
	switch msg["action"] {
	case "none":
		return map[string]any{"skipped": true}, nil
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
		if res.Err != nil {
			return nil, fmt.Errorf("%s", res.Err.Reason)
		}
		return map[string]any{"count": res.Success.Count}, nil
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
		if res.Err != nil {
			return nil, fmt.Errorf("%s", res.Err.Reason)
		}
		return map[string]any{"count": res.Success.Count}, nil
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
	case "send_form":
		form, err := into[string](msg["form"])
		if err != nil {
			return nil, err
		}
		if !(strings.HasPrefix(form, "<form>") && strings.HasSuffix(form, "</form>")){
			return nil, fmt.Errorf("form has bad syntax, expected to start with <form> and end with </form>, got %s", form)
		}
		return map[string]any{"form": form}, nil
	case "multipart":
		parts, err := intoList[any](msg["parts"])
		if err != nil {
			return nil, err
		}
		results := make([]any, 0, len(parts))
		for _, part := range parts {
			res, err := h.handlePost(repo, part)
			if err != nil {
				results = append(results, map[string]any{"error": err.Error()})
			} else {
				results = append(results, res)
			}
		}
		return map[string]any{"results": results}, nil
	}
	return nil, fmt.Errorf("unrecognized message action")
}

type pluginMetadataResponse struct {
	imagedata.Metadata
	Width  int `json:"width"`
	Height int `json:"height"`
}

func handlePluginGet(pl Plugin, repo repositories.ImageRepository, m PluginMessage) error {
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
		width, height, err := repo.GetImageResolution(id)
		if err != nil {
			return err
		}
		pl.IO().In <- PluginMessage{
			PUT,
			map[string]any{"action": "metadata", "data": pluginMetadataResponse{
				Metadata: meta,
				Width:    width,
				Height:   height,
			}},
		}
	case "search":
		res, err := pluginSearchRequest(msg, repo)
		if err != nil {
			return err
		}
		pl.IO().In <- PluginMessage{
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
		if webp == nil {
			return fmt.Errorf("no image")
		}
		pl.IO().In <- PluginMessage{
			PUT,
			map[string]any{"action": "image"},
		}
		pl.IO().In <- PluginMessage{BIN, webp}
	case "image_file":
		id, err := intoId(msg["id"])
		if err != nil {
			return err
		}
		file, err := repo.RetrieveImageFile(id)
		if err != nil {
			return err
		}
		pl.IO().In <- PluginMessage{
			PUT,
			map[string]any{"action": "image_file"},
		}
		pl.IO().In <- PluginMessage{BIN, file}
	case "tags":
		id, err := intoId(msg["id"])
		if err != nil {
			return err
		}
		tags, err := repo.RetrieveImageTags(id)
		if err != nil {
			return err
		}
		pl.IO().In <- PluginMessage{
			PUT,
			map[string]any{"action": "tags", "tags": tags},
		}
	case "all_tags":
		tags, err := repo.RetrieveAllTags()
		if err != nil {
			return err
		}
		pl.IO().In <- PluginMessage{
			PUT,
			map[string]any{"action": "all_tags", "tags": tags},
		}
	case "tags_query":
		res, err := pluginTagSearchRequest(msg, repo)
		if err != nil {
			return err
		}
		pl.IO().In <- PluginMessage{
			PUT,
			map[string]any{"action": "tags_query", "tags": res},
		}
	default:
		return fmt.Errorf("unknown request type")
	}
	return nil
}

func pluginSearchRequest(msg map[string]any, repo repositories.ImageRepository) ([]database.ImageId, error) {
	qParams, ok := msg["query"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't deserialize query parameters")
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
	qParams, ok := msg["query"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't deserialize query parameters")
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
