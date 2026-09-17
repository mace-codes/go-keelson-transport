package health

import (
	"net/http"

	"github.com/mace-codes/go-keelson-transport/utils/rest"
)

// HealthStatus represents the health status of the application.
type HealthStatus string

const (
	// Healthy indicates that the application is healthy.
	Healthy HealthStatus = "healthy"
	// Unhealthy indicates that the application is unhealthy.
	Unhealthy HealthStatus = "unhealthy"
	// Degraded indicates that the application is in a degraded state.
	Degraded HealthStatus = "degraded"
	// Live indicates that the application is live and running.
	Live HealthStatus = "live"
	// Dead indicates that the application is dead and not running.
	Dead HealthStatus = "dead"
)

// String returns the string representation of the HealthStatus.
func (hs HealthStatus) String() string {
	return string(hs)
}

// Result represents the result of a health check, including the overall status and individual component reports.
type Result struct {
	CriticalService bool `json:"criticalService"`
	ReporterResponse
}

// IsLive returns a Result indicating that the application is live and healthy.
func IsLive() http.HandlerFunc {
	type response struct {
		Status string `json:"status"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		resp := response{Status: Live.String()}
		rest.RespondJSON(w, r, http.StatusOK, nil, resp)
	}
}

// Health returns the current Health status of the app and dependencies
func Health(deps Dependencies) http.HandlerFunc {
	type response struct {
		Status  HealthStatus `json:"status"`
		Message string       `json:"message"`
		Results []Result     `json:"results"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var message string
		status := Healthy
		results := make([]Result, 0, len(deps.Critical())+len(deps.Optional()))

		for _, reporter := range deps.Critical() {
			result := reporter.Health()
			results = append(results, Result{CriticalService: true, ReporterResponse: result})
			if !result.Healthy {
				status = Unhealthy
				message += "(Critical) " + result.Component + ": " + result.Message + "\n"
			}
		}

		for _, reporter := range deps.Optional() {
			result := reporter.Health()
			results = append(results, Result{CriticalService: false, ReporterResponse: result})
			if !result.Healthy {
				status = Degraded
				message += "(Optional) " + result.Component + ": " + result.Message + "\n"
			}
		}

		if message == "" {
			message = "All components healthy"
		}

		resp := response{
			Status:  status,
			Message: message,
			Results: results,
		}

		code := http.StatusOK
		if status == Unhealthy {
			code = http.StatusServiceUnavailable
		}

		rest.RespondJSON(w, r, code, nil, resp)
	}
}

// IsReady reports the readiness of the app to handle requests.
func IsReady(deps Dependencies) http.HandlerFunc {
	type response struct {
		Status string `json:"status"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		for _, reporter := range deps.Critical() {
			if !reporter.Health().Healthy {
				resp := response{Status: Unhealthy.String()}
				rest.RespondJSON(w, r, http.StatusServiceUnavailable, nil, resp)
				return
			}
		}

		resp := response{Status: "ok"}
		rest.RespondJSON(w, r, http.StatusOK, nil, resp)
	}
}
