package transport

import (
	"net/http"
	"testing"

	"github.com/mace-codes/go-keelson-transport/utils/logger"
)

type startupConfig struct {
	host string
	port int
}

func (c startupConfig) Host() string { return c.host }
func (c startupConfig) Port() int    { return c.port }

// startupLogger captures the fields of the last entry it received.
type startupLogger struct {
	msg   string
	calls int
	seen  map[string]any
}

func (s *startupLogger) Debug(msg string, fields ...logger.Field) { s.record(msg, fields) }
func (s *startupLogger) Info(msg string, fields ...logger.Field)  { s.record(msg, fields) }
func (s *startupLogger) Warn(msg string, fields ...logger.Field)  { s.record(msg, fields) }
func (s *startupLogger) Error(msg string, fields ...logger.Field) { s.record(msg, fields) }
func (s *startupLogger) With(...logger.Field) logger.Logger       { return s }

func (s *startupLogger) record(msg string, fields []logger.Field) {
	s.calls++
	s.msg = msg
	s.seen = make(map[string]any, len(fields))

	for _, f := range fields {
		s.seen[f.Key] = f.Value
	}
}

func TestLogStartup(t *testing.T) {
	tests := []struct {
		name        string
		config      startupConfig
		addr        string
		wantAddress string
	}{
		{
			name:        "wildcard host",
			config:      startupConfig{host: "0.0.0.0", port: 8080},
			addr:        "0.0.0.0:8080",
			wantAddress: "http://0.0.0.0:8080",
		},
		{
			name:        "loopback host",
			config:      startupConfig{host: "127.0.0.1", port: 3000},
			addr:        "127.0.0.1:3000",
			wantAddress: "http://127.0.0.1:3000",
		},
		{
			name:        "empty host binds all interfaces",
			config:      startupConfig{host: "", port: 80},
			addr:        ":80",
			wantAddress: "http://:80",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := &startupLogger{}
			tr := &Transport[startupConfig]{config: tt.config, logger: log, router: http.NewServeMux()}

			tr.logStartup(tt.addr)

			if log.calls != 1 {
				t.Fatalf("logged %d entries, want 1", log.calls)
			}

			if log.msg != "starting service" {
				t.Errorf("msg = %q, want %q", log.msg, "starting service")
			}

			if log.seen["address"] != tt.wantAddress {
				t.Errorf("address = %v, want %q", log.seen["address"], tt.wantAddress)
			}

			if log.seen["host"] != tt.config.host {
				t.Errorf("host = %v, want %q", log.seen["host"], tt.config.host)
			}

			if log.seen["port"] != tt.config.port {
				t.Errorf("port = %v, want %d", log.seen["port"], tt.config.port)
			}
		})
	}
}

func TestLogStartupWithNoopLoggerIsSilent(t *testing.T) {
	tr := &Transport[startupConfig]{
		config: startupConfig{host: "0.0.0.0", port: 8080},
		logger: logger.Noop(),
		router: http.NewServeMux(),
	}

	// Must not panic and must produce no output.
	tr.logStartup("0.0.0.0:8080")
}
