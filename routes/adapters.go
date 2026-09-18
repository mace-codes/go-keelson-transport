package routes

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Adapter is a function type that takes an http.Handler and returns an http.Handler. It is used to adapt a router to the http.Handler interface.
type Adapter func(http.Handler) http.Handler

// AdapterFunc is a function type that takes an http.HandlerFunc and returns an http.HandlerFunc. It is used to adapt a router to the http.HandlerFunc interface.
type AdapterFunc func(http.Handler, ...Adapter) (http.HandlerFunc, error)

// ChiMuxAdapter is an adapter that allows a chi router to be used as a http.HandlerFunc.
func ChiMuxAdapter(hndlr http.Handler, adapters ...Adapter) (http.HandlerFunc, error) {
	for i := range adapters {
		adapter := adapters[len(adapters)-1-i]
		hndlr = adapter(hndlr)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		hndlr.ServeHTTP(w, r)
	}, nil
}

// GMuxAdapter is an adapter that allows a Gorilla Mux router to be used as a http.HandlerFunc.
func GMuxAdapter(hndlr http.Handler, adapters ...Adapter) (http.HandlerFunc, error) {
	for i := range adapters {
		adapter := mux.MiddlewareFunc(adapters[len(adapters)-1-i])
		hndlr = adapter(hndlr)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hndlr.ServeHTTP(w, r)
	}), nil
}
