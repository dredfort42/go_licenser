/*******************************************************************

		::          ::        +--------+-----------------------+
		  ::      ::          | Author | Dmitry Novikov        |
		::::::::::::::        | Email  | dredfort.42@gmail.com |
	  ::::  ::::::  ::::      +--------+-----------------------+
	::::::::::::::::::::::
	::  ::::::::::::::  ::    File     | licenser_test.go
	::  ::          ::  ::    Created  | 2025-08-08
		  ::::  ::::          Modified | 2025-08-08

	GitHub:   https://github.com/dredfort42
	LinkedIn: https://linkedin.com/in/novikov-da

*******************************************************************/

package licenser_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	licenser "github.com/dredfort42/go_licenser"
)

func TestNewManager(t *testing.T) {
	t.Run("ValidGeneratorMode", func(t *testing.T) {
		config := licenser.Config{
			KeySize:       2048,
			GeneratorMode: true,
		}

		manager, err := licenser.NewManager(config)
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		if manager == nil {
			t.Fatal("Manager should not be nil")
		}
	})

	t.Run("ValidatorModeWithPEMKeys", func(t *testing.T) {
		// First create a manager in generator mode to get keys
		genConfig := licenser.Config{
			KeySize:       1024, // Smaller key for faster tests
			GeneratorMode: true,
		}
		genManager, err := licenser.NewManager(genConfig)
		if err != nil {
			t.Fatalf("Failed to create generator manager: %v", err)
		}

		// Export keys
		_, publicKeyPEM, err := genManager.ExportKeys()
		if err != nil {
			t.Fatalf("Failed to export keys: %v", err)
		}

		// Create validator manager with public key
		validatorConfig := licenser.Config{
			PublicKeyPEM:  publicKeyPEM,
			GeneratorMode: false,
		}

		validatorManager, err := licenser.NewManager(validatorConfig)
		if err != nil {
			t.Fatalf("Failed to create validator manager: %v", err)
		}

		if validatorManager == nil {
			t.Fatal("Validator manager should not be nil")
		}
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		config := licenser.Config{
			GeneratorMode: false, // No keys provided
		}

		_, err := licenser.NewManager(config)
		if err == nil {
			t.Fatal("Expected error for invalid config")
		}
	})

	t.Run("WithKeyPaths", func(t *testing.T) {
		// Create temporary directory
		tempDir := t.TempDir()

		// Generate keys first
		genConfig := licenser.Config{
			KeySize:       1024,
			GeneratorMode: true,
		}
		genManager, err := licenser.NewManager(genConfig)
		if err != nil {
			t.Fatalf("Failed to create generator: %v", err)
		}

		privateKeyPath := filepath.Join(tempDir, "private.pem")
		publicKeyPath := filepath.Join(tempDir, "public.pem")

		err = genManager.SaveKeys(privateKeyPath, publicKeyPath)
		if err != nil {
			t.Fatalf("Failed to save keys: %v", err)
		}

		// Now test with key paths
		config := licenser.Config{
			PrivateKeyPath: privateKeyPath,
			PublicKeyPath:  publicKeyPath,
			GeneratorMode:  true,
		}

		manager, err := licenser.NewManager(config)
		if err != nil {
			t.Fatalf("Failed to create manager with key paths: %v", err)
		}

		if manager == nil {
			t.Fatal("Manager should not be nil")
		}
	})
}

