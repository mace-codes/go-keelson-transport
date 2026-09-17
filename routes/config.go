package routes

import (
	"net/http"
)

// RoutesConfig defines the configuration for a single route, including the HTTP method, path, handler, and any adapters to be applied to the handler.
type RoutesConfig struct {
	Path     string
	Methods  []string
	Handler  http.HandlerFunc
	Adapters []Adapter
}

// Factory is a function type that takes dependencies and returns a slice of RoutesConfig. It is used to create routes based on the provided dependencies.
type Factory[D any] func(deps D) ([]RoutesConfig, error)

// Registrar is a function type that takes a router and a slice of RoutesConfig, and registers the routes with the router.
type Registrar func(http.Handler, []RoutesConfig) error
