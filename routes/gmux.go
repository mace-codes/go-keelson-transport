package routes

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterGMUXRoutes registers the GMUX routes for the given router. It takes a router and a list of routes, and registers each route with the router.
func RegisterGMUXRoutes(router http.Handler, routes []RoutesConfig) error {
	mxr, ok := router.(*mux.Router)
	if !ok {
		return fmt.Errorf("router is not a *mux.Router")
	}

	for _, route := range routes {
		hdlr, err := GMuxAdapter(route.Handler, route.Adapters...)
		if err != nil {
			return fmt.Errorf("failed to adapt handler for route %s: %w", route.Path, err)
		}
		mxr.HandleFunc(route.Path, hdlr).Methods(route.Methods...)
	}

	return nil
}
