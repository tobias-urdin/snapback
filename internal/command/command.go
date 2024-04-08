package command

import (
	"github.com/tobias-urdin/snapback/internal/exporter"
	"github.com/tobias-urdin/snapback/internal/importer"

	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "snapback",
		Short:            "short",
		Long:             "TODO",
		TraverseChildren: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(exporter.NewCommand())
	cmd.AddCommand(importer.NewCommand())

	return cmd
}
