package zkteco

import (
	"errors"
	"testing"
)

func TestDeviceError(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		err := &DeviceError{Code: 2001, Message: "access denied"}
		expected := "zkteco: device error: access denied"
		if err.Error() != expected {
			t.Errorf("got %q, want %q", err.Error(), expected)
		}
	})

	t.Run("without message", func(t *testing.T) {
		err := &DeviceError{Code: 2001}
		expected := "zkteco: device error code 2001"
		if err.Error() != expected {
			t.Errorf("got %q, want %q", err.Error(), expected)
		}
	})

	t.Run("code zero", func(t *testing.T) {
		err := &DeviceError{Code: 0}
		expected := "zkteco: device error code 0"
		if err.Error() != expected {
			t.Errorf("got %q, want %q", err.Error(), expected)
		}
	})
}

func TestErrorHelpers(t *testing.T) {
	if !IsNotConnected(ErrNotConnected) {
		t.Error("IsNotConnected should return true for ErrNotConnected")
	}
	if IsNotConnected(ErrTimeout) {
		t.Error("IsNotConnected should return false for ErrTimeout")
	}

	if !IsTimeout(ErrTimeout) {
		t.Error("IsTimeout should return true for ErrTimeout")
	}

	if !IsUserNotFound(ErrUserNotFound) {
		t.Error("IsUserNotFound should return true for ErrUserNotFound")
	}

	// Test wrapped errors
	wrapped := errors.Join(ErrNotConnected, errors.New("extra context"))
	if !IsNotConnected(wrapped) {
		t.Error("IsNotConnected should work with wrapped errors")
	}
}
