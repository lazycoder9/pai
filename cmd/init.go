package cmd

import (
	"fmt"
	"os"

	"github.com/lazycoder9/pai/internal"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a .pai project structure",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			dir, _ := os.Getwd()
			if err := internal.Init(dir, name); err != nil {
				return err
			}
			fmt.Printf("Initialized .pai project %q in %s\n", name, dir)
			return nil
		},
	}
	cmd.Flags().String("name", "my-project", "Project name")
	rootCmd.AddCommand(cmd)
}
