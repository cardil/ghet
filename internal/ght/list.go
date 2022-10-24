package ght

import "github.com/spf13/cobra"

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed artifacts",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("list")
		},
	}
}
