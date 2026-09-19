[![CI](https://github.com/mace-codes/go-keelson-transport/actions/workflows/ci.yml/badge.svg)](https://github.com/mace-codes/go-keelson-transport/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mace-codes/go-keelson-transport.svg)](https://pkg.go.dev/github.com/mace-codes/go-keelson-transport)
[![License](https://img.shields.io/github/license/mace-codes/go-keelson-transport)](LICENSE)

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
| `logger.Logger` | `utils/logger` | Logging port — leveled methods plus structured `Field`s; adapters for zap, zerolog and logrus provided |
| `rest.RespondJSON` | `utils/rest` | Uniform JSON response helper, including error-shaped bodies |

`export.go` re-exports the public surface (`Reporter`, `RoutesConfig`, `Adapter`, `Logger`, `Field`, `RegisterChiRoutes`, `RegisterGMuxROutes`, `RequestLogger`, `DefaultLogger`, `NoopLogger`) so consumers import a single package rather than reaching into internals.

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
	"github.com/mace-codes/go-keelson-transport/utils/logger/zaplog"
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
		transport.WithLogger(zaplog.New(deps.logger)), // optional — omit for the zap default
	)
	if err != nil {
		panic(err)
	}
	panic(srv.ListenAndServe())
}
```

Swapping to Gorilla Mux is a two-argument change: `mux.NewRouter()` + `transport.RegisterGMuxRoutes`.

## Logging

Logging follows the same ports-and-adapters shape as routing: `utils/logger` declares the port, and each logging library gets an adapter in its own subpackage.

**zap is the default.** Omit `WithLogger` and the transport builds its own production zap logger — JSON to stderr at `Info` and above, ISO8601 timestamps. zerolog and logrus are one-line swaps; silence is opt-in.

```go
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
}
```

The default lives in the zap adapter as `zaplog.Default()`, re-exported as `transport.DefaultLogger`. To use anything else, pick an adapter at the composition root:

| Library | Adapter | Constructor |
|---|---|---|
| [zap](https://github.com/uber-go/zap) | `utils/logger/zaplog` | `zaplog.New(*zap.Logger)` |
| [zerolog](https://github.com/rs/zerolog) | `utils/logger/zrlog` | `zrlog.New(zerolog.Logger)` |
| [logrus](https://github.com/sirupsen/logrus) | `utils/logger/lrslog` | `lrslog.New(*logrus.Logger)` / `lrslog.NewEntry(*logrus.Entry)` |

```go
import "github.com/mace-codes/go-keelson-transport/utils/logger/zaplog"

srv, err := transport.NewTransport(
	deps, cfg, routes, chi.NewRouter(), transport.RegisterChiRoutes,
	transport.WithLogger(zaplog.New(zl)),
)
```

Swapping logger is a one-line change — `zrlog.New(zl)` or `lrslog.New(l)` in place of `zaplog.New`. Writing a fourth adapter means implementing five methods against your own library.

To silence the transport, pass the no-op logger explicitly:

```go
transport.WithLogger(transport.NoopLogger())
```

Each adapter skips its own stack frame where the library supports it, so zap's `caller` field reports your call site rather than `zaplog.go`. zerolog and logrus both disable caller reporting by default; if you enable it, expect the adapter frame to show — logrus offers no skip API, and zerolog's needs configuring on the logger you pass in.

Field constructors live in `utils/logger` (`String`, `Int`, `Int64`, `Float64`, `Bool`, `Duration`, `Time`, `Err`, `Any`) — they aren't re-exported from the root package, since names that generic would crowd it.

### Request logging

`logger.RequestLogger` is opt-in middleware logging method, path, status, bytes and duration per request — at `Error` level for 5xx, `Info` otherwise. Its signature matches `routes.Adapter`, so it composes with your own middleware in declared order:

```go
routes.RoutesConfig{
	Path:     "/widgets",
	Methods:  []string{http.MethodGet},
	Handler:  widgets.List(deps),
	Adapters: []routes.Adapter{logger.RequestLogger(log), authMiddleware},
}
```

It is never registered automatically, including on the built-in health routes — add it per route where you want it.

## Notes for reviewers

A few items worth flagging before this is treated as load-bearing infrastructure:

- **TLS is unimplemented.** `ListenAndServe` carries a `// TODO: handle serving TLS w/ Cert and Key` — currently plaintext-only; services likely terminate TLS upstream (LB/ingress), but that assumption should be stated explicitly rather than left implicit.
- **No middleware/timeout defaults** — `Adapter` composition is entirely the caller's responsibility; there's no built-in request timeout or panic recovery. Request logging is now *available* (`logger.RequestLogger`) but still opt-in per route, never registered by default.
- **zap is a hard dependency; zerolog and logrus are not.** Because zap is the default, the root package imports `zaplog`, so every consumer links zap (and picks it up as an indirect `require`) whether they use it or not. zerolog and logrus stay isolated: a consumer that imports neither `zrlog` nor `lrslog` links zero packages from them, and module pruning keeps them out of that consumer's `go.mod` entirely. Making zap optional again would mean either dropping the default or a build-tag/nested-module split.
