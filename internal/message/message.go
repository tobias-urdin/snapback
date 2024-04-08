package message

import (
	"github.com/fxamacker/cbor/v2"
)

const (
	// The message type number for error
	ErrorType = 1

	// The message type number for list pool request
	ListPoolRequestType = 2

	// The message type number for list pool response
	ListPoolResponseType = 3

	// The message type number for list snapshots request
	ListSnapshotsRequestType = 4

	// The message type number for list snapshots response
	ListSnapshotsResponseType = 5

	// The message type number for export request
	ExportRequestType = 6

	// The message type number for export response
	ExportResponseType = 7

	// The message type number for export chunk
	ExportChunkType = 8
)

// The message Type
type MessageType int

// The message version
type MessageVersion int

// The header that is the first part of all messages
type MessageHeader struct {
	// The type of the message
	Type MessageType `cbor:"1,keyasint"`

	// The version of the message type
	Version MessageVersion `cbor:"2,keyasint"`
}

// The message interface
type MessageInterface interface {
	// The type of the message
	Type() MessageType

	// The version of the message
	Version() MessageVersion

	// Marshal the message
	Marshal() ([]byte, error)
}

// The message that contains the header and data
type Message struct {
	// The message header
	Header MessageHeader `cbor:"1,keyasint"`

	// The message data
	Data cbor.RawMessage `cbor:"2,keyasint"`
}

// Unmarshal the message data
func (m *Message) Unmarshal(v interface{}) error {
	return cbor.Unmarshal(m.Data, v)
}

// Unmarshal data into a Message
func unmarshal(data []byte, msg *Message) error {
	if err := cbor.Unmarshal(data, msg); err != nil {
		return err
	}

	return nil
}

// The error message that returns an error
type ErrorMessage struct {
	// The error code
	ErrorCode int `cbor:"1,keyasint"`
}

// The error message type
func (e *ErrorMessage) Type() MessageType {
	return ErrorType
}

// The error message version
func (e *ErrorMessage) Version() MessageVersion {
	return MessageVersion(1)
}

