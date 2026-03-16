package protocol

import (
	"testing"
)

func TestCalculateChecksum(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"single byte", []byte{0x42}},
		{"two bytes", []byte{0x01, 0x02}},
		{"four bytes", []byte{0x01, 0x02, 0x03, 0x04}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateChecksum(tt.data)
			// Verify it's deterministic
			result2 := CalculateChecksum(tt.data)
			if result != result2 {
				t.Errorf("checksum not deterministic: got %d and %d", result, result2)
			}
		})
	}
}

func TestPacketEncodeDecodeRoundtrip(t *testing.T) {
	tests := []struct {
		name      string
		command   uint16
		sessionID uint16
		replyID   uint16
		data      []byte
	}{
		{"connect", CMD_CONNECT, 0, 0, nil},
		{"exit", CMD_EXIT, 1234, 5, nil},
		{"with data", CMD_OPTIONS_RRQ, 100, 1, []byte("~SerialNumber\x00")},
		{"ack ok", CMD_ACK_OK, 200, 3, []byte{0x01, 0x02, 0x03, 0x04}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := NewPacket(tt.command, tt.sessionID, tt.replyID, tt.data)
			encoded := original.Encode()

			decoded, err := Decode(encoded)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if decoded.Command != tt.command {
				t.Errorf("command: got %d, want %d", decoded.Command, tt.command)
			}
			if decoded.SessionID != tt.sessionID {
				t.Errorf("sessionID: got %d, want %d", decoded.SessionID, tt.sessionID)
			}
			if decoded.ReplyID != tt.replyID {
				t.Errorf("replyID: got %d, want %d", decoded.ReplyID, tt.replyID)
			}
			if !decoded.ChecksumValid {
				t.Error("checksum should be valid after encode/decode roundtrip")
			}

			if tt.data == nil {
				if len(decoded.Data) != 0 {
					t.Errorf("data: got %d bytes, want 0", len(decoded.Data))
				}
			} else {
				if len(decoded.Data) != len(tt.data) {
					t.Fatalf("data length: got %d, want %d", len(decoded.Data), len(tt.data))
				}
				for i := range tt.data {
					if decoded.Data[i] != tt.data[i] {
						t.Errorf("data[%d]: got 0x%02X, want 0x%02X", i, decoded.Data[i], tt.data[i])
					}
				}
			}
		})
	}
}

func TestDecodeInvalidPacket(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"too short", []byte{0x50, 0x50}},
		{"wrong magic byte 1", []byte{0x00, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{"wrong magic byte 2", []byte{0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode(tt.data)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestPacketIsAck(t *testing.T) {
	p := &Packet{Command: CMD_ACK_OK}
	if !p.IsAck() {
		t.Error("CMD_ACK_OK should be ack")
	}

	p.Command = CMD_ACK_DATA
	if !p.IsAck() {
		t.Error("CMD_ACK_DATA should be ack")
	}

	p.Command = CMD_ACK_ERROR
	if p.IsAck() {
		t.Error("CMD_ACK_ERROR should not be ack")
	}
}

func TestPacketIsError(t *testing.T) {
	p := &Packet{Command: CMD_ACK_ERROR}
	if !p.IsError() {
		t.Error("CMD_ACK_ERROR should be error")
	}

	p.Command = CMD_ACK_UNAUTH
	if !p.IsError() {
		t.Error("CMD_ACK_UNAUTH should be error")
	}

	p.Command = CMD_ACK_OK
	if p.IsError() {
		t.Error("CMD_ACK_OK should not be error")
	}
}
