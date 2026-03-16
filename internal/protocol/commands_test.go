package protocol

import "testing"

func TestVerifyTypeString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{VERIFY_PASSWORD, "PASSWORD"},
		{VERIFY_FINGERPRINT, "FINGERPRINT"},
		{VERIFY_CARD, "CARD"},
		{VERIFY_FINGERPRINT_PASSWORD, "FINGERPRINT+PASSWORD"},
		{VERIFY_FINGERPRINT_CARD, "FINGERPRINT+CARD"},
		{VERIFY_CARD_PASSWORD, "CARD+PASSWORD"},
		{VERIFY_FINGERPRINT_CARD_PWD, "FINGERPRINT+CARD+PASSWORD"},
		{VERIFY_PALM, "PALM"},
		{VERIFY_FACE_FINGERPRINT, "FACE+FINGERPRINT"},
		{VERIFY_FACE_PASSWORD, "FACE+PASSWORD"},
		{VERIFY_FACE_CARD, "FACE+CARD"},
		{VERIFY_PALM_FINGERPRINT, "PALM+FINGERPRINT"},
		{VERIFY_FACE_FINGERPRINT_CARD, "FACE+FINGERPRINT+CARD"},
		{VERIFY_FACE_FINGERPRINT_PWD, "FACE+FINGERPRINT+PASSWORD"},
		{VERIFY_FINGER_VEIN, "FINGER_VEIN"},
		{VERIFY_FACE, "FACE"},
		{99, "UNKNOWN"},
	}

	for _, tt := range tests {
		result := VerifyTypeString(tt.input)
		if result != tt.expected {
			t.Errorf("VerifyTypeString(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestStateString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{STATE_CHECK_IN, "CHECK_IN"},
		{STATE_CHECK_OUT, "CHECK_OUT"},
		{STATE_BREAK_OUT, "BREAK_OUT"},
		{STATE_BREAK_IN, "BREAK_IN"},
		{STATE_OT_IN, "OT_IN"},
		{STATE_OT_OUT, "OT_OUT"},
		{99, "UNKNOWN"},
	}

	for _, tt := range tests {
		result := StateString(tt.input)
		if result != tt.expected {
			t.Errorf("StateString(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestPrivilegeString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{PRIVILEGE_USER, "USER"},
		{PRIVILEGE_ENROLLER, "ENROLLER"},
		{PRIVILEGE_MANAGER, "MANAGER"},
		{PRIVILEGE_ADMIN, "ADMIN"},
		{99, "UNKNOWN"},
	}

	for _, tt := range tests {
		result := PrivilegeString(tt.input)
		if result != tt.expected {
			t.Errorf("PrivilegeString(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
