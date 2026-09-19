package transport

import (
	"github.com/mace-codes/go-keelson-transport/health"
	"github.com/mace-codes/go-keelson-transport/routes"
	"github.com/mace-codes/go-keelson-transport/utils/logger"
	"github.com/mace-codes/go-keelson-transport/utils/logger/zaplog"
	"github.com/mace-codes/go-keelson-transport/utils/rest"
)

type Reporter = health.Reporter
type ReporterResponse = health.ReporterResponse

var RegisterChiRoutes = routes.RegisterChiRoutes
var RegisterGMuxRoutes = routes.RegisterGMUXRoutes

type RoutesConfig = routes.RoutesConfig
type Adapter = routes.Adapter

type Logger = logger.Logger
type Field = logger.Field

var DefaultLogger = zaplog.Default
var NoopLogger = logger.Noop
var RequestLogger = logger.RequestLogger

// Field constructors (logger.String, logger.Int, logger.Err, ...) are
// deliberately not re-exported here — they would crowd this package's
// namespace with generic names. Import utils/logger directly for those.

var RespondJSON = rest.RespondJSON
