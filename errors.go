package zkteco

import (
	"errors"
	"fmt"
)

// Common errors returned by the package.
var (
	// ErrNotConnected is returned when trying to use a disconnected device.
	ErrNotConnected = errors.New("zkteco: not connected to device")

	// ErrAlreadyConnected is returned when trying to connect an already connected device.
	ErrAlreadyConnected = errors.New("zkteco: already connected")

	// ErrTimeout is returned when a connection or operation times out.
	ErrTimeout = errors.New("zkteco: operation timeout")

	// ErrInvalidResponse is returned when the device returns an unexpected response.
	ErrInvalidResponse = errors.New("zkteco: invalid response from device")

	// ErrConnectionFailed is returned when the connection to the device fails.
	ErrConnectionFailed = errors.New("zkteco: connection failed")

	// ErrCommandFailed is returned when a command fails on the device.
	ErrCommandFailed = errors.New("zkteco: command failed")

	// ErrUserNotFound is returned when a user is not found on the device.
	ErrUserNotFound = errors.New("zkteco: user not found")

	// ErrNoData is returned when no data is available.
	ErrNoData = errors.New("zkteco: no data available")

	// ErrDeviceBusy is returned when the device is busy.
	ErrDeviceBusy = errors.New("zkteco: device is busy")

	// ErrInvalidPacket is returned when a received packet is malformed.
	ErrInvalidPacket = errors.New("zkteco: invalid packet")

	// ErrChecksumMismatch is returned when packet checksum verification fails.
	ErrChecksumMismatch = errors.New("zkteco: checksum mismatch")
)

// DeviceError represents an error returned by the device.
type DeviceError struct {
	Code    uint16
	Message string
}

func (e *DeviceError) Error() string {
	if e.Message != "" {
		return "zkteco: device error: " + e.Message
	}
	return fmt.Sprintf("zkteco: device error code %d", e.Code)
}

// IsNotConnected returns true if the error indicates the device is not connected.
func IsNotConnected(err error) bool {
	return errors.Is(err, ErrNotConnected)
}

// IsTimeout returns true if the error indicates a timeout.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsUserNotFound returns true if the error indicates a user was not found.
func IsUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}
