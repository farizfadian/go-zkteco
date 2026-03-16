package zkteco

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/farizfadian/go-zkteco/internal/protocol"
)

// Device represents a connection to a ZKTeco attendance device.
type Device struct {
	mu sync.Mutex

	address string
	options *Options

	conn      net.Conn
	sessionID uint16
	replyID   uint16

	connected    bool
	serialNumber string
	deviceName   string
	platform     string
}

// connect establishes a TCP connection and starts a session.
func (d *Device) connect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.connected {
		return ErrAlreadyConnected
	}

	// Establish TCP connection
	conn, err := net.DialTimeout("tcp", d.address, d.options.Timeout)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	d.conn = conn
	d.replyID = 0

	// Send connect command
	packet := protocol.NewPacket(protocol.CMD_CONNECT, 0, d.replyID, nil)
	if err := d.sendPacket(packet); err != nil {
		d.conn.Close()
		return fmt.Errorf("failed to send connect command: %w", err)
	}

	// Receive response
	resp, err := d.receivePacket()
	if err != nil {
		d.conn.Close()
		return fmt.Errorf("failed to receive connect response: %w", err)
	}

	if !resp.IsAck() {
		d.conn.Close()
		return fmt.Errorf("%w: device rejected connection", ErrCommandFailed)
	}

	// Extract session ID from response
	d.sessionID = resp.SessionID
	d.replyID++
	d.connected = true

	d.log().Info("connected to device", "address", d.address, "sessionID", d.sessionID)

	return nil
}

// Disconnect closes the connection to the device.
func (d *Device) Disconnect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return nil
	}

	// Send exit command (best effort)
	packet := protocol.NewPacket(protocol.CMD_EXIT, d.sessionID, d.replyID, nil)
	_ = d.sendPacket(packet)

	// Close connection
	if d.conn != nil {
		d.conn.Close()
		d.conn = nil
	}

	d.connected = false
	d.log().Info("disconnected from device", "address", d.address)

	return nil
}

// IsConnected returns true if the device is connected.
func (d *Device) IsConnected() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.connected
}

// Address returns the device address.
func (d *Device) Address() string {
	return d.address
}

// SessionID returns the current session ID.
func (d *Device) SessionID() uint16 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.sessionID
}

