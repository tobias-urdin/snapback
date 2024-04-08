package exporter

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exporter",
		Short: "Run exporter",
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

	if err := runExporter(cmd, logger); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func runExporter(cmd *cobra.Command, logger *zap.Logger) error {
	logger.Info("starting exporter")

	exp := NewExporter(logger)

	if err := exp.Init(); err != nil {
		return err
	}
	defer exp.Close()

	return exp.Run()
}
