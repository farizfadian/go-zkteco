package zkteco

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/farizfadian/go-zkteco/internal/protocol"
)

// User represents a user registered on the device.
type User struct {
	UserID    int    // User ID (numeric)
	Name      string // User name
	Privilege int    // 0=User, 1=Enroller, 2=Manager, 14=Admin
	Password  string // User password (PIN)
	CardNo    string // RFID card number
	Enabled   bool   // Whether user is enabled
	GroupNo   int    // Group number
}

// PrivilegeString returns a human-readable string for the privilege level.
func (u User) PrivilegeString() string {
	return protocol.PrivilegeString(u.Privilege)
}

// String returns a string representation of the user.
func (u User) String() string {
	return fmt.Sprintf("User{ID: %d, Name: %s, Privilege: %s, Enabled: %t}",
		u.UserID, u.Name, u.PrivilegeString(), u.Enabled)
}

// GetUsers retrieves all users from the device.
func (d *Device) GetUsers() ([]User, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return nil, ErrNotConnected
	}

	// Disable device during data transfer
	_, _ = d.sendCommand(protocol.CMD_DISABLE_DEVICE, nil)
	defer d.sendCommand(protocol.CMD_ENABLE_DEVICE, nil)

	// Request user info
	data, err := d.readLargeData(protocol.CMD_USERINFO_RRQ)
	if err != nil {
		return nil, fmt.Errorf("failed to read users: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	users, err := parseUsers(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse users: %w", err)
	}

	d.log().Info("retrieved users", "count", len(users))
	return users, nil
}

// GetUser retrieves a specific user by ID.
func (d *Device) GetUser(userID int) (*User, error) {
	users, err := d.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if u.UserID == userID {
			return &u, nil
		}
	}

	return nil, ErrUserNotFound
}

// GetUserCount returns the number of users on the device.
func (d *Device) GetUserCount() (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return 0, ErrNotConnected
	}

	resp, err := d.sendCommand(protocol.CMD_GET_FREE_SIZES, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}

	if len(resp.Data) < 8 {
		return 0, ErrInvalidResponse
	}

	// User count is typically at offset 4
	count := int(binary.LittleEndian.Uint32(resp.Data[4:8]))
	return count, nil
}

// SetUser creates or updates a user on the device.
func (d *Device) SetUser(user User) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return ErrNotConnected
	}

	if user.UserID <= 0 {
		return fmt.Errorf("invalid user ID: %d", user.UserID)
	}

	data := encodeUser(user)

	_, err := d.sendCommand(protocol.CMD_USER_WRQ, data)
	if err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	d.log().Info("set user", "userID", user.UserID, "name", user.Name)
	return nil
}

// DeleteUser deletes a user from the device.
func (d *Device) DeleteUser(userID int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return ErrNotConnected
	}

	// Encode user ID
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, uint16(userID))

	_, err := d.sendCommand(protocol.CMD_DELETE_USER, data)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	d.log().Info("deleted user", "userID", userID)
	return nil
}

// DeleteAllUsers deletes all users from the device.
// WARNING: This permanently deletes all user data including fingerprints!
func (d *Device) DeleteAllUsers() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return ErrNotConnected
	}

	_, err := d.sendCommand(protocol.CMD_CLEAR_ADMIN, nil)
	if err != nil {
		return fmt.Errorf("failed to clear users: %w", err)
	}

	d.log().Info("deleted all users")
	return nil
}

// parseUsers parses raw user data from the device.
func parseUsers(data []byte) ([]User, error) {
	var users []User

	// Detect record size (commonly 28 or 72 bytes)
	recordSize := detectUserRecordSize(data)
	if recordSize == 0 {
		return nil, fmt.Errorf("unable to detect user record size")
	}

	for i := 0; i+recordSize <= len(data); i += recordSize {
		record := data[i : i+recordSize]
		user, err := parseUserRecord(record, recordSize)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// detectUserRecordSize attempts to detect the user record size.
func detectUserRecordSize(data []byte) int {
	// Common sizes: 28, 72 bytes
	sizes := []int{72, 28}

	for _, size := range sizes {
		if len(data)%size == 0 && len(data) >= size {
			return size
		}
	}

	// Default fallback
	if len(data) >= 28 {
		return 28
	}
	return 0
}

// parseUserRecord parses a single user record.
func parseUserRecord(record []byte, size int) (User, error) {
	var user User

	switch size {
	case 28:
		// Old format: UserID(2) + Privilege(1) + Password(8) + Name(8) + CardNo(4) + Group(1) + Timezone(2) + Flag(2)
		user.UserID = int(binary.LittleEndian.Uint16(record[0:2]))
		user.Privilege = int(record[2])
		user.Password = strings.TrimRight(string(record[3:11]), "\x00")
		user.Name = strings.TrimRight(string(record[11:19]), "\x00")
		user.CardNo = fmt.Sprintf("%d", binary.LittleEndian.Uint32(record[19:23]))
		user.GroupNo = int(record[23])
		user.Enabled = record[27]&0x01 != 0

	case 72:
		// New format (extended)
		// UserID(9 string) + Privilege(1) + Password(8) + Name(24) + CardNo(8) + Group(1) + ...
		userIDBytes := record[0:9]
		userIDStr := strings.TrimRight(string(userIDBytes), "\x00")
		var userID int
		fmt.Sscanf(userIDStr, "%d", &userID)
		user.UserID = userID

		user.Privilege = int(record[9])
		user.Password = strings.TrimRight(string(record[10:18]), "\x00")
		user.Name = strings.TrimRight(string(record[18:42]), "\x00")
		user.CardNo = strings.TrimRight(string(record[42:50]), "\x00")
		user.GroupNo = int(record[50])
		user.Enabled = true // Assume enabled in new format

	default:
		return user, fmt.Errorf("unsupported record size: %d", size)
	}

	// Validate
	if user.UserID <= 0 {
		return user, fmt.Errorf("invalid user ID")
	}

	// Clean name
	user.Name = strings.TrimSpace(user.Name)

	return user, nil
}

// encodeUser encodes a user for sending to the device.
// Format: UserID(2) + Privilege(1) + Password(8) + Name(8) + CardNo(4) + Group(1) + Timezone(2) + Flag(2) = 28 bytes
func encodeUser(user User) []byte {
	buf := make([]byte, 28)

	// UserID (2 bytes)
	binary.LittleEndian.PutUint16(buf[0:2], uint16(user.UserID))

	// Privilege (1 byte)
	buf[2] = byte(user.Privilege)

	// Password (8 bytes, null-padded)
	copy(buf[3:11], user.Password)

	// Name (8 bytes, null-padded)
	copy(buf[11:19], user.Name)

	// Card number (4 bytes)
	var cardNo uint32
	fmt.Sscanf(user.CardNo, "%d", &cardNo)
	binary.LittleEndian.PutUint32(buf[19:23], cardNo)

	// Group (1 byte)
	buf[23] = byte(user.GroupNo)

	// Timezone (2 bytes) - already zero

	// Flag (2 bytes) - enabled bit
	if user.Enabled {
		binary.LittleEndian.PutUint16(buf[26:28], 1)
	}

	return buf
}
