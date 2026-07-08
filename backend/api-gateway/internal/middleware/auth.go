package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Auth validira JWT token i prosleđuje identitet servisima kroz
// X-User-Id i X-User-Role header-e. Rute pod /api/auth su javne.
type Auth struct {
	secret []byte
}

func NewAuth(secret string) *Auth {
	return &Auth{secret: []byte(secret)}
}

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Klijent ne sme sam da postavi identitetske header-e.
		r.Header.Del("X-User-Id")
		r.Header.Del("X-User-Role")

		if strings.HasPrefix(r.URL.Path, "/api/auth/") || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "nedostaje Bearer token")
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("neočekivan algoritam potpisa: %v", t.Header["alg"])
			}
			return a.secret, nil
		})
		if err != nil || !token.Valid {
			writeError(w, http.StatusUnauthorized, "nevažeći token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeError(w, http.StatusUnauthorized, "nevažeći token")
			return
		}
		sub, _ := claims["sub"].(float64)
		role, _ := claims["role"].(string)
		if sub == 0 || role == "" {
			writeError(w, http.StatusUnauthorized, "token bez identiteta")
			return
		}

		r.Header.Set("X-User-Id", fmt.Sprintf("%.0f", sub))
		r.Header.Set("X-User-Role", role)
		next.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
