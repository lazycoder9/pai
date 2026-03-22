package internal

import (
	"path/filepath"
	"testing"
)

func TestParseLegacyEntityUsesIDAsSlugAndParentID(t *testing.T) {
	entity, err := Parse("id: auth-login\ntype: task\nparent: feature-auth\n\n---\n\nnotes\n", "tasks/backlog/auth-login.md")
	if err != nil {
		t.Fatalf("parse legacy entity: %v", err)
	}

	if entity.ID != "auth-login" {
		t.Fatalf("expected legacy id to be preserved, got %q", entity.ID)
	}
	if entity.Slug != "auth-login" {
		t.Fatalf("expected legacy slug fallback, got %q", entity.Slug)
	}
	if entity.ParentID != "feature-auth" {
		t.Fatalf("expected legacy parent fallback, got %q", entity.ParentID)
	}
}

func TestParseEntityReadsAffectsMetadata(t *testing.T) {
	entity, err := Parse("id: D-1\ntype: decision\naffects: I-1, T-3\n\n---\n\nnotes\n", "decisions/D-1-link-context.md")
	if err != nil {
		t.Fatalf("parse entity with affects: %v", err)
	}

	if len(entity.Affects) != 2 || entity.Affects[0] != "I-1" || entity.Affects[1] != "T-3" {
		t.Fatalf("expected affects to be parsed, got %#v", entity.Affects)
	}
}

func TestNextEntityIDAndFindEntitySupportTypedIDsAndSlugs(t *testing.T) {
	root := t.TempDir()
	if err := Init(root, "demo"); err != nil {
		t.Fatalf("init project: %v", err)
	}

	idea := &Entity{ID: "I-1", Slug: "project-memory", Type: "idea", Status: "raw"}
	if err := SaveEntity(root, idea); err != nil {
		t.Fatalf("save idea: %v", err)
	}

	task := &Entity{
		ID:       "T-2",
		Slug:     "auth-login",
		Type:     "task",
		Status:   "backlog",
		ParentID: "I-1",
	}
	if err := SaveEntity(root, task); err != nil {
		t.Fatalf("save task: %v", err)
	}

	nextID, err := NextEntityID(root, "task")
	if err != nil {
		t.Fatalf("next entity id: %v", err)
	}
	if nextID != "T-3" {
		t.Fatalf("expected next task id T-3, got %q", nextID)
	}

	foundByID, err := FindEntity(root, "T-2")
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if foundByID.Slug != "auth-login" {
		t.Fatalf("expected slug auth-login, got %q", foundByID.Slug)
	}

	foundBySlug, err := FindEntityByType(root, "task", "auth-login")
	if err != nil {
		t.Fatalf("find by slug: %v", err)
	}
	if foundBySlug.ID != "T-2" {
		t.Fatalf("expected id T-2, got %q", foundBySlug.ID)
	}

	expectedPath := filepath.Join("tasks", "backlog", "T-2-auth-login.md")
	if foundBySlug.FilePath != expectedPath {
		t.Fatalf("expected filepath %q, got %q", expectedPath, foundBySlug.FilePath)
	}
}

func TestGetEntityContextFindsReverseDecisionLinks(t *testing.T) {
	root := t.TempDir()
	if err := Init(root, "demo"); err != nil {
		t.Fatalf("init project: %v", err)
	}

	idea := &Entity{ID: "I-1", Slug: "project-memory", Type: "idea", Status: "raw"}
	feature := &Entity{ID: "F-1", Slug: "show-decision-context", Type: "feature", Status: "spec", ParentID: "I-1"}
	task := &Entity{ID: "T-1", Slug: "render-links", Type: "task", Status: "backlog", ParentID: "F-1"}
	decision := &Entity{ID: "D-1", Slug: "decisions-affect-work", Type: "decision", Affects: []string{"F-1"}}

	for _, entity := range []*Entity{idea, feature, task, decision} {
		if err := SaveEntity(root, entity); err != nil {
			t.Fatalf("save %s: %v", entity.DisplayName(), err)
		}
	}

	ctx, err := GetEntityContext(root, feature)
	if err != nil {
		t.Fatalf("get entity context: %v", err)
	}

	if len(ctx.Ancestors) != 1 || ctx.Ancestors[0].ID != "I-1" {
		t.Fatalf("expected idea ancestor, got %#v", ctx.Ancestors)
	}
	if len(ctx.Descendants) != 1 || ctx.Descendants[0].ID != "T-1" {
		t.Fatalf("expected task descendant, got %#v", ctx.Descendants)
	}
	if len(ctx.RelatedDecisions) != 1 || ctx.RelatedDecisions[0].ID != "D-1" {
		t.Fatalf("expected reverse decision link, got %#v", ctx.RelatedDecisions)
	}
}
