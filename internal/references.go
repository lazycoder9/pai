package internal

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func EntityPrefix(entityType string) string {
	switch entityType {
	case "idea":
		return "I"
	case "feature":
		return "F"
	case "task":
		return "T"
	case "decision":
		return "D"
	default:
		return strings.ToUpper(entityType[:1])
	}
}

func Slugify(value string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(value)))
	return strings.Join(fields, "-")
}

func (e *Entity) DisplayName() string {
	if e.Slug == "" || e.Slug == e.ID {
		return e.ID
	}
	return e.ID + " " + e.Slug
}

func entityFileName(id, slug string) string {
	if slug == "" || slug == id {
		return id
	}
	return id + "-" + slug
}

// EntityPath returns the file path for an entity relative to .pai/
func EntityPath(entityType, id, slug, status string) string {
	dir := TypeDir(entityType)
	fileName := entityFileName(id, slug) + ".md"
	if entityType == "task" {
		subdir := TaskStatusDir(status)
		return filepath.Join(dir, subdir, fileName)
	}
	return filepath.Join(dir, fileName)
}

func parseEntitySequence(id string) (string, int, bool) {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) != 2 {
		return "", 0, false
	}

	number, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, false
	}

	return parts[0], number, true
}

func lookupEntityByRef(entities []*Entity, ref string) *Entity {
	if ref == "" {
		return nil
	}

	for _, e := range entities {
		if e.ID == ref {
			return e
		}
	}
	for _, e := range entities {
		if e.Slug == ref {
			return e
		}
	}

	return nil
}

func canonicalParentID(e *Entity, entities []*Entity) string {
	if e.ParentID == "" {
		return ""
	}

	parent := lookupEntityByRef(entities, e.ParentID)
	if parent == nil {
		return e.ParentID
	}

	return parent.ID
}

func buildChildMap(entities []*Entity) map[string][]*Entity {
	childMap := make(map[string][]*Entity)
	for _, e := range entities {
		parentID := canonicalParentID(e, entities)
		if parentID == "" {
			continue
		}
		childMap[parentID] = append(childMap[parentID], e)
	}

	for parentID := range childMap {
		SortEntities(childMap[parentID])
	}

	return childMap
}

func SortEntities(entities []*Entity) {
	sort.Slice(entities, func(i, j int) bool {
		left := entities[i]
		right := entities[j]
		if entityTypeOrder(left.Type) != entityTypeOrder(right.Type) {
			return entityTypeOrder(left.Type) < entityTypeOrder(right.Type)
		}

		leftPrefix, leftNumber, leftOK := parseEntitySequence(left.ID)
		rightPrefix, rightNumber, rightOK := parseEntitySequence(right.ID)
		if leftOK && rightOK && leftPrefix == rightPrefix {
			if leftNumber != rightNumber {
				return leftNumber < rightNumber
			}
		} else if leftOK != rightOK {
			return leftOK
		}

		if left.Slug != right.Slug {
			return left.Slug < right.Slug
		}
		return left.ID < right.ID
	})
}

func entityTypeOrder(entityType string) int {
	switch entityType {
	case "idea":
		return 0
	case "feature":
		return 1
	case "task":
		return 2
	case "decision":
		return 3
	default:
		return 4
	}
}
