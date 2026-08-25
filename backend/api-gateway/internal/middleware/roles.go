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
//   - /api/vehicles (registracija/transfer/produženje)     → samo službenik MUP-a (uloga officer)
//   - /api/vehicles/{id}/report-theft|report-found|reports → vlasnik (citizen) ili službenik
//   - /api/vehicle-status/{plate}                          → samo službenik
//   - /api/plate-reservations POST (zahtev)                → građanin; GET/PUT (obrada) → službenik
func Roles(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsPublic(r.Method, r.URL.Path) {
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

	case strings.HasPrefix(path, "/api/vehicles"):
		// Prijava krađe/pronalaska i generisanje sopstvenog izveštaja može i
		// vlasnik vozila; registraciju, prenos i produženje obrađuje službenik.
		if strings.HasSuffix(path, "/report-theft") || strings.HasSuffix(path, "/report-found") || strings.HasSuffix(path, "/reports") {
			return role == "citizen" || role == "officer"
		}
		return role == "officer"

	case strings.HasPrefix(path, "/api/vehicle-status/"):
		return role == "officer"

	case strings.HasPrefix(path, "/api/plate-reservations"):
		// Zahtev za tablicu podnosi građanin; pregled reda čekanja i
		// odobravanje/odbijanje obrađuje službenik.
		if method == http.MethodPost {
			return role == "citizen"
		}
		return role == "officer"

	default:
		return true
	}
}
