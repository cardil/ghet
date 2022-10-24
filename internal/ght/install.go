package ght

import "github.com/spf13/cobra"

func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install artifact from GitHub release",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("install")
		},
	}
}
