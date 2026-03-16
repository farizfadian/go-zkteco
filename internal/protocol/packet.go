package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	// PacketHeaderSize is the size of the packet header in bytes
	PacketHeaderSize = 8

	// PacketMinSize is the minimum packet size (header only)
	PacketMinSize = 8

	// HeaderByte1 is the first magic byte
	HeaderByte1 = 0x50

	// HeaderByte2 is the second magic byte
	HeaderByte2 = 0x50
)

// Packet represents a ZKTeco protocol packet.
type Packet struct {
	Command       uint16 // Command ID
	SessionID     uint16 // Session identifier
	ReplyID       uint16 // Reply sequence number
	Data          []byte // Command-specific payload
	ChecksumValid bool   // Whether the checksum was valid on decode
}

// NewPacket creates a new packet with the given command and data.
func NewPacket(command uint16, sessionID uint16, replyID uint16, data []byte) *Packet {
	return &Packet{
		Command:   command,
		SessionID: sessionID,
		ReplyID:   replyID,
		Data:      data,
	}
}

// Encode encodes the packet into bytes for transmission.
// Format: Header(2) + DataLen(2) + Zeros(2) + Command(2) + Checksum(2) + SessionID(2) + ReplyID(2) + Data(N)
func (p *Packet) Encode() []byte {
	dataLen := len(p.Data)
	packetLen := PacketHeaderSize + 6 + dataLen // header(8) + session(2) + reply(2) + reserved(2) + data

	buf := make([]byte, packetLen)

	// Header bytes
	buf[0] = HeaderByte1
	buf[1] = HeaderByte2

	// Data length (includes everything after header except data length itself)
	// = session(2) + reply(2) + command(2) + checksum(2) + data(N) - but we use simplified
	binary.LittleEndian.PutUint16(buf[2:4], uint16(dataLen+8))

	// Zeros (reserved)
	buf[4] = 0x00
	buf[5] = 0x00

	// Command
	binary.LittleEndian.PutUint16(buf[6:8], p.Command)

	// Placeholder for checksum (will be calculated)
	buf[8] = 0x00
	buf[9] = 0x00

	// Session ID
	binary.LittleEndian.PutUint16(buf[10:12], p.SessionID)

	// Reply ID
	binary.LittleEndian.PutUint16(buf[12:14], p.ReplyID)

	// Data
	if dataLen > 0 {
		copy(buf[14:], p.Data)
	}

	// Calculate and set checksum
	checksum := CalculateChecksum(buf[6:])
	binary.LittleEndian.PutUint16(buf[8:10], checksum)

	return buf
}

// Decode decodes bytes into a packet.
func Decode(data []byte) (*Packet, error) {
	if len(data) < PacketMinSize {
		return nil, fmt.Errorf("packet too short: %d bytes", len(data))
	}

	// Verify header
	if data[0] != HeaderByte1 || data[1] != HeaderByte2 {
		return nil, fmt.Errorf("invalid header: 0x%02X 0x%02X", data[0], data[1])
	}

	// Parse fields
	// dataLen := binary.LittleEndian.Uint16(data[2:4])
	command := binary.LittleEndian.Uint16(data[6:8])
	checksum := binary.LittleEndian.Uint16(data[8:10])
	sessionID := binary.LittleEndian.Uint16(data[10:12])
	replyID := binary.LittleEndian.Uint16(data[12:14])

	// Verify checksum
	dataCopy := make([]byte, len(data)-6)
	copy(dataCopy, data[6:])
	dataCopy[2] = 0
	dataCopy[3] = 0
	expectedChecksum := CalculateChecksum(dataCopy)

	// Extract data payload
	var payload []byte
	if len(data) > 14 {
		payload = make([]byte, len(data)-14)
		copy(payload, data[14:])
	}

	return &Packet{
		Command:       command,
		SessionID:     sessionID,
		ReplyID:       replyID,
		Data:          payload,
		ChecksumValid: checksum == expectedChecksum,
	}, nil
}

// CalculateChecksum calculates the ZKTeco checksum for the given data.
func CalculateChecksum(data []byte) uint16 {
	var sum uint32

	// Sum 16-bit words
	for i := 0; i+1 < len(data); i += 2 {
		sum += uint32(binary.LittleEndian.Uint16(data[i : i+2]))
	}

	// Handle odd byte
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1])
	}

	// Fold 32-bit sum to 16-bit
	for sum > 0xFFFF {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}

	// One's complement
	return uint16(^sum)
}

// IsAck returns true if the packet is an acknowledgment.
func (p *Packet) IsAck() bool {
	return p.Command == CMD_ACK_OK || p.Command == CMD_ACK_DATA
}

// IsError returns true if the packet indicates an error.
func (p *Packet) IsError() bool {
	return p.Command == CMD_ACK_ERROR || p.Command == CMD_ACK_UNAUTH
}

// IsData returns true if the packet contains bulk data.
func (p *Packet) IsData() bool {
	return p.Command == CMD_DATA || p.Command == CMD_PREPARE_DATA
}

// String returns a string representation of the packet for debugging.
func (p *Packet) String() string {
	return fmt.Sprintf("Packet{Cmd: %d, Session: %d, Reply: %d, DataLen: %d}",
		p.Command, p.SessionID, p.ReplyID, len(p.Data))
}
