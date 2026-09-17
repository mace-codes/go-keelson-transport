package transport

import (
	"github.com/mace-codes/go-keelson-transport/health"
	"github.com/mace-codes/go-keelson-transport/routes"
)

type Reporter = health.Reporter
type ReporterResponse = health.ReporterResponse

var RegisterChiRoutes = routes.RegisterChiRoutes
var RegisterGMuxRoutes = routes.RegisterGMUXRoutes

type RoutesConfig = routes.RoutesConfig
type Adapter = routes.Adapter
