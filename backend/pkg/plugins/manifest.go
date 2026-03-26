package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type PluginHook struct {
	FriendlyName *string       `json:"friendly_name,omitempty"`
	Type         HookType      `json:"type"`
	Signature    HookSignature `json:"signature"`
}

type PluginManifest struct {
	Name          string                `json:"name"`
	FriendlyName  *string               `json:"friendly_name,omitempty"`
	Version       string                `json:"version"`
	Hooks         map[string]PluginHook `json:"hooks"`
	AbsPath       string                `json:"-"`
	MainFile      string                `json:"main_file"`
	Requirements  *string               `json:"requirements"`
	PythonVersion *string               `json:"python_version"`
	Disabled      bool                  `json:"default_disabled"`
}

type PluginInfo struct {
	Name         string                `json:"name"`
	FriendlyName *string               `json:"friendly_name,omitempty"`
	Version      string                `json:"version"`
	Enabled      bool                  `json:"enabled"`
	Hooks        map[string]PluginHook `json:"hooks"`
}

type HookInfo struct {
	Name string     `json:"name"`
	Hook PluginHook `json:"hook"`
}

// This function reads a manifest
func ReadManifest(directory string) (PluginManifest, error) {
	absPath, err := filepath.Abs(directory)
	if err != nil {
		return PluginManifest{}, fmt.Errorf("failed to read manifest at %v: %w", directory, err)
	}
	fullPath := path.Join(absPath, "manifest.json")
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
	manifest.AbsPath = absPath
	return manifest, nil
}

func (pi PluginInfo) HooksWithType(ty HookType) []HookInfo {
	var info []HookInfo
	for k, v := range pi.Hooks {
		if v.Type == ty {
			info = append(info, HookInfo{Name: k, Hook: v})
		}
	}
	return info
}

func (pi PluginInfo) HooksWithSignature(si HookSignature) []HookInfo {
	var info []HookInfo
	for k, v := range pi.Hooks {
		if v.Signature == si {
			info = append(info, HookInfo{Name: k, Hook: v})
		}
	}
	return info
}

func (pi PluginInfo) FilterHooks(ty HookType, si HookSignature) []HookInfo {
	var info []HookInfo
	for k, v := range pi.Hooks {
		if v.Signature == si && v.Type == ty {
			info = append(info, HookInfo{Name: k, Hook: v})
		}
	}
	return info
}

type HookType uint8

const (
	PostUpload HookType = iota
	ManualTrigger
	GenerateForm
)

type HookSignature uint8

const (
	ProcessOneImage HookSignature = iota
	ProcessImageBatch
	ProcessFormData
)

var stringHookMap map[string]HookType = map[string]HookType{
	"post_upload":    PostUpload,
	"manual_trigger": ManualTrigger,
	"generate_form": GenerateForm,
}
var hookStringMap map[HookType]string

var stringSignatureMap map[string]HookSignature = map[string]HookSignature{
	"single_image": ProcessOneImage,
	"image_batch":  ProcessImageBatch,
	"form_data": ProcessFormData,
}
var signatureStringMap map[HookSignature]string

func init() {
	hookStringMap = make(map[HookType]string)
	for k, v := range stringHookMap {
		hookStringMap[v] = k
	}
	signatureStringMap = make(map[HookSignature]string)
	for k, v := range stringSignatureMap {
		signatureStringMap[v] = k
	}
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

func (s *HookSignature) UnmarshalJSON(data []byte) error {
	var jsonValue string
	if err := json.Unmarshal(data, &jsonValue); err != nil {
		return fmt.Errorf("hook type should be a string: %w", err)
	}
	if signature, ok := stringSignatureMap[jsonValue]; ok {
		*s = signature
		return nil
	}
	return fmt.Errorf("invalid status value: %v", jsonValue)
}

func (s HookSignature) MarshalJSON() ([]byte, error) {
	if str, ok := signatureStringMap[s]; ok {
		return json.Marshal(str)
	}
	return nil, fmt.Errorf("unknown hook type")
}
