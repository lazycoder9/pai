package cmd

import (
	"fmt"
	"os"

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
			Use:   entityType + " <slug>",
			Short: "Edit a " + entityType,
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				slug := args[0]
				dir, _ := os.Getwd()
				root, err := internal.FindRoot(dir)
				if err != nil {
					return err
				}

				e, err := internal.FindEntityByType(root, entityType, slug)
				if err != nil {
					return err
				}

				// Delete old file first (path may change for tasks)
				oldPath := e.FilePath

				if s, _ := cmd.Flags().GetString("status"); s != "" {
					if entityType == "task" {
						// Tasks need to move between directories
						if err := internal.MoveTask(root, e, s); err != nil {
							return err
						}
						fmt.Printf("Updated %s %s (status: %s)\n", entityType, slug, s)
						return nil
					}
					e.Status = s
				}
				if p, _ := cmd.Flags().GetString("parent"); p != "" {
					e.Parent = p
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

				// Remove old file if path changed
				_ = oldPath
				if err := internal.SaveEntity(root, e); err != nil {
					return err
				}
				fmt.Printf("Updated %s: %s\n", entityType, slug)
				return nil
			},
		}
		cmd.Flags().String("status", "", "New status")
		cmd.Flags().String("parent", "", "New parent slug")
		cmd.Flags().String("tags", "", "New tags (comma-separated, replaces existing)")
		cmd.Flags().String("priority", "", "New priority")
		cmd.Flags().String("body", "", "New body content")
		editCmd.AddCommand(cmd)
	}

	rootCmd.AddCommand(editCmd)
}
