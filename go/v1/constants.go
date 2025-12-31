package v1

const (
	Version uint8 = 1

	// Magic marker: "RT" (Reverse Tunnel) - 0x52 0x54
	// Giúp detect protocol boundary, tránh decode sai khi lệch byte
	Magic0 uint8 = 0x52
	Magic1 uint8 = 0x54

	FrameAuth       uint8 = 0x01
	FrameOpenStream uint8 = 0x02
	FrameData       uint8 = 0x03
	FrameClose      uint8 = 0x04
	FrameHeartbeat  uint8 = 0x05

	// HeaderSize: Magic(2) + Version(1) + Type(1) + Flags(1) + StreamID(4) = 9 bytes
	HeaderSize   = 9
	MaxFrameSize = 16 * 1024 * 1024 // 16MB

	// StreamID quy ước:
	// StreamID == 0 → control frame (Auth, Heartbeat, Error global)
	// StreamID > 0  → data stream (OpenStream, Data, Close)
	StreamIDControl = uint32(0)

	// Frame flags - semantics cứng:
	FlagNone      uint8 = 0      // Không có flag
	FlagEndStream uint8 = 1 << 0 // FIN - stream kết thúc (HTTP response done, TCP half-close, WebSocket close)
	FlagAck       uint8 = 1 << 1 // Acknowledgment - xác nhận đã nhận
	FlagError     uint8 = 1 << 2 // Error - frame chứa lỗi

	// Flag rules:
	// - DATA + EndStream → đóng phía gửi (sender half-close)
	// - ERROR → payload phải chứa error code + message
	// - EndStream chỉ gửi 1 lần per stream
)
