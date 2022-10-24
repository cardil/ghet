package ght

import "github.com/spf13/cobra"

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove installed artifact",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("remove")
		},
	}
}
