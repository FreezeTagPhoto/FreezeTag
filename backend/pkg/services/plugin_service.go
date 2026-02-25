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
)

type PluginService interface {
	Plugins() []plugins.PluginInfo
	PluginInfo(plugin string) *plugins.PluginInfo
	SetEnabled(plugin string, enabled bool)
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

func (ps defaultPluginService) LaunchPlugin(plugin string, ctx context.Context) (*plugins.HookedPlugin, error) {
	manifest, exists := ps.plugins[plugin]
	if !exists {
		return nil, fmt.Errorf("plugin %v doesn't exist", plugin)
	}
	process, err := plugins.PluginFromManifest(*manifest, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}
	return &process, nil
}
