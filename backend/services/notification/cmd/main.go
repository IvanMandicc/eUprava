package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"euprava/notification/internal/handler"
	"euprava/notification/internal/repository"
	"euprava/notification/internal/service"
)

func main() {
	port := getenv("PORT", "8084")
	dbURL := getenv("DATABASE_URL", "postgres://euprava:euprava@localhost:5436/notification_db?sslmode=disable")

	ctx := context.Background()
	pool := mustConnect(ctx, dbURL)
	defer pool.Close()

	notifications, err := repository.NewPostgresNotificationRepository(ctx, pool)
	if err != nil {
		log.Fatalf("inicijalizacija šeme: %v", err)
	}
	svc := service.NewNotificationService(notifications)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "notification"}) })
	handler.NewNotificationHandler(svc).RegisterRoutes(r)

	log.Printf("Notification servis sluša na :%s", port)
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
