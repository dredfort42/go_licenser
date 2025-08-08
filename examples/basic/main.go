/*******************************************************************

		::          ::        +--------+-----------------------+
		  ::      ::          | Author | Dmitry Novikov        |
		::::::::::::::        | Email  | dredfort.42@gmail.com |
	  ::::  ::::::  ::::      +--------+-----------------------+
	::::::::::::::::::::::
	::  ::::::::::::::  ::    File     | main.go
	::  ::          ::  ::    Created  | 2025-08-08
		  ::::  ::::          Modified | 2025-08-08

	GitHub:   https://github.com/dredfort42
	LinkedIn: https://linkedin.com/in/novikov-da

*******************************************************************/

package main

import (
	"fmt"
	"log"
	"time"

	licenser "github.com/dredfort42/go_licenser"
)

func main() {
	fmt.Println("=== Go Licenser Basic Example ===")

	// Step 1: Create a License Manager
	fmt.Println("1. Creating License Manager...")
	config := licenser.Config{
		KeySize:       2048,
		GeneratorMode: true,
	}

	manager, err := licenser.NewManager(config)
	if err != nil {
		log.Fatal("Failed to create license manager:", err)
	}
	fmt.Println("✓ License Manager created successfully")

	// Step 2: Save keys for future use
	fmt.Println("\n2. Saving cryptographic keys...")
	err = manager.SaveKeys("examples/keys/private.pem", "examples/keys/public.pem")
	if err != nil {
		log.Fatal("Failed to save keys:", err)
	}
	fmt.Println("✓ Keys saved to examples/keys/")

	// Step 3: Create a service definition
	fmt.Println("\n3. Defining licensed services...")
	webService := licenser.Service{
		ID:          "web-api",
		Name:        "Web API Service",
		Description: "REST API access for web application",
	}

	dbService := licenser.Service{
		ID:          "database",
		Name:        "Database Service",
		Description: "Database access and operations",
	}

	// Step 4: Build a license using the fluent interface
	fmt.Println("\n4. Building license...")
	license := licenser.NewBuilder().
		WithCustomer("Acme Corporation").
		WithAppID("acme-web-app-v1").
		WithService(webService).
		WithService(dbService).
		WithLimit("api_calls", 10000).
		WithLimit("users", 100).
		WithLimit("storage_gb", 50).
		WithFeature("analytics", true).
		WithFeature("reporting", false).
		WithFeature("backup", true).
		WithExpirationDuration(365 * 24 * time.Hour). // 1 year
		WithEnvironment("production").
		WithVersion("1.0.0").
		Build()

	fmt.Printf("✓ License built for customer: %s\n", license.Customer)

	// Step 5: Generate and sign the license
	fmt.Println("\n5. Generating and signing license...")
	signedLicense, err := manager.GenerateLicense(&license)
	if err != nil {
		log.Fatal("Failed to generate license:", err)
	}
	fmt.Println("✓ License generated and signed")

	// Step 6: Save license to file
	fmt.Println("\n6. Saving license to file...")
	err = manager.SaveLicense(signedLicense, "examples/licenses/acme-corp.json")
	if err != nil {
		log.Fatal("Failed to save license:", err)
	}
	fmt.Println("✓ License saved to examples/licenses/acme-corp.json")

	// Step 7: Load and validate the license
	fmt.Println("\n7. Loading and validating license...")
	loadedLicense, err := manager.LoadLicense("examples/licenses/acme-corp.json")
	if err != nil {
		log.Fatal("Failed to load license:", err)
	}

	result := manager.ValidateLicense(loadedLicense)
	if !result.Valid {
		log.Printf("License validation failed: %v", result.Errors)
		return
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("⚠️  License validation warnings: %v\n", result.Warnings)
	} else {
		fmt.Println("✓ License is valid!")
	}

	// Step 8: Display license information
	fmt.Println("\n8. License Information:")
	info := manager.GetLicenseInfo(&loadedLicense.Data)
	fmt.Printf("   Customer: %s\n", info.Customer)
	fmt.Printf("   App ID: %s\n", info.AppID)
	fmt.Printf("   Status: %s\n", info.Status)
	fmt.Printf("   Environment: %s\n", info.Environment)
	fmt.Printf("   Version: %s\n", info.Version)
	fmt.Printf("   Services: %d\n", len(info.Services))
	fmt.Printf("   Features: %d\n", len(info.Features))
	fmt.Printf("   Limits: %d\n", len(info.Limits))
	fmt.Printf("   Expires: %s\n", info.TimeUntilExpiry)

	// Step 9: Check specific functionalities
	fmt.Println("\n9. Checking license capabilities...")

	// Check if license is active
	if manager.IsActive(&loadedLicense.Data) {
		fmt.Println("✓ License is currently active")
	}

	// Check service availability
	if licenser.HasServiceByID(&loadedLicense.Data, "web-api") {
		fmt.Println("✓ Web API service is licensed")
	}

	if licenser.HasServiceByID(&loadedLicense.Data, "database") {
		fmt.Println("✓ Database service is licensed")
	}

	// Check feature availability
	fmt.Printf("   Analytics enabled: %v\n", loadedLicense.Data.Features["analytics"])
	fmt.Printf("   Reporting enabled: %v\n", loadedLicense.Data.Features["reporting"])
	fmt.Printf("   Backup enabled: %v\n", loadedLicense.Data.Features["backup"])

	// Check usage limits
	fmt.Printf("   API call limit: %d\n", loadedLicense.Data.Limits["api_calls"])
	fmt.Printf("   User limit: %d\n", loadedLicense.Data.Limits["users"])
	fmt.Printf("   Storage limit: %d GB\n", loadedLicense.Data.Limits["storage_gb"])

	// Check if license is expiring soon (within 30 days)
	if licenser.IsExpiringSoon(&loadedLicense.Data, 30*24*time.Hour) {
		fmt.Println("⚠️  License expires within 30 days")
	} else {
		fmt.Println("✓ License has sufficient time remaining")
	}

	fmt.Println("\n=== Example completed successfully! ===")
}
