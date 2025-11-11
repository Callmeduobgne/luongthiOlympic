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

