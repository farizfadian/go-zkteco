# CLAUDE.md - go-zkteco SDK

> **PENTING**: File ini adalah konteks utama untuk Claude AI. Baca dan pahami sebelum melakukan perubahan apapun.

---

## 🎯 Project Identity

| Attribute | Value |
|-----------|-------|
| **Project Name** | go-zkteco |
| **Repository** | github.com/farizfadian/go-zkteco |
| **Type** | Go SDK / Library |
| **Purpose** | ZKTeco Attendance Device Communication |
| **License** | MIT (Open Source) |
| **Owner** | Fariz Fadian |

---

## 📋 Description

`go-zkteco` is a pure Go library for communicating with ZKTeco attendance devices (fingerprint, face recognition, RFID). It implements the ZKTeco proprietary TCP protocol without any external dependencies.

### Features
- Pure Go implementation (no CGO, no external SDK)
- Connect to ZKTeco devices via TCP
- Read attendance logs
- Manage users (create, update, delete)
- Read/set device time
- Cross-platform (Linux, Windows, macOS, ARM)

### Supported Devices
- ZKTeco Mini AC Plus
- ZKTeco SenseFace series
- ZKTeco SpeedFace series
- ZKTeco K-series (K40, K50, etc.)
- ZKTeco U-series
- Most ZKTeco devices with TCP/IP support on port 4370

---

## 🏗️ Architecture

```
go-zkteco/
├── CLAUDE.md                    # AI context (this file)
├── README.md                    # Public documentation
├── LICENSE                      # MIT License
├── go.mod                       # Go module (go 1.21, zero dependencies)
├── Makefile                     # Build commands
│
├── zkteco.go                    # Main entry point (Connect, MustConnect, Ping)
├── device.go                    # Device struct, TCP send/receive, bulk data
├── attendance.go                # Attendance operations + record parsing
├── user.go                      # User management + record parsing
├── info.go                      # Device info, time, capacity operations
├── errors.go                    # Error definitions + sentinel errors
├── errors_test.go               # Error tests
├── options.go                   # Functional options (WithTimeout, WithRetry, etc.)
│
├── internal/
│   └── protocol/
│       ├── packet.go            # Packet encode/decode + checksum calculation
│       ├── packet_test.go       # Packet roundtrip tests
│       ├── commands.go          # Command constants + string helpers
│       ├── commands_test.go     # Command string tests
│       ├── time.go              # ZKTeco time encoding/decoding
│       └── time_test.go         # Time roundtrip tests
│
└── cmd/
    └── example/
        └── main.go              # Usage example
```

---

## 🔌 Protocol Specification

### Connection
- **Transport**: TCP
- **Default Port**: 4370
- **Byte Order**: Little Endian

### Packet Structure

