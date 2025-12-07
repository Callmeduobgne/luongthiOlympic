//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/services/cache"
	"github.com/ibn-network/api-gateway/internal/services/fabric"
	"github.com/ibn-network/api-gateway/internal/utils"
)

func main() {
	fmt.Println("Testing connections to all services...")
	fmt.Println("")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Configuration loaded")

	// Create logger
	logger := utils.NewDevelopmentLogger()

	// Test PostgreSQL
	fmt.Println("\n📊 Testing PostgreSQL connection...")
	db, err := config.NewPostgresPool(&cfg.Database)
	if err != nil {
		fmt.Printf("❌ PostgreSQL connection failed: %v\n", err)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			fmt.Printf("❌ PostgreSQL ping failed: %v\n", err)
		} else {
			fmt.Println("✅ PostgreSQL connected successfully")
		}
		db.Close()
	}

	// Test Redis
	fmt.Println("\n🔴 Testing Redis connection...")
	redisService, err := cache.NewService(&cfg.Redis, logger)
	if err != nil {
		fmt.Printf("❌ Redis connection failed: %v\n", err)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := redisService.Health(ctx); err != nil {
			fmt.Printf("❌ Redis health check failed: %v\n", err)
		} else {
			fmt.Println("✅ Redis connected successfully")
		}
		redisService.Close()
	}

	// Test Fabric Gateway
	fmt.Println("\n⛓️  Testing Fabric Gateway connection...")
	fabricGateway, err := fabric.NewGatewayService(&cfg.Fabric, &cfg.CircuitBreaker, logger)
	if err != nil {
		fmt.Printf("❌ Fabric Gateway connection failed: %v\n", err)
		fmt.Println("\nTroubleshooting:")
		fmt.Println("- Check if Fabric network is running: docker ps | grep peer")
		fmt.Println("- Verify certificate paths in .env")
		fmt.Println("- Check peer endpoint is accessible")
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := fabricGateway.Health(ctx); err != nil {
			fmt.Printf("⚠️  Fabric Gateway connected but health check failed: %v\n", err)
			fmt.Println("  This is normal if no batches exist yet")
		} else {
			fmt.Println("✅ Fabric Gateway connected successfully")
		}
		fabricGateway.Close()
	}

	fmt.Println("\n✅ Connection test complete")
}

