package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrInvalidJSON is an error that indicates that the JSON in the request body is invalid.
var ErrInvalidJSON = fmt.Errorf("invalid JSON")

// Decoder is a function type that decodes an HTTP request body intoa value of type T.
type Decoder[T any] func(http.ResponseWriter, *http.Request, T) (T, error)

// DecodeJSON is a Decoder that decodes the JSON in the request body into a value of type T.
func DecodeJSON[T any](w http.ResponseWriter, r *http.Request, v T) (T, error) {
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("%w: %v", ErrInvalidJSON, err.Error())
	}
	return v, nil
}
