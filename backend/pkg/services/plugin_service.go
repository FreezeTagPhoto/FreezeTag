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
	"strings"
)

type PluginService interface {
	AllPlugins() []string
	AllHooks(plugin string) map[string]plugins.PluginHook
	HooksWithType(plugin string, ty plugins.HookType) map[string]plugins.HookSignature
	HooksWithSignature(plugin string, sig plugins.HookSignature) map[string]plugins.HookType
	Hooks(plugin string, ty plugins.HookType, sig plugins.HookSignature) []string
	RunPostUpload(plugin string, ctx context.Context, uploadJob []*repositories.ImageUploadSuccess) (<-chan error, error)
}

type defaultPluginService struct {
	imgRepo repositories.ImageRepository
	baseDir string
	plugins map[string]plugins.PluginManifest
	names   []string
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
	manifests := make(map[string]plugins.PluginManifest)
	seen := make(map[string]struct{})
	var names []string
	for _, e := range entries {
		if e.Name() == "freezetag-core" {
			// core plugin, not actually a plugin
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			// dot files are ignored (this allows you to 'disable' a plugin)
			continue
		}
		if e.IsDir() {
			manifest, err := plugins.ReadManifest(path.Join(baseDir, e.Name()))
			if err != nil {
				log.Printf("[ERR]  failed to read plugin manifest at %v: %s", e.Name(), err.Error())
			}
			if _, exists := seen[manifest.Name]; exists {
				log.Printf("[ERR]  detected duplicate plugins of the same name.")
			} else {
				seen[manifest.Name] = struct{}{}
				names = append(names, manifest.Name)
				manifests[manifest.Name] = manifest
			}
		}
	}
	return defaultPluginService{repo, baseDir, manifests, names}, nil
}

func (ps defaultPluginService) AllPlugins() []string {
	return ps.names
}

func (ps defaultPluginService) AllHooks(plugin string) map[string]plugins.PluginHook {
	man, exists := ps.plugins[plugin]
	if !exists {
		return nil
	}
	return man.Hooks
}

func (ps defaultPluginService) HooksWithType(plugin string, ty plugins.HookType) map[string]plugins.HookSignature {
	man, exists := ps.plugins[plugin]
	if !exists {
		return nil
	}
	oftype := make(map[string]plugins.HookSignature)
	for name, hook := range man.Hooks {
		if hook.Type == ty {
			oftype[name] = hook.Signature
		}
	}
	return oftype
}

func (ps defaultPluginService) HooksWithSignature(plugin string, sig plugins.HookSignature) map[string]plugins.HookType {
	man, exists := ps.plugins[plugin]
	if !exists {
		return nil
	}
	ofsig := make(map[string]plugins.HookType)
	for name, hook := range man.Hooks {
		if hook.Signature == sig {
			ofsig[name] = hook.Type
		}
	}
	return ofsig
}

func (ps defaultPluginService) Hooks(plugin string, ty plugins.HookType, sig plugins.HookSignature) []string {
	man, exists := ps.plugins[plugin]
	if !exists {
		return nil
	}
	var hooks []string
	for name, hook := range man.Hooks {
		if hook.Type == ty && hook.Signature == sig {
			hooks = append(hooks, name)
		}
	}
	return hooks
}

func (ps defaultPluginService) RunPostUpload(plugin string, ctx context.Context, uploadJob []*repositories.ImageUploadSuccess) (<-chan error, error) {
	// only process-image handlers for now, more eventually
	manifest, exists := ps.plugins[plugin]
	if !exists {
		return nil, fmt.Errorf("plugin %v doesn't exist", plugin)
	}
	process, err := plugins.PluginFromManifest(manifest, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}
	results := make(chan error)
	go func() {
		// hook functions of the same type will run sequentially for each job.
		hooks := ps.Hooks(plugin, plugins.PostUpload, plugins.ImageProcess)
	jobLoop:
		for _, job := range uploadJob {
			for _, hook := range hooks {
				err := plugins.ProcessImage(process, hook, job.Id, ps.imgRepo)
				if err != nil {
					results <- fmt.Errorf("hook %s failed to process image %v: %w", hook, job.Filename, err)
					if strings.Contains(err.Error(), "FATAL") {
						// unrecoverable error, don't bother continuing
						break jobLoop
					}
				}
			}
		}
		close(results)
		err := process.Shutdown()
		if err != nil {
			log.Printf("[WARN] failed to shut down plugin %v gracefully: %s", plugin, err.Error())
		}
	}()
	return results, nil
}
