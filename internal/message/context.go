package message

import (
	"go.uber.org/zap"
	"github.com/quic-go/quic-go"
)

// Message context
type Context struct {
	logger *zap.Logger
	message *Message
	stream quic.Stream
}

// Returns logger
func (c *Context) Logger() *zap.Logger {
	return c.logger
}

// Returns message
func (c *Context) Message() *Message {
	return c.message
}

// Returns stream
func (c *Context) Stream() quic.Stream {
	return c.stream
}

// Send a message
func (c *Context) Send(m MessageInterface) error {
	return Send(c.stream, m)
}
