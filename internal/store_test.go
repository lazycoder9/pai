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
