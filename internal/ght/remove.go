package ght

import "github.com/spf13/cobra"

func removeCmd(_ *Args) *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove an installed artifact",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println("remove")
		},
	}
}