func TestLicenseBuilder(t *testing.T) {
	t.Run("BasicBuilder", func(t *testing.T) {
		builder := licenser.NewBuilder()

		service := licenser.Service{
			ID:   "test-service",
			Name: "Test Service",
		}

		license := builder.
			WithCustomer("Test Customer").
			WithAppID("test-app").
			WithService(service).
			WithExpirationDuration(time.Hour * 24).
			Build()

		if license.Customer != "Test Customer" {
			t.Errorf("Expected customer 'Test Customer', got '%s'", license.Customer)
		}

		if license.AppID != "test-app" {
			t.Errorf("Expected app ID 'test-app', got '%s'", license.AppID)
		}

		if len(license.Services) != 1 {
			t.Errorf("Expected 1 service, got %d", len(license.Services))
		}
	})

	t.Run("CompleteBuilder", func(t *testing.T) {
		builder := licenser.NewBuilder()

		services := []licenser.Service{
			{ID: "service1", Name: "Service 1", Description: "First service"},
			{ID: "service2", Name: "Service 2", Description: "Second service"},
		}

		expirationTime := time.Now().Add(30 * 24 * time.Hour)

		license := builder.
			WithCustomer("Enterprise Customer").
			WithAppID("enterprise-app").
			WithServices(services).
			WithLimit("users", 100).
			WithLimit("requests", 10000).
			WithFeature("premium", true).
			WithFeature("analytics", false).
			WithExpirationTime(expirationTime).
			WithMetadata("department", "engineering").
			WithMetadata("contact", "admin@company.com").
			WithVersion("2.0").
			WithEnvironment("production").
			Build()

		if license.Customer != "Enterprise Customer" {
			t.Errorf("Expected customer 'Enterprise Customer', got '%s'", license.Customer)
		}

		if len(license.Services) != 2 {
			t.Errorf("Expected 2 services, got %d", len(license.Services))
		}

		if license.Limits["users"] != 100 {
			t.Errorf("Expected users limit 100, got %d", license.Limits["users"])
		}

		if !license.Features["premium"] {
			t.Error("Expected premium feature to be enabled")
		}

		if license.Features["analytics"] {
			t.Error("Expected analytics feature to be disabled")
		}

		if license.Version != "2.0" {
			t.Errorf("Expected version '2.0', got '%s'", license.Version)
		}

		if license.Environment != "production" {
			t.Errorf("Expected environment 'production', got '%s'", license.Environment)
		}
	})

	t.Run("BuilderValidation", func(t *testing.T) {
		builder := licenser.NewBuilder()

		// Test validation without required fields
		err := builder.Validate()
		if err == nil {
			t.Error("Expected validation error for incomplete license")
		}

		// Add customer and test again
		builder.WithCustomer("Test Customer")
		err = builder.Validate()
		if err == nil {
			t.Error("Expected validation error for license without app ID")
		}

		// Add app ID and test again
		builder.WithAppID("test-app")
		err = builder.Validate()
		if err == nil {
			t.Error("Expected validation error for license without services")
		}

		// Add service and test validation passes
		service := licenser.Service{ID: "test", Name: "Test Service"}
		builder.WithService(service)
		err = builder.Validate()
		if err != nil {
			t.Errorf("Expected validation to pass, got error: %v", err)
		}
	})

	t.Run("ExpirationMethods", func(t *testing.T) {
		builder := licenser.NewBuilder()

		// Test WithExpiration
		expirationTimestamp := time.Now().Add(time.Hour).Unix()
		license1 := builder.
			WithCustomer("Customer1").
			WithAppID("app1").
			WithService(licenser.Service{ID: "s1", Name: "Service 1"}).
			WithExpiration(expirationTimestamp).
			Build()

		if license1.ExpiresAt != expirationTimestamp {
			t.Errorf("Expected expiration %d, got %d", expirationTimestamp, license1.ExpiresAt)
		}

		// Test WithExpirationTime
		expirationTime := time.Now().Add(2 * time.Hour)
		license2 := licenser.NewBuilder().
			WithCustomer("Customer2").
			WithAppID("app2").
			WithService(licenser.Service{ID: "s2", Name: "Service 2"}).
			WithExpirationTime(expirationTime).
			Build()

		if license2.ExpiresAt != expirationTime.Unix() {
			t.Errorf("Expected expiration %d, got %d", expirationTime.Unix(), license2.ExpiresAt)
		}
	})
}

