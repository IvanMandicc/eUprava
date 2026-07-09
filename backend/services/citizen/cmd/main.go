package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"euprava/citizen/internal/handler"
	"euprava/citizen/internal/repository"
	"euprava/citizen/internal/service"
)

func main() {
	port := getenv("PORT", "8081")
	dbURL := getenv("DATABASE_URL", "postgres://euprava:euprava@localhost:5433/citizen_db?sslmode=disable")
	jwtSecret := getenv("JWT_SECRET", "super-tajni-kljuc-promeni-me")

	ctx := context.Background()
	pool := mustConnect(ctx, dbURL)
	defer pool.Close()

	// Dependency Injection: konkretne implementacije se "ubrizgavaju" ovde,
	// slojevi međusobno zavise samo od interfejsa.
	users, err := repository.NewPostgresUserRepository(ctx, pool)
	if err != nil {
		log.Fatalf("inicijalizacija šeme: %v", err)
	}
	authSvc := service.NewAuthService(users, jwtSecret)
	citizenSvc := service.NewCitizenService(users)

	if err := authSvc.SeedOfficer(ctx, getenv("OFFICER_EMAIL", "policija@euprava.rs"), getenv("OFFICER_PASSWORD", "policija123")); err != nil {
		log.Fatalf("seed policajca: %v", err)
	}
	if err := authSvc.SeedAdmin(ctx, getenv("ADMIN_EMAIL", "admin@euprava.rs"), getenv("ADMIN_PASSWORD", "admin123")); err != nil {
		log.Fatalf("seed administratora: %v", err)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "citizen"}) })
	handler.NewAuthHandler(authSvc).RegisterRoutes(r)
	handler.NewCitizenHandler(citizenSvc).RegisterRoutes(r)
	handler.NewUserHandler(authSvc, citizenSvc).RegisterRoutes(r)

	log.Printf("Citizen servis sluša na :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

// mustConnect pokušava konekciju više puta jer baza može još da se podiže.
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
