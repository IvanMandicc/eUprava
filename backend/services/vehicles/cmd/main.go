package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"euprava/vehicles/internal/client"
	"euprava/vehicles/internal/handler"
	"euprava/vehicles/internal/repository"
	"euprava/vehicles/internal/service"
)

func main() {
	port := getenv("PORT", "8085")
	dbURL := getenv("DATABASE_URL", "postgres://euprava:euprava@localhost:5435/vehicle_db?sslmode=disable")
	citizenURL := getenv("CITIZEN_SERVICE_URL", "http://localhost:8081")
	trafficURL := getenv("TRAFFIC_SERVICE_URL", "http://localhost:8082")
	notificationURL := getenv("NOTIFICATION_SERVICE_URL", "http://localhost:8084")

	ctx := context.Background()
	pool := mustConnect(ctx, dbURL)
	defer pool.Close()
	if err := repository.InitSchema(ctx, pool); err != nil {
		log.Fatalf("inicijalizacija šeme: %v", err)
	}

	// Dependency Injection: service sloj dobija interfejse, ne konkretne tipove.
	vehicles := repository.NewPostgresVehicleRepository(pool)
	transfers := repository.NewPostgresTransferRepository(pool)
	thefts := repository.NewPostgresTheftRepository(pool)
	reservations := repository.NewPostgresPlateReservationRepository(pool)
	reports := repository.NewPostgresReportRepository(pool)

	citizens := client.NewHTTPCitizenClient(citizenURL)
	traffic := client.NewHTTPTrafficClient(trafficURL)
	notifier := client.NewHTTPNotificationClient(notificationURL)

	vehicleSvc := service.NewVehicleService(vehicles, transfers, thefts, citizens, traffic, notifier)
	plateSvc := service.NewPlateService(reservations, vehicles, notifier)
	reportSvc := service.NewReportService(reports, vehicles, transfers)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "vehicles"}) })
	handler.NewVehicleHandler(vehicleSvc).RegisterRoutes(r)
	handler.NewMeHandler(vehicleSvc, plateSvc).RegisterRoutes(r)
	handler.NewPlateHandler(plateSvc).RegisterRoutes(r)
	handler.NewReportHandler(reportSvc).RegisterRoutes(r)

	log.Printf("Vehicles servis sluša na :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func mustConnect(ctx context.Context, url string) *pgxpool.Pool {
	var pool *pgxpool.Pool
	var err error
	for i := 0; i < 15; i++ {
		pool, err = pgxpool.New(ctx, url)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				return pool
			}
			pool.Close()
		}
		log.Printf("čekam bazu (%d/15): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("konekcija na bazu nije uspela: %v", err)
	return nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
