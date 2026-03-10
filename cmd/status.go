package cmd

import (
	"fmt"
	"os"

	"github.com/lazycoder9/pai/internal"
	"github.com/spf13/cobra"
)

func init() {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show project tree",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := os.Getwd()
			root, err := internal.FindRoot(dir)
			if err != nil {
				return err
			}

			ideas, _ := internal.ListEntities(root, "idea", "", "")
			features, _ := internal.ListEntities(root, "feature", "", "")
			tasks, _ := internal.ListEntities(root, "task", "", "")
			decisions, _ := internal.ListEntities(root, "decision", "", "")

			if len(ideas)+len(features)+len(tasks)+len(decisions) == 0 {
				fmt.Println("Project is empty.")
				return nil
			}

			internal.PrintTree(ideas, features, tasks, decisions)
			return nil
		},
	}
	rootCmd.AddCommand(statusCmd)
}
