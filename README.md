# Go Licenser

A powerful and flexible Go library for generating, signing, and validating software licenses. Perfect for commercial software applications that need robust license management with cryptographic security.

## Features

-   **Cryptographic Security**: RSA-based digital signatures for tamper-proof licenses
-   **Flexible License Structure**: Support for services, features, limits, and custom metadata
-   **Expiration Management**: Built-in support for license expiration and validation with detailed error reporting
-   **Key Management**: Generate, save, and load RSA key pairs
-   **Multiple Output Formats**: JSON-based license files with human-readable information
-   **Service-Based Licensing**: License specific services or features within your application
-   **Builder Pattern**: Fluent API for easy license creation

## Installation

```bash
go get github.com/dredfort42/go_licenser
```

## Quick Start

### 1. Basic License Generation

```go
package main

import (
    "fmt"
    "time"

    "github.com/dredfort42/go_licenser"
)

func main() {
    // Create a license manager in generator mode
    config := licenser.Config{
        GeneratorMode: true,
        KeySize:       2048,
    }

    manager, err := licenser.NewManager(config)
    if err != nil {
        panic(err)
    }

    // Define a service
    service := licenser.Service{
        ID:          "web-api",
        Name:        "Web API Service",
        Description: "REST API access",
    }

    // Build a license using the fluent builder
    license := licenser.NewBuilder().
        WithCustomer("Acme Corporation").
        WithAppID("my-app-v1").
        WithService(service).
        WithFeature("analytics", true).
        WithLimit("api_calls", 10000).
        WithExpirationDuration(365 * 24 * time.Hour). // 1 year
        Build()

    // Generate the signed license
    signedLicense, err := manager.GenerateLicense(&license)
    if err != nil {
        panic(err)
    }

    // Save to file
    err = manager.SaveLicense(signedLicense, "license.json")
    if err != nil {
        panic(err)
    }

    fmt.Println("License generated successfully!")
}
```

### 2. License Validation

```go
package main

import (
    "fmt"

    "github.com/dredfort42/go_licenser"
)

func main() {
    // Create a manager with public key for validation
    config := licenser.Config{
        PublicKeyPath: "public.pem",
    }

    manager, err := licenser.NewManager(config)
    if err != nil {
        panic(err)
    }

    // Load and validate license
    signedLicense, result, err := manager.LoadAndValidateLicense("license.json")
    if err != nil {
        panic(err)
    }

    if result.Valid {
        fmt.Println("License is valid!")

        // Check if a specific service is licensed
        if licenser.HasService(&signedLicense.Data, "web-api") {
            fmt.Println("Web API service is licensed")
        }

        // Get detailed license information
        info := manager.GetLicenseInfo(&signedLicense.Data)
        fmt.Printf("Customer: %s\n", info.Customer)
        fmt.Printf("Status: %s\n", info.Status)
        fmt.Printf("Expires: %s\n", info.TimeUntilExpiry)
    } else {
        fmt.Printf("License validation failed: %v\n", result.Errors)
    }
}
```

## Core Concepts

### License Structure

A license contains:

-   **Customer**: The name of the licensee
-   **AppID**: Unique identifier for your application
-   **Services**: List of licensed services/modules
-   **Features**: Boolean feature flags
-   **Limits**: Numerical usage limits
-   **Expiration**: Optional expiration timestamp
-   **Metadata**: Custom key-value data

### Manager Modes

The `Manager` can operate in two modes:

1. **Generator Mode** (`GeneratorMode: true`): Can generate and sign licenses (requires private key)
2. **Validator Mode** (`GeneratorMode: false`): Can only validate licenses (requires public key)

### Key Management

```go
// Generate and save keys
config := licenser.Config{GeneratorMode: true}
manager, _ := licenser.NewManager(config)

// Save keys to files
err := manager.SaveKeys("private.pem", "public.pem")

// Or export as PEM strings
privateKeyPEM := manager.ExportPrivateKey()
publicKeyPEM := manager.ExportPublicKey()
```

