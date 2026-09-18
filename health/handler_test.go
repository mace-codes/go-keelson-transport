package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/mace-codes/go-keelson-transport/health"
)

type fakeReporter struct {
	resp health.ReporterResponse
}

func (f fakeReporter) Health() health.ReporterResponse { return f.resp }

type fakeDeps struct {
	critical []health.Reporter
	optional []health.Reporter
}

func (d fakeDeps) Critical() []health.Reporter { return d.critical }
func (d fakeDeps) Optional() []health.Reporter { return d.optional }

func healthyResponse(name string) health.ReporterResponse {
	return health.ReporterResponse{Component: name, Healthy: true, Status: "ok"}
}

func unhealthyResponse(name, msg string) health.ReporterResponse {
	return health.ReporterResponse{Component: name, Healthy: false, Status: "down", Message: msg}
}

func healthyReporter(name string) health.Reporter {
	return fakeReporter{resp: healthyResponse(name)}
}

func unhealthyReporter(name, msg string) health.Reporter {
	return fakeReporter{resp: unhealthyResponse(name, msg)}
}

func TestHealth(t *testing.T) {
	tests := []struct {
		name        string
		deps        fakeDeps
		wantCode    int
		wantStatus  health.HealthStatus
		wantMessage string
		wantResults []health.Result
	}{
		{
			name:        "no dependencies",
			deps:        fakeDeps{},
			wantCode:    http.StatusOK,
			wantStatus:  health.Healthy,
			wantMessage: "All components healthy",
			wantResults: nil,
		},
		{
			name: "all healthy",
			deps: fakeDeps{
				critical: []health.Reporter{healthyReporter("db")},
				optional: []health.Reporter{healthyReporter("cache")},
			},
			wantCode:    http.StatusOK,
			wantStatus:  health.Healthy,
			wantMessage: "All components healthy",
			wantResults: []health.Result{
				{CriticalService: true, ReporterResponse: healthyResponse("db")},
				{CriticalService: false, ReporterResponse: healthyResponse("cache")},
			},
		},
		{
			name: "critical dependency down",
			deps: fakeDeps{
				critical: []health.Reporter{unhealthyReporter("db", "connection refused")},
				optional: []health.Reporter{healthyReporter("cache")},
			},
			wantCode:    http.StatusServiceUnavailable,
			wantStatus:  health.Unhealthy,
			wantMessage: "(Critical) db: connection refused\n",
			wantResults: []health.Result{
				{CriticalService: true, ReporterResponse: unhealthyResponse("db", "connection refused")},
				{CriticalService: false, ReporterResponse: healthyResponse("cache")},
			},
		},
		{
			name: "optional dependency down",
			deps: fakeDeps{
				critical: []health.Reporter{healthyReporter("db")},
				optional: []health.Reporter{unhealthyReporter("cache", "timeout")},
			},
			wantCode:    http.StatusOK,
			wantStatus:  health.Degraded,
			wantMessage: "(Optional) cache: timeout\n",
			wantResults: []health.Result{
				{CriticalService: true, ReporterResponse: healthyResponse("db")},
				{CriticalService: false, ReporterResponse: unhealthyResponse("cache", "timeout")},
			},
		},
		{
			name: "critical down takes precedence over optional down",
			deps: fakeDeps{
				critical: []health.Reporter{unhealthyReporter("db", "down")},
				optional: []health.Reporter{unhealthyReporter("cache", "down")},
			},
			wantCode:   http.StatusServiceUnavailable,
			wantStatus: health.Unhealthy,
			// Both failures are reported even though the status reflects only the critical one.
			wantMessage: "(Critical) db: down\n(Optional) cache: down\n",
			wantResults: []health.Result{
				{CriticalService: true, ReporterResponse: unhealthyResponse("db", "down")},
				{CriticalService: false, ReporterResponse: unhealthyResponse("cache", "down")},
			},
		},
		{
			name: "results preserve reporter order within each tier",
			deps: fakeDeps{
				critical: []health.Reporter{healthyReporter("db"), healthyReporter("queue")},
				optional: []health.Reporter{healthyReporter("cache"), healthyReporter("metrics")},
			},
			wantCode:    http.StatusOK,
			wantStatus:  health.Healthy,
			wantMessage: "All components healthy",
			wantResults: []health.Result{
				{CriticalService: true, ReporterResponse: healthyResponse("db")},
				{CriticalService: true, ReporterResponse: healthyResponse("queue")},
				{CriticalService: false, ReporterResponse: healthyResponse("cache")},
				{CriticalService: false, ReporterResponse: healthyResponse("metrics")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			health.Health(tt.deps)(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("status code = %d, want %d", w.Code, tt.wantCode)
			}

			var got struct {
				Status  health.HealthStatus `json:"status"`
				Message string              `json:"message"`
				Results []health.Result     `json:"results"`
			}
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if got.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", got.Status, tt.wantStatus)
			}
			if got.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", got.Message, tt.wantMessage)
			}

			if len(got.Results) != len(tt.wantResults) {
				t.Fatalf("len(results) = %d, want %d", len(got.Results), len(tt.wantResults))
			}
			for i, want := range tt.wantResults {
				if !reflect.DeepEqual(got.Results[i], want) {
					t.Errorf("results[%d] = %+v, want %+v", i, got.Results[i], want)
				}
			}
		})
	}
}

func TestIsReady(t *testing.T) {
	tests := []struct {
		name       string
		deps       fakeDeps
		wantCode   int
		wantStatus string
	}{
		{
			name:       "no critical dependencies",
			deps:       fakeDeps{},
			wantCode:   http.StatusOK,
			wantStatus: "ok",
		},
		{
			name:       "critical dependency healthy",
			deps:       fakeDeps{critical: []health.Reporter{healthyReporter("db")}},
			wantCode:   http.StatusOK,
			wantStatus: "ok",
		},
		{
			name:       "critical dependency unhealthy",
			deps:       fakeDeps{critical: []health.Reporter{unhealthyReporter("db", "down")}},
			wantCode:   http.StatusServiceUnavailable,
			wantStatus: health.Unhealthy.String(),
		},
		{
			name: "optional dependency unhealthy does not affect readiness",
			deps: fakeDeps{
				critical: []health.Reporter{healthyReporter("db")},
				optional: []health.Reporter{unhealthyReporter("cache", "down")},
			},
			wantCode:   http.StatusOK,
			wantStatus: "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
			w := httptest.NewRecorder()

			health.IsReady(tt.deps)(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("status code = %d, want %d", w.Code, tt.wantCode)
			}

			var got struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", got.Status, tt.wantStatus)
			}
		})
	}
}

func TestIsLive(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/is-live", nil)
	w := httptest.NewRecorder()

	health.IsLive()(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var got struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Status != health.Live.String() {
		t.Errorf("status = %q, want %q", got.Status, health.Live.String())
	}
}
