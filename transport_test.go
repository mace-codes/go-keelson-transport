package transport_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/go-chi/chi/v5"

	transport "github.com/mace-codes/go-keelson-transport"
	"github.com/mace-codes/go-keelson-transport/utils/logger"
)

type fakeConfig struct{}

func (fakeConfig) Host() string { return "0.0.0.0" }
func (fakeConfig) Port() int    { return 8080 }

type fakeDeps struct{}

func (fakeDeps) Critical() []transport.Reporter { return nil }
func (fakeDeps) Optional() []transport.Reporter { return nil }

func noRoutes(fakeDeps) ([]transport.RoutesConfig, error) { return nil, nil }

// capturingLogger records the fields of the last entry it received.
type capturingLogger struct {
	calls int
	msg   string
	seen  map[string]any
}

func (c *capturingLogger) Debug(msg string, fields ...logger.Field) { c.record(msg, fields) }
func (c *capturingLogger) Info(msg string, fields ...logger.Field)  { c.record(msg, fields) }
func (c *capturingLogger) Warn(msg string, fields ...logger.Field)  { c.record(msg, fields) }
func (c *capturingLogger) Error(msg string, fields ...logger.Field) { c.record(msg, fields) }
func (c *capturingLogger) With(...logger.Field) logger.Logger       { return c }

func (c *capturingLogger) record(msg string, fields []logger.Field) {
	c.calls++
	c.msg = msg
	c.seen = make(map[string]any, len(fields))

	for _, f := range fields {
		c.seen[f.Key] = f.Value
	}
}

func newTransport(t *testing.T, opts ...transport.Option) *transport.Transport[fakeConfig] {
	t.Helper()

	srv, err := transport.NewTransport(
		fakeDeps{}, fakeConfig{}, noRoutes, chi.NewRouter(), transport.RegisterChiRoutes, opts...,
	)
	if err != nil {
		t.Fatalf("NewTransport() error = %v, want nil", err)
	}

	return srv
}

func TestLoggerDefaultsToNoop(t *testing.T) {
	srv := newTransport(t)

	got := srv.Logger()
	if got == nil {
		t.Fatal("Logger() = nil, want the no-op default")
	}

	// Without WithLogger, the transport must stay silent by default —
	// comparing types is how an external test can tell the no-op logger
	// apart from a real adapter, since both are unexported.
	want := reflect.TypeOf(transport.NoopLogger())
	if gotType := reflect.TypeOf(got); gotType != want {
		t.Errorf("Logger() type = %s, want %s", gotType, want)
	}

	// The default must be usable without a nil check.
	got.Info("default logger is wired", logger.String("k", "v"))
}

func TestNoopLoggerExplicitMatchesDefault(t *testing.T) {
	// Supplying the no-op logger explicitly is equivalent to omitting
	// WithLogger altogether.
	srv := newTransport(t, transport.WithLogger(transport.NoopLogger()))

	want := reflect.TypeOf(transport.NoopLogger())
	if gotType := reflect.TypeOf(srv.Logger()); gotType != want {
		t.Errorf("Logger() type = %s, want %s", gotType, want)
	}
}

func TestWithLogger(t *testing.T) {
	log := &capturingLogger{}
	srv := newTransport(t, transport.WithLogger(log))

	if srv.Logger() != log {
		t.Errorf("Logger() = %v, want the supplied logger", srv.Logger())
	}
}

func TestWithLoggerIgnoresNil(t *testing.T) {
	log := &capturingLogger{}

	// A nil logger must not clobber an already-configured one.
	srv := newTransport(t, transport.WithLogger(log), transport.WithLogger(nil))

	if srv.Logger() != log {
		t.Errorf("Logger() = %v, want the supplied logger", srv.Logger())
	}
}

func TestNilOptionIsSkipped(t *testing.T) {
	srv := newTransport(t, nil, transport.WithLogger(&capturingLogger{}))

	if srv.Logger() == nil {
		t.Fatal("Logger() = nil, want the supplied logger")
	}
}

func TestRequestLoggerIsAssignableToAdapter(t *testing.T) {
	// The middleware must drop into RoutesConfig.Adapters without conversion.
	cfg := transport.RoutesConfig{
		Path:     "/widgets",
		Methods:  []string{http.MethodGet},
		Handler:  func(w http.ResponseWriter, r *http.Request) {},
		Adapters: []transport.Adapter{transport.RequestLogger(&capturingLogger{})},
	}

	if len(cfg.Adapters) != 1 {
		t.Fatalf("Adapters = %d, want 1", len(cfg.Adapters))
	}
}
