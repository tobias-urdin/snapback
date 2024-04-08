package exporter

import (
	"hash"
	"hash/crc32"

	"github.com/tobias-urdin/snapback/internal/message"

	"go.uber.org/zap"
)

func newChunkedWriter(ctx *message.Context) *chunkedWriter {
	return &chunkedWriter{
		ctx: ctx,
		h:   crc32.New(crc32.MakeTable(crc32.Castagnoli)),
	}
}

type chunkedWriter struct {
	ctx *message.Context
	h hash.Hash32
}

func (c *chunkedWriter) Write(p []byte) (int, error) {
	c.h.Reset()
	c.h.Write(p)

	sum := c.h.Sum32()
	c.ctx.Logger().Info("sending export chunk", zap.Any("crc", sum))

	chunk := message.ExportChunkV1{
		Payload: p,
		PayloadCRC: sum,
	}

	if err := c.ctx.Send(&chunk); err != nil {
		return 0, err
	}

	length := len(p)
	c.ctx.Logger().Info("going to next", zap.Any("len", length))

	return length, nil
}
