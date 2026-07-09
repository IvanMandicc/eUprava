package middleware

import (
	"net/http"
	"strings"
)

// IsPublic označava rute dostupne bez prijave: autentifikacija, health
// i otvoreni podaci (anonimna statistika i šifarnik prekršaja).
func IsPublic(method, path string) bool {
	switch {
	case strings.HasPrefix(path, "/api/auth/"), path == "/health":
		return true
	case method == http.MethodGet &&
		(strings.HasPrefix(path, "/api/open-data/") || path == "/api/violation-types"):
		return true // open data — javni, anonimni podaci
	default:
		return false
	}
}
