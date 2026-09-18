package logger_test

import (
	"fmt"

	"github.com/mace-codes/go-keelson-transport/utils/logger"
)

// entry is a single captured log entry.
type entry struct {
	level  string
	msg    string
	fields map[string]any
}

// recordingLogger is a logger.Logger that captures entries in memory, so tests
// can assert on level, message and fields without parsing encoded output.
type recordingLogger struct {
	entries *[]entry
	bound   []logger.Field
}

func newRecordingLogger() *recordingLogger {
	return &recordingLogger{entries: &[]entry{}}
}

func (r *recordingLogger) Debug(msg string, fields ...logger.Field) { r.log("debug", msg, fields) }
func (r *recordingLogger) Info(msg string, fields ...logger.Field)  { r.log("info", msg, fields) }
func (r *recordingLogger) Warn(msg string, fields ...logger.Field)  { r.log("warn", msg, fields) }
func (r *recordingLogger) Error(msg string, fields ...logger.Field) { r.log("error", msg, fields) }

func (r *recordingLogger) With(fields ...logger.Field) logger.Logger {
	bound := make([]logger.Field, 0, len(r.bound)+len(fields))
	bound = append(bound, r.bound...)
	bound = append(bound, fields...)

	return &recordingLogger{entries: r.entries, bound: bound}
}

func (r *recordingLogger) log(level, msg string, fields []logger.Field) {
	merged := make(map[string]any, len(r.bound)+len(fields))
	for _, f := range r.bound {
		merged[f.Key] = f.Value
	}
	for _, f := range fields {
		merged[f.Key] = f.Value
	}

	*r.entries = append(*r.entries, entry{level: level, msg: msg, fields: merged})
}

func (r *recordingLogger) captured() []entry {
	return *r.entries
}

// fmtValue renders a field value for comparison, so values of any type
// (including slices, which are not comparable with ==) can be asserted on.
func fmtValue(v any) string {
	return fmt.Sprintf("%T(%v)", v, v)
}