func TestLicenseGeneration(t *testing.T) {
	config := licenser.Config{
		KeySize:       1024, // Smaller key for faster tests
		GeneratorMode: true,
	}

	manager, err := licenser.NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	service := licenser.Service{
		ID:   "test-service",
		Name: "Test Service",
	}

	license := licenser.License{
		Customer: "Test Customer",
		AppID:    "test-app",
		Services: []licenser.Service{service},
		IssuedAt: time.Now().Unix(),
	}

	signedLicense, err := manager.GenerateLicense(&license)
	if err != nil {
		t.Fatalf("Failed to generate license: %v", err)
	}

	if signedLicense == nil {
		t.Fatal("Signed license should not be nil")
	}

	if signedLicense.Signature == "" {
		t.Error("Signature should not be empty")
	}

	if signedLicense.Algorithm == "" {
		t.Error("Algorithm should not be empty")
	}

	if signedLicense.CreatedAt == 0 {
		t.Error("CreatedAt should not be zero")
	}
}

func TestLicenseValidation(t *testing.T) {
	config := licenser.Config{
		KeySize:       1024,
		GeneratorMode: true,
	}

	manager, err := licenser.NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	t.Run("ValidLicense", func(t *testing.T) {
		service := licenser.Service{
			ID:   "test-service",
			Name: "Test Service",
		}

		license := licenser.License{
			Customer: "Test Customer",
			AppID:    "test-app",
			Services: []licenser.Service{service},
			IssuedAt: time.Now().Unix(),
		}

		signedLicense, err := manager.GenerateLicense(&license)
		if err != nil {
			t.Fatalf("Failed to generate license: %v", err)
		}

		result := manager.ValidateLicense(signedLicense)
		if !result.Valid {
			t.Errorf("License should be valid, errors: %v", result.Errors)
		}
	})

	t.Run("ExpiredLicense", func(t *testing.T) {
		service := licenser.Service{
			ID:   "test-service",
			Name: "Test Service",
		}

		license := licenser.License{
			Customer:  "Test Customer",
			AppID:     "test-app",
			Services:  []licenser.Service{service},
			IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
			ExpiresAt: time.Now().Add(-time.Hour).Unix(), // Expired
		}

		signedLicense, err := manager.GenerateLicense(&license)
		if err != nil {
			t.Fatalf("Failed to generate license: %v", err)
		}

		result := manager.ValidateLicense(signedLicense)
		if result.Valid {
			t.Error("License should be invalid due to expiration")
		}

		if len(result.Errors) == 0 {
			t.Error("Expected validation errors for expired license")
		}
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		service := licenser.Service{
			ID:   "test-service",
			Name: "Test Service",
		}

		license := licenser.License{
			Customer: "Test Customer",
			AppID:    "test-app",
			Services: []licenser.Service{service},
			IssuedAt: time.Now().Unix(),
		}

		signedLicense, err := manager.GenerateLicense(&license)
		if err != nil {
			t.Fatalf("Failed to generate license: %v", err)
		}

		// Tamper with the license data
		signedLicense.Data.Customer = "Tampered Customer"

		result := manager.ValidateLicense(signedLicense)
		if result.Valid {
			t.Error("License should be invalid due to signature mismatch")
		}
	})
}

