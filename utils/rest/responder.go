package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const (
	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
)

// Responder is a func that responds to an HTTP request with a given status code and body.
type Responder func(w http.ResponseWriter, r *http.Request, statusCode int, headers map[string]string, body any)

// respond is a helper function that writes the status code, headers, and body to the http.ResponseWriter.
func respond(w http.ResponseWriter, statusCode int, headers map[string]string, body []byte) {
	for k, v := range headers {
		w.Header().Set(k, v)
	}

	if body == nil {
		w.WriteHeader(statusCode)
		return
	}

	w.WriteHeader(statusCode)
	w.Write(body)
}

// RespondJSON is a Responder that responds with a JSON body and the given status code.
func RespondJSON(w http.ResponseWriter, r *http.Request, statusCode int, headers map[string]string, body any) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers[ContentType] = ApplicationJSON

	if statusCode == http.StatusNoContent {
		respond(w, statusCode, headers, nil)
		return
	}

	if body == nil {
		respond(w, statusCode, headers, []byte("{}"))
		return
	}

	if err, ok := body.(error); ok {
		resp, err := json.Marshal(map[string]any{
			"error": map[string]string{
				"code":    strconv.Itoa(statusCode),
				"message": err.Error(),
			},
		})
		if err != nil {
			respond(w, statusCode, headers, []byte(`{"error":{"code":"500","message":"`+err.Error()+`"}}`))
			return
		}
		respond(w, statusCode, headers, resp)
		return
	}

	bdy, err := json.Marshal(body)
	if err != nil {
		RespondJSON(w, r, http.StatusInternalServerError, nil, err)
		return
	}

	respond(w, statusCode, headers, bdy)
}
