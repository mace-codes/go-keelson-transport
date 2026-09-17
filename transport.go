package transport

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mace-codes/go-keelson-transport/health"
	"github.com/mace-codes/go-keelson-transport/routes"
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
}

// ServeHTTP handles incoming HTTP requests by delegating them to the router.
func (t *Transport[C]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.router.ServeHTTP(w, r)
}

// NewTransport creates a new Transport instance with the provided configuration and router.
func NewTransport[D health.Dependencies, C TransportConfig](deps D, config C, routesFactory routes.Factory[D], router http.Handler, routesRegistrar routes.Registrar) (*Transport[C], error) {
	t := &Transport[C]{
		config: config,
		router: router,
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
	log.Printf("Starting Service on http://%s:%d\n", t.config.Host(), t.config.Port())
	return http.ListenAndServe(fmt.Sprintf("%s:%d", t.config.Host(), t.config.Port()), t)
}

// TODO: handle serving TLS w/ Cert and Key