func TestFileOperations(t *testing.T) {
	config := licenser.Config{
		KeySize:       1024,
		GeneratorMode: true,
	}

	manager, err := licenser.NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tempDir := t.TempDir()

	t.Run("SaveAndLoadLicense", func(t *testing.T) {
		// Create license
		license := licenser.License{
			Customer: "File Test Customer",
			AppID:    "file-test-app",
			Services: []licenser.Service{{ID: "file-service", Name: "File Service"}},
			IssuedAt: time.Now().Unix(),
		}

		signedLicense, err := manager.GenerateLicense(&license)
		if err != nil {
			t.Fatalf("Failed to generate license: %v", err)
		}

		// Save license
		licensePath := filepath.Join(tempDir, "test-license.json")
		err = manager.SaveLicense(signedLicense, licensePath)
		if err != nil {
			t.Fatalf("Failed to save license: %v", err)
		}

		// Load license
		loadedLicense, err := manager.LoadLicense(licensePath)
		if err != nil {
			t.Fatalf("Failed to load license: %v", err)
		}

		if loadedLicense.Data.Customer != license.Customer {
			t.Errorf("Expected customer '%s', got '%s'", license.Customer, loadedLicense.Data.Customer)
		}
	})

	t.Run("LoadAndValidateLicense", func(t *testing.T) {
		// Create and save license
		license := licenser.License{
			Customer: "Validation Test Customer",
			AppID:    "validation-test-app",
			Services: []licenser.Service{{ID: "validation-service", Name: "Validation Service"}},
			IssuedAt: time.Now().Unix(),
		}

		signedLicense, err := manager.GenerateLicense(&license)
		if err != nil {
			t.Fatalf("Failed to generate license: %v", err)
		}

		licensePath := filepath.Join(tempDir, "validation-license.json")
		err = manager.SaveLicense(signedLicense, licensePath)
		if err != nil {
			t.Fatalf("Failed to save license: %v", err)
		}

		// Load and validate license
		loadedLicense, result, err := manager.LoadAndValidateLicense(licensePath)
		if err != nil {
			t.Fatalf("Failed to load and validate license: %v", err)
		}

		if !result.Valid {
			t.Errorf("License should be valid, errors: %v", result.Errors)
		}

		if loadedLicense.Data.Customer != license.Customer {
			t.Errorf("Expected customer '%s', got '%s'", license.Customer, loadedLicense.Data.Customer)
		}
	})

	t.Run("SaveAndLoadKeys", func(t *testing.T) {
		privateKeyPath := filepath.Join(tempDir, "private.pem")
		publicKeyPath := filepath.Join(tempDir, "public.pem")

		// Save keys
		err := manager.SaveKeys(privateKeyPath, publicKeyPath)
		if err != nil {
			t.Fatalf("Failed to save keys: %v", err)
		}

		// Verify files exist
		if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
			t.Error("Private key file should exist")
		}

		if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
			t.Error("Public key file should exist")
		}

		// Create new manager with loaded keys
		newConfig := licenser.Config{
			PrivateKeyPath: privateKeyPath,
			PublicKeyPath:  publicKeyPath,
			GeneratorMode:  true,
		}

		newManager, err := licenser.NewManager(newConfig)
		if err != nil {
			t.Fatalf("Failed to create manager with loaded keys: %v", err)
		}

		// Test that the new manager can generate and validate licenses
		license := licenser.License{
			Customer: "Key Test Customer",
			AppID:    "key-test-app",
			Services: []licenser.Service{{ID: "key-service", Name: "Key Service"}},
			IssuedAt: time.Now().Unix(),
		}

		signedLicense, err := newManager.GenerateLicense(&license)
		if err != nil {
			t.Fatalf("Failed to generate license with loaded keys: %v", err)
		}

		result := newManager.ValidateLicense(signedLicense)
		if !result.Valid {
			t.Errorf("License should be valid with loaded keys, errors: %v", result.Errors)
		}
	})

	t.Run("SavePublicKeyOnly", func(t *testing.T) {
		publicKeyPath := filepath.Join(tempDir, "public-only.pem")

		err := manager.SavePublicKey(publicKeyPath)
		if err != nil {
			t.Fatalf("Failed to save public key: %v", err)
		}

		if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
			t.Error("Public key file should exist")
		}
	})

	t.Run("LoadNonexistentFile", func(t *testing.T) {
		_, err := manager.LoadLicense("/nonexistent/path/license.json")
		if err == nil {
			t.Error("Expected error when loading nonexistent file")
		}
	})
}

