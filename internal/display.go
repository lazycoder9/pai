package internal

import (
	"fmt"
	"strings"
)

func PrintEntity(e *Entity, verbose bool) {
	fmt.Printf("  %s (%s)", e.ID, e.Type)
	if e.Status != "" {
		fmt.Printf(" [%s]", e.Status)
	}
	if e.Priority != "" {
		fmt.Printf(" priority:%s", e.Priority)
	}
	if len(e.Tags) > 0 {
		fmt.Printf(" tags:%s", strings.Join(e.Tags, ","))
	}
	if e.Parent != "" {
		fmt.Printf(" parent:%s", e.Parent)
	}
	fmt.Println()

	if verbose && e.Body != "" {
		lines := strings.Split(strings.TrimSpace(e.Body), "\n")
		for _, line := range lines {
			fmt.Printf("    %s\n", line)
		}
		fmt.Println()
	}
}

func PrintEntityFull(e *Entity) {
	fmt.Printf("id: %s\n", e.ID)
	fmt.Printf("type: %s\n", e.Type)
	if e.Status != "" {
		fmt.Printf("status: %s\n", e.Status)
	}
	if e.Parent != "" {
		fmt.Printf("parent: %s\n", e.Parent)
	}
	if len(e.Tags) > 0 {
		fmt.Printf("tags: %s\n", strings.Join(e.Tags, ", "))
	}
	if e.Priority != "" {
		fmt.Printf("priority: %s\n", e.Priority)
	}
	for k, v := range e.Extra {
		fmt.Printf("%s: %s\n", k, v)
	}
	if e.Body != "" {
		fmt.Println("---")
		fmt.Println()
		fmt.Print(e.Body)
	}
}

func PrintEntityWithRelated(e *Entity, related []*Entity) {
	fmt.Println("=== Entity ===")
	PrintEntityFull(e)
	fmt.Println()

	if len(related) == 0 {
		return
	}

	// Split related into ancestors and descendants
	var ancestors, descendants []*Entity
	foundSelf := false
	for _, r := range related {
		if r.ID == e.ID {
			foundSelf = true
			continue
		}
		_ = foundSelf // ancestors come before in the slice, descendants after
	}
	// Related is ordered: ancestors first, then descendants
	// Find the split point by checking parent chain
	for _, r := range related {
		if isAncestor(r, e, related) {
			ancestors = append(ancestors, r)
		} else {
			descendants = append(descendants, r)
		}
	}

	if len(ancestors) > 0 {
		fmt.Println("=== Parent Chain ===")
		for _, a := range ancestors {
			PrintEntity(a, false)
		}
		fmt.Println()
	}

	if len(descendants) > 0 {
		fmt.Println("=== Children ===")
		for _, d := range descendants {
			PrintEntity(d, false)
		}
		fmt.Println()
	}
}