// sendPacket sends a packet to the device.
func (d *Device) sendPacket(p *protocol.Packet) error {
	if d.conn == nil {
		return ErrNotConnected
	}

	data := p.Encode()

	if err := d.conn.SetWriteDeadline(time.Now().Add(d.options.Timeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	if _, err := d.conn.Write(data); err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}

	d.log().Debug("sent packet", "command", p.Command, "dataLen", len(p.Data))
	return nil
}

// receivePacket receives a packet from the device.
func (d *Device) receivePacket() (*protocol.Packet, error) {
	if d.conn == nil {
		return nil, ErrNotConnected
	}

	if err := d.conn.SetReadDeadline(time.Now().Add(d.options.Timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read header first (8 bytes: magic(2) + dataLen(2) + zeros(2) + command(2))
	header := make([]byte, 8)
	if _, err := io.ReadFull(d.conn, header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Get data length from header
	// dataLen = checksum(2) + sessionID(2) + replyID(2) + payload(N)
	dataLen := int(binary.LittleEndian.Uint16(header[2:4]))
	if dataLen < 6 {
		dataLen = 6 // minimum: checksum + sessionID + replyID
	}

	// Read the rest of the packet
	rest := make([]byte, dataLen)
	if _, err := io.ReadFull(d.conn, rest); err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	// Combine into full packet
	buf := make([]byte, 8+dataLen)
	copy(buf, header)
	copy(buf[8:], rest)

	packet, err := protocol.Decode(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode packet: %w", err)
	}

	if d.options.StrictChecksum && !packet.ChecksumValid {
		return nil, ErrChecksumMismatch
	}

	d.log().Debug("received packet", "command", packet.Command, "dataLen", len(packet.Data))
	return packet, nil
}

// sendCommand sends a command and waits for response.
func (d *Device) sendCommand(cmd uint16, data []byte) (*protocol.Packet, error) {
	packet := protocol.NewPacket(cmd, d.sessionID, d.replyID, data)

	if err := d.sendPacket(packet); err != nil {
		return nil, err
	}

	d.replyID++

	resp, err := d.receivePacket()
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("%w: command %d failed", ErrCommandFailed, cmd)
	}

	return resp, nil
}

// readLargeData reads bulk data from the device.
// Some commands return data in chunks using CMD_PREPARE_DATA / CMD_DATA protocol.
func (d *Device) readLargeData(cmd uint16) ([]byte, error) {
	// Send the command
	resp, err := d.sendCommand(cmd, nil)
	if err != nil {
		return nil, err
	}

	// Check if data is directly in response
	if resp.Command == protocol.CMD_ACK_OK || resp.Command == protocol.CMD_ACK_DATA {
		if len(resp.Data) >= 4 {
			// First 4 bytes might be data size
			dataSize := binary.LittleEndian.Uint32(resp.Data[:4])

			if dataSize == 0 {
				return nil, nil // No data
			}

			// Check if this is a prepare data response
			if resp.Command == protocol.CMD_ACK_DATA || dataSize > uint32(len(resp.Data)-4) {
				// Need to read bulk data
				return d.readBulkData(dataSize)
			}

			// Data is in response
			return resp.Data[4:], nil
		}
		return resp.Data, nil
	}

	// Check for prepare data command
	if resp.Command == protocol.CMD_PREPARE_DATA {
		if len(resp.Data) >= 4 {
			dataSize := binary.LittleEndian.Uint32(resp.Data[:4])
			return d.readBulkData(dataSize)
		}
	}

	return resp.Data, nil
}

// readBulkData reads bulk data using the DATA protocol.
func (d *Device) readBulkData(totalSize uint32) ([]byte, error) {
	data := make([]byte, 0, totalSize)

	for uint32(len(data)) < totalSize {
		// Request next chunk
		packet := protocol.NewPacket(protocol.CMD_DATA, d.sessionID, d.replyID, nil)
		if err := d.sendPacket(packet); err != nil {
			return nil, fmt.Errorf("failed to request data chunk: %w", err)
		}
		d.replyID++

		// Receive chunk
		resp, err := d.receivePacket()
		if err != nil {
			return nil, fmt.Errorf("failed to receive data chunk: %w", err)
		}

		if len(resp.Data) == 0 {
			break
		}

		data = append(data, resp.Data...)
	}

	// Free data buffer on device
	freePacket := protocol.NewPacket(protocol.CMD_FREE_DATA, d.sessionID, d.replyID, nil)
	_ = d.sendPacket(freePacket)
	d.replyID++

	return data, nil
}

// Enable enables the device (resumes normal operation).
func (d *Device) Enable() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return ErrNotConnected
	}

	_, err := d.sendCommand(protocol.CMD_ENABLE_DEVICE, nil)
	return err
}

// Disable disables the device (stops capturing attendance).
func (d *Device) Disable() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return ErrNotConnected
	}

	_, err := d.sendCommand(protocol.CMD_DISABLE_DEVICE, nil)
	return err
}

// Restart restarts the device.
func (d *Device) Restart() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return ErrNotConnected
	}

	_, err := d.sendCommand(protocol.CMD_RESTART, nil)
	if err != nil {
		return err
	}

	// Device will disconnect
	d.connected = false
	if d.conn != nil {
		d.conn.Close()
		d.conn = nil
	}

	return nil
}

// log returns the logger, defaulting to no-op if not set.
func (d *Device) log() Logger {
	if d.options.Logger != nil {
		return d.options.Logger
	}
	return defaultLogger()
}
