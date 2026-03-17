package cmd

import (
	"fmt"
	"os"

	"github.com/lazycoder9/pai/internal"
	"github.com/spf13/cobra"
)

func init() {
	startCmd := &cobra.Command{
		Use:   "start <ref>",
		Short: "Start a task (move to active)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			dir, _ := os.Getwd()
			root, err := internal.FindRoot(dir)
			if err != nil {
				return err
			}

			e, err := internal.FindEntityByType(root, "task", ref)
			if err != nil {
				return err
			}

			if err := internal.MoveTask(root, e, "active"); err != nil {
				return err
			}
			fmt.Printf("Started task: %s\n", e.DisplayName())
			return nil
		},
	}

	completeCmd := &cobra.Command{
		Use:   "complete <ref>",
		Short: "Complete a task (move to done)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref := args[0]
			dir, _ := os.Getwd()
			root, err := internal.FindRoot(dir)
			if err != nil {
				return err
			}

			e, err := internal.FindEntityByType(root, "task", ref)
			if err != nil {
				return err
			}

			if err := internal.MoveTask(root, e, "done"); err != nil {
				return err
			}
			fmt.Printf("Completed task: %s\n", e.DisplayName())
			return nil
		},
	}

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(completeCmd)
}
