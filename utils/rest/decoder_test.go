package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type payload struct {
	Name string `json:"name"`
}

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    payload
		wantErr bool
	}{
		{
			name: "valid JSON decodes into T",
			body: `{"name":"nic"}`,
			want: payload{Name: "nic"},
		},
		{
			name:    "malformed JSON returns wrapped error",
			body:    `{"name":`,
			wantErr: true,
		},
		{
			name:    "empty body returns wrapped error",
			body:    ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			got, err := DecodeJSON(w, req, payload{})

			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidJSON) {
					t.Errorf("err = %v, want wrapped %v", err, ErrInvalidJSON)
				}
				return
			}
			if got != tt.want {
				t.Errorf("got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
