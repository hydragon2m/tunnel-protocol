package v1

import (
	"bytes"
	"testing"
)

// BenchmarkFrameEncode benchmarks frame encoding
func BenchmarkFrameEncode(b *testing.B) {
	frame := &Frame{
		Version:  Version,
		Type:     FrameData,
		Flags:    FlagNone,
		StreamID: 123,
		Payload:  make([]byte, 1024), // 1KB payload
	}

	buf := new(bytes.Buffer)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := Encode(buf, frame); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFrameDecode benchmarks frame decoding
func BenchmarkFrameDecode(b *testing.B) {
	frame := &Frame{
		Version:  Version,
		Type:     FrameData,
		Flags:    FlagNone,
		StreamID: 123,
		Payload:  make([]byte, 1024),
	}

	buf := new(bytes.Buffer)
	if err := Encode(buf, frame); err != nil {
		b.Fatal(err)
	}
	encodedData := buf.Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(encodedData)
		if _, err := Decode(reader); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFrameEncodeDecode benchmarks full encode/decode cycle
func BenchmarkFrameEncodeDecode(b *testing.B) {
	frame := &Frame{
		Version:  Version,
		Type:     FrameData,
		Flags:    FlagNone,
		StreamID: 123,
		Payload:  make([]byte, 1024),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		if err := Encode(buf, frame); err != nil {
			b.Fatal(err)
		}

		if _, err := Decode(buf); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFrameEncode_SmallPayload benchmarks encoding with small payload
func BenchmarkFrameEncode_SmallPayload(b *testing.B) {
	frame := &Frame{
		Version:  Version,
		Type:     FrameData,
		Flags:    FlagNone,
		StreamID: 123,
		Payload:  make([]byte, 64), // 64 bytes
	}

	buf := new(bytes.Buffer)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := Encode(buf, frame); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFrameEncode_LargePayload benchmarks encoding with large payload
func BenchmarkFrameEncode_LargePayload(b *testing.B) {
	frame := &Frame{
		Version:  Version,
		Type:     FrameData,
		Flags:    FlagNone,
		StreamID: 123,
		Payload:  make([]byte, 64*1024), // 64KB
	}

	buf := new(bytes.Buffer)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := Encode(buf, frame); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBufferPool benchmarks buffer pool operations
func BenchmarkBufferPool_GetPut(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := GetBuffer(1024)
		PutBuffer(buf)
	}
}

// BenchmarkBufferPool_Concurrent benchmarks concurrent buffer pool access
func BenchmarkBufferPool_Concurrent(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := GetBuffer(1024)
			PutBuffer(buf)
		}
	})
}

// BenchmarkFrameEncode_Concurrent benchmarks concurrent frame encoding
func BenchmarkFrameEncode_Concurrent(b *testing.B) {
	frame := &Frame{
		Version:  Version,
		Type:     FrameData,
		Flags:    FlagNone,
		StreamID: 123,
		Payload:  make([]byte, 1024),
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		buf := new(bytes.Buffer)
		for pb.Next() {
			buf.Reset()
			if err := Encode(buf, frame); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkFrameDecode_Concurrent benchmarks concurrent frame decoding
func BenchmarkFrameDecode_Concurrent(b *testing.B) {
	frame := &Frame{
		Version:  Version,
		Type:     FrameData,
		Flags:    FlagNone,
		StreamID: 123,
		Payload:  make([]byte, 1024),
	}

	buf := new(bytes.Buffer)
	if err := Encode(buf, frame); err != nil {
		b.Fatal(err)
	}
	encodedData := buf.Bytes()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reader := bytes.NewReader(encodedData)
			if _, err := Decode(reader); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkControlFrames benchmarks control frame operations
func BenchmarkControlFrames(b *testing.B) {
	tests := []struct {
		name  string
		frame *Frame
	}{
		{
			name: "Auth",
			frame: &Frame{
				Version:  Version,
				Type:     FrameAuth,
				Flags:    FlagNone,
				StreamID: StreamIDControl,
				Payload:  []byte(`{"token":"test-token"}`),
			},
		},
		{
			name: "Heartbeat",
			frame: &Frame{
				Version:  Version,
				Type:     FrameHeartbeat,
				Flags:    FlagNone,
				StreamID: StreamIDControl,
				Payload:  nil,
			},
		},
		{
			name: "OpenStream",
			frame: &Frame{
				Version:  Version,
				Type:     FrameOpenStream,
				Flags:    FlagNone,
				StreamID: 1,
				Payload:  []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"),
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			buf := new(bytes.Buffer)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				buf.Reset()
				if err := Encode(buf, tt.frame); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
