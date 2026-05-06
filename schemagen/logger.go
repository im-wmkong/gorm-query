package schemagen

import (
	"log"
	"os"
)

// Logger defines the logging interface used by the schemagen generator.
type Logger interface {
	// Debug logs debug-level messages.
	Debug(format string, a ...any)
	// Info logs informational messages for key operations.
	Info(format string, a ...any)
	// Warn logs warning messages for non-fatal conditions.
	Warn(format string, a ...any)
}

// defaultLogger is the standard-library log based implementation (stderr).
type defaultLogger struct {
	l *log.Logger
}

// DefaultLogger creates a default Logger that writes to stderr.
//
// Example:
//
//	g := schemagen.New(schemagen.WithLogger(schemagen.DefaultLogger()))
//	_ = g
func DefaultLogger() Logger {
	return &defaultLogger{l: log.New(os.Stderr, "[schemagen] ", log.LstdFlags)}
}

func (d *defaultLogger) Debug(format string, a ...any) { d.l.Printf("[DEBUG] "+format, a...) }
func (d *defaultLogger) Info(format string, a ...any)  { d.l.Printf("[INFO]  "+format, a...) }
func (d *defaultLogger) Warn(format string, a ...any)  { d.l.Printf("[WARN]  "+format, a...) }

// nopLogger discards all logs.
type nopLogger struct{}

// NopLogger returns a Logger that discards all logs.
//
// Example:
//
//	g := schemagen.New(schemagen.WithLogger(schemagen.NopLogger()))
//	_ = g
func NopLogger() Logger                         { return nopLogger{} }
func (nopLogger) Debug(format string, a ...any) { _ = format; _ = a }
func (nopLogger) Info(format string, a ...any)  { _ = format; _ = a }
func (nopLogger) Warn(format string, a ...any)  { _ = format; _ = a }
