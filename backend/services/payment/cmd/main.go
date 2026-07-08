package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"euprava/payment/internal/client"
	"euprava/payment/internal/handler"
	"euprava/payment/internal/repository"
	"euprava/payment/internal/service"
)

func main() {
	port := getenv("PORT", "8083")
	dbURL := getenv("DATABASE_URL", "postgres://euprava:euprava@localhost:5435/payment_db?sslmode=disable")
	trafficURL := getenv("TRAFFIC_SERVICE_URL", "http://localhost:8082")
	notificationURL := getenv("NOTIFICATION_SERVICE_URL", "http://localhost:8084")

	ctx := context.Background()
	pool := mustConnect(ctx, dbURL)
	defer pool.Close()

	payments, err := repository.NewPostgresPaymentRepository(ctx, pool)
	if err != nil {
		log.Fatalf("inicijalizacija šeme: %v", err)
	}
	svc := service.NewPaymentService(payments,
		client.NewHTTPTrafficClient(trafficURL),
		client.NewHTTPNotificationClient(notificationURL))

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "payment"}) })
	handler.NewPaymentHandler(svc).RegisterRoutes(r)

	log.Printf("Payment servis sluša na :%s", port)
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
