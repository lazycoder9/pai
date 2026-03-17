package internal

import (
	"fmt"
	"sort"
	"strings"
)

func PrintEntity(e *Entity, verbose bool) {
	fmt.Printf("  %s (%s)", e.DisplayName(), e.Type)
	if e.Status != "" {
		fmt.Printf(" [%s]", e.Status)
	}
	if e.Priority != "" {
		fmt.Printf(" priority:%s", e.Priority)
	}
	if len(e.Tags) > 0 {
		fmt.Printf(" tags:%s", strings.Join(e.Tags, ","))
	}
	if e.ParentID != "" {
		fmt.Printf(" parent:%s", e.ParentID)
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
	fmt.Println(formatEntityLine(e, true, false))
	printEntityMetadata(e)
	printEntityBody(e)
}

func PrintEntityWithRelated(e *Entity, related []*Entity) {
	PrintEntityFull(e)

	if len(related) == 0 {
		return
	}

	entities := map[string]*Entity{
		e.ID: e,
	}
	for _, r := range related {
		entities[r.ID] = r
	}

	entityList := make([]*Entity, 0, len(entities))
	for _, candidate := range entities {
		entityList = append(entityList, candidate)
	}

	root := e
	for root.ParentID != "" {
		parent := lookupEntityByRef(entityList, root.ParentID)
		if parent == nil {
			break
		}
		root = parent
	}

	childMap := buildChildMap(entityList)

	fmt.Printf("\n%s── Context ──%s\n", colorDim, colorReset)
	printContextNode(root, childMap, "", true, e.ID)
}

// ANSI color helpers
const (
	colorReset    = "\033[0m"
	colorDim      = "\033[2m"
	colorBold     = "\033[1m"
	colorCyan     = "\033[36m"
	colorHiYellow = "\033[93m"
	colorGreen    = "\033[32m"
	colorYellow   = "\033[33m"
	colorBlue     = "\033[34m"
	colorMagenta  = "\033[35m"
	colorWhite    = "\033[37m"
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

func formatEntityLine(e *Entity, emphasizeName bool, markCurrent bool) string {
	icon := typeIcon(e.Type)
	name := e.DisplayName()
	if markCurrent {
		name = colorHiYellow + colorBold + e.DisplayName() + colorReset
	} else if emphasizeName {
		name = colorBold + e.DisplayName() + colorReset
	}

	line := fmt.Sprintf("%s %s", icon, name)
	if e.Status != "" {
		line += " " + statusStyle(e.Status)
	}
	if markCurrent {
		line = colorHiYellow + "→ " + colorReset + line
	}

	return line
}

func printEntityMetadata(e *Entity) {
	printMetaLine("id", e.ID)
	if e.Slug != "" {
		printMetaLine("slug", e.Slug)
	}
	printMetaLine("type", e.Type)
	if e.FilePath != "" {
		printMetaLine("path", e.FilePath)
	}
	if e.ParentID != "" {
		printMetaLine("parent_id", e.ParentID)
	}
	if len(e.Tags) > 0 {
		printMetaLine("tags", strings.Join(e.Tags, ", "))
	}
	if e.Priority != "" {
		printMetaLine("priority", e.Priority)
	}

	keys := make([]string, 0, len(e.Extra))
	for k := range e.Extra {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		printMetaLine(k, e.Extra[k])
	}
}

func printMetaLine(label, value string) {
	fmt.Printf("  %s%-10s%s %s\n", colorDim, label+":", colorReset, value)
}

func printEntityBody(e *Entity) {
	if strings.TrimSpace(e.Body) == "" {
		return
	}

	fmt.Printf("\n%s── Notes ──%s\n", colorDim, colorReset)
	lines := strings.Split(strings.TrimRight(e.Body, "\n"), "\n")
	for _, line := range lines {
		if line == "" {
			fmt.Println()
			continue
		}
		fmt.Printf("  %s\n", line)
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
	topLevel := make(map[string][]*Entity)

	all := make([]*Entity, 0)
	all = append(all, ideas...)
	all = append(all, features...)
	all = append(all, tasks...)
	all = append(all, decisions...)

	childMap := buildChildMap(all)
	for _, e := range all {
		if canonicalParentID(e, all) == "" {
			topLevel[e.Type] = append(topLevel[e.Type], e)
		}
	}
	for entityType := range topLevel {
		SortEntities(topLevel[entityType])
	}

	printed := make(map[string]bool)

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
			printTreeNode(e, childMap, "", printed)
		}
	}

	hasOrphans := false
	for _, e := range topLevel["feature"] {
		if !printed[e.ID] && !allLeafsDone(e, childMap) {
			if !hasOrphans {
				fmt.Println()
				hasOrphans = true
			}
			printTreeNode(e, childMap, "", printed)
		}
	}

	hasOrphans = false
	for _, e := range topLevel["task"] {
		if !printed[e.ID] && !allLeafsDone(e, childMap) {
			if !hasOrphans {
				fmt.Println()
				hasOrphans = true
			}
			printTreeNode(e, childMap, "", printed)
		}
	}

	if len(topLevel["decision"]) > 0 {
		fmt.Println()
		fmt.Printf("%s── Decisions ──%s\n", colorDim, colorReset)
		for _, e := range topLevel["decision"] {
			if !printed[e.ID] {
				printed[e.ID] = true
				fmt.Printf("  %s\n", formatEntityLine(e, false, false))
			}
		}
	}
}

func printTreeNode(e *Entity, childMap map[string][]*Entity, prefix string, printed map[string]bool) {
	if printed[e.ID] {
		return
	}
	printed[e.ID] = true

	icon := typeIcon(e.Type)
	status := ""
	if e.Status != "" {
		status = " " + statusStyle(e.Status)
	}

	name := colorBold + e.DisplayName() + colorReset
	fmt.Printf("%s%s %s%s\n", prefix, icon, name, status)

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

	fmt.Printf("%s%s %s%s\n", line, icon, e.DisplayName(), status)

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

func printContextNode(e *Entity, childMap map[string][]*Entity, prefix string, emphasizeName bool, currentID string) {
	fmt.Printf("%s%s\n", prefix, formatEntityLine(e, emphasizeName, e.ID == currentID))

	children := childMap[e.ID]
	for i, child := range children {
		isLast := i == len(children)-1
		connector := "├── "
		childPrefix := prefix + "│   "
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		}
		printContextChildNode(child, childMap, prefix+connector, childPrefix, currentID)
	}
}

func printContextChildNode(e *Entity, childMap map[string][]*Entity, line, prefix string, currentID string) {
	fmt.Printf("%s%s\n", line, formatEntityLine(e, false, e.ID == currentID))

	children := childMap[e.ID]
	for i, child := range children {
		isLast := i == len(children)-1
		connector := "├── "
		childPrefix := prefix + "│   "
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		}
		printContextChildNode(child, childMap, prefix+connector, childPrefix, currentID)
	}
}
