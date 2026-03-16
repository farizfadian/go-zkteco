package zkteco

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/farizfadian/go-zkteco/internal/protocol"
)

// AttendanceLog represents a single attendance record from the device.
type AttendanceLog struct {
	UserID     int       // User ID in device
	Time       time.Time // Punch time
	State      int       // 0=CheckIn, 1=CheckOut, 2=BreakOut, 3=BreakIn, 4=OTIn, 5=OTOut
	VerifyType int       // 0=Password, 1=Fingerprint, 2=Card, 15=Face
	WorkCode   int       // Work code (if supported)
}

// StateString returns a human-readable string for the attendance state.
func (a AttendanceLog) StateString() string {
	return protocol.StateString(a.State)
}

// VerifyTypeString returns a human-readable string for the verify type.
func (a AttendanceLog) VerifyTypeString() string {
	return protocol.VerifyTypeString(a.VerifyType)
}

// String returns a string representation of the attendance log.
func (a AttendanceLog) String() string {
	return fmt.Sprintf("AttendanceLog{UserID: %d, Time: %s, State: %s, Verify: %s}",
		a.UserID, a.Time.Format("2006-01-02 15:04:05"), a.StateString(), a.VerifyTypeString())
}

// GetAttendance retrieves all attendance logs from the device.
func (d *Device) GetAttendance() ([]AttendanceLog, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return nil, ErrNotConnected
	}

	// Disable device during data transfer (optional but recommended)
	_, _ = d.sendCommand(protocol.CMD_DISABLE_DEVICE, nil)
	defer d.sendCommand(protocol.CMD_ENABLE_DEVICE, nil)

	// Request attendance logs
	data, err := d.readLargeData(protocol.CMD_ATTLOG_RRQ)
	if err != nil {
		return nil, fmt.Errorf("failed to read attendance logs: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	logs, err := parseAttendanceLogs(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse attendance logs: %w", err)
	}

	d.log().Info("retrieved attendance logs", "count", len(logs))
	return logs, nil
}

// GetAttendanceSince retrieves attendance logs since the given time.
// Note: This filters client-side as most devices don't support server-side filtering.
func (d *Device) GetAttendanceSince(since time.Time) ([]AttendanceLog, error) {
	logs, err := d.GetAttendance()
	if err != nil {
		return nil, err
	}

	var filtered []AttendanceLog
	for _, log := range logs {
		if !log.Time.Before(since) {
			filtered = append(filtered, log)
		}
	}

	return filtered, nil
}

// GetAttendanceCount returns the number of attendance records on the device.
func (d *Device) GetAttendanceCount() (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return 0, ErrNotConnected
	}

	resp, err := d.sendCommand(protocol.CMD_GET_FREE_SIZES, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get attendance count: %w", err)
	}

	if len(resp.Data) < 24 {
		return 0, ErrInvalidResponse
	}

	// Log count is at offset 16 (varies by device)
	count := int(binary.LittleEndian.Uint32(resp.Data[16:20]))
	return count, nil
}

// ClearAttendance clears all attendance logs from the device.
// WARNING: This permanently deletes all attendance data!
func (d *Device) ClearAttendance() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return ErrNotConnected
	}

	_, err := d.sendCommand(protocol.CMD_CLEAR_ATTLOG, nil)
	if err != nil {
		return fmt.Errorf("failed to clear attendance logs: %w", err)
	}

	d.log().Info("cleared attendance logs")
	return nil
}

// parseAttendanceLogs parses raw attendance data from the device.
// The format varies by device model. This handles the most common formats.
func parseAttendanceLogs(data []byte) ([]AttendanceLog, error) {
	var logs []AttendanceLog

	// Try different record sizes (8, 16, 40 bytes are common)
	recordSize := detectRecordSize(data)
	if recordSize == 0 {
		return nil, fmt.Errorf("unable to detect record size")
	}

	for i := 0; i+recordSize <= len(data); i += recordSize {
		record := data[i : i+recordSize]
		log, err := parseAttendanceRecord(record, recordSize)
		if err != nil {
			// Skip invalid records
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// detectRecordSize attempts to detect the attendance record size.
func detectRecordSize(data []byte) int {
	// Common sizes: 8, 16, 40 bytes
	sizes := []int{40, 16, 8}

	for _, size := range sizes {
		if len(data)%size == 0 && len(data) >= size {
			// Try to parse first record
			if _, err := parseAttendanceRecord(data[:size], size); err == nil {
				return size
			}
		}
	}

	// Default fallback
	if len(data) >= 16 {
		return 16
	}
	return 0
}

// parseAttendanceRecord parses a single attendance record.
func parseAttendanceRecord(record []byte, size int) (AttendanceLog, error) {
	var log AttendanceLog

	switch size {
	case 8:
		// Simple format: UserID(2) + Time(4) + State(1) + Verify(1)
		log.UserID = int(binary.LittleEndian.Uint16(record[0:2]))
		timeVal := binary.LittleEndian.Uint32(record[2:6])
		log.Time = protocol.ParseAttendanceTime(timeVal)
		log.State = int(record[6])
		log.VerifyType = int(record[7])

	case 16:
		// Medium format: UserID(4) + Time(4) + State(1) + Verify(1) + Reserved(6)
		log.UserID = int(binary.LittleEndian.Uint32(record[0:4]))
		timeVal := binary.LittleEndian.Uint32(record[4:8])
		log.Time = protocol.ParseAttendanceTime(timeVal)
		log.State = int(record[8])
		log.VerifyType = int(record[9])

	case 40:
		// Extended format (newer devices)
		// UserID(9) + Timestamp(4) + State(1) + Verify(1) + WorkCode(1) + Reserved(24)
		// UserID might be string
		userIDBytes := record[0:9]
		// Find null terminator or parse as string
		for i, b := range userIDBytes {
			if b == 0 {
				userIDBytes = userIDBytes[:i]
				break
			}
		}
		// Try to parse as number
		var userID int
		for _, b := range userIDBytes {
			if b >= '0' && b <= '9' {
				userID = userID*10 + int(b-'0')
			}
		}
		log.UserID = userID

		timeVal := binary.LittleEndian.Uint32(record[24:28])
		log.Time = protocol.DecodeTime(timeVal)
		log.State = int(record[28])
		log.VerifyType = int(record[29])
		log.WorkCode = int(record[30])

	default:
		return log, fmt.Errorf("unsupported record size: %d", size)
	}

	// Validate
	if log.UserID <= 0 || log.Time.Year() < 2000 || log.Time.Year() > 2100 {
		return log, fmt.Errorf("invalid record data")
	}

	return log, nil
}
