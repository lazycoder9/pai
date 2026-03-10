package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const PaiDir = ".pai"

type ProjectState struct {
	Project   string `json:"project"`
	CreatedAt string `json:"created_at"`
}

// FindRoot walks up from dir looking for .pai/
func FindRoot(dir string) (string, error) {
	for {
		candidate := filepath.Join(dir, PaiDir)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf(".pai directory not found (run 'pai init' first)")
		}
		dir = parent
	}
}

func PaiPath(root string) string {
	return filepath.Join(root, PaiDir)
}

// Init creates the .pai folder structure
func Init(dir, name string) error {
	base := filepath.Join(dir, PaiDir)
	if _, err := os.Stat(base); err == nil {
		return fmt.Errorf(".pai already exists in %s", dir)
	}

	dirs := []string{
		"ideas",
		"features",
		"tasks/backlog",
		"tasks/active",
		"tasks/done",
		"decisions",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(base, d), 0o755); err != nil {
			return err
		}
	}

	// context.md
	if err := os.WriteFile(filepath.Join(base, "context.md"), []byte("# "+name+"\n\nProject overview for agents.\n"), 0o644); err != nil {
		return err
	}
	// architecture.md
	if err := os.WriteFile(filepath.Join(base, "architecture.md"), []byte("# Architecture\n\nTechnical architecture decisions.\n"), 0o644); err != nil {
		return err
	}
	// roadmap.md
	if err := os.WriteFile(filepath.Join(base, "roadmap.md"), []byte("# Roadmap\n\nHigh-level project direction.\n"), 0o644); err != nil {
		return err
	}
	// state.json
	state := ProjectState{
		Project:   name,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	data, _ := json.MarshalIndent(state, "", "  ")
	if err := os.WriteFile(filepath.Join(base, "state.json"), data, 0o644); err != nil {
		return err
	}

	return nil
}

// SaveEntity writes an entity to the correct location
func SaveEntity(root string, e *Entity) error {
	relPath := EntityPath(e.Type, e.ID, e.Status)
	fullPath := filepath.Join(PaiPath(root), relPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}

	e.FilePath = relPath
	return os.WriteFile(fullPath, []byte(e.Serialize()), 0o644)
}

// FindEntity finds an entity by slug across all type directories
func FindEntity(root, slug string) (*Entity, error) {
	base := PaiPath(root)
	var found *Entity

	err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		// Skip top-level md files (context.md, architecture.md, roadmap.md)
		rel, _ := filepath.Rel(base, path)
		if !strings.Contains(rel, string(filepath.Separator)) {
			return nil
		}

		e, parseErr := ParseFile(path)
		if parseErr != nil {
			return nil
		}
		if e.ID == slug {
			e.FilePath = rel
			found = e
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, fmt.Errorf("entity %q not found", slug)
	}
	return found, nil
}

// FindEntityByType finds an entity by slug within a specific type directory
func FindEntityByType(root, entityType, slug string) (*Entity, error) {
	base := PaiPath(root)
	dir := filepath.Join(base, TypeDir(entityType))

	var found *Entity
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		e, parseErr := ParseFile(path)
		if parseErr != nil {
			return nil
		}
		if e.ID == slug {
			rel, _ := filepath.Rel(base, path)
			e.FilePath = rel
			found = e
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if found == nil {
		return nil, fmt.Errorf("%s %q not found", entityType, slug)
	}
	return found, nil
}

// ListEntities lists all entities, optionally filtered
func ListEntities(root, entityType, status, tag string) ([]*Entity, error) {
	base := PaiPath(root)

	var searchDir string
	if entityType != "" {
		searchDir = filepath.Join(base, TypeDir(entityType))
	} else {
		searchDir = base
	}

	var entities []*Entity
	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		rel, _ := filepath.Rel(base, path)
		// Skip top-level files when listing all
		if entityType == "" && !strings.Contains(rel, string(filepath.Separator)) {
			return nil
		}

		e, parseErr := ParseFile(path)
		if parseErr != nil {
			return nil
		}
		e.FilePath = rel

		if status != "" && e.Status != status {
			return nil
		}
		if tag != "" && !hasTag(e.Tags, tag) {
			return nil
		}

		entities = append(entities, e)
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return entities, nil
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// DeleteEntity removes an entity file
func DeleteEntity(root string, e *Entity) error {
	fullPath := filepath.Join(PaiPath(root), e.FilePath)
	return os.Remove(fullPath)
}

// MoveTask moves a task file to a new status subdirectory
func MoveTask(root string, e *Entity, newStatus string) error {
	oldPath := filepath.Join(PaiPath(root), e.FilePath)
	e.Status = newStatus
	newRel := EntityPath(e.Type, e.ID, newStatus)
	newPath := filepath.Join(PaiPath(root), newRel)

	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return err
	}

	data := e.Serialize()
	if err := os.WriteFile(newPath, []byte(data), 0o644); err != nil {
		return err
	}
	os.Remove(oldPath)
	e.FilePath = newRel
	return nil
}

// GetRelated finds all entities related to the given one (up and down the chain)
func GetRelated(root string, e *Entity) ([]*Entity, error) {
	var related []*Entity

	// Traverse UP: follow parent chain
	current := e
	for current.Parent != "" {
		parent, err := FindEntity(root, current.Parent)
		if err != nil {
			break
		}
		related = append([]*Entity{parent}, related...)
		current = parent
	}

	// Traverse DOWN: find children
	children, err := findChildren(root, e.ID)
	if err == nil {
		related = append(related, children...)
	}

	return related, nil
}

func findChildren(root, parentID string) ([]*Entity, error) {
	all, err := ListEntities(root, "", "", "")
	if err != nil {
		return nil, err
	}

	var children []*Entity
	for _, e := range all {
		if e.Parent == parentID {
			children = append(children, e)
			// Recurse
			grandchildren, err := findChildren(root, e.ID)
			if err == nil {
				children = append(children, grandchildren...)
			}
		}
	}
	return children, nil
}
