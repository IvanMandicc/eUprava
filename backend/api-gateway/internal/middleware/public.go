package middleware

import (
	"net/http"
	"strings"
)

// IsPublic označava rute dostupne bez prijave: autentifikacija, health,
// otvoreni podaci (anonimna statistika i šifarnik prekršaja) i javna
// verifikacija izveštaja o vozilu.
func IsPublic(method, path string) bool {
	switch {
	case strings.HasPrefix(path, "/api/auth/"), path == "/health":
		return true
	case method == http.MethodGet &&
		(strings.HasPrefix(path, "/api/open-data/") || path == "/api/violation-types"):
		return true // open data — javni, anonimni podaci
	case method == http.MethodGet && strings.HasPrefix(path, "/api/reports/verify/"):
		return true // javna provera autentičnosti izveštaja o vozilu, bez prijave
	default:
		return false
	}
}
