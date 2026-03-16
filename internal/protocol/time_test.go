package protocol

import (
	"testing"
	"time"
)

func TestEncodeDecodeTimeRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
	}{
		{"epoch", time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)},
		{"typical", time.Date(2024, 6, 15, 14, 30, 0, 0, time.Local)},
		{"new year", time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local)},
		{"end of day", time.Date(2023, 12, 31, 23, 59, 59, 0, time.Local)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeTime(tt.time)
			decoded := DecodeTime(encoded)

			if !decoded.Equal(tt.time) {
				t.Errorf("roundtrip failed: got %v, want %v", decoded, tt.time)
			}
		})
	}
}

func TestEncodeDecodeTimeBytesRoundtrip(t *testing.T) {
	original := time.Date(2024, 3, 15, 10, 30, 45, 0, time.Local)

	encoded := EncodeTimeBytes(original)
	if len(encoded) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(encoded))
	}

	decoded := DecodeTimeBytes(encoded)
	if !decoded.Equal(original) {
		t.Errorf("roundtrip failed: got %v, want %v", decoded, original)
	}
}

func TestDecodeTimeBytesTooShort(t *testing.T) {
	result := DecodeTimeBytes([]byte{0x01, 0x02})
	if !result.IsZero() {
		t.Errorf("expected zero time for short input, got %v", result)
	}
}

func TestPackUnpackAttendanceTimeRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
	}{
		{"morning", time.Date(2024, 6, 15, 8, 0, 0, 0, time.Local)},
		{"afternoon", time.Date(2024, 6, 15, 14, 30, 20, 0, time.Local)},
		{"midnight", time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packed := PackAttendanceTime(tt.time)
			unpacked := ParseAttendanceTime(packed)

			// Note: seconds lose 1-bit precision (stored as second/2)
			expectedSecond := (tt.time.Second() / 2) * 2
			expected := time.Date(
				tt.time.Year(), tt.time.Month(), tt.time.Day(),
				tt.time.Hour(), tt.time.Minute(), expectedSecond,
				0, time.Local,
			)

			if !unpacked.Equal(expected) {
				t.Errorf("roundtrip failed: got %v, want %v", unpacked, expected)
			}
		})
	}
}

func TestDecodeDateTimeBytes(t *testing.T) {
	data := []byte{24, 6, 15, 14, 30, 45} // 2024-06-15 14:30:45
	result := DecodeDateTimeBytes(data)

	expected := time.Date(2024, 6, 15, 14, 30, 45, 0, time.Local)
	if !result.Equal(expected) {
		t.Errorf("got %v, want %v", result, expected)
	}
}

func TestDecodeDateTimeBytesTooShort(t *testing.T) {
	result := DecodeDateTimeBytes([]byte{24, 6, 15})
	if !result.IsZero() {
		t.Errorf("expected zero time for short input, got %v", result)
	}
}

func TestEncodeDateTimeBytes(t *testing.T) {
	input := time.Date(2024, 6, 15, 14, 30, 45, 0, time.Local)
	result := EncodeDateTimeBytes(input)

	expected := []byte{24, 6, 15, 14, 30, 45}
	if len(result) != len(expected) {
		t.Fatalf("length: got %d, want %d", len(result), len(expected))
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("byte[%d]: got %d, want %d", i, result[i], expected[i])
		}
	}
}
