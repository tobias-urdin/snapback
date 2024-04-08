package importer

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "importer",
		Short: "Run importer",
		Long:  "TODO",
		Run:   runCommand,
	}
	return cmd
}

func runCommand(cmd *cobra.Command, args []string) {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	if err := runImporter(cmd, logger); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func runImporter(cmd *cobra.Command, logger *zap.Logger) error {
	logger.Info("starting importer")

	imp := NewImporter(logger)

	if err := imp.Init(); err != nil {
		return err
	}
	defer imp.Close()

	return imp.Run()
}
