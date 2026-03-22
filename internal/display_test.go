package internal

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestPrintEntityFull(t *testing.T) {
	entity := &Entity{
		ID:       "F-3",
		Slug:     "detail-view",
		Type:     "feature",
		Status:   "spec",
		ParentID: "I-2",
		Tags:     []string{"cli", "ux"},
		Priority: "high",
		Extra: map[string]string{
			"owner":    "alice",
			"severity": "medium",
		},
		Body:     "Make `pai get` easier to scan.\n\nKeep metadata visible.",
		FilePath: "features/F-3-detail-view.md",
	}

	output := captureStdout(t, func() {
		PrintEntityFull(entity)
	})
	output = stripANSI(output)

	assertContains(t, output, "🔧 F-3 detail-view spec")
	assertContains(t, output, "  id:        F-3")
	assertContains(t, output, "  slug:      detail-view")
	assertContains(t, output, "  type:      feature")
	assertContains(t, output, "  path:      features/F-3-detail-view.md")
	assertContains(t, output, "  parent_id: I-2")
	assertContains(t, output, "  tags:      cli, ux")
	assertContains(t, output, "  priority:  high")

	ownerIndex := strings.Index(output, "  owner:     alice")
	severityIndex := strings.Index(output, "  severity:  medium")
	if ownerIndex == -1 || severityIndex == -1 {
		t.Fatalf("expected sorted extra metadata in output:\n%s", output)
	}
	if ownerIndex > severityIndex {
		t.Fatalf("expected owner before severity in output:\n%s", output)
	}

	assertContains(t, output, "── Notes ──")
	assertContains(t, output, "  Make `pai get` easier to scan.")
	assertContains(t, output, "  Keep metadata visible.")
}

func TestPrintEntityWithRelatedShowsContextTree(t *testing.T) {
	idea := &Entity{ID: "I-2", Slug: "improve-status", Type: "idea", Status: "raw"}
	feature := &Entity{
		ID:       "F-3",
		Slug:     "detail-view",
		Type:     "feature",
		Status:   "spec",
		ParentID: "I-2",
		FilePath: "features/F-3-detail-view.md",
	}
	task := &Entity{ID: "T-8", Slug: "renderer-tests", Type: "task", Status: "active", ParentID: "F-3"}
	subtask := &Entity{ID: "T-9", Slug: "manual-check", Type: "task", Status: "backlog", ParentID: "T-8"}
	decision := &Entity{ID: "D-2", Slug: "render-structured-context", Type: "decision", Affects: []string{"F-3"}}

	output := captureStdout(t, func() {
		PrintEntityWithContext(feature, &EntityContext{
			Ancestors:        []*Entity{idea},
			Descendants:      []*Entity{task, subtask},
			RelatedDecisions: []*Entity{decision},
		})
	})
	output = stripANSI(output)

	assertContains(t, output, "── Context ──")
	assertContains(t, output, "💡 I-2 improve-status raw")
	assertContains(t, output, "└── → 🔧 F-3 detail-view spec")
	assertContains(t, output, "    └── 📌 T-8 renderer-tests active")
	assertContains(t, output, "        └── 📌 T-9 manual-check backlog")
	assertContains(t, output, "── Related Decisions ──")
	assertContains(t, output, "📋 D-2 render-structured-context")
}

func TestPrintDecisionContextShowsAffectedEntities(t *testing.T) {
	decision := &Entity{
		ID:      "D-3",
		Slug:    "link-decisions-to-work",
		Type:    "decision",
		Affects: []string{"I-2", "T-8"},
	}
	idea := &Entity{ID: "I-2", Slug: "improve-status", Type: "idea", Status: "raw"}
	task := &Entity{ID: "T-8", Slug: "renderer-tests", Type: "task", Status: "active"}

	output := captureStdout(t, func() {
		PrintEntityWithContext(decision, &EntityContext{
			AffectedEntities: []*Entity{idea, task},
		})
	})
	output = stripANSI(output)

	assertContains(t, output, "  affects:   I-2, T-8")
	assertContains(t, output, "── Affects ──")
	assertContains(t, output, "💡 I-2 improve-status raw")
	assertContains(t, output, "📌 T-8 renderer-tests active")
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stdout = writer

	defer func() {
		os.Stdout = oldStdout
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}

	return buf.String()
}

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

func assertContains(t *testing.T, output, want string) {
	t.Helper()
	if !strings.Contains(output, want) {
		t.Fatalf("expected %q in output:\n%s", want, output)
	}
}
