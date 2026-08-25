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
	vehiclesURL := getenv("VEHICLES_SERVICE_URL", "http://localhost:8085")

	routes := []proxy.Route{
		{Prefix: "/api/auth/", Target: citizenURL},
		{Prefix: "/api/citizens", Target: citizenURL},
		{Prefix: "/api/users", Target: citizenURL},
		{Prefix: "/api/drivers", Target: trafficURL},
		{Prefix: "/api/violation-types", Target: trafficURL},
		{Prefix: "/api/open-data/", Target: trafficURL},
		{Prefix: "/api/violations", Target: trafficURL},
		{Prefix: "/api/fines", Target: trafficURL},
		{Prefix: "/api/penalty-points", Target: trafficURL},
		{Prefix: "/api/payments", Target: paymentURL},
		{Prefix: "/api/notifications", Target: notificationURL},
		// MUP-vozila: registar vozila, personalizovane tablice i javna
		// verifikacija izveštaja — peti mikroservis u sistemu.
		{Prefix: "/api/vehicles", Target: vehiclesURL},
		{Prefix: "/api/vehicle-status/", Target: vehiclesURL},
		{Prefix: "/api/plate-reservations", Target: vehiclesURL},
		{Prefix: "/api/reports/", Target: vehiclesURL},
		// /api/me/ rute idu ka oba servisa: /me/driver,violations,fines →
		// Traffic Police; /me/vehicles,plate-reservations → Vehicles.
		// Redosled je bitan — proxy.New bira prvi prefiks koji se poklapa.
		{Prefix: "/api/me/vehicles", Target: vehiclesURL},
		{Prefix: "/api/me/plate-reservations", Target: vehiclesURL},
		{Prefix: "/api/me/", Target: trafficURL},
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
