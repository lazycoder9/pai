package internal

import (
	"fmt"
	"strings"
)

type EntityContext struct {
	Ancestors        []*Entity
	Descendants      []*Entity
	RelatedDecisions []*Entity
	AffectedEntities []*Entity
	AffectedContext  []*Entity
}

func ResolveAffects(root, raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}

	var affects []string
	seen := make(map[string]bool)
	for _, ref := range splitComma(raw) {
		target, err := FindEntity(root, ref)
		if err != nil {
			return nil, err
		}
		if target.Type == "decision" {
			return nil, fmt.Errorf("decision %q cannot affect another decision", target.DisplayName())
		}
		if !seen[target.ID] {
			seen[target.ID] = true
			affects = append(affects, target.ID)
		}
	}

	return affects, nil
}

func GetEntityContext(root string, e *Entity) (*EntityContext, error) {
	all, err := ListEntities(root, "", "", "")
	if err != nil {
		return nil, err
	}

	byID := make(map[string]*Entity, len(all))
	for _, entity := range all {
		byID[entity.ID] = entity
	}

	ctx := &EntityContext{}
	childMap := buildChildMap(all)

	if e.Type == "decision" {
		seen := make(map[string]bool)
		var walkChildren func(parentID string)
		walkChildren = func(parentID string) {
			for _, child := range childMap[parentID] {
				if seen[child.ID] {
					continue
				}
				seen[child.ID] = true
				ctx.AffectedContext = append(ctx.AffectedContext, child)
				walkChildren(child.ID)
			}
		}

		for _, ref := range e.Affects {
			target := lookupEntityByRef(all, ref)
			if target != nil {
				ctx.AffectedEntities = append(ctx.AffectedEntities, target)
				if !seen[target.ID] {
					seen[target.ID] = true
					ctx.AffectedContext = append(ctx.AffectedContext, target)
				}
				current := target
				for current.ParentID != "" {
					parent := byID[canonicalParentID(current, all)]
					if parent == nil || seen[parent.ID] {
						break
					}
					seen[parent.ID] = true
					ctx.AffectedContext = append(ctx.AffectedContext, parent)
					current = parent
				}
				walkChildren(target.ID)
			}
		}
		SortEntities(ctx.AffectedEntities)
		SortEntities(ctx.AffectedContext)
		return ctx, nil
	}

	seen := map[string]bool{e.ID: true}

	current := e
	for current.ParentID != "" {
		parent := byID[canonicalParentID(current, all)]
		if parent == nil || seen[parent.ID] {
			break
		}
		ctx.Ancestors = append([]*Entity{parent}, ctx.Ancestors...)
		seen[parent.ID] = true
		current = parent
	}

	var walkChildren func(parentID string)
	walkChildren = func(parentID string) {
		for _, child := range childMap[parentID] {
			if seen[child.ID] {
				continue
			}
			ctx.Descendants = append(ctx.Descendants, child)
			seen[child.ID] = true
			walkChildren(child.ID)
		}
	}
	walkChildren(e.ID)

	for _, candidate := range all {
		if candidate.Type != "decision" {
			continue
		}
		for _, ref := range candidate.Affects {
			if ref == e.ID {
				ctx.RelatedDecisions = append(ctx.RelatedDecisions, candidate)
				break
			}
		}
	}
	SortEntities(ctx.RelatedDecisions)

	return ctx, nil
}

func splitComma(s string) []string {
	var parts []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}
