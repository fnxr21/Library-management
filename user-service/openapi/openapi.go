// Package openapi serves the generated API spec (make openapi).
package openapi

import (
	_ "embed"
	"net/http"
)

//go:embed library.swagger.json
var spec []byte

// Handler serves the OpenAPI v2 spec for every public endpoint.
func Handler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(spec)
}
