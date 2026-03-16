package zkteco

import "time"

// Options contains configuration options for the device connection.
type Options struct {
	// Timeout is the connection and read/write timeout
	Timeout time.Duration

	// Password is the device communication key (usually empty or "0")
	Password string

	// Logger is an optional logger for debug output
	Logger Logger

	// RetryCount is the number of retries on transient failures
	RetryCount int

	// RetryDelay is the delay between retries
	RetryDelay time.Duration

	// StrictChecksum enables strict checksum validation on received packets.
	// Some devices don't follow checksum strictly, so this is disabled by default.
	StrictChecksum bool
}

// Option is a function that configures Options.
type Option func(*Options)

// WithTimeout sets the connection and read/write timeout.
func WithTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.Timeout = d
	}
}

// WithPassword sets the device communication key.
func WithPassword(p string) Option {
	return func(o *Options) {
		o.Password = p
	}
}

// WithLogger sets a custom logger for debug output.
func WithLogger(l Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

// WithRetry sets the retry count and delay for transient failures.
func WithRetry(count int, delay time.Duration) Option {
	return func(o *Options) {
		o.RetryCount = count
		o.RetryDelay = delay
	}
}

// WithStrictChecksum enables strict checksum validation on received packets.
func WithStrictChecksum(strict bool) Option {
	return func(o *Options) {
		o.StrictChecksum = strict
	}
}

// Logger is the interface for custom loggers.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// nopLogger is a no-operation logger (default).
type nopLogger struct{}

func (nopLogger) Debug(msg string, args ...any) {}
func (nopLogger) Info(msg string, args ...any)  {}
func (nopLogger) Warn(msg string, args ...any)  {}
func (nopLogger) Error(msg string, args ...any) {}

// defaultLogger returns a no-op logger.
func defaultLogger() Logger {
	return nopLogger{}
}
