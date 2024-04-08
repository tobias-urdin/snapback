package exporter

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"

	"github.com/tobias-urdin/snapback/internal/message"

	"go.uber.org/zap"
	"github.com/quic-go/quic-go"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

// Exporter address
// TODO(tobias.urdin): Hardcoded
const exporterAddr = "localhost:4242"

// Exporter
type Exporter struct {
	// Logger
	logger *zap.Logger

	// Message handler
	handler *message.MessageHandler

	// Rados
	conn *rados.Conn
}

// Create a new exporter
func NewExporter(logger *zap.Logger) *Exporter {
	return &Exporter{
		logger: logger,
	}
}

// Initialize exporter
func (e *Exporter) Init() error {
	e.logger.Info("initialize exporter")

	e.handler = message.NewHandler(e.logger)

	e.handler.AddHandler(message.ErrorType, 1, e.handleErrorV1)
	e.handler.AddHandler(message.ListPoolRequestType, 1, e.handleListPoolRequestV1)
	e.handler.AddHandler(message.ListSnapshotsRequestType, 1, e.handleListSnapshotsRequestV1)
	e.handler.AddHandler(message.ExportRequestType, 1, e.handleExportRequestV1)

	e.logger.Info("connecting to rados")

	conn, err := rados.NewConn()
	if err != nil {
		return err
	}
	e.conn = conn

	// TODO(tobias.urdin): Dont use default config, use options to determine
	// what ceph config file and cephx user should be used
	if err := e.conn.ReadDefaultConfigFile(); err != nil {
		panic(err)
	}

	if err := e.conn.Connect(); err != nil {
		return err
	}

	return nil
}

// Close exporter
func (e *Exporter) Close() {
	e.logger.Info("close exporter")

	if e.conn != nil {
		e.conn.Shutdown()
	}
}

// Handle error message version 1
func (e *Exporter) handleErrorV1(ctx *message.Context) error {
	return errors.New("not implemented")
}

// Handle list pool message version 1
func (e *Exporter) handleListPoolRequestV1(ctx *message.Context) error {
	msg := ctx.Message()

	var listMsg message.ListPoolRequestV1
	if err := msg.Unmarshal(&listMsg); err != nil {
		return err
	}

	ctx.Logger().Info("listpool request message", zap.Any("msg", listMsg))

	ioctx, err := e.conn.OpenIOContext(listMsg.Pool)
	if err != nil {
		return err
	}
	defer ioctx.Destroy()

	names, err := rbd.GetImageNames(ioctx)
	if err != nil {
		return err
	}

	ctx.Logger().Info("sending list pool response with names", zap.Any("names", names))

	resp := message.ListPoolResponseV1{
		Names: names,
	}

	return ctx.Send(&resp)
}

// Handle list snapshots message version 1
func (e *Exporter) handleListSnapshotsRequestV1(ctx *message.Context) error {
	msg := ctx.Message()

	var listMsg message.ListSnapshotsRequestV1
	if err := msg.Unmarshal(&listMsg); err != nil {
		return err
	}

	ctx.Logger().Info("listsnapshots request message", zap.Any("msg", listMsg))

	ioctx, err := e.conn.OpenIOContext(listMsg.Pool)
	if err != nil {
		return err
	}
	defer ioctx.Destroy()

	image, err := rbd.OpenImageReadOnly(ioctx, listMsg.Image, rbd.NoSnapshot)
	if err != nil {
		return err
	}
	defer image.Close()

	snaps, err := image.GetSnapshotNames()
	if err != nil {
		return err
	}

	result := make([]string, len(snaps))

	for _, snap := range snaps {
		result = append(result, snap.Name)
	}

	resp := message.ListSnapshotsResponseV1{
		Pool: listMsg.Pool,
		Image: listMsg.Image,
		Snapshots: result,
	}

	return ctx.Send(&resp)
}


// Handle export request version 1
func (e *Exporter) handleExportRequestV1(ctx *message.Context) error {
	msg := ctx.Message()

	var req message.ExportRequestV1
	if err := msg.Unmarshal(&req); err != nil {
		return err
	}

	ctx.Logger().Info("incoming export request", zap.Any("msg", req))

	w := newChunkedWriter(ctx)
	if err := exportDiff(&req, w); err != nil {
		return err
	}

	resp := message.ExportResponseV1{
		Pool: req.Pool,
		Image: req.Image,
		Snapshot: req.Snapshot,
	}

	if err := ctx.Send(&resp); err != nil {
		return err
	}

	return nil
}

// Generate TLS config
func (e Exporter) generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE",
		Bytes: certDER,
	})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"snapback"},
	}, nil
}

// Handle a new connection
func (e *Exporter) onConn(ctx context.Context, conn quic.Connection) {
	addr := conn.RemoteAddr()

	connLogger := e.logger.With(
		zap.String("connection", addr.String()))

	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			e.logger.Error(err.Error())
			break
		}

		go e.onStream(connLogger, stream)
	}
}

// Handle a new stream on a connection
func (e *Exporter) onStream(connLogger *zap.Logger, stream quic.Stream) {
	logger := connLogger.With(
		zap.Any("stream", stream.StreamID()))

	defer func() {
		stream.Close()
		logger.Info("stream closed")
	}()

	logger.Info("new stream opened")

	if err := e.handler.Run(logger, stream); err != nil {
		logger.Error("handler error")
		logger.Error(err.Error())
	}
}

// Run the exporter
func (e *Exporter) Run() error {
	tlsConfig, err := e.generateTLSConfig()
	if err != nil {
		return err
	}

	listener, err := quic.ListenAddr(exporterAddr, tlsConfig, nil)
	if err != nil {
		return err
	}
	defer listener.Close()

	e.logger.Info("listening on address", zap.String("address", exporterAddr))

	sigC := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			conn, err := listener.Accept(ctx)
			if err != nil {
				e.logger.Error(err.Error())
				break
			}

			go e.onConn(ctx, conn)
		}
	}()

	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	e.logger.Info("signal captured, exiting...")

	cancel()

	return nil
}
