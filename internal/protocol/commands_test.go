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
