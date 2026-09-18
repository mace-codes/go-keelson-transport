package lrslog

import (
	"github.com/sirupsen/logrus"

	"github.com/mace-codes/go-keelson-transport/utils/logger"
)

// New adapts a *logrus.Logger to the logger.Logger port.
func New(l *logrus.Logger) logger.Logger {
	if l == nil {
		return logger.Noop()
	}

	return adapter{entry: logrus.NewEntry(l)}
}

// NewEntry adapts a *logrus.Entry, preserving any fields already attached to it.
func NewEntry(e *logrus.Entry) logger.Logger {
	if e == nil {
		return logger.Noop()
	}

	return adapter{entry: e}
}

type adapter struct {
	entry *logrus.Entry
}

func (a adapter) Debug(msg string, fields ...logger.Field) {
	a.withFields(fields).Debug(msg)
}

func (a adapter) Info(msg string, fields ...logger.Field) {
	a.withFields(fields).Info(msg)
}

func (a adapter) Warn(msg string, fields ...logger.Field) {
	a.withFields(fields).Warn(msg)
}

func (a adapter) Error(msg string, fields ...logger.Field) {
	a.withFields(fields).Error(msg)
}

func (a adapter) With(fields ...logger.Field) logger.Logger {
	if len(fields) == 0 {
		return a
	}

	return adapter{entry: a.withFields(fields)}
}

func (a adapter) withFields(fields []logger.Field) *logrus.Entry {
	if len(fields) == 0 {
		return a.entry
	}

	lfs := make(logrus.Fields, len(fields))
	for _, f := range fields {
		lfs[f.Key] = f.Value
	}

	return a.entry.WithFields(lfs)
}
