package zkteco

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/farizfadian/go-zkteco/internal/protocol"
)

// DeviceInfo contains information about the device.
type DeviceInfo struct {
	SerialNumber    string
	DeviceName      string
	Platform        string
	FirmwareVersion string
	MACAddress      string
	ProductTime     string
	Manufacturer    string
}

// String returns a string representation of the device info.
func (d DeviceInfo) String() string {
	return fmt.Sprintf("DeviceInfo{SN: %s, Name: %s, Platform: %s, Firmware: %s}",
		d.SerialNumber, d.DeviceName, d.Platform, d.FirmwareVersion)
}

// GetDeviceInfo retrieves comprehensive device information.
func (d *Device) GetDeviceInfo() (*DeviceInfo, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return nil, ErrNotConnected
	}

	info := &DeviceInfo{}

	// Get various options
	if sn, err := d.getOption("~SerialNumber"); err == nil {
		info.SerialNumber = sn
	}
	if name, err := d.getOption("~DeviceName"); err == nil {
		info.DeviceName = name
	}
	if platform, err := d.getOption("~Platform"); err == nil {
		info.Platform = platform
	}
	if firmware, err := d.getOption("FPVersion"); err == nil {
		info.FirmwareVersion = firmware
	}
	if mac, err := d.getOption("MAC"); err == nil {
		info.MACAddress = mac
	}

	return info, nil
}

// GetSerialNumber returns the device serial number.
func (d *Device) GetSerialNumber() (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return "", ErrNotConnected
	}

	return d.getOption("~SerialNumber")
}

// GetDeviceName returns the device name.
func (d *Device) GetDeviceName() (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return "", ErrNotConnected
	}

	return d.getOption("~DeviceName")
}

// GetFirmwareVersion returns the firmware version.
func (d *Device) GetFirmwareVersion() (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return "", ErrNotConnected
	}

	return d.getOption("FPVersion")
}

// GetPlatform returns the device platform.
func (d *Device) GetPlatform() (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return "", ErrNotConnected
	}

	return d.getOption("~Platform")
}

// GetTime returns the current time on the device.
func (d *Device) GetTime() (time.Time, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return time.Time{}, ErrNotConnected
	}

	resp, err := d.sendCommand(protocol.CMD_GET_TIME, nil)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get device time: %w", err)
	}

	if len(resp.Data) < 4 {
		return time.Time{}, ErrInvalidResponse
	}

	return protocol.DecodeTimeBytes(resp.Data), nil
}

// SetTime sets the device time.
func (d *Device) SetTime(t time.Time) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return ErrNotConnected
	}

	data := protocol.EncodeTimeBytes(t)

	_, err := d.sendCommand(protocol.CMD_SET_TIME, data)
	if err != nil {
		return fmt.Errorf("failed to set device time: %w", err)
	}

	d.log().Info("set device time", "time", t.Format(time.RFC3339))
	return nil
}

// SyncTime synchronizes the device time with the local system time.
func (d *Device) SyncTime() error {
	return d.SetTime(time.Now())
}

// GetCapacity returns the device capacity information.
type Capacity struct {
	UserCount       int
	UserCapacity    int
	LogCount        int
	LogCapacity     int
	FPCount         int // Fingerprint count
	FPCapacity      int
	FaceCount       int
	FaceCapacity    int
	PasswordCount   int
	CardCount       int
}

// GetCapacity returns the device capacity and current usage.
func (d *Device) GetCapacity() (*Capacity, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return nil, ErrNotConnected
	}

	resp, err := d.sendCommand(protocol.CMD_GET_FREE_SIZES, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get capacity: %w", err)
	}

	if len(resp.Data) < 24 {
		return nil, ErrInvalidResponse
	}

	// Parse response - common format with 4-byte values
	cap := &Capacity{
		UserCount:    int(binary.LittleEndian.Uint32(resp.Data[0:4])),
		UserCapacity: int(binary.LittleEndian.Uint32(resp.Data[4:8])),
		FPCount:      int(binary.LittleEndian.Uint32(resp.Data[8:12])),
		FPCapacity:   int(binary.LittleEndian.Uint32(resp.Data[12:16])),
		LogCount:     int(binary.LittleEndian.Uint32(resp.Data[16:20])),
		LogCapacity:  int(binary.LittleEndian.Uint32(resp.Data[20:24])),
	}

	return cap, nil
}

// getOption reads a device option by name.
func (d *Device) getOption(name string) (string, error) {
	// Build request data
	data := []byte(name + "\x00")

	resp, err := d.sendCommand(protocol.CMD_OPTIONS_RRQ, data)
	if err != nil {
		return "", err
	}

	// Response format: "name=value"
	result := strings.TrimRight(string(resp.Data), "\x00")

	// Parse "name=value" format
	parts := strings.SplitN(result, "=", 2)
	if len(parts) != 2 {
		// Some devices return just the value
		return strings.TrimSpace(result), nil
	}

	return strings.TrimSpace(parts[1]), nil
}

// setOption writes a device option.
func (d *Device) setOption(name, value string) error {
	data := []byte(fmt.Sprintf("%s=%s\x00", name, value))

	_, err := d.sendCommand(protocol.CMD_OPTIONS_WRQ, data)
	return err
}
