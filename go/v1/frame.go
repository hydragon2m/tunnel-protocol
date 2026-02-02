package v1

import (
	"encoding/binary"
	"io"
)

type Frame struct {
	Version  uint8
	Type     uint8
	Flags    uint8
	StreamID uint32
	Payload  []byte
}

func Encode(w io.Writer, f *Frame) error {
	if f.Version != Version {
		return NewError(ErrCodeInvalidVersion, "invalid protocol version")
	}

	// Header: Magic(2) + Version(1) + Type(1) + Flags(1) + StreamID(4) = HeaderSize bytes
	// Length = số byte đọc tiếp theo, KHÔNG bao gồm chính length field
	length := uint32(HeaderSize + len(f.Payload))
	if length > MaxFrameSize {
		return NewError(ErrCodeFrameTooLarge, "frame too large")
	}

	// Use a logical buffer size: 4 (length) + length (rest of frame)
	totalSize := 4 + int(length)
	buf := make([]byte, totalSize)

	// 1. Length (4 bytes)
	binary.BigEndian.PutUint32(buf[0:4], length)

	// 2. Magic (2 bytes)
	buf[4] = Magic0
	buf[5] = Magic1

	// 3. Header fields
	buf[6] = f.Version
	buf[7] = f.Type
	buf[8] = f.Flags

	// 4. StreamID (4 bytes)
	binary.BigEndian.PutUint32(buf[9:13], f.StreamID)

	// 5. Payload
	if len(f.Payload) > 0 {
		copy(buf[13:], f.Payload)
	}

	// Atomic Write: Write everything in one syscall
	// Note: w.Write might still return partial write if underlying socket buffer is full,
	// but for small frames this is much better than 5 separate syscalls.
	// For full atomicity on network, we rely on the OS TCP stack coalescing, but sending
	// one block gives it the best chance.
	_, err := w.Write(buf)
	return err
}

// Decode reads a frame from reader
// NOTE: Caller MUST ensure r has a ReadDeadline set or is cancellable to prevent infinite blocking.
// If r is a net.Conn, call SetReadDeadline() before passing it here.
func Decode(r io.Reader) (*Frame, error) {
	// 1. Read Length
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	// 2. Validate Size
	if length < HeaderSize || length > MaxFrameSize {
		return nil, NewError(ErrCodeBadFrame, "invalid frame size")
	}

	// 3. Allocate Buffer (Standard Decode uses new allocation)
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	// 4. Parse
	return ParseFrame(buf)
}

// ParseFrame parses a frame from a byte slice containing: Magic + Header + StreamID + Payload
// The slice must NOT include the length field.
func ParseFrame(buf []byte) (*Frame, error) {
	if len(buf) < HeaderSize {
		return nil, NewError(ErrCodeBadFrame, "buffer too small")
	}

	// Validate magic marker
	if buf[0] != Magic0 || buf[1] != Magic1 {
		return nil, NewError(ErrCodeBadFrame, "invalid magic marker")
	}

	// Validate version
	if buf[2] != Version {
		return nil, NewError(ErrCodeInvalidVersion, "invalid protocol version")
	}

	// Validate frame type
	frameType := buf[3]
	if !IsValidFrameType(frameType) {
		return nil, NewError(ErrCodeBadFrame, "invalid frame type")
	}

	// StreamID
	streamID := binary.BigEndian.Uint32(buf[5:9])

	// Copy payload to separate slice?
	// If call passed a pooled buffer, we might want to copy if the frame outlives the buffer return.
	// But `ParseFrame` assumes ownership or that caller handles it.
	// For safety in standard `Decode`, `make` was used so it's safe.
	// Here `buf` is a slice. The returned Frame.Payload will slice into it.

	return &Frame{
		Version:  buf[2],
		Type:     frameType,
		Flags:    buf[4],
		StreamID: streamID,
		Payload:  buf[HeaderSize:],
	}, nil
}

// ReadFrameLength reads the 4-byte length prefix
func ReadFrameLength(r io.Reader) (uint32, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return 0, err
	}
	return length, nil
}

// IsControlFrame kiểm tra frame có phải control frame không
// Control frame: StreamID == 0 (Auth, Heartbeat, Error global)
func (f *Frame) IsControlFrame() bool {
	return f.StreamID == StreamIDControl
}

// IsDataStream kiểm tra frame có phải data stream không
// Data stream: StreamID > 0 (OpenStream, Data, Close)
func (f *Frame) IsDataStream() bool {
	return f.StreamID > StreamIDControl
}

// HasFlag kiểm tra frame có flag cụ thể không
func (f *Frame) HasFlag(flag uint8) bool {
	return (f.Flags & flag) != 0
}

// IsEndStream kiểm tra frame có flag EndStream không
func (f *Frame) IsEndStream() bool {
	return f.HasFlag(FlagEndStream)
}

// IsError kiểm tra frame có flag Error không
func (f *Frame) IsError() bool {
	return f.HasFlag(FlagError)
}

// IsAck kiểm tra frame có flag Ack không
func (f *Frame) IsAck() bool {
	return f.HasFlag(FlagAck)
}

// IsValidFrameType kiểm tra frame type có hợp lệ không
func IsValidFrameType(frameType uint8) bool {
	switch frameType {
	case FrameAuth, FrameOpenStream, FrameData, FrameClose, FrameHeartbeat:
		return true
	default:
		return false
	}
}
