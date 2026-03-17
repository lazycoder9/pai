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

	if err := os.WriteFile(filepath.Join(base, "context.md"), []byte("# "+name+"\n\nProject overview for agents.\n"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(base, "architecture.md"), []byte("# Architecture\n\nTechnical architecture decisions.\n"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(base, "roadmap.md"), []byte("# Roadmap\n\nHigh-level project direction.\n"), 0o644); err != nil {
		return err
	}

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
	relPath := EntityPath(e.Type, e.ID, e.Slug, e.Status)
	fullPath := filepath.Join(PaiPath(root), relPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}

	e.FilePath = relPath
	return os.WriteFile(fullPath, []byte(e.Serialize()), 0o644)
}

func findByRef(entities []*Entity, ref string) (*Entity, error) {
	var slugMatches []*Entity

	for _, e := range entities {
		if e.ID == ref {
			return e, nil
		}
	}
	for _, e := range entities {
		if e.Slug == ref {
			slugMatches = append(slugMatches, e)
		}
	}

	if len(slugMatches) == 1 {
		return slugMatches[0], nil
	}
	if len(slugMatches) > 1 {
		return nil, fmt.Errorf("reference %q is ambiguous; use a typed id or specify the entity type", ref)
	}
	return nil, nil
}

// FindEntity finds an entity by id or slug across all type directories
func FindEntity(root, ref string) (*Entity, error) {
	entities, err := ListEntities(root, "", "", "")
	if err != nil {
		return nil, err
	}

	match, err := findByRef(entities, ref)
	if err != nil {
		return nil, err
	}
	if match == nil {
		return nil, fmt.Errorf("entity %q not found", ref)
	}
	return match, nil
}

// FindEntityByType finds an entity by id or slug within a specific type directory
func FindEntityByType(root, entityType, ref string) (*Entity, error) {
	entities, err := ListEntities(root, entityType, "", "")
	if err != nil {
		return nil, err
	}

	match, err := findByRef(entities, ref)
	if err != nil {
		return nil, err
	}
	if match == nil {
		return nil, fmt.Errorf("%s %q not found", entityType, ref)
	}
	return match, nil
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

	SortEntities(entities)
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

func NextEntityID(root, entityType string) (string, error) {
	entities, err := ListEntities(root, entityType, "", "")
	if err != nil {
		return "", err
	}

	prefix := EntityPrefix(entityType)
	maxNumber := 0
	for _, e := range entities {
		entityPrefix, number, ok := parseEntitySequence(e.ID)
		if !ok || entityPrefix != prefix {
			continue
		}
		if number > maxNumber {
			maxNumber = number
		}
	}

	return fmt.Sprintf("%s-%d", prefix, maxNumber+1), nil
}

func EnsureUniqueSlug(root, entityType, slug, excludeID string) error {
	entities, err := ListEntities(root, entityType, "", "")
	if err != nil {
		return err
	}

	for _, existing := range entities {
		if existing.ID == excludeID {
			continue
		}
		if existing.Slug == slug {
			return fmt.Errorf("%s slug %q already exists", entityType, slug)
		}
	}

	return nil
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
	newRel := EntityPath(e.Type, e.ID, e.Slug, newStatus)
	newPath := filepath.Join(PaiPath(root), newRel)

	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return err
	}

	data := e.Serialize()
	if err := os.WriteFile(newPath, []byte(data), 0o644); err != nil {
		return err
	}
	_ = os.Remove(oldPath)
	e.FilePath = newRel
	return nil
}

// GetRelated finds all entities related to the given one (up and down the chain)
func GetRelated(root string, e *Entity) ([]*Entity, error) {
	all, err := ListEntities(root, "", "", "")
	if err != nil {
		return nil, err
	}

	childMap := buildChildMap(all)
	var related []*Entity
	seen := map[string]bool{
		e.ID: true,
	}

	current := e
	for current.ParentID != "" {
		parent := lookupEntityByRef(all, current.ParentID)
		if parent == nil || seen[parent.ID] {
			break
		}
		related = append([]*Entity{parent}, related...)
		seen[parent.ID] = true
		current = parent
	}

	var walkChildren func(parentID string)
	walkChildren = func(parentID string) {
		for _, child := range childMap[parentID] {
			if seen[child.ID] {
				continue
			}
			related = append(related, child)
			seen[child.ID] = true
			walkChildren(child.ID)
		}
	}
	walkChildren(e.ID)

	return related, nil
}
