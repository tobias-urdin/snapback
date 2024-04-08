package importer

import (
	"os"
	"os/signal"
	"syscall"
	"time"
	"context"
	"crypto/tls"

	"github.com/tobias-urdin/snapback/internal/message"

	"go.uber.org/zap"
	"github.com/quic-go/quic-go"
)

// Exporter address
// TODO(tobias.urdin): Hardcoded
const exporterAddr = "localhost:4242"

// Importer
type Importer struct {
	// Logger
	logger *zap.Logger

	// Message handler
	handler *message.MessageHandler

	// Connection
	conn quic.Connection
}

// Create a new importer
func NewImporter(logger *zap.Logger) *Importer {
	return &Importer{
		logger: logger,
	}
}

// Initialize importer
func (i *Importer) Init() error {
	i.logger.Info("initialize importer")
        i.handler = message.NewHandler(i.logger)

        return nil
}

// Close importer
func (i *Importer) Close() {
	i.logger.Info("close importer")
}

// Generate TLS config
func (i *Importer) generateTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"snapback"},
	}
}

// Run one iteration of the importer
func (i *Importer) run(ctx context.Context) error {
	stream, err := i.conn.OpenStreamSync(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	logger := i.logger.With(
		zap.Any("stream", stream.StreamID()))

	logger.Info("starting import")

	listMsg := message.ListPoolRequestV1{
		Pool: "nova",
	}

	if err := message.Send(stream, &listMsg); err != nil {
		return err
	}

	var respMsg message.Message
	if err := i.handler.Read(stream, &respMsg); err != nil {
		return err
	}

	var respList message.ListPoolResponseV1
	if err := respMsg.Unmarshal(&respList); err != nil {
		return err
	}

	logger.Info("list pool response message", zap.Any("msg", respList))

	cb := func(msg *message.Message) error {
		var chunk message.ExportChunkV1
		if err := msg.Unmarshal(&chunk); err != nil {
			return err
		}

		logger.Info("got export chunk", zap.Any("crc", chunk.PayloadCRC))
		return nil
	}

	for _, image := range respList.Names {
		logger.Info("get snapshots for image", zap.String("image", image))
		listSnapMsg := message.ListSnapshotsRequestV1{
			Pool: "nova",
			Image: image,
		}

		if err := message.Send(stream, &listSnapMsg); err != nil {
			logger.Error("failed to send message", zap.String("error", err.Error()))
			continue
		}

		var respSnapMsg message.Message
		if err := i.handler.Read(stream, &respSnapMsg); err != nil {
			logger.Error("failed to read message", zap.String("error", err.Error()))
			continue
		}

		var respSnaps message.ListSnapshotsResponseV1
		if err := respSnapMsg.Unmarshal(&respSnaps); err != nil {
			logger.Error("failed to unmarshal snaps", zap.String("error", err.Error()))
			continue
		}

		logger.Info("found snapshots for image", zap.String("image", image), zap.Any("snapshots", respSnaps.Snapshots))

		// TODO(tobias.urdin): Hack just testing
		for _, snap := range respSnaps.Snapshots {
			exp := message.ExportRequestV1{
				Pool: "nova",
				Image: image,
				Snapshot: snap,
			}

			if err := message.Send(stream, &exp); err != nil {
				logger.Error("failed to send export message", zap.String("error", err.Error()))
				continue
			}

			rawResp, err := i.handler.ReadChunks(stream, cb)
			if err != nil {
				logger.Error("failed to read chunks", zap.String("error", err.Error()))
				continue
			}

			var resp message.ExportResponseV1
			if err := rawResp.Unmarshal(&resp); err != nil {
				logger.Error("failed to read export response message", zap.String("error", err.Error()))
				continue
			}

			logger.Info("got export response", zap.Any("msg", resp))
		}
	}

	return nil
}

// Run the importer
func (i *Importer) Run() error {
	i.logger.Info("connecting to exporter", zap.String("address", exporterAddr))

	sigC := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	conn, err := quic.DialAddr(ctx, exporterAddr, i.generateTLSConfig(), nil)
	if err != nil {
		return err
	}
	i.conn = conn
	defer conn.CloseWithError(0, "Goodbye")

	// TODO(tobias.urdin): Validate that we trust the exporters certificate fingerprint

	// Start the import loop that we run on an interval
	go func(ctx context.Context) {
		i.logger.Info("starting import loop")

		for {
			if err := i.run(ctx); err != nil {
				i.logger.Error("import run failed", zap.String("error", err.Error()))
			}

			i.logger.Info("next run in 10 seconds")
			time.Sleep(10 * time.Second)
		}

		i.logger.Info("stopping import loop")
	}(ctx)

	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	i.logger.Info("signal captured, exiting...")

	cancel()

	return nil
}
