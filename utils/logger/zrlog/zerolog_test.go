package zrlog_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/rs/zerolog"

	"github.com/mace-codes/go-keelson-transport/utils/logger"
	"github.com/mace-codes/go-keelson-transport/utils/logger/zrlog"
)

// newBuffered returns an adapter writing newline-delimited JSON into buf.
func newBuffered() (logger.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	zl := zerolog.New(buf).Level(zerolog.DebugLevel)

	return zrlog.New(zl), buf
}

// decode parses the newline-delimited JSON zerolog wrote.
func decode(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	var entries []map[string]any

	dec := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	for dec.More() {
		var e map[string]any
		if err := dec.Decode(&e); err != nil {
			t.Fatalf("decoding zerolog output %q: %v", buf.String(), err)
		}

		entries = append(entries, e)
	}

	return entries
}

func TestLevelsMapToZerolog(t *testing.T) {
	tests := []struct {
		name  string
		log   func(logger.Logger)
		want  string
		wantM string
	}{
		{
			name:  "debug",
			log:   func(l logger.Logger) { l.Debug("d") },
			want:  "debug",
			wantM: "d",
		},
		{
			name:  "info",
			log:   func(l logger.Logger) { l.Info("i") },
			want:  "info",
			wantM: "i",
		},
		{
			name:  "warn",
			log:   func(l logger.Logger) { l.Warn("w") },
			want:  "warn",
			wantM: "w",
		},
		{
			name:  "error",
			log:   func(l logger.Logger) { l.Error("e") },
			want:  "error",
			wantM: "e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, buf := newBuffered()
			tt.log(log)

			entries := decode(t, buf)
			if len(entries) != 1 {
				t.Fatalf("captured %d entries, want 1", len(entries))
			}

			if entries[0]["level"] != tt.want {
				t.Errorf("level = %v, want %q", entries[0]["level"], tt.want)
			}

			if entries[0]["message"] != tt.wantM {
				t.Errorf("message = %v, want %q", entries[0]["message"], tt.wantM)
			}
		})
	}
}

func TestFieldsReachZerolog(t *testing.T) {
	log, buf := newBuffered()

	log.Info("listening",
		logger.String("host", "0.0.0.0"),
		logger.Int("port", 8080),
		logger.Bool("tls", false),
		logger.Err(errors.New("boom")),
	)

	entries := decode(t, buf)
	if len(entries) != 1 {
		t.Fatalf("captured %d entries, want 1", len(entries))
	}

	got := entries[0]
	if got["host"] != "0.0.0.0" {
		t.Errorf("host = %v, want %q", got["host"], "0.0.0.0")
	}

	if got["port"] != float64(8080) {
		t.Errorf("port = %v, want %d", got["port"], 8080)
	}

	if got["tls"] != false {
		t.Errorf("tls = %v, want false", got["tls"])
	}

	if got[logger.ErrorKey] != "boom" {
		t.Errorf("%s = %v, want %q", logger.ErrorKey, got[logger.ErrorKey], "boom")
	}
}

func TestWithBindsFieldsToEveryEntry(t *testing.T) {
	log, buf := newBuffered()

	child := log.With(logger.String("component", "transport"))
	child.Info("first")
	child.Warn("second", logger.Int("attempt", 2))
	log.Info("parent")

	entries := decode(t, buf)
	if len(entries) != 3 {
		t.Fatalf("captured %d entries, want 3", len(entries))
	}

	for i := range 2 {
		if got := entries[i]["component"]; got != "transport" {
			t.Errorf("entry %d component = %v, want %q", i, got, "transport")
		}
	}

	if got := entries[1]["attempt"]; got != float64(2) {
		t.Errorf("attempt = %v, want 2", got)
	}

	// The parent must not inherit the child's bound fields.
	if got, ok := entries[2]["component"]; ok {
		t.Errorf("parent entry leaked component = %v", got)
	}
}

func TestWithNoFieldsStillLogs(t *testing.T) {
	log, buf := newBuffered()

	log.With().Info("still logs")

	if entries := decode(t, buf); len(entries) != 1 {
		t.Fatalf("captured %d entries, want 1", len(entries))
	}
}

func TestDisabledLevelIsDropped(t *testing.T) {
	buf := &bytes.Buffer{}
	log := zrlog.New(zerolog.New(buf).Level(zerolog.WarnLevel))

	// Below the configured level: the event must be discarded, not written.
	log.Debug("dropped", logger.String("k", "v"))
	log.Info("dropped too")

	if buf.Len() != 0 {
		t.Errorf("wrote %q, want nothing", buf.String())
	}

	log.Warn("kept")

	if entries := decode(t, buf); len(entries) != 1 {
		t.Fatalf("captured %d entries, want 1", len(entries))
	}
}
