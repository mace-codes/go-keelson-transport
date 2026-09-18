package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       any
		wantBody   string
	}{
		{
			name:       "no content status omits body",
			statusCode: http.StatusNoContent,
			body:       map[string]string{"ignored": "true"},
			wantBody:   "",
		},
		{
			name:       "nil body writes empty object",
			statusCode: http.StatusOK,
			body:       nil,
			wantBody:   `{}`,
		},
		{
			name:       "struct body is marshaled",
			statusCode: http.StatusCreated,
			body:       map[string]string{"status": "ok"},
			wantBody:   `{"status":"ok"}`,
		},
		{
			name:       "error body is wrapped with the status code",
			statusCode: http.StatusBadRequest,
			body:       errors.New("boom"),
			wantBody:   `{"error":{"code":"400","message":"boom"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			RespondJSON(w, req, tt.statusCode, nil, tt.body)

			if w.Code != tt.statusCode {
				t.Errorf("status code = %d, want %d", w.Code, tt.statusCode)
			}
			if got := w.Header().Get(ContentType); got != ApplicationJSON {
				t.Errorf("Content-Type = %q, want %q", got, ApplicationJSON)
			}
			if got := w.Body.String(); got != tt.wantBody {
				t.Errorf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func TestRespondJSON_PreservesCallerHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	RespondJSON(w, req, http.StatusOK, map[string]string{"X-Request-Id": "abc123"}, nil)

	if got := w.Header().Get("X-Request-Id"); got != "abc123" {
		t.Errorf("X-Request-Id = %q, want %q", got, "abc123")
	}
	if got := w.Header().Get(ContentType); got != ApplicationJSON {
		t.Errorf("Content-Type = %q, want %q", got, ApplicationJSON)
	}
}
