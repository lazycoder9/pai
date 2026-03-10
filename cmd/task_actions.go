package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ula-t/pai/internal"
)

func init() {
	startCmd := &cobra.Command{
		Use:   "start <slug>",
		Short: "Start a task (move to active)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			dir, _ := os.Getwd()
			root, err := internal.FindRoot(dir)
			if err != nil {
				return err
			}

			e, err := internal.FindEntityByType(root, "task", slug)
			if err != nil {
				return err
			}

			if err := internal.MoveTask(root, e, "active"); err != nil {
				return err
			}
			fmt.Printf("Started task: %s\n", slug)
			return nil
		},
	}

	completeCmd := &cobra.Command{
		Use:   "complete <slug>",
		Short: "Complete a task (move to done)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			dir, _ := os.Getwd()
			root, err := internal.FindRoot(dir)
			if err != nil {
				return err
			}

			e, err := internal.FindEntityByType(root, "task", slug)
			if err != nil {
				return err
			}

			if err := internal.MoveTask(root, e, "done"); err != nil {
				return err
			}
			fmt.Printf("Completed task: %s\n", slug)
			return nil
		},
	}

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(completeCmd)
}
