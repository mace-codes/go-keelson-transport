package logger_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mace-codes/go-keelson-transport/utils/logger"
)

func TestFieldConstructors(t *testing.T) {
	err := errors.New("boom")
	now := time.Now()

	tests := []struct {
		name  string
		field logger.Field
		want  logger.Field
	}{
		{
			name:  "string",
			field: logger.String("host", "0.0.0.0"),
			want:  logger.Field{Key: "host", Value: "0.0.0.0"},
		},
		{
			name:  "int",
			field: logger.Int("port", 8080),
			want:  logger.Field{Key: "port", Value: 8080},
		},
		{
			name:  "int64",
			field: logger.Int64("bytes", int64(64)),
			want:  logger.Field{Key: "bytes", Value: int64(64)},
		},
		{
			name:  "float64",
			field: logger.Float64("ratio", 0.5),
			want:  logger.Field{Key: "ratio", Value: 0.5},
		},
		{
			name:  "bool",
			field: logger.Bool("tls", true),
			want:  logger.Field{Key: "tls", Value: true},
		},
		{
			name:  "duration",
			field: logger.Duration("took", time.Second),
			want:  logger.Field{Key: "took", Value: time.Second},
		},
		{
			name:  "time",
			field: logger.Time("at", now),
			want:  logger.Field{Key: "at", Value: now},
		},
		{
			name:  "err uses ErrorKey",
			field: logger.Err(err),
			want:  logger.Field{Key: logger.ErrorKey, Value: err},
		},
		{
			name:  "any",
			field: logger.Any("meta", []string{"a"}),
			want:  logger.Field{Key: "meta", Value: []string{"a"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.field.Key != tt.want.Key {
				t.Errorf("Key = %q, want %q", tt.field.Key, tt.want.Key)
			}

			if got, want := fmtValue(tt.field.Value), fmtValue(tt.want.Value); got != want {
				t.Errorf("Value = %s, want %s", got, want)
			}
		})
	}
}

func TestNoopDiscardsEntries(t *testing.T) {
	log := logger.Noop()

	// Every method must be safe to call, including on a child logger.
	log.Debug("debug", logger.String("k", "v"))
	log.Info("info")
	log.Warn("warn", logger.Int("n", 1))
	log.Error("error", logger.Err(errors.New("boom")))

	child := log.With(logger.String("scope", "test"))
	if child == nil {
		t.Fatal("With returned nil Logger")
	}

	child.Info("from child")
}

func TestNoopSatisfiesLogger(t *testing.T) {
	var _ logger.Logger = logger.Noop()
}
