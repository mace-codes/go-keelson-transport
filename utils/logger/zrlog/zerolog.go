package zrlog

import (
	"github.com/rs/zerolog"

	"github.com/mace-codes/go-keelson-transport/utils/logger"
)

// New adapts a zerolog.Logger to the logger.Logger port.
func New(l zerolog.Logger) logger.Logger {
	return adapter{log: l}
}

type adapter struct {
	log zerolog.Logger
}

func (a adapter) Debug(msg string, fields ...logger.Field) {
	emit(a.log.Debug(), msg, fields)
}

func (a adapter) Info(msg string, fields ...logger.Field) {
	emit(a.log.Info(), msg, fields)
}

func (a adapter) Warn(msg string, fields ...logger.Field) {
	emit(a.log.Warn(), msg, fields)
}

func (a adapter) Error(msg string, fields ...logger.Field) {
	emit(a.log.Error(), msg, fields)
}

func (a adapter) With(fields ...logger.Field) logger.Logger {
	if len(fields) == 0 {
		return a
	}

	return adapter{log: a.log.With().Fields(convert(fields)).Logger()}
}

// emit attaches fields to an event and sends it. The event must always be
// consumed — zerolog leaks a pooled buffer for any event never dispatched.
func emit(e *zerolog.Event, msg string, fields []logger.Field) {
	if len(fields) > 0 {
		e = e.Fields(convert(fields))
	}

	e.Msg(msg)
}

func convert(fields []logger.Field) map[string]any {
	m := make(map[string]any, len(fields))
	for _, f := range fields {
		m[f.Key] = f.Value
	}

	return m
}
