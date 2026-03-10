package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ula-t/pai/internal"
)

func init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List entities",
	}

	types := []string{"ideas", "features", "tasks", "decisions"}
	singular := map[string]string{
		"ideas": "idea", "features": "feature",
		"tasks": "task", "decisions": "decision",
	}

	for _, t := range types {
		typePlural := t
		entityType := singular[t]
		cmd := &cobra.Command{
			Use:   typePlural,
			Short: "List " + typePlural,
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				dir, _ := os.Getwd()
				root, err := internal.FindRoot(dir)
				if err != nil {
					return err
				}

				status, _ := cmd.Flags().GetString("status")
				tag, _ := cmd.Flags().GetString("tag")

				entities, err := internal.ListEntities(root, entityType, status, tag)
				if err != nil {
					return err
				}
				if len(entities) == 0 {
					fmt.Printf("No %s found.\n", typePlural)
					return nil
				}
				fmt.Printf("%s:\n", typePlural)
				for _, e := range entities {
					internal.PrintEntity(e, false)
				}
				return nil
			},
		}
		cmd.Flags().String("status", "", "Filter by status")
		cmd.Flags().String("tag", "", "Filter by tag")
		listCmd.AddCommand(cmd)
	}

	rootCmd.AddCommand(listCmd)
}
