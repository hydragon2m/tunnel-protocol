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

	// NOTE: Hiện tại encode từng phần (length, magic, header, streamID, payload).
	// Nếu connection bị close giữa chừng → frame có thể bị cắt.
	// Phase sau: nên dùng buffer tạm để atomic write:
	//   buf := make([]byte, 4+length)
	//   // build full frame vào buf
	//   w.Write(buf) // atomic

	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return err
	}

	// Write magic marker: "RT" (0x52 0x54)
	if _, err := w.Write([]byte{Magic0, Magic1}); err != nil {
		return err
	}

	// Write header: Version, Type, Flags
	header := []byte{f.Version, f.Type, f.Flags}
	if _, err := w.Write(header); err != nil {
		return err
	}

	// Write StreamID (4 bytes, big-endian)
	if err := binary.Write(w, binary.BigEndian, f.StreamID); err != nil {
		return err
	}

	// Write Payload
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

	// Minimum frame size: Magic(2) + Version(1) + Type(1) + Flags(1) + StreamID(4) = HeaderSize bytes
	if length < HeaderSize || length > MaxFrameSize {
		return nil, NewError(ErrCodeBadFrame, "invalid frame size")
	}

	// NOTE: io.ReadFull có thể block vô hạn nếu agent chết giữa chừng.
	// Khi dùng thực tế, reader phải:
	// - Gắn deadline: conn.SetReadDeadline(time.Now().Add(timeout))
	// - Hoặc chạy trong goroutine có context với cancel
	//
	// NOTE: Hiện tại allocate buffer mỗi frame. Khi traffic lớn (hàng nghìn stream):
	// - GC pressure tăng
	// - Latency tăng
	// Phase sau: nên dùng sync.Pool hoặc reuse buffer per-connection
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	// Validate magic marker: "RT" (0x52 0x54)
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

	// Parse StreamID (4 bytes, big-endian, offset 5-8)
	streamID := binary.BigEndian.Uint32(buf[5:9])

	// StreamID validation:
	// - StreamID == 0: control frame (Auth, Heartbeat, Error global) - OK
	// - StreamID > 0: data stream - OK
	// - StreamID validation theo frame type sẽ được xử lý ở layer trên

	return &Frame{
		Version:  buf[2],
		Type:     frameType,
		Flags:    buf[4],
		StreamID: streamID,
		Payload:  buf[HeaderSize:],
	}, nil
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
