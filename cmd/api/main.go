package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/dragondarkon/bqredis-crud/internal/delivery/http"
	"github.com/dragondarkon/bqredis-crud/internal/repository"
	"github.com/dragondarkon/bqredis-crud/internal/usecase"
	"github.com/dragondarkon/bqredis-crud/pkg/config"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize context
	ctx := context.Background()

	// Initialize BigQuery client
	bqClient, err := bigquery.NewClient(ctx, cfg.GoogleCloudProject)
	if err != nil {
		log.Fatalf("Failed to create BigQuery client: %v", err)
	}
	defer bqClient.Close()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	defer redisClient.Close()

	// Test Redis connection
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize repositories
	primaryRepo := repository.NewBigQueryRepository(bqClient, cfg.GoogleCloudProject, cfg.BigQueryDataset, cfg.BigQueryTable)
	cacheRepo := repository.NewRedisRepository(redisClient, primaryRepo, cfg.RedisTTL)

	// Initialize use case with primary and cache repositories
	userUseCase := usecase.NewUserUseCase(primaryRepo, cacheRepo)

	// Initialize Echo framework
	e := echo.New()

	// Setup routes
	http.SetupRoutes(e, userUseCase)

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		if err := e.Start(addr); err != nil {
			log.Printf("Shutting down the server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
