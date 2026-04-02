package services

import (
	"context"
	"fmt"
	"freezetag/backend/pkg/plugins"
	"freezetag/backend/pkg/repositories"
	"log"
	"os"
	"path"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

type PluginService interface {
	Plugins() []plugins.PluginInfo
	PluginInfo(plugin string) *plugins.PluginInfo
	SetEnabled(plugin string, enabled bool)
	ReadConfiguration(plugin string) (map[string]plugins.PublicConfigField, error)
	ChangeConfiguration(plugin string, changes map[string]any) error
	LaunchPlugin(plugin string, ctx context.Context) (*plugins.HookedPlugin, error)
}

type defaultPluginService struct {
	imgRepo repositories.ImageRepository
	baseDir string
	plugins map[string]*plugins.PluginManifest
}

func InitDefaultPluginService(dir string, repo repositories.ImageRepository) (defaultPluginService, error) {
	baseDir, err := filepath.Abs(dir)
	if err != nil {
		return defaultPluginService{}, fmt.Errorf("failed to read plugin directory: %w", err)
	}
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return defaultPluginService{}, fmt.Errorf("failed to read plugin directory: %w", err)
	}
	manifests := make(map[string]*plugins.PluginManifest)
	seen := make(map[string]struct{})
	for _, e := range entries {
		if e.Name() == "freezetag-core" {
			// core plugin, not actually a plugin
			continue
		}
		if e.IsDir() {
			manifest, err := plugins.ReadManifest(path.Join(baseDir, e.Name()))
			if err != nil {
				log.Printf("[ERR]  failed to read plugin manifest at %v: %s", e.Name(), err.Error())
				continue
			}
			if _, exists := seen[manifest.Name]; exists {
				log.Printf("[ERR]  detected duplicate plugins of the same name.")
			} else {
				seen[manifest.Name] = struct{}{}
				manifests[manifest.Name] = &manifest
			}
		}
	}
	service := defaultPluginService{repo, baseDir, manifests}
	// make sure all configs are initialized
	for _, plugin := range service.Plugins() {
		_, _ = service.ReadConfiguration(plugin.Name)
	}
	return defaultPluginService{repo, baseDir, manifests}, nil
}

func (ps defaultPluginService) Plugins() []plugins.PluginInfo {
	info := make([]plugins.PluginInfo, 0, len(ps.plugins))
	for k, v := range ps.plugins {
		info = append(info, plugins.PluginInfo{
			Name:         k,
			FriendlyName: v.FriendlyName,
			Version:      v.Version,
			Enabled:      !v.Disabled,
			Hooks:        v.Hooks,
			Configurable: v.Config != nil,
		})
	}
	return info
}

func (ps defaultPluginService) PluginInfo(plugin string) *plugins.PluginInfo {
	man, exists := ps.plugins[plugin]
	if !exists {
		return nil
	}
	return &plugins.PluginInfo{
		Name:         plugin,
		FriendlyName: man.FriendlyName,
		Version:      man.Version,
		Enabled:      !man.Disabled,
		Hooks:        man.Hooks,
	}
}

func (ps defaultPluginService) SetEnabled(plugin string, enabled bool) {
	man, exists := ps.plugins[plugin]
	if !exists {
		return
	}
	man.Disabled = !enabled
}

func (ps defaultPluginService) getConfigFile(plugin string) (string, error) {
	manifest, exists := ps.plugins[plugin]
	if !exists {
		return "", fmt.Errorf("plugin %v doesn't exist", plugin)
	}
	if manifest.Config == nil {
		return "", nil
	}
	return path.Join(manifest.AbsPath, manifest.Config.File), nil
}

func readConfig(file string, layout []plugins.PluginConfigField) (map[string]any, error) {
	configContents, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		defaultConfig := make(map[string]any)
		for _, field := range layout {
			if field.DefaultValue != nil {
				defaultConfig[field.Name] = field.DefaultValue
			} else {
				defaultConfig[field.Name] = ""
			}
		}
		err = writeConfig(file, defaultConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to write default config: %w", err)
		}
		configContents, err = os.ReadFile(file)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin config: %w", err)
	}
	var config map[string]any
	err = toml.Unmarshal(configContents, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize plugin config: %w", err)
	}
	return config, nil
}

func writeConfig(file string, config map[string]any) error {
	configContents, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize plugin config: %w", err)
	}
	err = os.WriteFile(file, configContents, 0666)
	if err != nil {
		return fmt.Errorf("failed to write plugin config: %w", err)
	}
	return nil
}

func (ps defaultPluginService) ChangeConfiguration(plugin string, changes map[string]any) error {
	configFile, err := ps.getConfigFile(plugin)
	if err != nil {
		return err
	}
	if configFile == "" {
		return fmt.Errorf("plugin isn't configurable")
	}
	layout := ps.plugins[plugin].Config.Fields
	config, err := readConfig(configFile, layout)
	if err != nil {
		return err
	}
	for field, value := range changes {
		_, exists := config[field]
		if !exists {
			return fmt.Errorf("plugin %v doesn't have config field %v", plugin, field)
		}
		config[field] = value
	}
	err = writeConfig(configFile, config)
	if err != nil {
		return err
	}
	return nil
}

func (ps defaultPluginService) ReadConfiguration(plugin string) (map[string]plugins.PublicConfigField, error) {
	configFile, err := ps.getConfigFile(plugin)
	if err != nil {
		return nil, err
	}
	if configFile == "" {
		return nil, nil
	}
	layout := ps.plugins[plugin].Config.Fields
	config, err := readConfig(configFile, layout)
	if err != nil {
		return nil, err
	}
	publicConfig := make(map[string]plugins.PublicConfigField)
	for _, field := range layout {
		var publicValue any
		if field.Protected {
			publicValue = nil
		} else {
			publicValue = config[field.Name]
		}
		publicConfig[field.Name] = plugins.PublicConfigField{
			Value:        publicValue,
			DefaultValue: field.DefaultValue,
			Protected:    field.Protected,
			Name:         field.FriendlyName,
			Description:  field.Description,
		}
	}
	return publicConfig, nil
}

func (ps defaultPluginService) LaunchPlugin(plugin string, ctx context.Context) (*plugins.HookedPlugin, error) {
	manifest, exists := ps.plugins[plugin]
	if !exists {
		return nil, fmt.Errorf("plugin %v doesn't exist", plugin)
	}
	process, err := plugins.PluginFromManifest(*manifest, ctx, ps.imgRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}
	return &process, nil
}
