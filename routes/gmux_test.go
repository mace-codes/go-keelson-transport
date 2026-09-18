package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/mux"
	"github.com/mace-codes/go-keelson-transport/routes"
)

var chiHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestRegisterGMUXRoutes(t *testing.T) {
	tests := []struct {
		name       string
		router     http.Handler
		routes     []routes.RoutesConfig
		wantErr    bool
		wantStatus int
	}{
		{
			name:    "wrong router type returns error",
			router:  chi.NewRouter(),
			routes:  []routes.RoutesConfig{{Path: "/ping", Methods: []string{http.MethodGet}, Handler: chiHandler}},
			wantErr: true,
		},
		{
			name:       "registers a GET route on a gorilla mux router",
			router:     mux.NewRouter(),
			routes:     []routes.RoutesConfig{{Path: "/ping", Methods: []string{http.MethodGet}, Handler: chiHandler}},
			wantErr:    false,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := routes.RegisterGMUXRoutes(tt.router, tt.routes)
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

// markingAdapter records name into order when the handler chain executes,
// so tests can assert adapters run in the declared order.
func markingAdapter(name string, order *[]string) routes.Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*order = append(*order, name)
			next.ServeHTTP(w, r)
		})
	}
}

func TestChiMuxAdapterOrder(t *testing.T) {
	tests := []struct {
		name          string
		adapterNames  []string
		wantExecOrder []string
	}{
		{name: "no adapters", adapterNames: nil, wantExecOrder: nil},
		{name: "single adapter", adapterNames: []string{"A"}, wantExecOrder: []string{"A"}},
		{name: "multiple adapters run in declared order", adapterNames: []string{"A", "B", "C"}, wantExecOrder: []string{"A", "B", "C"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var order []string
			adapters := make([]routes.Adapter, len(tt.adapterNames))
			for i, name := range tt.adapterNames {
				adapters[i] = markingAdapter(name, &order)
			}

			hdlr, err := routes.ChiMuxAdapter(okHandler, adapters...)
			if err != nil {
				t.Fatalf("chiMuxAdapter() error = %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			hdlr(w, req)

			if len(order) != len(tt.wantExecOrder) {
				t.Fatalf("order = %v, want %v", order, tt.wantExecOrder)
			}
			for i := range order {
				if order[i] != tt.wantExecOrder[i] {
					t.Errorf("order[%d] = %q, want %q", i, order[i], tt.wantExecOrder[i])
				}
			}
		})
	}
}

func TestGMuxAdapterOrder(t *testing.T) {
	tests := []struct {
		name          string
		adapterNames  []string
		wantExecOrder []string
	}{
		{name: "no adapters", adapterNames: nil, wantExecOrder: nil},
		{name: "single adapter", adapterNames: []string{"A"}, wantExecOrder: []string{"A"}},
		{name: "multiple adapters run in declared order", adapterNames: []string{"A", "B", "C"}, wantExecOrder: []string{"A", "B", "C"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var order []string
			adapters := make([]routes.Adapter, len(tt.adapterNames))
			for i, name := range tt.adapterNames {
				adapters[i] = markingAdapter(name, &order)
			}

			hdlr, err := routes.GMuxAdapter(okHandler, adapters...)
			if err != nil {
				t.Fatalf("gMuxAdapter() error = %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			hdlr(w, req)

			if len(order) != len(tt.wantExecOrder) {
				t.Fatalf("order = %v, want %v", order, tt.wantExecOrder)
			}
			for i := range order {
				if order[i] != tt.wantExecOrder[i] {
					t.Errorf("order[%d] = %q, want %q", i, order[i], tt.wantExecOrder[i])
				}
			}
		})
	}
}
