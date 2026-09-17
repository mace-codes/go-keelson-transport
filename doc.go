// Package transport provides a pluggable HTTP transport layer for
// Keelson-based services. It defines a minimal, framework-agnostic
// routing port with interchangeable adapters for Chi and Gorilla Mux,
// so infrastructure can swap routers via configuration without
// leaking either dependency into application code.
package transport
