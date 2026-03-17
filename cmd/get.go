package cmd

import (
	"os"

	"github.com/lazycoder9/pai/internal"
	"github.com/spf13/cobra"
)

func init() {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get an entity by slug",
	}

	types := []string{"idea", "feature", "task", "decision"}

	for _, t := range types {
		entityType := t
		cmd := &cobra.Command{
			Use:   entityType + " <ref>",
			Short: "Get a " + entityType,
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

				all, _ := cmd.Flags().GetBool("all")
				if all {
					related, err := internal.GetRelated(root, e)
					if err != nil {
						return err
					}
					internal.PrintEntityWithRelated(e, related)
				} else {
					internal.PrintEntityFull(e)
				}
				return nil
			},
		}
		cmd.Flags().Bool("all", false, "Show all related entities (parent chain + children)")
		getCmd.AddCommand(cmd)
	}

	// Also allow `pai get <ref>` without specifying type
	getCmd.Args = cobra.ExactArgs(1)
	getCmd.RunE = func(cmd *cobra.Command, args []string) error {
		ref := args[0]
		dir, _ := os.Getwd()
		root, err := internal.FindRoot(dir)
		if err != nil {
			return err
		}

		e, err := internal.FindEntity(root, ref)
		if err != nil {
			return err
		}

		all, _ := cmd.Flags().GetBool("all")
		if all {
			related, relErr := internal.GetRelated(root, e)
			if relErr != nil {
				return relErr
			}
			internal.PrintEntityWithRelated(e, related)
		} else {
			internal.PrintEntityFull(e)
		}
		return nil
	}
	getCmd.Flags().Bool("all", false, "Show all related entities (parent chain + children)")

	rootCmd.AddCommand(getCmd)
}