func isAncestor(candidate, target *Entity, all []*Entity) bool {
	current := target
	for current.Parent != "" {
		if current.Parent == candidate.ID {
			return true
		}
		// Find parent in all
		found := false
		for _, e := range all {
			if e.ID == current.Parent {
				current = e
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return false
}

// ANSI color helpers
const (
	colorReset  = "\033[0m"
	colorDim    = "\033[2m"
	colorBold   = "\033[1m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorMagenta = "\033[35m"
	colorWhite  = "\033[37m"
)

func typeIcon(t string) string {
	switch t {
	case "idea":
		return "💡"
	case "feature":
		return "🔧"
	case "task":
		return "📌"
	case "decision":
		return "📋"
	default:
		return "•"
	}
}

func statusStyle(status string) string {
	switch status {
	case "raw":
		return colorDim + status + colorReset
	case "explored", "spec":
		return colorYellow + status + colorReset
	case "tasks_generated":
		return colorCyan + status + colorReset
	case "backlog":
		return colorBlue + status + colorReset
	case "active":
		return colorGreen + colorBold + status + colorReset
	case "done":
		return colorGreen + status + colorReset
	default:
		return status
	}
}

// allLeafsDone returns true if every leaf in the subtree has status "done".
// A childless entity is a leaf: it's considered done only if its status is "done".
// A non-done leaf (e.g. a raw idea with no children) is NOT considered done.
func allLeafsDone(e *Entity, childMap map[string][]*Entity) bool {
	children := childMap[e.ID]
	if len(children) == 0 {
		return e.Status == "done"
	}
	for _, child := range children {
		if !allLeafsDone(child, childMap) {
			return false
		}
	}
	return true
}

// PrintTree prints a tree view of all entities grouped by type
func PrintTree(ideas, features, tasks, decisions []*Entity) {
	// Build parent->children map
	childMap := make(map[string][]*Entity)
	topLevel := make(map[string][]*Entity) // type -> top-level entities

	all := make([]*Entity, 0)
	all = append(all, ideas...)
	all = append(all, features...)
	all = append(all, tasks...)
	all = append(all, decisions...)

	for _, e := range all {
		if e.Parent != "" {
			childMap[e.Parent] = append(childMap[e.Parent], e)
		} else {
			topLevel[e.Type] = append(topLevel[e.Type], e)
		}
	}

	printed := make(map[string]bool)

	// Ideas (and their children)
	if len(topLevel["idea"]) > 0 {
		first := true
		for _, e := range topLevel["idea"] {
			if allLeafsDone(e, childMap) {
				continue
			}
			if !first {
				fmt.Println()
			}
			first = false
			printTreeNode(e, childMap, "", true, printed)
		}
	}

	// Orphaned features
	hasOrphans := false
	for _, e := range topLevel["feature"] {
		if !printed[e.ID] && !allLeafsDone(e, childMap) {
			if !hasOrphans {
				fmt.Println()
				hasOrphans = true
			}
			printTreeNode(e, childMap, "", true, printed)
		}
	}

	// Orphaned tasks
	hasOrphans = false
	for _, e := range topLevel["task"] {
		if !printed[e.ID] && !allLeafsDone(e, childMap) {
			if !hasOrphans {
				fmt.Println()
				hasOrphans = true
			}
			printTreeNode(e, childMap, "", true, printed)
		}
	}

	// Decisions section
	if len(topLevel["decision"]) > 0 {
		fmt.Println()
		fmt.Printf("%s── Decisions ──%s\n", colorDim, colorReset)
		for _, e := range topLevel["decision"] {
			if !printed[e.ID] {
				printed[e.ID] = true
				fmt.Printf("  📋 %s%s%s\n", colorDim, e.ID, colorReset)
			}
		}
	}
}

func printTreeNode(e *Entity, childMap map[string][]*Entity, prefix string, isRoot bool, printed map[string]bool) {
	if printed[e.ID] {
		return
	}
	printed[e.ID] = true

	icon := typeIcon(e.Type)

	status := ""
	if e.Status != "" {
		status = " " + statusStyle(e.Status)
	}

	name := colorBold + e.ID + colorReset

	fmt.Printf("%s%s %s%s\n", prefix, icon, name, status)

	// Filter out children whose subtrees are all done
	var visible []*Entity
	for _, child := range childMap[e.ID] {
		if !allLeafsDone(child, childMap) {
			visible = append(visible, child)
		}
	}
	for i, child := range visible {
		isLast := i == len(visible)-1
		connector := "├── "
		childPrefix := prefix + "│   "
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		}
		printTreeChildNode(child, childMap, prefix+connector, childPrefix, printed)
	}
}

func printTreeChildNode(e *Entity, childMap map[string][]*Entity, line, prefix string, printed map[string]bool) {
	if printed[e.ID] {
		return
	}
	printed[e.ID] = true

	icon := typeIcon(e.Type)

	status := ""
	if e.Status != "" {
		status = " " + statusStyle(e.Status)
	}

	fmt.Printf("%s%s %s%s\n", line, icon, e.ID, status)

	// Filter out children whose subtrees are all done
	var visible []*Entity
	for _, child := range childMap[e.ID] {
		if !allLeafsDone(child, childMap) {
			visible = append(visible, child)
		}
	}
	for i, child := range visible {
		isLast := i == len(visible)-1
		connector := "├── "
		childPrefix := prefix + "│   "
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		}
		printTreeChildNode(child, childMap, prefix+connector, childPrefix, printed)
	}
}
