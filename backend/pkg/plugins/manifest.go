package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type PluginHook struct {
	Type HookType `json:"type"`
}

type PluginManifest struct {
	Name         string                `json:"name"`
	Version      string                `json:"version"`
	Hooks        map[string]PluginHook `json:"hooks"`
	MainFile     string                `json:"main_file"`
	Requirements *string               `json:"requirements"`
}

func ReadManifest(directory string) (PluginManifest, error) {
	fullPath := path.Join(directory, "manifest.json")
	content, err := os.ReadFile(path.Join(directory, "manifest.json"))
	if err != nil {
		return PluginManifest{}, fmt.Errorf("failed to read manifest at %v: %w", fullPath, err)
	}
	manifest := PluginManifest{
		Version: "0",
	}
	err = json.Unmarshal(content, &manifest)
	if err != nil {
		return PluginManifest{}, fmt.Errorf("failed to parse manifest at %v: %w", fullPath, err)
	}
	return manifest, nil
}

type HookType uint8

const (
	PostUpload HookType = iota
)

var stringHookMap map[string]HookType = map[string]HookType{
	"post_upload": PostUpload,
}

var hookStringMap map[HookType]string = map[HookType]string{
	PostUpload: "post_upload",
}

func (h *HookType) UnmarshalJSON(data []byte) error {
	var jsonValue string
	if err := json.Unmarshal(data, &jsonValue); err != nil {
		return fmt.Errorf("hook type should be a string: %w", err)
	}
	if hook, ok := stringHookMap[jsonValue]; ok {
		*h = hook
		return nil
	}
	return fmt.Errorf("invalid status value: %v", jsonValue)
}

func (h HookType) MarshalJSON() ([]byte, error) {
	if str, ok := hookStringMap[h]; ok {
		return json.Marshal(str)
	}
	return nil, fmt.Errorf("unknown hook type")
}
