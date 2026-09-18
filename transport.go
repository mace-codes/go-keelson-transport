package transport

import (
	"fmt"
	"net/http"

	"github.com/mace-codes/go-keelson-transport/health"
	"github.com/mace-codes/go-keelson-transport/routes"
	"github.com/mace-codes/go-keelson-transport/utils/logger"
)

// TransportConfig interface is used by a Transport to configure host and port for the server.
type TransportConfig interface {
	Host() string
	Port() int
}

// Transport defines the properties of the http transport layer, including the host and port to listen on, and the handler to serve requests.
type Transport[C TransportConfig] struct {
	config TransportConfig
	router http.Handler
	logger logger.Logger
}

// options holds the optional settings a Transport can be constructed with.
type options struct {
	logger logger.Logger
}

// Option configures a Transport at construction time. Options are applied in the order they are passed to NewTransport.
type Option func(*options)

// WithLogger supplies the Logger the transport logs through. Without it the
// transport is silent — no logging implementation is imposed on consumers.
// A nil logger is ignored, leaving the default in place.
//
// Adapters for zap, zerolog and logrus live under utils/logger:
//
//	transport.WithLogger(zaplog.New(zl))
func WithLogger(l logger.Logger) Option {
	return func(o *options) {
		if l == nil {
			return
		}

		o.logger = l
	}
}

// Logger returns the Logger the transport was built with, or a no-op Logger if
// none was supplied. It never returns nil, so callers can log unconditionally.
func (t *Transport[C]) Logger() logger.Logger {
	return t.logger
}

// ServeHTTP handles incoming HTTP requests by delegating them to the router.
func (t *Transport[C]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

// NewTransport creates a new Transport instance with the provided configuration and router.
func NewTransport[D health.Dependencies, C TransportConfig](deps D, config C, routesFactory routes.Factory[D], router http.Handler, routesRegistrar routes.Registrar, opts ...Option) (*Transport[C], error) {
	o := options{logger: logger.Noop()}
	for _, opt := range opts {
		if opt == nil {
			continue
		}

		opt(&o)
	}

	t := &Transport[C]{
		config: config,
		router: router,
		logger: o.logger,
	}

	rts, err := routesFactory(deps)
	if err != nil {
		return nil, err
	}

	defaultHealthRoutes := []routes.RoutesConfig{
		{
			Path:     "/health",
			Methods:  []string{http.MethodGet},
			Handler:  health.Health(deps),
			Adapters: nil,
		},
		{
			Path:     "/health/is-live",
			Methods:  []string{http.MethodGet},
			Handler:  health.IsLive(),
			Adapters: nil,
		},
		{
			Path:    "/health/ready",
			Methods: []string{http.MethodGet},
			Handler: health.IsReady(deps),
		},
	}

	rts = append(rts, defaultHealthRoutes...)

	err = routesRegistrar(t.router, rts)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// ListenAndServe starts a server on the given host and port, and serves requests using the provided handler.
func (t *Transport[C]) ListenAndServe() error {
	addr := fmt.Sprintf("%s:%d", t.config.Host(), t.config.Port())
	t.logStartup(addr)

	return http.ListenAndServe(addr, t)
}

// logStartup emits the single startup entry. Split out from ListenAndServe so
// it can be asserted on without binding a socket.
func (t *Transport[C]) logStartup(addr string) {
	t.logger.Info("starting service",
		logger.String("address", "http://"+addr),
		logger.String("host", t.config.Host()),
		logger.Int("port", t.config.Port()),
	)
}

// TODO: handle serving TLS w/ Cert and Key
