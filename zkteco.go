// Package zkteco provides a Go client for communicating with ZKTeco attendance devices.
//
// Basic usage:
//
//	device, err := zkteco.Connect("192.168.1.201:4370")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer device.Disconnect()
//
//	logs, err := device.GetAttendance()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, log := range logs {
//	    fmt.Printf("User %d punched at %s\n", log.UserID, log.Time)
//	}
package zkteco

import (
	"fmt"
	"net"
	"time"
)

const (
	// DefaultPort is the default ZKTeco TCP port
	DefaultPort = 4370

	// DefaultTimeout is the default connection/read timeout
	DefaultTimeout = 10 * time.Second
)

// Connect establishes a connection to a ZKTeco device.
// The address should be in the format "host:port" or just "host" (uses default port 4370).
//
// Example:
//
//	device, err := zkteco.Connect("192.168.1.201")
//	device, err := zkteco.Connect("192.168.1.201:4370")
//	device, err := zkteco.Connect("192.168.1.201", zkteco.WithTimeout(5*time.Second))
func Connect(address string, opts ...Option) (*Device, error) {
	// Apply default options
	options := &Options{
		Timeout:    DefaultTimeout,
		Password:   "",
		RetryCount: 3,
		RetryDelay: time.Second,
	}

	// Apply custom options
	for _, opt := range opts {
		opt(options)
	}

	// Parse address
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		// No port specified, use default
		host = address
		port = fmt.Sprintf("%d", DefaultPort)
	}

	addr := net.JoinHostPort(host, port)

	// Create device
	device := &Device{
		address: addr,
		options: options,
	}

	// Connect
	if err := device.connect(); err != nil {
		return nil, err
	}

	return device, nil
}

// MustConnect is like Connect but panics on error.
// Useful for initialization in main() or tests.
func MustConnect(address string, opts ...Option) *Device {
	device, err := Connect(address, opts...)
	if err != nil {
		panic(fmt.Sprintf("zkteco: failed to connect to %s: %v", address, err))
	}
	return device
}

// Ping checks if a ZKTeco device is reachable at the given address.
// This is a quick connectivity check without establishing a full session.
func Ping(address string, timeout time.Duration) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		host = address
		port = fmt.Sprintf("%d", DefaultPort)
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return fmt.Errorf("device not reachable: %w", err)
	}
	conn.Close()
	return nil
}
