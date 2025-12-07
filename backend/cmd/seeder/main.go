package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const (
	BaseURL = "http://localhost:9090/api/v1"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	MSPID    string `json:"msp_id"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

type BatchRequest struct {
	BatchID       string `json:"batch_id"`
	FarmName      string `json:"farm_name"`
	HarvestDate   string `json:"harvest_date"`
	Certification string `json:"certification"`
	CertificateID string `json:"certificate_id"`
}

type PackageRequest struct {
	PackageID      string  `json:"package_id"`
	BatchID        string  `json:"batch_id"`
	Weight         float64 `json:"weight"`
	ProductionDate string  `json:"production_date"`
	ExpiryDate     string  `json:"expiry_date"`
}

var (
	teaTypes = []struct {
		Code string
		Name string
	}{
		{"TN", "Chè Xanh Thái Nguyên"},
		{"ST", "Chè Shan Tuyết"},
		{"OL", "Chè Ô Long"},
		{"SE", "Chè Sen Tây Hồ"},
		{"LA", "Chè Lài Bảo Lộc"},
	}
	
	client = &http.Client{Timeout: 30 * time.Second}
	token  string
)

func main() {
	fmt.Println("🌱 Starting to seed 100 sample data entries...")

	// Authenticate
	if err := authenticate(); err != nil {
		fmt.Printf("❌ Authentication failed: %v\n", err)
		return
	}

	batchesCreated := 0
	packagesCreated := 0

	for i := 1; i <= 20; i++ {
		tea := teaTypes[(i-1)%len(teaTypes)]
		batchID := fmt.Sprintf("BATCH_%s_%03d", tea.Code, i)
		
		// Check if batch exists
		exists, err := checkBatchExists(batchID)
		if err != nil {
			fmt.Printf("⚠️ Failed to check batch %s: %v\n", batchID, err)
		}

		if !exists {
			// Create Batch
			err := createBatch(batchID, tea.Name)
			if err != nil {
				fmt.Printf("❌ Failed to create batch %s: %v\n", batchID, err)
				continue
			}
			batchesCreated++
			fmt.Printf("✅ Created batch: %s (%s)\n", batchID, tea.Name)
			
			// Wait for batch to be committed
			time.Sleep(2 * time.Second)
		} else {
			fmt.Printf("ℹ️ Batch %s already exists, skipping creation\n", batchID)
		}

		// Create 5 Packages for this Batch
		for j := 1; j <= 5; j++ {
			pkgID := fmt.Sprintf("PKG_%s_%03d_%02d", tea.Code, i, j)
			err := createPackage(pkgID, batchID, tea.Name)
			if err != nil {
				fmt.Printf("  ❌ Failed to create package %s: %v\n", pkgID, err)
			} else {
				packagesCreated++
				fmt.Printf("  📦 Created package: %s\n", pkgID)
			}
			// Small delay to avoid overwhelming the server
			time.Sleep(500 * time.Millisecond)
		}
		
		// Delay between batches
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n🎉 Seeding complete!")
	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   - Batches created: %d\n", batchesCreated)
	fmt.Printf("   - Packages created: %d\n", packagesCreated)
	fmt.Printf("   - Total data points: %d\n", batchesCreated+packagesCreated)
}

func checkBatchExists(batchID string) (bool, error) {
	httpReq, err := http.NewRequest("GET", BaseURL+"/teatrace/batches/"+batchID, nil)
	if err != nil {
		return false, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(httpReq)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	
	return false, fmt.Errorf("status code %d", resp.StatusCode)
}

func authenticate() error {
	// Generate random email to avoid conflict
	rand.Seed(time.Now().UnixNano())
	email := fmt.Sprintf("seeder_%d@ibn.vn", rand.Intn(100000))
	password := "password123"
	
	fmt.Printf("👤 Attempting to register new user: %s\n", email)

	// Register
	registerReq := RegisterRequest{
		Email:    email,
		Password: password,
		Role:     "admin",
		MSPID:    "Org1MSP",
	}
	body, _ := json.Marshal(registerReq)
	resp, err := client.Post(BaseURL+"/auth/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		// Read body to see error
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed with status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	fmt.Println("✅ Registration successful, logging in...")

	// Login
	loginReq := LoginRequest{Email: email, Password: password}
	body, _ = json.Marshal(loginReq)
	resp, err = client.Post(BaseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var authResp AuthResponse
		if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
			return err
		}
		token = authResp.AccessToken
		fmt.Println("🔑 Logged in successfully")
		return nil
	}

	return fmt.Errorf("login failed after registration with status: %d", resp.StatusCode)
}

func createBatch(batchID, productType string) error {
	req := BatchRequest{
		BatchID:       batchID,
		FarmName:      "HTX Chè Tân Cương",
		HarvestDate:   time.Now().Format("2006-01-02"),
		Certification: "VIETGAP",
		CertificateID: fmt.Sprintf("CERT_%s", batchID),
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", BaseURL+"/teatrace/batches", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		// Read body to see error
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status code %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func createPackage(packageID, batchID, productType string) error {
	req := PackageRequest{
		PackageID:      packageID,
		BatchID:        batchID,
		Weight:         500.0,
		ProductionDate: time.Now().Format("2006-01-02"),
		ExpiryDate:     time.Now().AddDate(2, 0, 0).Format("2006-01-02"),
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", BaseURL+"/teatrace/packages", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		// Read body to see error
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status code %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
