package v1

import (
	"bytes"
	"testing"
)

func TestFrameEncodeDecode(t *testing.T) {
	tests := []struct {
		name  string
		frame *Frame
	}{
		{
			name: "Control frame - Auth",
			frame: &Frame{
				Version:  Version,
				Type:     FrameAuth,
				Flags:    FlagNone,
				StreamID: StreamIDControl,
				Payload:  []byte(`{"token":"test-token"}`),
			},
		},
		{
			name: "Data frame - OpenStream",
			frame: &Frame{
				Version:  Version,
				Type:     FrameOpenStream,
				Flags:    FlagNone,
				StreamID: 1,
				Payload:  []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"),
			},
		},
		{
			name: "Data frame - With EndStream flag",
			frame: &Frame{
				Version:  Version,
				Type:     FrameData,
				Flags:    FlagEndStream,
				StreamID: 1,
				Payload:  []byte("Response body"),
			},
		},
		{
			name: "Data frame - With Error flag",
			frame: &Frame{
				Version:  Version,
				Type:     FrameData,
				Flags:    FlagError,
				StreamID: 1,
				Payload:  []byte("Error message"),
			},
		},
		{
			name: "Heartbeat frame",
			frame: &Frame{
				Version:  Version,
				Type:     FrameHeartbeat,
				Flags:    FlagNone,
				StreamID: StreamIDControl,
				Payload:  nil,
			},
		},
		{
			name: "Empty payload",
			frame: &Frame{
				Version:  Version,
				Type:     FrameClose,
				Flags:    FlagNone,
				StreamID: 1,
				Payload:  nil,
			},
		},
		{
			name: "Large payload",
			frame: &Frame{
				Version:  Version,
				Type:     FrameData,
				Flags:    FlagNone,
				StreamID: 1,
				Payload:  make([]byte, 1024*1024), // 1MB
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode frame
			var buf bytes.Buffer
			if err := Encode(&buf, tt.frame); err != nil {
				t.Fatalf("Encode failed: %v", err)
			}
			
			// Decode frame
			decoded, err := Decode(&buf)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}
			
			// Verify fields
			if decoded.Version != tt.frame.Version {
				t.Errorf("Version mismatch: got %d, want %d", decoded.Version, tt.frame.Version)
			}
			
			if decoded.Type != tt.frame.Type {
				t.Errorf("Type mismatch: got %d, want %d", decoded.Type, tt.frame.Type)
			}
			
			if decoded.Flags != tt.frame.Flags {
				t.Errorf("Flags mismatch: got %d, want %d", decoded.Flags, tt.frame.Flags)
			}
			
			if decoded.StreamID != tt.frame.StreamID {
				t.Errorf("StreamID mismatch: got %d, want %d", decoded.StreamID, tt.frame.StreamID)
			}
			
			if !bytes.Equal(decoded.Payload, tt.frame.Payload) {
				t.Errorf("Payload mismatch: got %d bytes, want %d bytes", len(decoded.Payload), len(tt.frame.Payload))
			}
		})
	}
}

func TestFrameMagicMarker(t *testing.T) {
	// Create valid frame
	frame := &Frame{
		Version:  Version,
		Type:     FrameAuth,
		Flags:    FlagNone,
		StreamID: StreamIDControl,
		Payload:  []byte("test"),
	}
	
	var buf bytes.Buffer
	if err := Encode(&buf, frame); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	
	// Verify magic marker
	data := buf.Bytes()
	if data[4] != Magic0 || data[5] != Magic1 {
		t.Errorf("Magic marker not found at correct position")
	}
	
	// Corrupt magic marker
	data[4] = 0xFF
	data[5] = 0xFF
	
	// Try to decode - should fail
	_, err := Decode(bytes.NewReader(data))
	if err == nil {
		t.Error("Expected error for invalid magic marker, got nil")
	}
}

func TestFrameVersionValidation(t *testing.T) {
	// Create frame with invalid version
	frame := &Frame{
		Version:  0xFF, // Invalid version
		Type:     FrameAuth,
		Flags:    FlagNone,
		StreamID: StreamIDControl,
		Payload:  []byte("test"),
	}
	
	var buf bytes.Buffer
	// Encode should fail for invalid version
	if err := Encode(&buf, frame); err == nil {
		t.Error("Expected Encode to fail for invalid version, got nil")
	}
	
	// Create valid frame and manually corrupt version
	validFrame := &Frame{
		Version:  Version,
		Type:     FrameAuth,
		Flags:    FlagNone,
		StreamID: StreamIDControl,
		Payload:  []byte("test"),
	}
	
	buf.Reset()
	if err := Encode(&buf, validFrame); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	
	// Corrupt version byte
	data := buf.Bytes()
	data[6] = 0xFF // Corrupt version
	
	// Try to decode - should fail
	_, err := Decode(bytes.NewReader(data))
	if err == nil {
		t.Error("Expected error for invalid version, got nil")
	}
}