func TestKeyExport(t *testing.T) {
	config := licenser.Config{
		KeySize:       1024,
		GeneratorMode: true,
	}

	manager, err := licenser.NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	t.Run("ExportKeys", func(t *testing.T) {
		privateKey, publicKey, err := manager.ExportKeys()
		if err != nil {
			t.Fatalf("Failed to export keys: %v", err)
		}

		if privateKey == "" {
			t.Error("Private key should not be empty")
		}

		if publicKey == "" {
			t.Error("Public key should not be empty")
		}

		// Verify keys contain PEM headers
		if !contains(privateKey, "BEGIN RSA PRIVATE KEY") {
			t.Error("Private key should contain PEM header")
		}

		if !contains(publicKey, "BEGIN PUBLIC KEY") {
			t.Error("Public key should contain PEM header")
		}
	})

	t.Run("ExportPrivateKey", func(t *testing.T) {
		privateKey := manager.ExportPrivateKey()
		if privateKey == "" {
			t.Error("Private key should not be empty")
		}

		if !contains(privateKey, "BEGIN RSA PRIVATE KEY") {
			t.Error("Private key should contain PEM header")
		}
	})

	t.Run("ExportPublicKey", func(t *testing.T) {
		publicKey := manager.ExportPublicKey()
		if publicKey == "" {
			t.Error("Public key should not be empty")
		}

		if !contains(publicKey, "BEGIN PUBLIC KEY") {
			t.Error("Public key should contain PEM header")
		}
	})

	t.Run("GetPublicKey", func(t *testing.T) {
		publicKey := manager.GetPublicKey()
		if publicKey == nil {
			t.Error("Public key should not be nil")
		}
	})
}

