# go-keelson-transport
Pluggable HTTP router adapter for Keelson-based services. Defines a minimal, framework-agnostic router port and provides interchangeable adapters for Gorilla Mux and Chi, so infrastructure can swap routers via config without leaking either dependency into application code. Part of the go-keelson architecture family.
