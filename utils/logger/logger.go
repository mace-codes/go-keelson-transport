package logger

// ErrorKey is the field key used by Err for the error value.
const ErrorKey = "error"

// Field is a single structured key/value pair attached to a log entry.
// Adapters translate it into their own library's native field type.
type Field struct {
	Key   string
	Value any
}

// Logger is the logging port used by the transport layer.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
}