func TestExpirationFunctions(t *testing.T) {
	config := licenser.Config{
		KeySize:       1024,
		GeneratorMode: true,
	}

	manager, err := licenser.NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	t.Run("CheckExpiration", func(t *testing.T) {
		// Test active license
		activeLicense := &licenser.License{
			Customer:  "Test Customer",
			AppID:     "test-app",
			Services:  []licenser.Service{{ID: "test", Name: "Test"}},
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		}

		err := manager.CheckExpiration(activeLicense)
		if err != nil {
			t.Errorf("Active license should not have expiration error: %v", err)
		}

		// Test expired license
		expiredLicense := &licenser.License{
			Customer:  "Test Customer",
			AppID:     "test-app",
			Services:  []licenser.Service{{ID: "test", Name: "Test"}},
			IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
			ExpiresAt: time.Now().Add(-time.Hour).Unix(),
		}

		err = manager.CheckExpiration(expiredLicense)
		if err == nil {
			t.Error("Expired license should have expiration error")
		}
	})

	t.Run("GetLicenseInfo", func(t *testing.T) {
		license := &licenser.License{
			Customer: "Info Test Customer",
			AppID:    "info-test-app",
			Services: []licenser.Service{
				{ID: "service1", Name: "Service 1"},
				{ID: "service2", Name: "Service 2"},
			},
			Limits:      map[string]int{"users": 50},
			Features:    map[string]bool{"premium": true},
			IssuedAt:    time.Now().Unix(),
			ExpiresAt:   time.Now().Add(30 * 24 * time.Hour).Unix(),
			Metadata:    map[string]string{"department": "engineering"},
			Version:     "1.0",
			Environment: "production",
		}

		info := manager.GetLicenseInfo(license)
		if info == nil {
			t.Fatal("License info should not be nil")
		}

		if info.Customer != license.Customer {
			t.Errorf("Expected customer '%s', got '%s'", license.Customer, info.Customer)
		}

		if info.AppID != license.AppID {
			t.Errorf("Expected app ID '%s', got '%s'", license.AppID, info.AppID)
		}

		if len(info.Services) != 2 {
			t.Errorf("Expected 2 services, got %d", len(info.Services))
		}

		if info.Status == "" {
			t.Error("Status should not be empty")
		}

		if info.TimeUntilExpiry == "" {
			t.Error("Time until expiry should not be empty")
		}
	})
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("HasService", func(t *testing.T) {
		service1 := licenser.Service{ID: "service-1", Name: "Service One"}
		service2 := licenser.Service{ID: "service-2", Name: "Service Two"}

		license := &licenser.License{
			Customer: "Test Customer",
			AppID:    "test-app",
			Services: []licenser.Service{service1, service2},
			IssuedAt: time.Now().Unix(),
		}

		if !licenser.HasService(license, "service-1") {
			t.Error("Should have service-1")
		}

		if !licenser.HasServiceByID(license, "service-2") {
			t.Error("Should have service-2 by ID")
		}

		if !licenser.HasServiceByName(license, "Service One") {
			t.Error("Should have 'Service One' by name")
		}

		if licenser.HasService(license, "non-existent") {
			t.Error("Should not have non-existent service")
		}

		if licenser.HasServiceByID(license, "non-existent") {
			t.Error("Should not have non-existent service by ID")
		}

		if licenser.HasServiceByName(license, "Non-existent Service") {
			t.Error("Should not have non-existent service by name")
		}
	})

	t.Run("IsExpiringSoon", func(t *testing.T) {
		// License expiring in 1 hour
		soonLicense := &licenser.License{
			Customer:  "Test Customer",
			AppID:     "test-app",
			Services:  []licenser.Service{{ID: "test", Name: "Test"}},
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		}

		if !licenser.IsExpiringSoon(soonLicense, 2*time.Hour) {
			t.Error("License should be expiring soon within 2 hours")
		}

		if licenser.IsExpiringSoon(soonLicense, 30*time.Minute) {
			t.Error("License should not be expiring soon within 30 minutes")
		}

		// License without expiration
		noExpirationLicense := &licenser.License{
			Customer: "Test Customer",
			AppID:    "test-app",
			Services: []licenser.Service{{ID: "test", Name: "Test"}},
			IssuedAt: time.Now().Unix(),
		}

		if licenser.IsExpiringSoon(noExpirationLicense, time.Hour) {
			t.Error("License without expiration should not be expiring soon")
		}
	})

	t.Run("CalculateRemainingTime", func(t *testing.T) {
		futureTime := time.Now().Add(2 * time.Hour).Unix()
		remaining := licenser.CalculateRemainingTime(futureTime)

		if remaining <= time.Hour {
			t.Error("Remaining time should be more than 1 hour")
		}

		pastTime := time.Now().Add(-time.Hour).Unix()
		remainingPast := licenser.CalculateRemainingTime(pastTime)

		if remainingPast > 0 {
			t.Error("Remaining time for past time should be 0 or negative")
		}
	})

	t.Run("FormatTimeUntilExpiry", func(t *testing.T) {
		futureTime := time.Now().Add(25 * time.Hour).Unix()
		formatted := licenser.FormatTimeUntilExpiry(futureTime)

		if formatted == "" {
			t.Error("Formatted time should not be empty")
		}

		// For 25 hours, should contain "d" or "h" for day/hour
		if !strings.Contains(formatted, "d") && !strings.Contains(formatted, "h") && !strings.Contains(formatted, "m") {
			t.Logf("Formatted time: %s", formatted)
			t.Error("Formatted time should contain time units")
		}

		// Test with 0 (never expires)
		neverExpires := licenser.FormatTimeUntilExpiry(0)
		if neverExpires != "License never expired" {
			t.Errorf("Expected 'License never expired', got '%s'", neverExpires)
		}

		// Test with past time
		pastTime := time.Now().Add(-time.Hour).Unix()
		expired := licenser.FormatTimeUntilExpiry(pastTime)
		if expired != "License expired" {
			t.Errorf("Expected 'License expired', got '%s'", expired)
		}
	})

	t.Run("FormatExpiry", func(t *testing.T) {
		futureTime := time.Now().Add(24 * time.Hour).Unix()
		formatted := licenser.FormatExpiry(futureTime)

		if formatted == "" {
			t.Error("Formatted expiry should not be empty")
		}

		// Test with 0 (never expires)
		neverExpires := licenser.FormatExpiry(0)
		if neverExpires != "License never expired" {
			t.Errorf("Expected 'License never expired', got '%s'", neverExpires)
		}
	})

	t.Run("GetLicenseStatus", func(t *testing.T) {
		// Active license
		activeLicense := &licenser.License{
			Customer:  "Test Customer",
			AppID:     "test-app",
			Services:  []licenser.Service{{ID: "test", Name: "Test"}},
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		}

		status := licenser.GetLicenseStatus(activeLicense)
		if status != "active" {
			t.Errorf("Expected status 'active', got '%s'", status)
		}

		// Expired license
		expiredLicense := &licenser.License{
			Customer:  "Test Customer",
			AppID:     "test-app",
			Services:  []licenser.Service{{ID: "test", Name: "Test"}},
			IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
			ExpiresAt: time.Now().Add(-time.Hour).Unix(),
		}

		expiredStatus := licenser.GetLicenseStatus(expiredLicense)
		if expiredStatus != "expired" {
			t.Errorf("Expected status 'expired', got '%s'", expiredStatus)
		}

		// Never expires license
		neverExpiresLicense := &licenser.License{
			Customer: "Test Customer",
			AppID:    "test-app",
			Services: []licenser.Service{{ID: "test", Name: "Test"}},
			IssuedAt: time.Now().Unix(),
		}

		neverExpiresStatus := licenser.GetLicenseStatus(neverExpiresLicense)
		if neverExpiresStatus != "active" {
			t.Errorf("Expected status 'active' for never expires license, got '%s'", neverExpiresStatus)
		}
	})
}

