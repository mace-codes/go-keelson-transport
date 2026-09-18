package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/mux"
	"github.com/mace-codes/go-keelson-transport/routes"
)

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestRegisterChiRoutes(t *testing.T) {
	tests := []struct {
		name       string
		router     http.Handler
		routes     []routes.RoutesConfig
		wantErr    bool
		wantStatus int
	}{
		{
			name:    "wrong router type returns error",
			router:  mux.NewRouter(),
			routes:  []routes.RoutesConfig{{Path: "/ping", Methods: []string{http.MethodGet}, Handler: okHandler}},
			wantErr: true,
		},
		{
			name:       "registers a GET route on a chi router",
			router:     chi.NewRouter(),
			routes:     []routes.RoutesConfig{{Path: "/ping", Methods: []string{http.MethodGet}, Handler: okHandler}},
			wantErr:    false,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := routes.RegisterChiRoutes(tt.router, tt.routes)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			req := httptest.NewRequest(http.MethodGet, tt.routes[0].Path, nil)
			w := httptest.NewRecorder()
			tt.router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
