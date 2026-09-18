package zaplog

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/mace-codes/go-keelson-transport/utils/logger"
)

// New adapts a *zap.Logger to the logger.Logger port.
//
// One caller frame is skipped, so zap's "caller" field points at the code that
// called Info/Warn/etc. rather than at this adapter. Every entry passes through
// exactly one adapter method, so the skip is always correct — including on
// child loggers returned by With.
func New(l *zap.Logger) logger.Logger {
	if l == nil {
		return logger.Noop()
	}

	return adapter{log: l.WithOptions(zap.AddCallerSkip(1))}
}

// Default returns the transport's default logger: zap's production config —
// JSON encoding to stderr at Info level and above, with ISO8601 timestamps.
//
// It is what a Transport uses when no WithLogger option is supplied. Callers
// who want different behaviour build their own *zap.Logger and pass it to New.
//
// The error is zap's own; it surfaces only if the log sink cannot be opened.
func Default() (logger.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	l, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("building default zap logger: %w", err)
	}

	return New(l), nil
}

type adapter struct {
	log *zap.Logger
}

func (a adapter) Debug(msg string, fields ...logger.Field) {
	a.log.Debug(msg, convert(fields)...)
}

func (a adapter) Info(msg string, fields ...logger.Field) {
	a.log.Info(msg, convert(fields)...)
}

func (a adapter) Warn(msg string, fields ...logger.Field) {
	a.log.Warn(msg, convert(fields)...)
}

func (a adapter) Error(msg string, fields ...logger.Field) {
	a.log.Error(msg, convert(fields)...)
}

func (a adapter) With(fields ...logger.Field) logger.Logger {
	if len(fields) == 0 {
		return a
	}

	return adapter{log: a.log.With(convert(fields)...)}
}

// convert maps port fields onto zap fields. zap.Any already dispatches on the
// concrete type, including error values, so typed keys survive the round trip.
func convert(fields []logger.Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	zfs := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zfs = append(zfs, zap.Any(f.Key, f.Value))
	}

	return zfs
}
