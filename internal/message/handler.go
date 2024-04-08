package message

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"
	"github.com/quic-go/quic-go"
)

// The message handler func
type MessageHandlerFunc func(*Context) error

// The message handlers type
type MessageHandlers map[MessageType]map[MessageVersion]MessageHandlerFunc

// MessageHandler is used to handle incoming messages
type MessageHandler struct {
	// Logger
	logger *zap.Logger

	// Handler mappings
	handlers MessageHandlers
}

// Returns a new MessageHandler
func NewHandler(logger *zap.Logger) *MessageHandler {
	return &MessageHandler{
		logger: logger,
		handlers: make(MessageHandlers, 0),
	}
}

// Add a message handler
func (mh *MessageHandler) AddHandler(msgType MessageType, msgVersion MessageVersion, f MessageHandlerFunc) {
	if _, ok := mh.handlers[msgType]; !ok {
		mh.handlers[msgType] = make(map[MessageVersion]MessageHandlerFunc, 0)
	}

	mh.handlers[msgType][msgVersion] = f
}

// Get a message handler
func (mh *MessageHandler) getHandler(msg *Message) MessageHandlerFunc {
	versionHandlers, ok := mh.handlers[msg.Header.Type]
	if !ok {
		mh.logger.Error("no handler for message type", zap.Any("type", msg.Header.Type))
		return nil
	}

	handlerFunc, ok := versionHandlers[msg.Header.Version]
	if !ok {
		mh.logger.Error("no handler for message version of type", zap.Any("type", msg.Header.Type), zap.Any("version", msg.Header.Version))
		return nil
	}

	return handlerFunc
}

// Read the stream and return the message
func (mh *MessageHandler) read(r *bufio.Reader, msg *Message) error {
	// Read until newline which is what we use in our protocol
	// to tell the end of the data for this message
	buf, _, err := r.ReadLine()
	if err != nil {
		return err
	}

	// Unmarshal the raw message into a Message
	if err := unmarshal(buf, msg); err != nil {
		return err
	}

	return nil
}

// Read the stream and return the message
func (mh *MessageHandler) Read(stream quic.Stream, msg *Message) error {
	r := bufio.NewReader(stream)

	return mh.read(r, msg)
}

// Read chunks from the stream
func (mh *MessageHandler) ReadChunks(stream quic.Stream, cb func(*Message) error) (*Message, error) {
	r := bufio.NewReader(stream)

	for {
		var msg Message
		if err := mh.read(r, &msg); err != nil {
			return nil, err
		}

		if msg.Header.Type != ExportChunkType {
			if msg.Header.Type != ExportResponseType {
				return nil, fmt.Errorf("chunk stream did not end with a response, type: %d", msg.Header.Type)
			}

			return &msg, nil
		}

		if err := cb(&msg); err != nil {
			return nil, err
		}
	}

	return nil, errors.New("read chunks never finished")
}

// This runs the mssage handler that reads messages from the stream and gives the
// messages to the handler that is registered for the message.
func (mh *MessageHandler) Run(logger *zap.Logger, stream quic.Stream) error {
	r := bufio.NewReader(stream)

        for {
		var msg Message
		if err := mh.read(r, &msg); err != nil {
			if err == io.EOF {
				return nil
			}

			logger.Error("failed to read message", zap.String("error", err.Error()))
			// TODO(tobias.urdin): Send error
			return err
		}

		ctx := Context{
			logger: logger,
			message: &msg,
			stream: stream,
		}

		handlerFunc := mh.getHandler(&msg)
		if handlerFunc == nil {
			// TODO(tobias.urdin): Send error to sender, could not handle message
			return errors.New("could not handle message")
		}

		if err := handlerFunc(&ctx); err != nil {
			// TODO(tobias.urdin): Send error to sender, failed to handle message
			logger.Error("failed to handle message", zap.String("error", err.Error()))
			return err
		}
        }

	return nil
}
