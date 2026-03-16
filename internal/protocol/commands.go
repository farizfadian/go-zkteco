package protocol

// ZKTeco command constants
const (
	// Connection commands
	CMD_CONNECT        uint16 = 1000 // Establish connection
	CMD_EXIT           uint16 = 1001 // Close connection
	CMD_ENABLE_DEVICE  uint16 = 1002 // Enable device
	CMD_DISABLE_DEVICE uint16 = 1003 // Disable device (stops capturing)
	CMD_RESTART        uint16 = 1004 // Restart device
	CMD_POWEROFF       uint16 = 1005 // Power off device

	// Response commands
	CMD_ACK_OK      uint16 = 2000 // Command successful
	CMD_ACK_ERROR   uint16 = 2001 // Command failed
	CMD_ACK_DATA    uint16 = 2002 // Data acknowledgment
	CMD_ACK_RETRY   uint16 = 2003 // Retry command
	CMD_ACK_REPEAT  uint16 = 2004 // Repeat command
	CMD_ACK_UNAUTH  uint16 = 2005 // Unauthorized

	// Data transfer commands
	CMD_PREPARE_DATA uint16 = 1500 // Prepare bulk data transfer
	CMD_DATA         uint16 = 1501 // Bulk data packet
	CMD_FREE_DATA    uint16 = 1502 // Free bulk data buffer

	// Device info commands
	CMD_GET_TIME       uint16 = 201 // Get device time
	CMD_SET_TIME       uint16 = 202 // Set device time
	CMD_OPTIONS_RRQ    uint16 = 11  // Read device options
	CMD_OPTIONS_WRQ    uint16 = 12  // Write device options
	CMD_INFO_RRQ       uint16 = 17  // Read device info (serial, etc.)
	CMD_GET_FREE_SIZES uint16 = 50  // Get free storage sizes

	// Attendance commands
	CMD_ATTLOG_RRQ   uint16 = 13 // Read attendance logs
	CMD_CLEAR_ATTLOG uint16 = 15 // Clear attendance logs

	// User commands
	CMD_USER_WRQ     uint16 = 8  // Write user info
	CMD_USERINFO_RRQ uint16 = 9  // Read all users info
	CMD_DELETE_USER  uint16 = 18 // Delete user
	CMD_DELETE_ADMIN uint16 = 20 // Delete admin privilege
	CMD_CLEAR_ADMIN  uint16 = 19 // Clear all admin privileges

	// User template commands (fingerprint/face)
	CMD_USERTEMP_RRQ uint16 = 88 // Read user template
	CMD_USERTEMP_WRQ uint16 = 87 // Write user template
	CMD_DELETE_TEMP  uint16 = 89 // Delete user template

	// Real-time event commands
	CMD_REG_EVENT uint16 = 500 // Register for real-time events

	// Miscellaneous
	CMD_QUERY_USERCOUNT uint16 = 16  // Query user count
	CMD_QUERY_LOGCOUNT  uint16 = 22  // Query log count
	CMD_WRITE_LCD       uint16 = 66  // Write to LCD screen
	CMD_CLEAR_LCD       uint16 = 67  // Clear LCD screen
	CMD_BEEP            uint16 = 44  // Make beep sound
	CMD_GET_VERSION     uint16 = 1100 // Get firmware version
)

// Verify types (how the user was verified)
const (
	VERIFY_PASSWORD               = 0
	VERIFY_FINGERPRINT            = 1
	VERIFY_CARD                   = 2
	VERIFY_FINGERPRINT_PASSWORD   = 3
	VERIFY_FINGERPRINT_CARD       = 4
	VERIFY_CARD_PASSWORD          = 5
	VERIFY_FINGERPRINT_CARD_PWD   = 6
	VERIFY_PALM                   = 7
	VERIFY_FACE_FINGERPRINT       = 8
	VERIFY_FACE_PASSWORD          = 9
	VERIFY_FACE_CARD              = 10
	VERIFY_PALM_FINGERPRINT       = 11
	VERIFY_FACE_FINGERPRINT_CARD  = 12
	VERIFY_FACE_FINGERPRINT_PWD   = 13
	VERIFY_FINGER_VEIN            = 14
	VERIFY_FACE                   = 15
)

// Attendance states
const (
	STATE_CHECK_IN  = 0
	STATE_CHECK_OUT = 1
	STATE_BREAK_OUT = 2
	STATE_BREAK_IN  = 3
	STATE_OT_IN     = 4
	STATE_OT_OUT    = 5
)

// User privileges
const (
	PRIVILEGE_USER     = 0
	PRIVILEGE_ENROLLER = 1
	PRIVILEGE_MANAGER  = 2
	PRIVILEGE_ADMIN    = 14
)

// VerifyTypeString returns a human-readable string for the verify type.
func VerifyTypeString(verifyType int) string {
	switch verifyType {
	case VERIFY_PASSWORD:
		return "PASSWORD"
	case VERIFY_FINGERPRINT:
		return "FINGERPRINT"
	case VERIFY_CARD:
		return "CARD"
	case VERIFY_FINGERPRINT_PASSWORD:
		return "FINGERPRINT+PASSWORD"
	case VERIFY_FINGERPRINT_CARD:
		return "FINGERPRINT+CARD"
	case VERIFY_CARD_PASSWORD:
		return "CARD+PASSWORD"
	case VERIFY_FINGERPRINT_CARD_PWD:
		return "FINGERPRINT+CARD+PASSWORD"
	case VERIFY_PALM:
		return "PALM"
	case VERIFY_FACE_FINGERPRINT:
		return "FACE+FINGERPRINT"
	case VERIFY_FACE_PASSWORD:
		return "FACE+PASSWORD"
	case VERIFY_FACE_CARD:
		return "FACE+CARD"
	case VERIFY_PALM_FINGERPRINT:
		return "PALM+FINGERPRINT"
	case VERIFY_FACE_FINGERPRINT_CARD:
		return "FACE+FINGERPRINT+CARD"
	case VERIFY_FACE_FINGERPRINT_PWD:
		return "FACE+FINGERPRINT+PASSWORD"
	case VERIFY_FINGER_VEIN:
		return "FINGER_VEIN"
	case VERIFY_FACE:
		return "FACE"
	default:
		return "UNKNOWN"
	}
}

// StateString returns a human-readable string for the attendance state.
func StateString(state int) string {
	switch state {
	case STATE_CHECK_IN:
		return "CHECK_IN"
	case STATE_CHECK_OUT:
		return "CHECK_OUT"
	case STATE_BREAK_OUT:
		return "BREAK_OUT"
	case STATE_BREAK_IN:
		return "BREAK_IN"
	case STATE_OT_IN:
		return "OT_IN"
	case STATE_OT_OUT:
		return "OT_OUT"
	default:
		return "UNKNOWN"
	}
}

// PrivilegeString returns a human-readable string for the user privilege.
func PrivilegeString(privilege int) string {
	switch privilege {
	case PRIVILEGE_USER:
		return "USER"
	case PRIVILEGE_ENROLLER:
		return "ENROLLER"
	case PRIVILEGE_MANAGER:
		return "MANAGER"
	case PRIVILEGE_ADMIN:
		return "ADMIN"
	default:
		return "UNKNOWN"
	}
}
