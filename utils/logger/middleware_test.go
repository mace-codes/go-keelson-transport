package logger_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mace-codes/go-keelson-transport/utils/logger"
)

func TestRequestLogger(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		target     string
		handler    http.HandlerFunc
		wantLevel  string
		wantMsg    string
		wantStatus int
		wantBytes  int
	}{
		{
			name:   "implicit 200 when handler never writes a header",
			method: http.MethodGet,
			target: "/health",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"status":"live"}`))
			},
			wantLevel:  "info",
			wantMsg:    "request",
			wantStatus: http.StatusOK,
			wantBytes:  17,
		},
		{
			name:       "no body at all still reports 200",
			method:     http.MethodGet,
			target:     "/empty",
			handler:    func(w http.ResponseWriter, r *http.Request) {},
			wantLevel:  "info",
			wantMsg:    "request",
			wantStatus: http.StatusOK,
			wantBytes:  0,
		},
		{
			name:   "explicit client error logs at info",
			method: http.MethodPost,
			target: "/widgets",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			wantLevel:  "info",
			wantMsg:    "request",
			wantStatus: http.StatusBadRequest,
			wantBytes:  0,
		},
		{
			name:   "server error logs at error",
			method: http.MethodGet,
			target: "/boom",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("boom"))
			},
			wantLevel:  "error",
			wantMsg:    "request failed",
			wantStatus: http.StatusInternalServerError,
			wantBytes:  4,
		},
		{
			name:   "first WriteHeader wins",
			method: http.MethodGet,
			target: "/double",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantLevel:  "info",
			wantMsg:    "request",
			wantStatus: http.StatusTeapot,
			wantBytes:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecordingLogger()

			mw := logger.RequestLogger(rec)
			mw(tt.handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(tt.method, tt.target, nil))

			got := rec.captured()
			if len(got) != 1 {
				t.Fatalf("captured %d entries, want 1", len(got))
			}

			e := got[0]
			if e.level != tt.wantLevel {
				t.Errorf("level = %q, want %q", e.level, tt.wantLevel)
			}

			if e.msg != tt.wantMsg {
				t.Errorf("msg = %q, want %q", e.msg, tt.wantMsg)
			}

			if e.fields["method"] != tt.method {
				t.Errorf("method = %v, want %q", e.fields["method"], tt.method)
			}

			if e.fields["path"] != tt.target {
				t.Errorf("path = %v, want %q", e.fields["path"], tt.target)
			}

			if e.fields["status"] != tt.wantStatus {
				t.Errorf("status = %v, want %d", e.fields["status"], tt.wantStatus)
			}

			if e.fields["bytes"] != tt.wantBytes {
				t.Errorf("bytes = %v, want %d", e.fields["bytes"], tt.wantBytes)
			}

			if _, ok := e.fields["duration"].(time.Duration); !ok {
				t.Errorf("duration = %T, want time.Duration", e.fields["duration"])
			}
		})
	}
}

func TestRequestLoggerPassesResponseThrough(t *testing.T) {
	mw := logger.RequestLogger(newRecordingLogger())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"ok":true}`))
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/widgets", nil))

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	if body := w.Body.String(); body != `{"ok":true}` {
		t.Errorf("body = %q, want %q", body, `{"ok":true}`)
	}
}

func TestRequestLoggerFlushThroughResponseController(t *testing.T) {
	mw := logger.RequestLogger(newRecordingLogger())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("chunk"))

		// Unwrap on the recorder is what lets this reach the real writer.
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Errorf("Flush() error = %v, want nil", err)
		}
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/stream", nil))

	if !w.Flushed {
		t.Error("response was not flushed through the middleware")
	}
}

func TestRequestLoggerNilLoggerDoesNotPanic(t *testing.T) {
	mw := logger.RequestLogger(nil)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
