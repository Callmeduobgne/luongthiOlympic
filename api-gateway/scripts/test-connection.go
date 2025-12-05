/*
 * Copyright (c) 2025 IBN Network
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 */

// Copyright 2024 IBN Network (ICTU Blockchain Network)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

