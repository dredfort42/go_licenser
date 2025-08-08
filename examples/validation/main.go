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

	licenser "github.com/dredfort42/go_licenser"
)

func main() {
	fmt.Println("=== License Validation Example ===")

	// Step 1: Create a license manager with only the public key
	fmt.Println("1. Creating License Manager for validation...")
	config := licenser.Config{
		PublicKeyPath: "../basic/examples/keys/public.pem",
	}

	manager, err := licenser.NewManager(config)
	if err != nil {
		log.Fatal("Failed to create license manager:", err)
	}
	fmt.Println("✓ License Manager created for validation")

	// Step 2: Load and validate an existing license
	fmt.Println("\n2. Loading and validating license...")
	licensePath := "../basic/examples/licenses/acme-corp.json"

	signedLicense, result, err := manager.LoadAndValidateLicense(licensePath)
	if err != nil {
		log.Fatal("Failed to load or validate license:", err)
	}

	if !result.Valid {
		log.Fatalf("License validation failed: %v", result.Errors)
	}

	fmt.Println("✓ License is valid!")

	// Step 3: Display key information from the validated license
	fmt.Println("\n3. Displaying Validated License Information:")
	info := manager.GetLicenseInfo(&signedLicense.Data)
	fmt.Printf("   Customer: %s\n", info.Customer)
	fmt.Printf("   App ID: %s\n", info.AppID)
	fmt.Printf("   Status: %s\n", info.Status)
	fmt.Printf("   Expires: %s\n", info.TimeUntilExpiry)

	// Step 4: Check specific features or services
	fmt.Println("\n4. Checking License Capabilities:")
	if licenser.HasService(&signedLicense.Data, "web-api") {
		fmt.Println("✓ Web API service is available.")
	}

	if feature, ok := signedLicense.Data.Features["analytics"]; ok && feature {
		fmt.Println("✓ Analytics feature is enabled.")
	} else {
		fmt.Println("✗ Analytics feature is disabled.")
	}

	fmt.Println("\n=== Validation example completed successfully! ===")
}
