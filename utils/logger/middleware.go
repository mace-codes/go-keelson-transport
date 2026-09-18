package logger

import (
	"net/http"
	"time"
)

// RequestLogger returns middleware that logs one entry per request: method,
// path, status code, bytes written and duration. Responses of 5xx log at Error
// level, everything else at Info.
//
// It is opt-in — the Transport registers no middleware of its own. The returned
// func has the same signature as routes.Adapter, so it drops straight into
// RoutesConfig.Adapters
func RequestLogger(log Logger) func(http.Handler) http.Handler {
	if log == nil {
		log = Noop()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &recorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			fields := []Field{
				String("method", r.Method),
				String("path", r.URL.Path),
				Int("status", rec.status),
				Int("bytes", rec.written),
				Duration("duration", time.Since(start)),
			}

			if rec.status >= http.StatusInternalServerError {
				log.Error("request failed", fields...)
				return
			}

			log.Info("request", fields...)
		})
	}
}

// recorder wraps an http.ResponseWriter to capture the status code and response size without buffering the body.
type recorder struct {
	http.ResponseWriter
	status      int
	written     int
	wroteHeader bool
}

func (rec *recorder) WriteHeader(code int) {
	if rec.wroteHeader {
		return
	}

	rec.status = code
	rec.wroteHeader = true
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *recorder) Write(b []byte) (int, error) {
	if !rec.wroteHeader {
		rec.WriteHeader(http.StatusOK)
	}

	n, err := rec.ResponseWriter.Write(b)
	rec.written += n

	return n, err
}

// Unwrap exposes the underlying ResponseWriter to http.ResponseController, so
// flushing and hijacking keep working through this middleware.
func (rec *recorder) Unwrap() http.ResponseWriter {
	return rec.ResponseWriter
}
