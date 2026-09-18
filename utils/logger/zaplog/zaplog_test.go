package zaplog_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/mace-codes/go-keelson-transport/utils/logger"
	"github.com/mace-codes/go-keelson-transport/utils/logger/zaplog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func newObserved() (logger.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.DebugLevel)
	return zaplog.New(zap.New(core)), logs
}

func TestLevelsMapToZap(t *testing.T) {
	tests := []struct {
		name  string
		log   func(logger.Logger)
		want  zapcore.Level
		wantM string
	}{
		{
			name:  "debug",
			log:   func(l logger.Logger) { l.Debug("d") },
			want:  zapcore.DebugLevel,
			wantM: "d",
		},
		{
			name:  "info",
			log:   func(l logger.Logger) { l.Info("i") },
			want:  zapcore.InfoLevel,
			wantM: "i",
		},
		{
			name:  "warn",
			log:   func(l logger.Logger) { l.Warn("w") },
			want:  zapcore.WarnLevel,
			wantM: "w",
		},
		{
			name:  "error",
			log:   func(l logger.Logger) { l.Error("e") },
			want:  zapcore.ErrorLevel,
			wantM: "e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, logs := newObserved()
			tt.log(log)

			all := logs.All()
			if len(all) != 1 {
				t.Fatalf("captured %d entries, want 1", len(all))
			}

			if all[0].Level != tt.want {
				t.Errorf("level = %v, want %v", all[0].Level, tt.want)
			}

			if all[0].Message != tt.wantM {
				t.Errorf("message = %q, want %q", all[0].Message, tt.wantM)
			}
		})
	}
}

func TestFieldsReachZap(t *testing.T) {
	log, logs := newObserved()

	log.Info("listening",
		logger.String("host", "0.0.0.0"),
		logger.Int("port", 8080),
		logger.Bool("tls", false),
	)

	all := logs.All()
	if len(all) != 1 {
		t.Fatalf("captured %d entries, want 1", len(all))
	}

	got := all[0].ContextMap()
	if got["host"] != "0.0.0.0" {
		t.Errorf("host = %v, want %q", got["host"], "0.0.0.0")
	}

	if got["port"] != int64(8080) {
		t.Errorf("port = %v, want %d", got["port"], 8080)
	}

	if got["tls"] != false {
		t.Errorf("tls = %v, want false", got["tls"])
	}
}

func TestErrFieldBecomesZapError(t *testing.T) {
	log, logs := newObserved()

	log.Error("failed", logger.Err(errors.New("boom")))

	got := logs.All()[0].ContextMap()
	if got[logger.ErrorKey] != "boom" {
		t.Errorf("%s = %v, want %q", logger.ErrorKey, got[logger.ErrorKey], "boom")
	}
}

func TestWithBindsFieldsToEveryEntry(t *testing.T) {
	log, logs := newObserved()

	child := log.With(logger.String("component", "transport"))
	child.Info("first")
	child.Warn("second", logger.Int("attempt", 2))

	all := logs.All()
	if len(all) != 2 {
		t.Fatalf("captured %d entries, want 2", len(all))
	}

	for i, e := range all {
		if got := e.ContextMap()["component"]; got != "transport" {
			t.Errorf("entry %d component = %v, want %q", i, got, "transport")
		}
	}

	if got := all[1].ContextMap()["attempt"]; got != int64(2) {
		t.Errorf("attempt = %v, want 2", got)
	}

	// The parent must not inherit the child's bound fields.
	log.Info("parent")

	if got, ok := logs.All()[2].ContextMap()["component"]; ok {
		t.Errorf("parent entry leaked component = %v", got)
	}
}

func TestWithNoFieldsReturnsSameLogger(t *testing.T) {
	log, logs := newObserved()

	log.With().Info("still logs")

	if len(logs.All()) != 1 {
		t.Fatalf("captured %d entries, want 1", len(logs.All()))
	}
}

func TestDefault(t *testing.T) {
	log, err := zaplog.Default()
	if err != nil {
		t.Fatalf("Default() error = %v, want nil", err)
	}

	if log == nil {
		t.Fatal("Default() = nil, want a zap-backed Logger")
	}

	// Must be the zap adapter, not the no-op logger — otherwise the transport
	// would silently default to discarding entries.
	if got, want := reflect.TypeOf(log), reflect.TypeOf(zaplog.New(zap.NewNop())); got != want {
		t.Errorf("Default() type = %s, want %s", got, want)
	}

	if got := reflect.TypeOf(log); got == reflect.TypeOf(logger.Noop()) {
		t.Errorf("Default() returned the no-op logger")
	}

	// Usable without further setup. This writes one line to stderr, which is
	// the point — the default is not silent.
	log.Info("default logger is usable", logger.String("test", t.Name()))
	log.With(logger.String("component", "transport")).Warn("child logger is usable")
}

func TestNilZapLoggerFallsBackToNoop(t *testing.T) {
	log := zaplog.New(nil)
	if log == nil {
		t.Fatal("New(nil) returned nil Logger")
	}

	// Must not panic.
	log.Info("discarded", logger.String("k", "v"))
	log.With(logger.Int("n", 1)).Error("also discarded")
}

func TestCallerPointsAtCallSiteNotAdapter(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	log := zaplog.New(zap.New(core, zap.AddCaller()))

	log.Info("from the test")
	log.With(logger.String("component", "transport")).Error("from a child")

	all := logs.All()
	if len(all) != 2 {
		t.Fatalf("captured %d entries, want 2", len(all))
	}

	// Without AddCallerSkip these would report zaplog.go, making the caller
	// field useless for every consumer of the port.
	for i, e := range all {
		if got := e.Caller.TrimmedPath(); !strings.HasPrefix(got, "zaplog/zaplog_test.go:") {
			t.Errorf("entry %d caller = %q, want this test file", i, got)
		}
	}
}