func TestFrameStreamID(t *testing.T) {
	tests := []struct {
		name     string
		streamID uint32
		isControl bool
	}{
		{
			name:      "Control frame",
			streamID:  StreamIDControl,
			isControl: true,
		},
		{
			name:      "Data stream",
			streamID:  1,
			isControl: false,
		},
		{
			name:      "Large stream ID",
			streamID:  0xFFFFFFFF,
			isControl: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := &Frame{
				Version:  Version,
				Type:     FrameData,
				Flags:    FlagNone,
				StreamID: tt.streamID,
				Payload:  []byte("test"),
			}
			
			if frame.IsControlFrame() != tt.isControl {
				t.Errorf("IsControlFrame() = %v, want %v", frame.IsControlFrame(), tt.isControl)
			}
			
			if frame.IsDataStream() == tt.isControl {
				t.Errorf("IsDataStream() = %v, want %v", frame.IsDataStream(), !tt.isControl)
			}
		})
	}
}

func TestFrameFlags(t *testing.T) {
	frame := &Frame{
		Version:  Version,
		Type:     FrameData,
		Flags:    FlagEndStream | FlagAck,
		StreamID: 1,
		Payload:  []byte("test"),
	}
	
	if !frame.HasFlag(FlagEndStream) {
		t.Error("Expected FlagEndStream to be set")
	}
	
	if !frame.HasFlag(FlagAck) {
		t.Error("Expected FlagAck to be set")
	}
	
	if frame.HasFlag(FlagError) {
		t.Error("Expected FlagError to not be set")
	}
	
	if !frame.IsEndStream() {
		t.Error("Expected IsEndStream() to return true")
	}
	
	if !frame.IsAck() {
		t.Error("Expected IsAck() to return true")
	}
	
	if frame.IsError() {
		t.Error("Expected IsError() to return false")
	}
}

func TestFrameMaxSize(t *testing.T) {
	// Create frame with payload exceeding MaxFrameSize
	largePayload := make([]byte, MaxFrameSize+1)
	
	frame := &Frame{
		Version:  Version,
		Type:     FrameData,
		Flags:    FlagNone,
		StreamID: 1,
		Payload:  largePayload,
	}
	
	var buf bytes.Buffer
	err := Encode(&buf, frame)
	if err == nil {
		t.Error("Expected error for frame exceeding MaxFrameSize, got nil")
	}
}

func TestFrameDecodeIncomplete(t *testing.T) {
	// Test with length that says we need more data than available
	// Create a frame with length > available data
	incomplete := []byte{
		0x00, 0x00, 0x00, 0x20, // Length = 32 (but we only have 9 bytes)
		Magic0, Magic1,          // Magic
		Version,                 // Version
		FrameAuth,               // Type
		FlagNone,                // Flags
		0x00, 0x00, 0x00, 0x00, // StreamID
		// Missing rest of frame (should have 23 more bytes)
	}
	
	_, err := Decode(bytes.NewReader(incomplete))
	if err == nil {
		t.Error("Expected error for incomplete frame, got nil")
	}
	
	// Test with correct length but truncated data
	correctFrame := &Frame{
		Version:  Version,
		Type:     FrameAuth,
		Flags:    FlagNone,
		StreamID: StreamIDControl,
		Payload:  []byte("test"),
	}
	
	var buf bytes.Buffer
	if err := Encode(&buf, correctFrame); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	
	// Truncate buffer - remove some bytes from payload
	data := buf.Bytes()
	truncated := data[:len(data)-2] // Remove last 2 bytes
	
	_, err = Decode(bytes.NewReader(truncated))
	if err == nil {
		t.Error("Expected error for truncated frame, got nil")
	}
}

func TestFrameDecodeEOF(t *testing.T) {
	// Empty reader
	_, err := Decode(bytes.NewReader(nil))
	if err == nil {
		t.Error("Expected error for EOF, got nil")
	}
	
	// Reader with only length
	_, err = Decode(bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x09}))
	if err == nil {
		t.Error("Expected error for incomplete frame, got nil")
	}
}

func BenchmarkFrameEncode(b *testing.B) {
	frame := &Frame{
		Version:  Version,
		Type:     FrameData,
		Flags:    FlagNone,
		StreamID: 1,
		Payload:  make([]byte, 1024),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = Encode(&buf, frame)
	}
}

func BenchmarkFrameDecode(b *testing.B) {
	frame := &Frame{
		Version:  Version,
		Type:     FrameData,
		Flags:    FlagNone,
		StreamID: 1,
		Payload:  make([]byte, 1024),
	}
	
	var buf bytes.Buffer
	_ = Encode(&buf, frame)
	data := buf.Bytes()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decode(bytes.NewReader(data))
	}
}

