package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"euprava/gateway/internal/middleware"
	"euprava/gateway/internal/proxy"
)

func main() {
	port := getenv("PORT", "8080")
	citizenURL := getenv("CITIZEN_SERVICE_URL", "http://localhost:8081")
	trafficURL := getenv("TRAFFIC_SERVICE_URL", "http://localhost:8082")
	paymentURL := getenv("PAYMENT_SERVICE_URL", "http://localhost:8083")
	notificationURL := getenv("NOTIFICATION_SERVICE_URL", "http://localhost:8084")

	routes := []proxy.Route{
		{Prefix: "/api/auth/", Target: citizenURL},
		{Prefix: "/api/citizens", Target: citizenURL},
		{Prefix: "/api/users", Target: citizenURL},
		{Prefix: "/api/drivers", Target: trafficURL},
		{Prefix: "/api/violation-types", Target: trafficURL},
		{Prefix: "/api/violations", Target: trafficURL},
		{Prefix: "/api/fines", Target: trafficURL},
		{Prefix: "/api/penalty-points", Target: trafficURL},
		{Prefix: "/api/me/", Target: trafficURL},
		{Prefix: "/api/payments", Target: paymentURL},
		{Prefix: "/api/notifications", Target: notificationURL},
	}

	p, err := proxy.New(routes)
	if err != nil {
		log.Fatalf("neispravna konfiguracija ruta: %v", err)
	}

	auth := middleware.NewAuth(getenv("JWT_SECRET", "super-tajni-kljuc-promeni-me"))

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "api-gateway"})
	})
	mux.Handle("/api/", p)

	// Lanac middleware-a: CORS → JWT autentifikacija → autorizacija po ulozi → proxy.
	handler := middleware.CORS(auth.Middleware(middleware.Roles(mux)))

	log.Printf("API Gateway sluša na :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
