package openapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Bundler struct {
	baseDir    string
	components map[string]map[string]any
}

func NewBundler(specPath string) *Bundler {
	return &Bundler{
		baseDir:    filepath.Dir(specPath),
		components: make(map[string]map[string]any),
	}
}

func (b *Bundler) Bundle(specPath string) ([]byte, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var spec map[string]any
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := b.resolveExternalRefs(spec, b.baseDir, ""); err != nil {
		return nil, fmt.Errorf("failed to resolve references: %w", err)
	}

	if len(b.components) > 0 {
		if _, exists := spec["components"]; !exists {
			spec["components"] = make(map[string]any)
		}
		components := spec["components"].(map[string]any)

		for refPath, content := range b.components {
			parts := strings.Split(strings.TrimPrefix(refPath, "#/components/"), "/")
			if len(parts) >= 2 {
				componentType := parts[0]
				componentName := parts[1]

				if _, exists := components[componentType]; !exists {
					components[componentType] = make(map[string]any)
				}
				componentTypeMap := components[componentType].(map[string]any)
				componentTypeMap[componentName] = content
			}
		}
	}

	bundled, err := yaml.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return bundled, nil
}

func (b *Bundler) resolveExternalRefs(node any, currentDir string, parentKey string) error {
	switch v := node.(type) {
	case map[string]any:
		if ref, ok := v["$ref"].(string); ok {
			if strings.HasPrefix(ref, "./") || strings.HasPrefix(ref, "../") {
				content, shouldExpand, err := b.loadExternalRef(ref, currentDir, parentKey)
				if err != nil {
					return fmt.Errorf("failed to resolve ref '%s': %w", ref, err)
				}

				if shouldExpand {
					delete(v, "$ref")
					for k, val := range content {
						v[k] = val
					}
				} else {
					componentPath := b.getComponentPath(filepath.Join(currentDir, ref), parentKey)
					v["$ref"] = componentPath
				}
				return nil
			}
			return nil
		}

		for key, val := range v {
			if err := b.resolveExternalRefs(val, currentDir, key); err != nil {
				return err
			}
		}

	case []any:
		for _, item := range v {
			if err := b.resolveExternalRefs(item, currentDir, parentKey); err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *Bundler) loadExternalRef(ref string, currentDir string, parentKey string) (map[string]any, bool, error) {
	fullPath := filepath.Join(currentDir, ref)
	shouldExpand := strings.Contains(ref, "/paths/")

	var componentPath string
	if !shouldExpand {
		componentPath = b.getComponentPath(fullPath, parentKey)

		if cached, exists := b.components[componentPath]; exists {
			return cached, false, nil
		}
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read referenced file %s (full path: %s): %w", ref, fullPath, err)
	}

	var content map[string]any
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, false, fmt.Errorf("failed to parse referenced file %s: %w", ref, err)
	}

	newDir := filepath.Dir(fullPath)
	if err := b.resolveExternalRefs(content, newDir, parentKey); err != nil {
		return nil, false, err
	}

	if !shouldExpand {
		b.components[componentPath] = content
	}

	return content, shouldExpand, nil
}

func (b *Bundler) getComponentPath(fullPath string, parentKey string) string {
	relPath, _ := filepath.Rel(b.baseDir, fullPath)
	relPath = filepath.ToSlash(relPath)

	relPath = strings.TrimSuffix(relPath, ".yaml")
	relPath = strings.TrimSuffix(relPath, ".yml")

	if strings.Contains(relPath, "schemas/") {
		parts := strings.Split(relPath, "/")
		name := parts[len(parts)-1]
		return "#/components/schemas/" + name
	} else if strings.Contains(relPath, "responses/") {
		parts := strings.Split(relPath, "/")
		name := parts[len(parts)-1]
		return "#/components/responses/" + name
	} else if strings.Contains(relPath, "paths/") {
		return ""
	}

	parts := strings.Split(relPath, "/")
	name := parts[len(parts)-1]
	return "#/components/schemas/" + name
}

func (b *Bundler) BundleToFile(specPath, outputPath string) error {
	bundled, err := b.Bundle(specPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, bundled, 0644); err != nil {
		return fmt.Errorf("failed to write bundled file: %w", err)
	}

	return nil
}
