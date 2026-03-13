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
		ID:       "detail-view",
		Type:     "feature",
		Status:   "spec",
		Parent:   "improve-status",
		Tags:     []string{"cli", "ux"},
		Priority: "high",
		Extra: map[string]string{
			"owner":    "alice",
			"severity": "medium",
		},
		Body:     "Make `pai get` easier to scan.\n\nKeep metadata visible.",
		FilePath: "features/detail-view.md",
	}

	output := captureStdout(t, func() {
		PrintEntityFull(entity)
	})
	output = stripANSI(output)

	assertContains(t, output, "🔧 detail-view spec")
	assertContains(t, output, "  type:    feature")
	assertContains(t, output, "  path:    features/detail-view.md")
	assertContains(t, output, "  parent:  improve-status")
	assertContains(t, output, "  tags:    cli, ux")
	assertContains(t, output, "  priority: high")

	ownerIndex := strings.Index(output, "  owner:   alice")
	severityIndex := strings.Index(output, "  severity: medium")
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
	idea := &Entity{ID: "improve-status", Type: "idea", Status: "raw"}
	feature := &Entity{
		ID:       "detail-view",
		Type:     "feature",
		Status:   "spec",
		Parent:   "improve-status",
		FilePath: "features/detail-view.md",
	}
	task := &Entity{ID: "renderer-tests", Type: "task", Status: "active", Parent: "detail-view"}
	subtask := &Entity{ID: "manual-check", Type: "task", Status: "backlog", Parent: "renderer-tests"}

	output := captureStdout(t, func() {
		PrintEntityWithRelated(feature, []*Entity{idea, task, subtask})
	})
	output = stripANSI(output)

	assertContains(t, output, "── Context ──")
	assertContains(t, output, "💡 improve-status raw")
	assertContains(t, output, "└── → 🔧 detail-view spec")
	assertContains(t, output, "    └── 📌 renderer-tests active")
	assertContains(t, output, "        └── 📌 manual-check backlog")
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
