package v1

import (
	"encoding/binary"
	"io"
)

type Frame struct {
	Version uint8
	Type    uint8
	Flags   uint8
	Payload []byte
}

func Encode(w io.Writer, f *Frame) error {
	if f.Version != Version {
		return NewError(ErrCodeInvalidVersion, "invalid protocol version")
	}

	length := uint32(3 + len(f.Payload)) // v + type + flags
	if length > MaxFrameSize {
		return NewError(ErrCodeFrameTooLarge, "frame too large")
	}

	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return err
	}

	header := []byte{f.Version, f.Type, f.Flags}
	if _, err := w.Write(header); err != nil {
		return err
	}

	if len(f.Payload) > 0 {
		_, err := w.Write(f.Payload)
		return err
	}

	return nil
}

func Decode(r io.Reader) (*Frame, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	if length < 3 || length > MaxFrameSize {
		return nil, NewError(ErrCodeFrameTooLarge, "frame too large")
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	if buf[0] != Version {
		return nil, NewError(ErrCodeInvalidVersion, "invalid protocol version")
	}

	return &Frame{
		Version: buf[0],
		Type:    buf[1],
		Flags:   buf[2],
		Payload: buf[3:],
	}, nil
}
