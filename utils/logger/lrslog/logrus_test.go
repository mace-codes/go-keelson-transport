package lrslog_test

import (
	"errors"
	"testing"

	"github.com/mace-codes/go-keelson-transport/utils/logger"
	"github.com/mace-codes/go-keelson-transport/utils/logger/lrslog"
	"github.com/sirupsen/logrus"

	logrustest "github.com/sirupsen/logrus/hooks/test"
)

// newHooked returns an adapter whose entries are captured by a test hook.
func newHooked() (logger.Logger, *logrustest.Hook) {
	base, hook := logrustest.NewNullLogger()
	base.SetLevel(logrus.DebugLevel)

	return lrslog.New(base), hook
}

func TestLevelsMapToLogrus(t *testing.T) {
	tests := []struct {
		name  string
		log   func(logger.Logger)
		want  logrus.Level
		wantM string
	}{
		{
			name:  "debug",
			log:   func(l logger.Logger) { l.Debug("d") },
			want:  logrus.DebugLevel,
			wantM: "d",
		},
		{
			name:  "info",
			log:   func(l logger.Logger) { l.Info("i") },
			want:  logrus.InfoLevel,
			wantM: "i",
		},
		{
			name:  "warn",
			log:   func(l logger.Logger) { l.Warn("w") },
			want:  logrus.WarnLevel,
			wantM: "w",
		},
		{
			name:  "error",
			log:   func(l logger.Logger) { l.Error("e") },
			want:  logrus.ErrorLevel,
			wantM: "e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, hook := newHooked()
			tt.log(log)

			if len(hook.Entries) != 1 {
				t.Fatalf("captured %d entries, want 1", len(hook.Entries))
			}

			if hook.Entries[0].Level != tt.want {
				t.Errorf("level = %v, want %v", hook.Entries[0].Level, tt.want)
			}

			if hook.Entries[0].Message != tt.wantM {
				t.Errorf("message = %q, want %q", hook.Entries[0].Message, tt.wantM)
			}
		})
	}
}

func TestFieldsReachLogrus(t *testing.T) {
	log, hook := newHooked()

	err := errors.New("boom")
	log.Info("listening",
		logger.String("host", "0.0.0.0"),
		logger.Int("port", 8080),
		logger.Bool("tls", false),
		logger.Err(err),
	)

	if len(hook.Entries) != 1 {
		t.Fatalf("captured %d entries, want 1", len(hook.Entries))
	}

	got := hook.Entries[0].Data
	if got["host"] != "0.0.0.0" {
		t.Errorf("host = %v, want %q", got["host"], "0.0.0.0")
	}

	if got["port"] != 8080 {
		t.Errorf("port = %v, want %d", got["port"], 8080)
	}

	if got["tls"] != false {
		t.Errorf("tls = %v, want false", got["tls"])
	}

	if got[logger.ErrorKey] != err {
		t.Errorf("%s = %v, want %v", logger.ErrorKey, got[logger.ErrorKey], err)
	}
}

func TestWithBindsFieldsToEveryEntry(t *testing.T) {
	log, hook := newHooked()

	child := log.With(logger.String("component", "transport"))
	child.Info("first")
	child.Warn("second", logger.Int("attempt", 2))
	log.Info("parent")

	if len(hook.Entries) != 3 {
		t.Fatalf("captured %d entries, want 3", len(hook.Entries))
	}

	for i := range 2 {
		if got := hook.Entries[i].Data["component"]; got != "transport" {
			t.Errorf("entry %d component = %v, want %q", i, got, "transport")
		}
	}

	if got := hook.Entries[1].Data["attempt"]; got != 2 {
		t.Errorf("attempt = %v, want 2", got)
	}

	// The parent must not inherit the child's bound fields.
	if got, ok := hook.Entries[2].Data["component"]; ok {
		t.Errorf("parent entry leaked component = %v", got)
	}
}

func TestWithNoFieldsStillLogs(t *testing.T) {
	log, hook := newHooked()

	log.With().Info("still logs")

	if len(hook.Entries) != 1 {
		t.Fatalf("captured %d entries, want 1", len(hook.Entries))
	}
}

func TestNewEntryPreservesExistingFields(t *testing.T) {
	base, hook := logrustest.NewNullLogger()
	entry := base.WithField("service", "widgets")

	lrslog.NewEntry(entry).Info("hello", logger.Int("port", 8080))

	if len(hook.Entries) != 1 {
		t.Fatalf("captured %d entries, want 1", len(hook.Entries))
	}

	got := hook.Entries[0].Data
	if got["service"] != "widgets" {
		t.Errorf("service = %v, want %q", got["service"], "widgets")
	}

	if got["port"] != 8080 {
		t.Errorf("port = %v, want %d", got["port"], 8080)
	}
}

func TestNilLoggersFallBackToNoop(t *testing.T) {
	tests := []struct {
		name string
		log  logger.Logger
	}{
		{name: "nil *logrus.Logger", log: lrslog.New(nil)},
		{name: "nil *logrus.Entry", log: lrslog.NewEntry(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.log == nil {
				t.Fatal("returned nil Logger")
			}

			// Must not panic.
			tt.log.Info("discarded", logger.String("k", "v"))
			tt.log.With(logger.Int("n", 1)).Error("also discarded")
		})
	}
}
