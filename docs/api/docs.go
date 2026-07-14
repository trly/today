// Package apidocs serves generated API documentation.
package apidocs

import (
	_ "embed"
	"net/http"
)

//go:embed today-v1.md
var todayV1 []byte

// Handler returns the generated API documentation.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		_, _ = w.Write(todayV1)
	})
}