```
┌────────────────────────────────────────────────────────────────┐
│  ZKTECO PACKET FORMAT                                          │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  Offset   Size   Field         Description                    │
│  ──────────────────────────────────────────────────────────── │
│  0        2      Header        Magic bytes (0x50, 0x50)       │
│  2        2      DataLength    Length of data after header    │
│  4        2      Zeros         Reserved (0x00, 0x00)          │
│  6        2      Command       Command ID                     │
│  8        2      Checksum      Packet checksum                │
│  10       2      SessionID     Session identifier             │
│  12       2      ReplyID       Reply sequence number          │
│  14       N      Data          Command-specific payload       │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### Commands

| Command | ID (Dec) | ID (Hex) | Description |
|---------|----------|----------|-------------|
| CMD_CONNECT | 1000 | 0x03E8 | Establish connection |
| CMD_EXIT | 1001 | 0x03E9 | Close connection |
| CMD_ENABLE_DEVICE | 1002 | 0x03EA | Enable device |
| CMD_DISABLE_DEVICE | 1003 | 0x03EB | Disable device |
| CMD_RESTART | 1004 | 0x03EC | Restart device |
| CMD_POWEROFF | 1005 | 0x03ED | Power off device |
| CMD_ACK_OK | 2000 | 0x07D0 | Acknowledgment OK |
| CMD_ACK_ERROR | 2001 | 0x07D1 | Acknowledgment Error |
| CMD_ACK_DATA | 2002 | 0x07D2 | Data acknowledgment |
| CMD_PREPARE_DATA | 1500 | 0x05DC | Prepare bulk data |
| CMD_DATA | 1501 | 0x05DD | Bulk data transfer |
| CMD_FREE_DATA | 1502 | 0x05DE | Free bulk data |
| CMD_GET_TIME | 201 | 0x00C9 | Get device time |
| CMD_SET_TIME | 202 | 0x00CA | Set device time |
| CMD_ATTLOG_RRQ | 13 | 0x000D | Read attendance logs |
| CMD_CLEAR_ATTLOG | 15 | 0x000F | Clear attendance logs |
| CMD_USER_WRQ | 8 | 0x0008 | Write user |
| CMD_USERINFO_RRQ | 9 | 0x0009 | Read user info |
| CMD_DELETE_USER | 18 | 0x0012 | Delete user |
| CMD_OPTIONS_RRQ | 11 | 0x000B | Read options |
| CMD_OPTIONS_WRQ | 12 | 0x000C | Write options |

### Checksum Algorithm

```go
func calculateChecksum(data []byte) uint16 {
    var sum uint32
    for i := 0; i+1 < len(data); i += 2 {
        sum += uint32(binary.LittleEndian.Uint16(data[i : i+2]))
    }
    if len(data)%2 == 1 {
        sum += uint32(data[len(data)-1])
    }
    for sum > 0xFFFF {
        sum = (sum & 0xFFFF) + (sum >> 16)
    }
    return uint16(^sum)
}
```

### Time Encoding
ZKTeco uses seconds since 2000-01-01 00:00:00 (local time).

```go
var zkEpoch = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)

func encodeTime(t time.Time) uint32 {
    return uint32(t.Sub(zkEpoch).Seconds())
}

func decodeTime(seconds uint32) time.Time {
    return zkEpoch.Add(time.Duration(seconds) * time.Second)
}
```

---

## 💻 Public API

### Basic Usage

```go
import "github.com/farizfadian/go-zkteco"

// Connect to device
device, err := zkteco.Connect("192.168.1.201:4370")
if err != nil {
    log.Fatal(err)
}
defer device.Disconnect()

// Get attendance logs
logs, err := device.GetAttendance()
if err != nil {
    log.Fatal(err)
}

for _, log := range logs {
    fmt.Printf("User %d punched at %s\n", log.UserID, log.Time)
}
```

### With Options

```go
device, err := zkteco.Connect("192.168.1.201:4370",
    zkteco.WithTimeout(10*time.Second),
    zkteco.WithPassword("12345"),
    zkteco.WithLogger(myLogger),
)
```

### Device Interface

```go
type Device interface {
    // Connection
    Connect() error
    Disconnect() error
    IsConnected() bool
    
    // Device info
    GetSerialNumber() (string, error)
    GetFirmwareVersion() (string, error)
    GetDeviceName() (string, error)
    GetTime() (time.Time, error)
    SetTime(t time.Time) error
    
    // Attendance
    GetAttendance() ([]AttendanceLog, error)
    GetAttendanceSince(since time.Time) ([]AttendanceLog, error)
    ClearAttendance() error
    
    // Users
    GetUsers() ([]User, error)
    GetUser(userID int) (*User, error)
    SetUser(user User) error
    DeleteUser(userID int) error
    
    // Device control
    Enable() error
    Disable() error
    Restart() error
}
```

---

## 🔧 Data Structures

### AttendanceLog

```go
type AttendanceLog struct {
    UserID     int       // User ID in device
    Time       time.Time // Punch time
    State      int       // 0=CheckIn, 1=CheckOut, 2=BreakOut, 3=BreakIn, 4=OTIn, 5=OTOut
    VerifyType int       // 0=Password, 1=FP, 2=Card, 3-6=Combos, 7=Palm, 8-13=Face combos, 14=Vein, 15=Face
    WorkCode   int       // Work code (if supported)
}

