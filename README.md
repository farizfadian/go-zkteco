# go-zkteco

[![Go Reference](https://pkg.go.dev/badge/github.com/farizfadian/go-zkteco.svg)](https://pkg.go.dev/github.com/farizfadian/go-zkteco)
[![Go Report Card](https://goreportcard.com/badge/github.com/farizfadian/go-zkteco)](https://goreportcard.com/report/github.com/farizfadian/go-zkteco)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A pure Go library for communicating with ZKTeco attendance devices (fingerprint, face recognition, RFID). No CGO, no external SDK required.

## Features

- ✅ Pure Go implementation (cross-platform)
- ✅ Connect to ZKTeco devices via TCP
- ✅ Read attendance logs
- ✅ Manage users (create, read, delete)
- ✅ Read/set device time
- ✅ Get device information
- ✅ Thread-safe operations
- ✅ Configurable timeout and retry

## Supported Devices

Tested with:
- ZKTeco Mini AC Plus
- ZKTeco SenseFace series
- ZKTeco SpeedFace series

Should work with most ZKTeco devices with TCP/IP support on port 4370.

## Installation

```bash
go get github.com/farizfadian/go-zkteco
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/farizfadian/go-zkteco"
)

func main() {
    // Connect to device
    device, err := zkteco.Connect("192.168.1.201")
    if err != nil {
        log.Fatal(err)
    }
    defer device.Disconnect()

    // Get device info
    info, err := device.GetDeviceInfo()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Connected to: %s\n", info.SerialNumber)

    // Get attendance logs
    logs, err := device.GetAttendance()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d attendance records\n", len(logs))
    for _, record := range logs {
        fmt.Printf("  User %d: %s (%s via %s)\n",
            record.UserID,
            record.Time.Format("2006-01-02 15:04:05"),
            record.StateString(),
            record.VerifyTypeString())
    }
}
```

## API Reference

### Connection

```go
// Connect with default settings
device, err := zkteco.Connect("192.168.1.201")

// Connect with custom port
device, err := zkteco.Connect("192.168.1.201:4370")

// Connect with options
device, err := zkteco.Connect("192.168.1.201",
    zkteco.WithTimeout(10*time.Second),
    zkteco.WithPassword("0"),
    zkteco.WithRetry(3, time.Second),
)

// Check connectivity without establishing session
err := zkteco.Ping("192.168.1.201", 5*time.Second)

// Disconnect
device.Disconnect()
```

### Attendance

```go
// Get all attendance logs
logs, err := device.GetAttendance()

// Get logs since a specific time
logs, err := device.GetAttendanceSince(time.Now().Add(-24*time.Hour))

// Get attendance count
count, err := device.GetAttendanceCount()

// Clear all attendance (DANGER!)
err := device.ClearAttendance()
```

### AttendanceLog Structure

```go
type AttendanceLog struct {
    UserID     int       // User ID in device
    Time       time.Time // Punch time
    State      int       // 0=CheckIn, 1=CheckOut, 2=BreakOut, 3=BreakIn
    VerifyType int       // 0=Password, 1=Fingerprint, 2=Card, 15=Face
    WorkCode   int       // Work code
}

// Human-readable helpers
log.StateString()      // "CHECK_IN", "CHECK_OUT", etc.
log.VerifyTypeString() // "FINGERPRINT", "FACE", etc.
```

### Users

```go
// Get all users
users, err := device.GetUsers()

// Get specific user
user, err := device.GetUser(123)

// Create/update user
err := device.SetUser(zkteco.User{
    UserID:    123,
    Name:      "John Doe",
    Privilege: 0,  // 0=User, 14=Admin
    Password:  "1234",
    CardNo:    "12345678",
    Enabled:   true,
})

// Delete user
err := device.DeleteUser(123)

// Get user count
count, err := device.GetUserCount()
```

### Device Info & Control

```go
// Get device information
info, err := device.GetDeviceInfo()
fmt.Println(info.SerialNumber)
fmt.Println(info.FirmwareVersion)

// Get/set device time
t, err := device.GetTime()
err := device.SetTime(time.Now())
err := device.SyncTime() // Sync with local time

// Device control
err := device.Enable()   // Resume operation
err := device.Disable()  // Pause operation
err := device.Restart()  // Restart device
```

## Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithTimeout(d)` | 10s | Connection/read timeout |
| `WithPassword(p)` | "" | Device communication key |
| `WithRetry(n, d)` | 3, 1s | Retry count and delay |
| `WithLogger(l)` | nil | Custom logger |

## Error Handling

```go
// Check specific errors
if errors.Is(err, zkteco.ErrNotConnected) {
    // Handle disconnection
}

if errors.Is(err, zkteco.ErrTimeout) {
    // Handle timeout
}

if errors.Is(err, zkteco.ErrUserNotFound) {
    // Handle user not found
}
```

## Thread Safety

All methods on `Device` are thread-safe and can be called from multiple goroutines.

## Logging

Implement the `Logger` interface for custom logging:

```go
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}

device, err := zkteco.Connect("192.168.1.201",
    zkteco.WithLogger(myLogger))
```

## Protocol

This library implements the ZKTeco proprietary TCP protocol:
- TCP connection on port 4370
- Binary packet format with checksums
- Session-based communication

See [docs/PROTOCOL.md](docs/PROTOCOL.md) for technical details.

## Related Projects

- [go-fingerspot](https://github.com/farizfadian/go-fingerspot) - Fingerspot/Solutions device SDK
- [bizcore-attendance-bridge](https://github.com/farizfadian/bizcore-attendance-bridge) - Multi-brand attendance middleware

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Author

Fariz Fadian - [@farizfadian](https://github.com/farizfadian)