// Marshal error version 1 to message
func (e *ErrorMessage) Marshal() ([]byte, error) {
	data, err := cbor.Marshal(e)
	if err != nil {
		return nil, err
	}

	msg := Message{
		Header: MessageHeader{
			Type: e.Type(),
			Version: e.Version(),
		},
		Data: data,
	}

	res, err := cbor.Marshal(&msg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// The list pool request version 1
type ListPoolRequestV1 struct {
	// The pool name we want to list images on
	Pool string `cbor:"1,keyasint"`
}

// The list pool request message type
func (l *ListPoolRequestV1) Type() MessageType {
	return ListPoolRequestType
}

// The list pool request message version
func (l *ListPoolRequestV1) Version() MessageVersion {
	return MessageVersion(1)
}

// Marshal list pool request version 1 to message
func (l *ListPoolRequestV1) Marshal() ([]byte, error) {
	data, err := cbor.Marshal(l)
	if err != nil {
		return nil, err
	}

	msg := Message{
		Header: MessageHeader{
			Type: l.Type(),
			Version: l.Version(),
		},
		Data: data,
	}

	res, err := cbor.Marshal(&msg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// The list pool response version 1
type ListPoolResponseV1 struct {
	// The image names
	Names []string `cbor:"1,keyasint"`
}

// The list pool response type
func (l *ListPoolResponseV1) Type() MessageType {
	return ListPoolResponseType
}

// The list pool response version
func (l *ListPoolResponseV1) Version() MessageVersion {
	return MessageVersion(1)
}

// Marshal list pool response version 1 to message
func (l *ListPoolResponseV1) Marshal() ([]byte, error) {
	data, err := cbor.Marshal(l)
	if err != nil {
		return nil, err
	}

	msg := Message{
		Header: MessageHeader{
			Type: l.Type(),
			Version: l.Version(),
		},
		Data: data,
	}

	res, err := cbor.Marshal(&msg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// The list snapshots request version 1
type ListSnapshotsRequestV1 struct {
	// The pool name
	Pool string `cbor:"1,keyasint"`

	// The image name
	Image string `cbor:"2,keyasint"`
}

// The list snapshot request type
func (l *ListSnapshotsRequestV1) Type() MessageType {
	return ListSnapshotsRequestType
}

// The list snapshot request version
func (l *ListSnapshotsRequestV1) Version() MessageVersion {
	return MessageVersion(1)
}

// Marshal list snapshots request version 1 to message
func (l *ListSnapshotsRequestV1) Marshal() ([]byte, error) {
	data, err := cbor.Marshal(l)
	if err != nil {
		return nil, err
	}

	msg := Message{
		Header: MessageHeader{
			Type: l.Type(),
			Version: l.Version(),
		},
		Data: data,
	}

	res, err := cbor.Marshal(&msg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// The list snapshots response version 1
type ListSnapshotsResponseV1 struct {
	// The pool name
	Pool string `cbor:"1,keyasint"`

	// The image name
	Image string `cbor:"2,keyasint"`

	// The snapshots
	Snapshots []string `cbor:"3,keyasint"`
}

// The list snapshots response type
func (l *ListSnapshotsResponseV1) Type() MessageType {
	return ListSnapshotsResponseType
}

// The list snapshots response version
func (l *ListSnapshotsResponseV1) Version() MessageVersion {
	return MessageVersion(1)
}

// Marshal list snapshots response version 1 to message
func (l *ListSnapshotsResponseV1) Marshal() ([]byte, error) {
	data, err := cbor.Marshal(l)
	if err != nil {
		return nil, err
	}

	msg := Message{
		Header: MessageHeader{
			Type: l.Type(),
			Version: l.Version(),
		},
		Data: data,
	}

	res, err := cbor.Marshal(&msg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// The export request version 1
type ExportRequestV1 struct {
	// The pool name
	Pool string `cbor:"1,keyasint"`

	// The image name we want to export from the pool
	Image string `cbor:"2,keyasint"`

	// The snapshot on the image we want to export
	Snapshot string `cbor:"3,keyasint"`
}

// The export request type
func (e *ExportRequestV1) Type() MessageType {
	return ExportRequestType
}

// The export message version
func (e *ExportRequestV1) Version() MessageVersion {
	return MessageVersion(1)
}

// Marshal the export request version 1 to message
func (e *ExportRequestV1) Marshal() ([]byte, error) {
	data, err := cbor.Marshal(e)
	if err != nil {
		return nil, err
	}

	msg := Message{
		Header: MessageHeader{
			Type: e.Type(),
			Version: e.Version(),
		},
		Data: data,
	}

	res, err := cbor.Marshal(&msg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// The export response version 1
type ExportResponseV1 struct {
	// The pool name
	Pool string `cbor:"1,keyasint"`

	// The image name
	Image string `cbor:"2,keyasint"`

	// The snapshot name
	Snapshot string `cbor:"3,keyasint"`
}

// The export response type
func (e *ExportResponseV1) Type() MessageType {
	return ExportResponseType
}

// The export response version
func (e *ExportResponseV1) Version() MessageVersion {
	return MessageVersion(1)
}

// Marshal the export response version 1 to message
func (e *ExportResponseV1) Marshal() ([]byte, error) {
	data, err := cbor.Marshal(e)
	if err != nil {
		return nil, err
	}

	msg := Message{
		Header: MessageHeader{
			Type: e.Type(),
			Version: e.Version(),
		},
		Data: data,
	}

	res, err := cbor.Marshal(&msg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// The export chunk version 1
type ExportChunkV1 struct {
	// The bytes payload for this chunk
	Payload []byte `cbor:"2,keyasint"`

	// The CRC32 of the payload for this chunk
	PayloadCRC uint32 `cbor:"3,keyasint"`
}

// The export chunk type
func (e *ExportChunkV1) Type() MessageType {
	return ExportChunkType
}

// The export chunk version
func (e *ExportChunkV1) Version() MessageVersion {
	return MessageVersion(1)
}

// Marshal the export chunk version 1 to message
func (e *ExportChunkV1) Marshal() ([]byte, error) {
	data, err := cbor.Marshal(e)
	if err != nil {
		return nil, err
	}

	msg := Message{
		Header: MessageHeader{
			Type: e.Type(),
			Version: e.Version(),
		},
		Data: data,
	}

	res, err := cbor.Marshal(&msg)
	if err != nil {
		return nil, err
	}

	return res, nil
}
