package health

// ReporterResponse represents the response structure for a health check, including the overall status and individual component reports.
type ReporterResponse struct {
	Component string         `json:"components"`
	Healthy   bool           `json:"healthy"`
	Status    string         `json:"status"`
	Message   string         `json:"message,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Reporter interface defines a method to report the health status of a component.
type Reporter interface {
	Health() ReporterResponse
}
