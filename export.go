package transport

import (
	"github.com/mace-codes/go-keelson-transport/health"
	"github.com/mace-codes/go-keelson-transport/routes"
	"github.com/mace-codes/go-keelson-transport/utils/logger"
	"github.com/mace-codes/go-keelson-transport/utils/logger/zaplog"
)

type Reporter = health.Reporter
type ReporterResponse = health.ReporterResponse

var RegisterChiRoutes = routes.RegisterChiRoutes
var RegisterGMuxRoutes = routes.RegisterGMUXRoutes

type RoutesConfig = routes.RoutesConfig
type Adapter = routes.Adapter

// Logger is the logging port. Concrete adapters live in utils/logger/zaplog,
// utils/logger/zrlog (zerolog) and utils/logger/lrslog (logrus) — import the
// one matching the logging library the application already uses.
type Logger = logger.Logger
type Field = logger.Field

// DefaultLogger is what a Transport uses when WithLogger is omitted: zap's
// production config, JSON to stderr at Info level and above.
var DefaultLogger = zaplog.Default

// NoopLogger discards every entry. Pass it through WithLogger to opt out of
// the zap default and silence the transport.
var NoopLogger = logger.Noop

// RequestLogger is opt-in per-request logging middleware, assignable to Adapter.
var RequestLogger = logger.RequestLogger

// Field constructors (logger.String, logger.Int, logger.Err, ...) are
// deliberately not re-exported here — they would crowd this package's
// namespace with generic names. Import utils/logger directly for those.
