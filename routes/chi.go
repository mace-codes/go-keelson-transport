package routes

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterChiRoutes registers the Chi routes for the given router. It takes a router and a list of routes, and registers each route with the router.
func RegisterChiRoutes(router http.Handler, routes []RoutesConfig) error {
	chir, ok := router.(*chi.Mux)
	if !ok {
		return fmt.Errorf("router is not a *chi.Mux")
	}

	for _, route := range routes {
		hdlr, err := chiMuxAdapter(route.Handler, route.Adapters...)
		if err != nil {
			return fmt.Errorf("failed to adapt handler for route %s: %w", route.Path, err)
		}

		for _, method := range route.Methods {
			chir.Method(method, route.Path, hdlr)
		}
	}

	return nil
}
