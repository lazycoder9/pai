package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/lazycoder9/pai/internal"
	"github.com/spf13/cobra"
)

func init() {
	types := []string{"idea", "feature", "task", "decision"}

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new entity",
	}

	for _, t := range types {
		entityType := t
		cmd := &cobra.Command{
			Use:   entityType + " <slug>",
			Short: "Add a new " + entityType,
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				slug := args[0]
				dir, _ := os.Getwd()
				root, err := internal.FindRoot(dir)
				if err != nil {
					return err
				}

				status, _ := cmd.Flags().GetString("status")
				parent, _ := cmd.Flags().GetString("parent")
				tags, _ := cmd.Flags().GetString("tags")
				priority, _ := cmd.Flags().GetString("priority")
				body, _ := cmd.Flags().GetString("body")
				if !cmd.Flags().Changed("body") {
					body = readStdin()
				}

				if status == "" {
					status = internal.DefaultStatus(entityType)
				}

				if entityType == "decision" {
					slug = internal.Slugify(slug)
				}
				if err := internal.EnsureUniqueSlug(root, entityType, slug, ""); err != nil {
					return err
				}

				id, err := internal.NextEntityID(root, entityType)
				if err != nil {
					return err
				}

				parentID := ""
				if parent != "" {
					parentEntity, err := internal.FindEntity(root, parent)
					if err != nil {
						return err
					}
					parentID = parentEntity.ID
				}

				e := &internal.Entity{
					ID:       id,
					Slug:     slug,
					Type:     entityType,
					Status:   status,
					ParentID: parentID,
					Priority: priority,
					Body:     body,
				}

				if tags != "" {
					for _, t := range splitTags(tags) {
						e.Tags = append(e.Tags, t)
					}
				}

				if err := internal.SaveEntity(root, e); err != nil {
					return err
				}
				fmt.Printf("Created %s: %s\n", entityType, e.DisplayName())
				return nil
			},
		}
		cmd.Flags().String("status", "", "Status (default depends on type)")
		cmd.Flags().String("parent", "", "Parent entity id or slug")
		cmd.Flags().String("tags", "", "Comma-separated tags")
		cmd.Flags().String("priority", "", "Priority (low, medium, high)")
		cmd.Flags().String("body", "", "Body content for the entity")
		addCmd.AddCommand(cmd)
	}

	rootCmd.AddCommand(addCmd)
}

func readStdin() string {
	info, err := os.Stdin.Stat()
	if err != nil {
		return ""
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		return ""
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}
	return string(data)
}

func splitTags(s string) []string {
	var tags []string
	for _, t := range splitComma(s) {
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func splitComma(s string) []string {
	var parts []string
	for _, p := range append([]string{}, split(s, ",")...) {
		parts = append(parts, trim(p))
	}
	return parts
}

func split(s, sep string) []string {
	if s == "" {
		return nil
	}
	result := []string{}
	for len(s) > 0 {
		idx := indexOf(s, sep)
		if idx < 0 {
			result = append(result, s)
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	return result
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
