package ght

import (
	"github.com/cardil/ghet/pkg/metadata"
	"github.com/spf13/cobra"
)

func versionCmd(_ *Args) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println(metadata.Version)
		},
	}
}
