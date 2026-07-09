package middleware

import (
	"net/http"
	"strings"
)

// Roles sprovodi pravila pristupa po ulozi (autorizacija).
//
// Pravila:
//   - /api/drivers, /api/violations, /api/penalty-points  → samo policajac
//   - /api/citizens/{id} (osim /me)                       → samo policajac
//   - /api/fines (pregled svih)                           → samo policajac; izmene su interne
//   - /api/payments                                       → samo građanin
//   - /api/notifications POST                             → interno (blokirano spolja)
//   - /api/me/*                                           → svaki ulogovan korisnik (svoji podaci)
func Roles(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/auth/") || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		role := r.Header.Get("X-User-Role")
		if !allowed(r.Method, r.URL.Path, role) {
			writeError(w, http.StatusForbidden, "nemate dozvolu za ovu akciju")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func allowed(method, path, role string) bool {
	switch {
	case strings.HasPrefix(path, "/api/me/"), path == "/api/citizens/me":
		return true // sopstveni podaci

	case strings.HasPrefix(path, "/api/users"):
		// Upravljanje korisnicima (pregled, kreiranje policajaca) — samo administrator.
		return role == "admin"

	case strings.HasPrefix(path, "/api/drivers"),
		strings.HasPrefix(path, "/api/violations"),
		strings.HasPrefix(path, "/api/penalty-points"),
		strings.HasPrefix(path, "/api/citizens"):
		return role == "officer"

	case strings.HasPrefix(path, "/api/fines"):
		// Plaćanje kazne (PUT /fines/{id}/pay) je interni poziv Payment servisa
		// direktno ka Traffic Police servisu — spolja dozvoljavamo samo pregled policajcu.
		return method == http.MethodGet && role == "officer"

	case strings.HasPrefix(path, "/api/payments"):
		return role == "citizen"

	case strings.HasPrefix(path, "/api/notifications"):
		// Kreiranje obaveštenja je interno (servis → servis); spolja samo čitanje i označavanje.
		return method != http.MethodPost

	default:
		return true
	}
}
