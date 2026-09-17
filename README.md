# go-keelson-transport

A pluggable HTTP transport layer for Keelson-based services. It defines a minimal, framework-agnostic **port** for routing and health-checking, with interchangeable **adapters** for Chi and Gorilla Mux — so infrastructure can swap routers via configuration without leaking either dependency into application code.

Part of the `go-keelson` architecture family (DDD / hexagonal — ports & adapters).

## Why this exists

Services accumulate hard dependencies on their router of choice — handler signatures, middleware chains, and route registration all end up coupled to (say) Chi's API. This package inverts that: application code depends on the `routes.Registrar` / `routes.Factory` **interfaces**, not on any router implementation. The router is supplied at the composition root (`main.go`), and swapping Chi for Gorilla Mux — or adding a third adapter — requires zero changes to business logic.

## Core concepts

| Concept | Package | Role |
|---|---|---|
| `Transport[C]` | `transport` | Generic wrapper around an `http.Handler`; owns config, wires routes, exposes `ListenAndServe` |
| `routes.Factory[D]` | `routes` | `func(deps D) ([]RoutesConfig, error)` — the application supplies its own routes |
| `routes.Registrar` | `routes` | `func(http.Handler, []RoutesConfig) error` — binds `RoutesConfig` to a concrete router (Chi/GMux provided) |
| `routes.Adapter` | `routes` | Middleware wrapper, applied in declared order regardless of which router is behind it |
| `health.Dependencies` | `health` | `Critical() []Reporter` / `Optional() []Reporter` — feeds the built-in health endpoints |
| `health.Reporter` | `health` | Single-method interface (`Health() ReporterResponse`) any dependency implements to report its own status |
| `rest.RespondJSON` | `utils/rest` | Uniform JSON response helper, including error-shaped bodies |

`export.go` re-exports the public surface (`Reporter`, `RoutesConfig`, `Adapter`, `RegisterChiRoutes`, `RegisterGMuxROutes`) so consumers import a single package rather than reaching into internals.

## Built-in health endpoints

`NewTransport` registers three routes automatically, on top of whatever the application's `Factory` returns:

| Path | Purpose |
|---|---|
| `GET /health` | Aggregates all `Critical()` and `Optional()` reporters into one status payload |
| `GET /health/is-live` | Static liveness probe — no dependency checks |
| `GET /health/ready` | Readiness probe gated on critical dependencies |

Overall `/health` status is `healthy`, `degraded` (an optional dependency is down), or `unhealthy` (a critical one is down).

## Quick start

```go
package main

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	transport "github.com/mace-codes/go-keelson-transport"
)

func main() {
	cfg := config{host: "0.0.0.0", port: 8080}

	deps := dependencies{
		logger:           zap.NewExample(),
		criticalServices: []transport.Reporter{db{}},
		optionalServices: []transport.Reporter{cache{}},
	}

	srv, err := transport.NewTransport(
		deps, cfg, routes, chi.NewRouter(), transport.RegisterChiRoutes,
	)
	if err != nil {
		panic(err)
	}
	panic(srv.ListenAndServe())
}
```

Swapping to Gorilla Mux is a two-argument change: `mux.NewRouter()` + `transport.RegisterGMuxROutes`.

## Notes for reviewers

A few items worth flagging before this is treated as load-bearing infrastructure:

- **TLS is unimplemented.** `ListenAndServe` carries a `// TODO: handle serving TLS w/ Cert and Key` — currently plaintext-only; services likely terminate TLS upstream (LB/ingress), but that assumption should be stated explicitly rather than left implicit.
- **`health.IsReady` has a control-flow bug.** On an unhealthy critical dependency it calls `RespondJSON` but does not `return`, then falls through and calls `RespondJSON` again with `"ok"`. The second call double-writes the response (a logged "superfluous WriteHeader" warning, with only the first status code actually honored by the client) — net effect, an unhealthy critical dependency doesn't reliably fail readiness. Worth a fix before this gates deploys/orchestration.
- **No middleware/timeout defaults** — `Adapter` composition is entirely the caller's responsibility; there's no built-in request timeout, panic recovery, or logging middleware supplied out of the box.
