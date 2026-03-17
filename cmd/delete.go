package cmd

import (
	"fmt"
	"os"

	"github.com/lazycoder9/pai/internal"
	"github.com/spf13/cobra"
)

func init() {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an entity",
	}

	types := []string{"idea", "feature", "task", "decision"}

	for _, t := range types {
		entityType := t
		cmd := &cobra.Command{
			Use:   entityType + " <ref>",
			Short: "Delete a " + entityType,
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

				if err := internal.DeleteEntity(root, e); err != nil {
					return err
				}
				fmt.Printf("Deleted %s: %s\n", entityType, e.DisplayName())
				return nil
			},
		}
		deleteCmd.AddCommand(cmd)
	}

	rootCmd.AddCommand(deleteCmd)
}
