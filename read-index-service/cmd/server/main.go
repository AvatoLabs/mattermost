package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattermost/mattermost-read-index-service/internal/api"
	"github.com/mattermost/mattermost-read-index-service/internal/consumer"
	"github.com/mattermost/mattermost-read-index-service/internal/index"
)

func main() {
	log.Println("Starting Mattermost Read Index Service...")

	// 配置
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")
	port := getEnv("PORT", "8066")
	windowSize := getEnvInt("WINDOW_SIZE", 1000)

	// 创建索引服务
	indexService := index.NewService(windowSize)

	// 创建 Redis 消费者
	redisConsumer := consumer.NewRedisConsumer(redisURL, indexService)

	// 创建 HTTP API
	apiServer := api.NewServer(indexService, port)

	// 启动服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动 Redis 消费者
	go func() {
		if err := redisConsumer.Start(ctx); err != nil {
			log.Printf("Redis consumer error: %v", err)
		}
	}()

	// 启动 HTTP 服务器
	go func() {
		log.Printf("HTTP server listening on :%s", port)
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	cancel() // 停止 Redis 消费者
	log.Println("Service stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}