## API Reference

### Core Types

#### `Config`

Configuration for the license manager:

```go
type Config struct {
    PrivateKeyPath string // Path to private key file
    PrivateKeyPEM  string // PEM-encoded private key
    PublicKeyPath  string // Path to public key file
    PublicKeyPEM   string // PEM-encoded public key
    KeySize        int    // RSA key size (default: 2048)
    GeneratorMode  bool   // Enable license generation
}
```

#### `License`

Core license data structure:

```go
type License struct {
    Customer    string            // Customer name
    AppID       string            // Application ID
    Services    []Service         // Licensed services
    Limits      map[string]int    // Usage limits
    Features    map[string]bool   // Feature flags
    IssuedAt    int64             // Issue timestamp
    ExpiresAt   int64             // Expiration timestamp
    Metadata    map[string]string // Custom metadata
    Version     string            // License version
    Environment string            // Target environment
}
```

#### `Service`

Represents a licensable service or module:

```go
type Service struct {
    ID          string            // Unique identifier
    Name        string            // Display name
    Description string            // Optional description
    Metadata    map[string]string // Service-specific metadata
}
```

### Builder API

The `Builder` provides a fluent interface for license creation:

```go
license := licenser.NewBuilder().
    WithCustomer("Customer Name").
    WithAppID("app-id").
    WithService(service).
    WithFeature("feature_name", true).
    WithLimit("limit_name", 1000).
    WithExpirationDuration(time.Hour * 24 * 365).
    WithMetadata("key", "value").
    WithVersion("1.0.0").
    WithEnvironment("production").
    Build()
```

### Utility Functions

```go
// Check if license includes a service
hasService := licenser.HasService(&license, "service-id")
hasServiceByID := licenser.HasServiceByID(&license, "service-id")
hasServiceByName := licenser.HasServiceByName(&license, "Service Name")

// Check expiration
isExpired := licenser.IsExpiringSoon(&license, time.Hour * 24 * 30) // 30 days
remaining := licenser.CalculateRemainingTime(license.ExpiresAt)
status := licenser.GetLicenseStatus(&license)
```

## Examples

The `examples/` directory contains comprehensive examples:

-   **`examples/basic/`**: Complete license generation and validation workflow
-   **`examples/config/`**: Different license types and configurations
-   **`examples/validation/`**: License validation and capability checking

> Before running any examples, make sure RSA keys are generated.  
> You can generate the required keys by running the basic example in `examples/basic/`.

Run any example:

```bash
cd examples/basic && go run main.go
cd examples/config && go run main.go
cd examples/validation && go run main.go
```

## Security Considerations

-   **Private Key Protection**: Keep private keys secure and never distribute them with your application
-   **Public Key Distribution**: Only distribute public keys with your application for license validation
-   **Key Size**: Use at least 2048-bit RSA keys (default)
-   **License Validation**: Always validate licenses on application startup and periodically during runtime
-   **Signature Verification**: The library uses RSA-SHA256 signatures for cryptographic security

## Best Practices

1. **Separate License Generation**: Generate licenses on a secure server, not in client applications
2. **Key Management**: Use proper key storage solutions for production environments
3. **Validation Frequency**: Validate licenses at startup and periodically (but not too frequently to avoid performance impact)
4. **Error Handling**: Gracefully handle license validation failures
5. **Expiration Warnings**: Warn users about upcoming license expiration
6. **Service Checks**: Check specific service licensing before enabling features

## Requirements

-   Go 1.24 or later

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Author

**Dmitry Novikov**

-   Email: [dredfort.42@gmail.com](mailto:dredfort.42@gmail.com)
-   GitHub: [@dredfort42](https://github.com/dredfort42)
-   LinkedIn: [novikov-da](https://linkedin.com/in/novikov-da)
