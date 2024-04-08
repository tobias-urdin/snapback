package message

import (
        "github.com/quic-go/quic-go"
)

// Send a message
func Send(stream quic.Stream, m MessageInterface) error {
        encoded, err := m.Marshal()
        if err != nil {
                return err
        }

        // NOTE(tobias.urdin): Our protocol handles newlines as the end
        // of a message and the start of new messages so we need to end
        // all message with a newline.
        encoded = append(encoded, "\n"...)

        if _, err := stream.Write(encoded); err != nil {
                return err
        }

        return nil
}

