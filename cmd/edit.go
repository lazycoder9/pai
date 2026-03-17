package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lazycoder9/pai/internal"
	"github.com/spf13/cobra"
)

func init() {
	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit an entity's metadata",
	}

	types := []string{"idea", "feature", "task", "decision"}

	for _, t := range types {
		entityType := t
		cmd := &cobra.Command{
			Use:   entityType + " <ref>",
			Short: "Edit a " + entityType,
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				ref := args[0]
				dir, _ := os.Getwd()
				root, err := internal.FindRoot(dir)
				if err != nil {
					return err
				}

				e, err := internal.FindEntityByType(root, entityType, ref)
				if err != nil {
					return err
				}

				oldPath := e.FilePath

				if slug, _ := cmd.Flags().GetString("slug"); slug != "" {
					if entityType == "decision" {
						slug = internal.Slugify(slug)
					}
					if err := internal.EnsureUniqueSlug(root, entityType, slug, e.ID); err != nil {
						return err
					}
					e.Slug = slug
				}
				if p, _ := cmd.Flags().GetString("parent"); p != "" {
					parentEntity, err := internal.FindEntity(root, p)
					if err != nil {
						return err
					}
					e.ParentID = parentEntity.ID
				}
				if t, _ := cmd.Flags().GetString("tags"); t != "" {
					e.Tags = nil
					for _, tag := range splitTags(t) {
						e.Tags = append(e.Tags, tag)
					}
				}
				if p, _ := cmd.Flags().GetString("priority"); p != "" {
					e.Priority = p
				}
				if cmd.Flags().Changed("body") {
					e.Body, _ = cmd.Flags().GetString("body")
				} else if s := readStdin(); s != "" {
					e.Body = s
				}

				if s, _ := cmd.Flags().GetString("status"); s != "" {
					if entityType == "task" {
						if err := internal.MoveTask(root, e, s); err != nil {
							return err
						}
						fmt.Printf("Updated %s: %s\n", entityType, e.DisplayName())
						return nil
					}
					e.Status = s
				}

				if err := internal.SaveEntity(root, e); err != nil {
					return err
				}
				if oldPath != "" && oldPath != e.FilePath {
					_ = os.Remove(filepath.Join(internal.PaiPath(root), oldPath))
				}
				fmt.Printf("Updated %s: %s\n", entityType, e.DisplayName())
				return nil
			},
		}
		cmd.Flags().String("status", "", "New status")
		cmd.Flags().String("slug", "", "New slug")
		cmd.Flags().String("parent", "", "New parent id or slug")
		cmd.Flags().String("tags", "", "New tags (comma-separated, replaces existing)")
		cmd.Flags().String("priority", "", "New priority")
		cmd.Flags().String("body", "", "New body content")
		editCmd.AddCommand(cmd)
	}

	rootCmd.AddCommand(editCmd)
}