func (a AttendanceLog) StateString() string {
    switch a.State {
    case 0: return "CHECK_IN"
    case 1: return "CHECK_OUT"
    case 2: return "BREAK_OUT"
    case 3: return "BREAK_IN"
    case 4: return "OT_IN"
    case 5: return "OT_OUT"
    default: return "UNKNOWN"
    }
}

// VerifyTypeString() supports all 16 types:
// 0=PASSWORD, 1=FINGERPRINT, 2=CARD, 3=FINGERPRINT+PASSWORD,
// 4=FINGERPRINT+CARD, 5=CARD+PASSWORD, 6=FINGERPRINT+CARD+PASSWORD,
// 7=PALM, 8=FACE+FINGERPRINT, 9=FACE+PASSWORD, 10=FACE+CARD,
// 11=PALM+FINGERPRINT, 12=FACE+FINGERPRINT+CARD,
// 13=FACE+FINGERPRINT+PASSWORD, 14=FINGER_VEIN, 15=FACE
```

### User

```go
type User struct {
    UserID    int
    Name      string
    Privilege int    // 0=User, 1=Enroller, 2=Manager, 14=Admin
    Password  string
    CardNo    string
    Enabled   bool
}
```

---

## ⚙️ Configuration Options

```go
type Options struct {
    Timeout        time.Duration  // Connection/read timeout
    Password       string         // Device communication key
    Logger         Logger         // Custom logger
    RetryCount     int            // Retry on failure
    RetryDelay     time.Duration  // Delay between retries
    StrictChecksum bool           // Validate packet checksums (default: false)
}

// Option functions
func WithTimeout(d time.Duration) Option
func WithPassword(p string) Option
func WithLogger(l Logger) Option
func WithRetry(count int, delay time.Duration) Option
func WithStrictChecksum(strict bool) Option
```

---

## 🧪 Testing

```bash
# Unit tests (no device needed)
go test ./...

# Integration tests (requires device)
ZKTECO_IP=192.168.1.201 go test ./... -tags=integration

# Coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## 📦 Building

```bash
# Build example
go build -o bin/example ./cmd/example

# Cross compile
GOOS=linux GOARCH=amd64 go build -o bin/example-linux ./cmd/example
GOOS=windows GOARCH=amd64 go build -o bin/example.exe ./cmd/example
GOOS=linux GOARCH=arm64 go build -o bin/example-arm64 ./cmd/example
```

---

## ⚠️ Known Limitations

1. **Face templates**: Reading/writing face templates not yet implemented
2. **Fingerprint templates**: Reading/writing fingerprint templates not yet implemented
3. **Real-time events**: Push mode not implemented (use polling)
4. **Some firmware variations**: Packet format may vary slightly
5. **Password option**: `WithPassword()` accepted but not yet sent during connect handshake
6. **context.Context**: Not yet supported - timeout only via Options
7. **GetCapacity**: Field offsets may vary by device model (best-effort parsing)

---

## 📝 Coding Standards

### Error Handling

```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to read attendance: %w", err)
}

// Use custom error types for specific cases
var (
    ErrNotConnected    = errors.New("not connected to device")
    ErrTimeout         = errors.New("connection timeout")
    ErrInvalidResponse = errors.New("invalid response from device")
    ErrUserNotFound    = errors.New("user not found")
)
```

### Logging

```go
// Use interface for flexibility
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Error(msg string, args ...any)
}
```

### Thread Safety

```go
// Device struct is thread-safe
type zkDevice struct {
    mu   sync.Mutex
    conn net.Conn
    // ...
}

func (d *zkDevice) GetAttendance() ([]AttendanceLog, error) {
    d.mu.Lock()
    defer d.mu.Unlock()
    // ...
}
```

---

## 🔗 Related Projects

| Project | Purpose |
|---------|---------|
| `go-fingerspot` | Fingerspot/Solutions device SDK |
| `bizcore-attendance-bridge` | BizCore middleware using both SDKs |

---

## 📚 References

- ZKTeco official documentation (limited)
- Community reverse engineering efforts
- Existing implementations (Python, PHP, Node.js)

---

*Last Updated: March 2026*
*Maintained by: Fariz Fadian*
