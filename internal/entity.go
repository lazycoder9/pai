package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Entity struct {
	// Parsed metadata
	ID       string
	Slug     string
	Type     string // idea, feature, task, decision
	Status   string
	ParentID string
	Affects  []string
	Tags     []string
	Priority string

	// Extra metadata fields not in the known set
	Extra map[string]string

	// Free-form content below ---
	Body string

	// File path (relative to .pai/)
	FilePath string
}

var knownFields = map[string]bool{
	"id": true, "slug": true, "type": true, "status": true,
	"parent": true, "parent_id": true, "affects": true, "tags": true, "priority": true,
}

func ParseFile(path string) (*Entity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(string(data), path)
}

func Parse(content string, filePath string) (*Entity, error) {
	e := &Entity{
		FilePath: filePath,
		Extra:    make(map[string]string),
	}

	parts := strings.SplitN(content, "\n---", 2)
	header := parts[0]
	if len(parts) == 2 {
		e.Body = strings.TrimLeft(parts[1], "-\n")
	}

	scanner := bufio.NewScanner(strings.NewReader(header))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		key, val, ok := parseKV(line)
		if !ok {
			continue
		}
		switch key {
		case "id":
			e.ID = val
		case "slug":
			e.Slug = val
		case "type":
			e.Type = val
		case "status":
			e.Status = val
		case "parent_id":
			e.ParentID = val
		case "parent":
			if e.ParentID == "" {
				e.ParentID = val
			}
		case "affects":
			for _, ref := range strings.Split(val, ",") {
				ref = strings.TrimSpace(ref)
				if ref != "" {
					e.Affects = append(e.Affects, ref)
				}
			}
		case "tags":
			for _, t := range strings.Split(val, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					e.Tags = append(e.Tags, t)
				}
			}
		case "priority":
			e.Priority = val
		default:
			e.Extra[key] = val
		}
	}

	if e.Slug == "" {
		e.Slug = e.ID
	}
	if e.ID == "" {
		e.ID = e.Slug
	}

	return e, nil
}

func parseKV(line string) (string, string, bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}
	key := strings.TrimSpace(line[:idx])
	val := strings.TrimSpace(line[idx+1:])
	return key, val, true
}

func (e *Entity) Serialize() string {
	var b strings.Builder

	if e.ID != "" {
		fmt.Fprintf(&b, "id: %s\n", e.ID)
	}
	if e.Slug != "" {
		fmt.Fprintf(&b, "slug: %s\n", e.Slug)
	}
	if e.Type != "" {
		fmt.Fprintf(&b, "type: %s\n", e.Type)
	}
	if e.Status != "" {
		fmt.Fprintf(&b, "status: %s\n", e.Status)
	}
	if e.ParentID != "" {
		fmt.Fprintf(&b, "parent_id: %s\n", e.ParentID)
	}
	if len(e.Affects) > 0 {
		fmt.Fprintf(&b, "affects: %s\n", strings.Join(e.Affects, ", "))
	}
	if len(e.Tags) > 0 {
		fmt.Fprintf(&b, "tags: %s\n", strings.Join(e.Tags, ", "))
	}
	if e.Priority != "" {
		fmt.Fprintf(&b, "priority: %s\n", e.Priority)
	}
	for k, v := range e.Extra {
		fmt.Fprintf(&b, "%s: %s\n", k, v)
	}

	b.WriteString("\n---\n")

	if e.Body != "" {
		b.WriteString("\n")
		b.WriteString(e.Body)
	}

	return b.String()
}

// TypeDir returns the directory name for an entity type
func TypeDir(entityType string) string {
	switch entityType {
	case "idea":
		return "ideas"
	case "feature":
		return "features"
	case "task":
		return "tasks"
	case "decision":
		return "decisions"
	default:
		return entityType + "s"
	}
}

// DefaultStatus returns the default status for an entity type
func DefaultStatus(entityType string) string {
	switch entityType {
	case "idea":
		return "raw"
	case "feature":
		return "spec"
	case "task":
		return "backlog"
	case "decision":
		return ""
	default:
		return ""
	}
}

// TaskStatusDir returns the subdirectory for task status
func TaskStatusDir(status string) string {
	switch status {
	case "backlog", "active", "done":
		return status
	default:
		return "backlog"
	}
}
