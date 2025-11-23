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
	"fmt"

	"github.com/ibn-network/api-gateway/internal/utils"
)

func main() {
	fmt.Println("Generating API Key...")
	fmt.Println("")

	// Generate API key
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		fmt.Printf("❌ Failed to generate API key: %v\n", err)
		return
	}

	// Hash the API key
	keyHash := utils.HashAPIKey(apiKey)

	fmt.Println("✅ API Key generated successfully")
	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Printf("API Key (give to client): %s\n", apiKey)
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("")
	fmt.Println("Store this hash in database:")
	fmt.Printf("Key Hash: %s\n", keyHash)
	fmt.Println("")
	fmt.Println("⚠️  WARNING: The API key is shown only once. Save it securely!")
}