func TestExpiration(t *testing.T) {
	config := licenser.Config{
		KeySize:       1024,
		GeneratorMode: true,
	}

	manager, err := licenser.NewManager(config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test active license
	activeLicense := &licenser.License{
		Customer:  "Test Customer",
		AppID:     "test-app",
		Services:  []licenser.Service{{ID: "test", Name: "Test"}},
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}

	if manager.IsExpired(activeLicense) {
		t.Error("License should not be expired")
	}

	if !manager.IsActive(activeLicense) {
		t.Error("License should be active")
	}

	// Test expired license
	expiredLicense := &licenser.License{
		Customer:  "Test Customer",
		AppID:     "test-app",
		Services:  []licenser.Service{{ID: "test", Name: "Test"}},
		IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
		ExpiresAt: time.Now().Add(-time.Hour).Unix(),
	}

	if !manager.IsExpired(expiredLicense) {
		t.Error("License should be expired")
	}

	if manager.IsActive(expiredLicense) {
		t.Error("License should not be active")
	}

	// Test license that never expires
	neverExpiresLicense := &licenser.License{
		Customer: "Test Customer",
		AppID:    "test-app",
		Services: []licenser.Service{{ID: "test", Name: "Test"}},
		IssuedAt: time.Now().Unix(),
		// ExpiresAt is 0, meaning never expires
	}

	if manager.IsExpired(neverExpiresLicense) {
		t.Error("License that never expires should not be expired")
	}

	if !manager.IsActive(neverExpiresLicense) {
		t.Error("License that never expires should be active")
	}
}

func TestErrorCases(t *testing.T) {
	t.Run("GenerateLicenseWithoutGeneratorMode", func(t *testing.T) {
		// Create public key for validator mode
		genConfig := licenser.Config{
			KeySize:       1024,
			GeneratorMode: true,
		}
		genManager, err := licenser.NewManager(genConfig)
		if err != nil {
			t.Fatalf("Failed to create generator manager: %v", err)
		}

		_, publicKeyPEM, err := genManager.ExportKeys()
		if err != nil {
			t.Fatalf("Failed to export keys: %v", err)
		}

		// Create validator manager
		validatorConfig := licenser.Config{
			PublicKeyPEM:  publicKeyPEM,
			GeneratorMode: false,
		}

		validatorManager, err := licenser.NewManager(validatorConfig)
		if err != nil {
			t.Fatalf("Failed to create validator manager: %v", err)
		}

		license := licenser.License{
			Customer: "Test Customer",
			AppID:    "test-app",
			Services: []licenser.Service{{ID: "test", Name: "Test"}},
			IssuedAt: time.Now().Unix(),
		}

		_, err = validatorManager.GenerateLicense(&license)
		if err == nil {
			t.Error("Expected error when generating license without generator mode")
		}
	})

	t.Run("InvalidJSONFile", func(t *testing.T) {
		config := licenser.Config{
			KeySize:       1024,
			GeneratorMode: true,
		}

		manager, err := licenser.NewManager(config)
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// Create a file with invalid JSON
		tempDir := t.TempDir()
		invalidJSONPath := filepath.Join(tempDir, "invalid.json")
		err = os.WriteFile(invalidJSONPath, []byte("invalid json content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid JSON file: %v", err)
		}

		_, err = manager.LoadLicense(invalidJSONPath)
		if err == nil {
			t.Error("Expected error when loading invalid JSON file")
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Benchmark tests
func BenchmarkNewManager(b *testing.B) {
	config := licenser.Config{
		KeySize:       2048,
		GeneratorMode: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager, err := licenser.NewManager(config)
		if err != nil {
			b.Fatalf("Failed to create manager: %v", err)
		}
		_ = manager
	}
}

func BenchmarkLicenseGeneration(b *testing.B) {
	config := licenser.Config{
		KeySize:       2048,
		GeneratorMode: true,
	}

	manager, err := licenser.NewManager(config)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}

	license := licenser.License{
		Customer: "Benchmark Customer",
		AppID:    "benchmark-app",
		Services: []licenser.Service{{ID: "benchmark", Name: "Benchmark Service"}},
		IssuedAt: time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GenerateLicense(&license)
		if err != nil {
			b.Fatalf("Failed to generate license: %v", err)
		}
	}
}

func BenchmarkLicenseValidation(b *testing.B) {
	config := licenser.Config{
		KeySize:       2048,
		GeneratorMode: true,
	}

	manager, err := licenser.NewManager(config)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}

	license := licenser.License{
		Customer: "Benchmark Customer",
		AppID:    "benchmark-app",
		Services: []licenser.Service{{ID: "benchmark", Name: "Benchmark Service"}},
		IssuedAt: time.Now().Unix(),
	}

	signedLicense, err := manager.GenerateLicense(&license)
	if err != nil {
		b.Fatalf("Failed to generate license: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := manager.ValidateLicense(signedLicense)
		if !result.Valid {
			b.Fatalf("License validation failed: %v", result.Errors)
		}
	}
}

func BenchmarkLicenseBuilder(b *testing.B) {
	services := []licenser.Service{
		{ID: "service1", Name: "Service 1"},
		{ID: "service2", Name: "Service 2"},
		{ID: "service3", Name: "Service 3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		license := licenser.NewBuilder().
			WithCustomer("Benchmark Customer").
			WithAppID("benchmark-app").
			WithServices(services).
			WithLimit("users", 1000).
			WithFeature("premium", true).
			WithExpirationDuration(30*24*time.Hour).
			WithMetadata("environment", "production").
			WithVersion("1.0").
			Build()
		_ = license
	}
}

func BenchmarkSaveLoadLicense(b *testing.B) {
	config := licenser.Config{
		KeySize:       1024, // Smaller key for faster benchmarks
		GeneratorMode: true,
	}

	manager, err := licenser.NewManager(config)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}

	license := licenser.License{
		Customer: "Benchmark Customer",
		AppID:    "benchmark-app",
		Services: []licenser.Service{{ID: "benchmark", Name: "Benchmark Service"}},
		IssuedAt: time.Now().Unix(),
	}

	signedLicense, err := manager.GenerateLicense(&license)
	if err != nil {
		b.Fatalf("Failed to generate license: %v", err)
	}

	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		licensePath := filepath.Join(tempDir, fmt.Sprintf("benchmark-license-%d.json", i))

		err := manager.SaveLicense(signedLicense, licensePath)
		if err != nil {
			b.Fatalf("Failed to save license: %v", err)
		}

		_, err = manager.LoadLicense(licensePath)
		if err != nil {
			b.Fatalf("Failed to load license: %v", err)
		}
	}
}

func BenchmarkUtilityFunctions(b *testing.B) {
	services := make([]licenser.Service, 100) // Large number of services
	for i := 0; i < 100; i++ {
		services[i] = licenser.Service{
			ID:   fmt.Sprintf("service-%d", i),
			Name: fmt.Sprintf("Service %d", i),
		}
	}

	license := &licenser.License{
		Customer: "Benchmark Customer",
		AppID:    "benchmark-app",
		Services: services,
		IssuedAt: time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test service lookup performance
		licenser.HasServiceByID(license, "service-50")
		licenser.HasServiceByName(license, "Service 75")
		licenser.HasService(license, "service-25")
	}
}
