package v1

const (
	Version uint8 = 1

	FrameAuth       uint8 = 0x01
	FrameOpenStream uint8 = 0x02
	FrameData       uint8 = 0x03
	FrameClose      uint8 = 0x04
	FrameHeartbeat  uint8 = 0x05

	MaxFrameSize = 16 * 1024 * 1024 // 16MB
)
