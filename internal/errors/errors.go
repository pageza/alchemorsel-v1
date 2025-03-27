package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Error represents a structured application error
type Error struct {
	Code     string
	Message  string
	Err      error
	Stack    string
	Fields   []zap.Field
	Severity zapcore.Level
}

// New creates a new Error with the given code and message
func New(code string, message string) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Severity: zapcore.ErrorLevel,
		Stack:    getStack(),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code string, message string) *Error {
	if err == nil {
		return nil
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		// If it's already our error type, update it
		appErr.Code = code
		appErr.Message = message
		appErr.Err = err
		return appErr
	}

	return &Error{
		Code:     code,
		Message:  message,
		Err:      err,
		Severity: zapcore.ErrorLevel,
		Stack:    getStack(),
	}
}

// WithFields adds zap fields to the error
func (e *Error) WithFields(fields ...zap.Field) *Error {
	e.Fields = append(e.Fields, fields...)
	return e
}

// WithSeverity sets the error severity level
func (e *Error) WithSeverity(level zapcore.Level) *Error {
	e.Severity = level
	return e
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap implements the errors.Unwrap interface
func (e *Error) Unwrap() error {
	return e.Err
}

// Log logs the error using zap logger
func (e *Error) Log(logger *zap.Logger) {
	// Add standard fields
	fields := append([]zap.Field{
		zap.String("error_code", e.Code),
		zap.String("error_message", e.Message),
		zap.String("stack_trace", e.Stack),
	}, e.Fields...)

	// Add wrapped error if present
	if e.Err != nil {
		fields = append(fields, zap.Error(e.Err))
	}

	// Log at appropriate severity level
	switch e.Severity {
	case zapcore.DebugLevel:
		logger.Debug("Application error", fields...)
	case zapcore.InfoLevel:
		logger.Info("Application error", fields...)
	case zapcore.WarnLevel:
		logger.Warn("Application error", fields...)
	case zapcore.ErrorLevel:
		logger.Error("Application error", fields...)
	case zapcore.DPanicLevel:
		logger.DPanic("Application error", fields...)
	case zapcore.PanicLevel:
		logger.Panic("Application error", fields...)
	case zapcore.FatalLevel:
		logger.Fatal("Application error", fields...)
	}
}

// getStack returns the current stack trace
func getStack() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var stack strings.Builder
	for {
		frame, more := frames.Next()
		stack.WriteString(fmt.Sprintf("%s:%d\n", frame.File, frame.Line))
		if !more {
			break
		}
	}
	return stack.String()
}

// Common error codes
const (
	ErrInternal     = "INTERNAL_ERROR"
	ErrValidation   = "VALIDATION_ERROR"
	ErrNotFound     = "NOT_FOUND"
	ErrUnauthorized = "UNAUTHORIZED"
	ErrForbidden    = "FORBIDDEN"
	ErrConflict     = "CONFLICT"
	ErrTimeout      = "TIMEOUT"
	ErrDatabase     = "DATABASE_ERROR"
	ErrNetwork      = "NETWORK_ERROR"
	ErrConfig       = "CONFIG_ERROR"
)

// Common error constructors
func NewInternalError(message string) *Error {
	return New(ErrInternal, message)
}

func NewValidationError(message string) *Error {
	return New(ErrValidation, message)
}

func NewNotFoundError(message string) *Error {
	return New(ErrNotFound, message)
}

func NewUnauthorizedError(message string) *Error {
	return New(ErrUnauthorized, message)
}

func NewForbiddenError(message string) *Error {
	return New(ErrForbidden, message)
}

func NewConflictError(message string) *Error {
	return New(ErrConflict, message)
}

func NewTimeoutError(message string) *Error {
	return New(ErrTimeout, message)
}

func NewDatabaseError(message string) *Error {
	return New(ErrDatabase, message)
}

func NewNetworkError(message string) *Error {
	return New(ErrNetwork, message)
}

func NewConfigError(message string) *Error {
	return New(ErrConfig, message)
}
