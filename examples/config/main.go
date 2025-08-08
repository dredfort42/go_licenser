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
	fmt.Println("=== Simple Configuration Example ===\n")

	// Step 1: Create manager with existing keys
	fmt.Println("1. Creating License Manager with existing keys...")
	config := licenser.Config{
		PrivateKeyPath: "../basic/examples/keys/private.pem",
		PublicKeyPath:  "../basic/examples/keys/public.pem",
		GeneratorMode:  true,
	}

	manager, err := licenser.NewManager(config)
	if err != nil {
		log.Fatal("Failed to create license manager:", err)
	}
	fmt.Println("✓ License Manager created with existing keys")

	// Step 2: Create different license types
	fmt.Println("\n2. Creating different license types...")

	// Basic license
	basicLicense := licenser.NewBuilder().
		WithCustomer("Small Business").
		WithAppID("basic-app").
		WithService(licenser.Service{ID: "basic", Name: "Basic Service"}).
		WithLimit("users", 5).
		WithExpirationDuration(30 * 24 * time.Hour). // 30 days
		Build()

	// Premium license
	premiumLicense := licenser.NewBuilder().
		WithCustomer("Enterprise Corp").
		WithAppID("premium-app").
		WithService(licenser.Service{ID: "api", Name: "API Service"}).
		WithService(licenser.Service{ID: "analytics", Name: "Analytics Service"}).
		WithLimit("users", 1000).
		WithLimit("api_calls", 100000).
		WithFeature("reporting", true).
		WithFeature("backup", true).
		WithExpirationDuration(365 * 24 * time.Hour). // 1 year
		Build()

	// Perpetual license
	perpetualLicense := licenser.NewBuilder().
		WithCustomer("Lifetime Customer").
		WithAppID("lifetime-app").
		WithService(licenser.Service{ID: "all", Name: "All Services"}).
		WithFeature("everything", true).
		// No expiration set - perpetual license
		Build()

	fmt.Println("✓ Created basic, premium, and perpetual licenses")

	// Step 3: Generate and save licenses
	fmt.Println("\n3. Generating and saving licenses...")

	licenses := map[string]*licenser.License{
		"basic":     &basicLicense,
		"premium":   &premiumLicense,
		"perpetual": &perpetualLicense,
	}

	for name, license := range licenses {
		signedLicense, err := manager.GenerateLicense(license)
		if err != nil {
			log.Printf("Failed to generate %s license: %v", name, err)
			continue
		}

		filename := fmt.Sprintf("%s-license.json", name)
		err = manager.SaveLicense(signedLicense, filename)
		if err != nil {
			log.Printf("Failed to save %s license: %v", name, err)
			continue
		}

		fmt.Printf("✓ Generated and saved %s license to %s\n", name, filename)
	}

	// Step 4: Load and validate all licenses
	fmt.Println("\n4. Loading and validating licenses...")

	for name := range licenses {
		filename := fmt.Sprintf("%s-license.json", name)

		signedLicense, result, err := manager.LoadAndValidateLicense(filename)
		if err != nil {
			log.Printf("Failed to load %s license: %v", name, err)
			continue
		}

		if !result.Valid {
			log.Printf("%s license validation failed: %v", name, result.Errors)
			continue
		}

		info := manager.GetLicenseInfo(&signedLicense.Data)
		fmt.Printf("✓ %s license - Customer: %s, Status: %s, Expires: %s\n",
			name, info.Customer, info.Status, info.TimeUntilExpiry)
	}

	// Step 5: Demonstrate utility functions
	fmt.Println("\n5. Demonstrating utility functions...")

	// Load the premium license for testing
	premiumSigned, _, err := manager.LoadAndValidateLicense("premium-license.json")
	if err != nil {
		log.Fatal("Failed to load premium license for testing:", err)
	}

	// Test service checks
	if licenser.HasService(&premiumSigned.Data, "api") {
		fmt.Println("✓ Premium license includes API service")
	}

	if licenser.HasServiceByName(&premiumSigned.Data, "Analytics Service") {
		fmt.Println("✓ Premium license includes Analytics service by name")
	}

	// Test expiration functions
	remaining := licenser.CalculateRemainingTime(premiumSigned.Data.ExpiresAt)
	fmt.Printf("✓ Premium license has %v remaining\n", remaining)

	if licenser.IsExpiringSoon(&premiumSigned.Data, 30*24*time.Hour) {
		fmt.Println("⚠️  Premium license expires within 30 days")
	} else {
		fmt.Println("✓ Premium license has plenty of time remaining")
	}

	// Step 6: Export keys for backup
	fmt.Println("\n6. Exporting keys for backup...")
	privateKey, publicKey, err := manager.ExportKeys()
	if err != nil {
		log.Fatal("Failed to export keys:", err)
	}

	fmt.Printf("✓ Private key length: %d bytes\n", len(privateKey))
	fmt.Printf("✓ Public key length: %d bytes\n", len(publicKey))

	fmt.Println("\n=== Configuration example completed successfully! ===")
}
