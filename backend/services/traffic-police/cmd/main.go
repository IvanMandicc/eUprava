package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"euprava/traffic-police/internal/client"
	"euprava/traffic-police/internal/handler"
	"euprava/traffic-police/internal/repository"
	"euprava/traffic-police/internal/service"
)

func main() {
	port := getenv("PORT", "8082")
	dbURL := getenv("DATABASE_URL", "postgres://euprava:euprava@localhost:5434/traffic_db?sslmode=disable")
	citizenURL := getenv("CITIZEN_SERVICE_URL", "http://localhost:8081")
	notificationURL := getenv("NOTIFICATION_SERVICE_URL", "http://localhost:8084")
	vehiclesURL := getenv("VEHICLES_SERVICE_URL", "http://localhost:8085")

	ctx := context.Background()
	pool := mustConnect(ctx, dbURL)
	defer pool.Close()
	if err := repository.InitSchema(ctx, pool); err != nil {
		log.Fatalf("inicijalizacija šeme: %v", err)
	}

	// Dependency Injection: service sloj dobija interfejse, ne konkretne tipove.
	drivers := repository.NewPostgresDriverRepository(pool)
	violations := repository.NewPostgresViolationRepository(pool)
	fines := repository.NewPostgresFineRepository(pool)
	citizens := client.NewHTTPCitizenClient(citizenURL)
	notifier := client.NewHTTPNotificationClient(notificationURL)
	vehicles := client.NewHTTPVehiclesClient(vehiclesURL)

	driverSvc := service.NewDriverService(drivers, citizens)
	violationSvc := service.NewViolationService(drivers, violations, fines, notifier, vehicles)
	fineSvc := service.NewFineService(fines, violations, notifier)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "traffic-police"}) })
	handler.NewDriverHandler(driverSvc, violationSvc).RegisterRoutes(r)
	handler.NewViolationHandler(violationSvc).RegisterRoutes(r)
	handler.NewFineHandler(fineSvc).RegisterRoutes(r)
	handler.NewMeHandler(driverSvc, violationSvc, fineSvc).RegisterRoutes(r)

	log.Printf("Traffic Police servis sluša na :%s", port)
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
