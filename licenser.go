/*******************************************************************

		::          ::        +--------+-----------------------+
		  ::      ::          | Author | Dmitry Novikov        |
		::::::::::::::        | Email  | dredfort.42@gmail.com |
	  ::::  ::::::  ::::      +--------+-----------------------+
	::::::::::::::::::::::
	::  ::::::::::::::  ::    File     | licenser.go
	::  ::          ::  ::    Created  | 2025-08-08
		  ::::  ::::          Modified | 2025-08-08

	GitHub:   https://github.com/dredfort42
	LinkedIn: https://linkedin.com/in/novikov-da

*******************************************************************/

package licenser

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Common errors.
var (
	ErrInvalidPrivateKey     = errors.New("invalid private key")
	ErrInvalidPublicKey      = errors.New("invalid public key")
	ErrNoPublicKey           = errors.New("no public key provided")
	ErrLicenseExpired        = errors.New("license has expired")
	ErrInvalidSignature      = errors.New("invalid signature")
	ErrSignatureVerification = errors.New("signature verification failed")
	ErrGeneratorModeRequired = errors.New("generator mode is required")
	ErrCustomerRequired      = errors.New("customer name is required")
	ErrAppIDRequired         = errors.New("application ID is required")
	ErrNoServicesAllowed     = errors.New("at least one service must be allowed")
)

// Constants.
const (
	DefaultKeySize      = 2048
	StatusActive        = "active"
	StatusExpired       = "expired"
	LicenseExpired      = "License expired"
	LicenseNeverExpired = "License never expired"
)

// Service represents a licensed service.
type Service struct {
	ID          string            `json:"id"`                    // Unique identifier for the service
	Name        string            `json:"name"`                  // Human-readable name of the service
	Description string            `json:"description,omitempty"` // Optional description of the service
	Metadata    map[string]string `json:"metadata,omitempty"`    // Optional metadata associated with the service
}

// License contains core license information.
type License struct {
	Customer    string            `json:"customer"`              // Name of the customer
	AppID       string            `json:"app_id"`                // Application ID
	Services    []Service         `json:"services"`              // List of licensed services
	Limits      map[string]int    `json:"limits,omitempty"`      // Usage limits
	Features    map[string]bool   `json:"features,omitempty"`    // Feature flags
	IssuedAt    int64             `json:"issued_at"`             // License issuance timestamp
	ExpiresAt   int64             `json:"expires_at,omitempty"`  // License expiration timestamp
	Metadata    map[string]string `json:"metadata,omitempty"`    // Optional metadata associated with the license
	Version     string            `json:"version,omitempty"`     // License version
	Environment string            `json:"environment,omitempty"` // License environment
}

// SignedLicense represents a complete signed license.
type SignedLicense struct {
	Data      License `json:"data"`                // License data
	Signature string  `json:"signature"`           // License signature
	KeyID     string  `json:"key_id,omitempty"`    // Key ID used for signing
	Algorithm string  `json:"algorithm,omitempty"` // Signing algorithm
	CreatedAt int64   `json:"created_at"`          // License creation timestamp
}

// LicenseInfo contains formatted license information for display.
type LicenseInfo struct {
	Customer        string            `json:"customer"`              // Customer name
	AppID           string            `json:"app_id"`                // Application ID
	IssuedAt        time.Time         `json:"issued_at"`             // Issuance timestamp
	ExpiresAt       *time.Time        `json:"expires_at,omitempty"`  // Expiration timestamp
	Status          string            `json:"status"`                // License status
	TimeUntilExpiry string            `json:"time_until_expiry"`     // Time until expiration
	Services        []Service         `json:"services"`              // Licensed services
	Limits          map[string]int    `json:"limits,omitempty"`      // Usage limits
	Features        map[string]bool   `json:"features,omitempty"`    // Feature flags
	Metadata        map[string]string `json:"metadata,omitempty"`    // Optional metadata
	Version         string            `json:"version,omitempty"`     // License version
	Environment     string            `json:"environment,omitempty"` // License environment
}

// Config holds configuration for the license manager.
type Config struct {
	PrivateKeyPath string `json:"private_key_path,omitempty"` // Path to the private key file
	PrivateKeyPEM  string `json:"private_key_pem,omitempty"`  // PEM-encoded private key
	PublicKeyPath  string `json:"public_key_path,omitempty"`  // Path to the public key file
	PublicKeyPEM   string `json:"public_key_pem,omitempty"`   // PEM-encoded public key
	KeySize        int    `json:"key_size,omitempty"`         // Size of the key in bits
	GeneratorMode  bool   `json:"generator_mode,omitempty"`   // Whether to operate in generator mode
}

// ValidationResult contains the result of license validation.
type ValidationResult struct {
	Valid    bool     `json:"valid"`              // Indicates if the license is valid
	Errors   []string `json:"errors,omitempty"`   // List of validation errors
	Warnings []string `json:"warnings,omitempty"` // List of validation warnings
}

// Builder provides a fluent interface for building licenses.
type Builder struct {
	license License
}

// Manager handles license generation and validation.
type Manager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	config     Config
}

// NewManager creates a new license manager.
func NewManager(config Config) (*Manager, error) {
	m := &Manager{config: config}

	if config.KeySize == 0 {
		m.config.KeySize = DefaultKeySize
	}

	var err error

	if config.GeneratorMode {
		// Load or generate private key
		switch {
		case config.PrivateKeyPEM != "":
			m.privateKey, err = parsePrivateKeyFromPEM(config.PrivateKeyPEM)
		case config.PrivateKeyPath != "":
			m.privateKey, err = loadPrivateKeyFromFile(config.PrivateKeyPath)
		default:
			m.privateKey, err = rsa.GenerateKey(rand.Reader, m.config.KeySize)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to setup private key: %w", err)
		}

		m.publicKey = &m.privateKey.PublicKey
	}

	// Load public key if specified separately
	if config.PublicKeyPEM != "" {
		m.publicKey, err = parsePublicKeyFromPEM(config.PublicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	} else if config.PublicKeyPath != "" {
		m.publicKey, err = loadPublicKeyFromFile(config.PublicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load public key: %w", err)
		}
	}

	if m.publicKey == nil {
		return nil, ErrNoPublicKey
	}

	return m, nil
}

// GenerateLicense creates a signed license.
func (m *Manager) GenerateLicense(license *License) (*SignedLicense, error) {
	if !m.config.GeneratorMode {
		return nil, ErrGeneratorModeRequired
	}

	if license.Customer == "" {
		return nil, ErrCustomerRequired
	}

	if license.AppID == "" {
		return nil, ErrAppIDRequired
	}

	if len(license.Services) == 0 {
		return nil, ErrNoServicesAllowed
	}

	if license.IssuedAt == 0 {
		license.IssuedAt = time.Now().Unix()
	}

	data, err := json.Marshal(license)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal license: %w", err)
	}

	signature, err := m.signData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign license: %w", err)
	}

	return &SignedLicense{
		Data:      *license,
		Signature: signature,
		CreatedAt: time.Now().Unix(),
		Algorithm: "RS256",
	}, nil
}

// ValidateLicense validates a signed license.
func (m *Manager) ValidateLicense(signedLicense *SignedLicense) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Verify signature
	data, err := json.Marshal(signedLicense.Data)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "failed to marshal license data")

		return result
	}

	if err := m.verifySignature(data, signedLicense.Signature); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, "signature verification failed")
	}

	// Check expiration
	if signedLicense.Data.ExpiresAt > 0 && time.Now().Unix() > signedLicense.Data.ExpiresAt {
		result.Valid = false
		result.Errors = append(result.Errors, "license has expired")
	}

	// Basic validation
	if signedLicense.Data.Customer == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "customer is required")
	}

	if signedLicense.Data.AppID == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "app ID is required")
	}

	if len(signedLicense.Data.Services) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "at least one service is required")
	}

	return result
}

// SaveLicense saves a license to file.
func (m *Manager) SaveLicense(signedLicense *SignedLicense, filePath string) error {
	data, err := json.MarshalIndent(signedLicense, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal license: %w", err)
	}

	return os.WriteFile(filePath, data, 0600)
}

// LoadLicense loads a license from file.
func (m *Manager) LoadLicense(filePath string) (*SignedLicense, error) {
	// #nosec G304
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read license file: %w", err)
	}

	var signedLicense SignedLicense
	if err := json.Unmarshal(data, &signedLicense); err != nil {
		return nil, fmt.Errorf("failed to unmarshal license: %w", err)
	}

	return &signedLicense, nil
}

// LoadAndValidateLicense loads and validates a license in one call.
func (m *Manager) LoadAndValidateLicense(filePath string) (*SignedLicense, *ValidationResult, error) {
	signedLicense, err := m.LoadLicense(filePath)
	if err != nil {
		return nil, nil, err
	}

	result := m.ValidateLicense(signedLicense)

	return signedLicense, result, nil
}

// SaveKeys saves private and public keys to files.
func (m *Manager) SaveKeys(privateKeyPath, publicKeyPath string) error {
	// Save private key
	privateKeyPEM := m.ExportPrivateKey()
	if err := os.WriteFile(privateKeyPath, []byte(privateKeyPEM), 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	// Save public key
	publicKeyPEM := m.ExportPublicKey()
	if err := os.WriteFile(publicKeyPath, []byte(publicKeyPEM), 0600); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	return nil
}

// SavePublicKey saves the public key to a file.
func (m *Manager) SavePublicKey(filePath string) error {
	if m.publicKey == nil {
		return ErrNoPublicKey
	}

	publicKeyPEM := m.ExportPublicKey()

	return os.WriteFile(filePath, []byte(publicKeyPEM), 0600)
}

// ExportKeys returns both private and public keys as PEM strings.
func (m *Manager) ExportKeys() (privateKey string, publicKey string, err error) {
	return m.ExportPrivateKey(), m.ExportPublicKey(), nil
}

// ExportPrivateKey exports the private key as PEM.
func (m *Manager) ExportPrivateKey() string {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(m.privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return string(privateKeyPEM)
}

// ExportPublicKey exports the public key as PEM.
func (m *Manager) ExportPublicKey() string {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(m.publicKey)
	if err != nil {
		return ""
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM)
}

// GetPublicKey returns the RSA public key.
func (m *Manager) GetPublicKey() *rsa.PublicKey {
	return m.publicKey
}

// IsExpired checks if a license is expired.
func (m *Manager) IsExpired(license *License) bool {
	return license.ExpiresAt > 0 && time.Now().Unix() > license.ExpiresAt
}

// IsActive checks if a license is currently active.
func (m *Manager) IsActive(license *License) bool {
	return !m.IsExpired(license)
}

// CheckExpiration returns an error if the license is expired.
func (m *Manager) CheckExpiration(license *License) error {
	if m.IsExpired(license) {
		return ErrLicenseExpired
	}

	return nil
}

// GetLicenseInfo creates formatted license information.
func (m *Manager) GetLicenseInfo(license *License) *LicenseInfo {
	info := &LicenseInfo{
		Customer:    license.Customer,
		AppID:       license.AppID,
		IssuedAt:    time.Unix(license.IssuedAt, 0),
		Services:    license.Services,
		Limits:      license.Limits,
		Features:    license.Features,
		Metadata:    license.Metadata,
		Version:     license.Version,
		Environment: license.Environment,
	}

	if license.ExpiresAt > 0 {
		expiresAt := time.Unix(license.ExpiresAt, 0)
		info.ExpiresAt = &expiresAt

		if m.IsExpired(license) {
			info.Status = StatusExpired
			info.TimeUntilExpiry = LicenseExpired
		} else {
			info.Status = StatusActive
			remaining := time.Until(expiresAt)
			info.TimeUntilExpiry = formatDuration(remaining)
		}
	} else {
		info.Status = StatusActive
		info.TimeUntilExpiry = LicenseNeverExpired
	}

	return info
}

// NewBuilder creates a new license builder.
func NewBuilder() *Builder {
	return &Builder{
		license: License{
			Services: make([]Service, 0),
			Limits:   make(map[string]int),
			Features: make(map[string]bool),
			Metadata: make(map[string]string),
		},
	}
}

// WithCustomer sets the customer name.
func (b *Builder) WithCustomer(customer string) *Builder {
	b.license.Customer = customer

	return b
}

// WithAppID sets the application ID.
func (b *Builder) WithAppID(appID string) *Builder {
	b.license.AppID = appID

	return b
}

// WithService adds a service to the license.
func (b *Builder) WithService(service Service) *Builder {
	b.license.Services = append(b.license.Services, service)

	return b
}

// WithServices sets all services.
func (b *Builder) WithServices(services []Service) *Builder {
	b.license.Services = services

	return b
}

// WithLimit adds a limit.
func (b *Builder) WithLimit(key string, value int) *Builder {
	b.license.Limits[key] = value

	return b
}

// WithFeature adds a feature flag.
func (b *Builder) WithFeature(key string, enabled bool) *Builder {
	b.license.Features[key] = enabled

	return b
}

// WithExpiration sets the expiration timestamp.
func (b *Builder) WithExpiration(expiresAt int64) *Builder {
	b.license.ExpiresAt = expiresAt

	return b
}

// WithExpirationTime sets the expiration time.
func (b *Builder) WithExpirationTime(expiresAt time.Time) *Builder {
	b.license.ExpiresAt = expiresAt.Unix()

	return b
}

// WithExpirationDuration sets expiration relative to now.
func (b *Builder) WithExpirationDuration(duration time.Duration) *Builder {
	b.license.ExpiresAt = time.Now().Add(duration).Unix()

	return b
}

// WithMetadata adds metadata.
func (b *Builder) WithMetadata(key, value string) *Builder {
	b.license.Metadata[key] = value

	return b
}

// WithVersion sets the version.
func (b *Builder) WithVersion(version string) *Builder {
	b.license.Version = version

	return b
}

// WithEnvironment sets the environment.
func (b *Builder) WithEnvironment(environment string) *Builder {
	b.license.Environment = environment

	return b
}

// Build returns the built license.
func (b *Builder) Build() License {
	if b.license.IssuedAt == 0 {
		b.license.IssuedAt = time.Now().Unix()
	}

	return b.license
}

// Validate validates the license being built.
func (b *Builder) Validate() error {
	if b.license.Customer == "" {
		return ErrCustomerRequired
	}

	if b.license.AppID == "" {
		return ErrAppIDRequired
	}

	if len(b.license.Services) == 0 {
		return ErrNoServicesAllowed
	}

	return nil
}

// Helper functions

func (m *Manager) signData(data []byte) (string, error) {
	hash := sha256.Sum256(data)

	signature, err := rsa.SignPKCS1v15(rand.Reader, m.privateKey, 0, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (m *Manager) verifySignature(data []byte, signatureStr string) error {
	signature, err := base64.StdEncoding.DecodeString(signatureStr)
	if err != nil {
		return ErrInvalidSignature
	}

	hash := sha256.Sum256(data)

	return rsa.VerifyPKCS1v15(m.publicKey, 0, hash[:], signature)
}

func loadPrivateKeyFromFile(filePath string) (*rsa.PrivateKey, error) {
	// #nosec G304
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return parsePrivateKeyFromPEM(string(data))
}

func loadPublicKeyFromFile(filePath string) (*rsa.PublicKey, error) {
	// #nosec G304
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return parsePublicKeyFromPEM(string(data))
}

func parsePrivateKeyFromPEM(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, ErrInvalidPrivateKey
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func parsePublicKeyFromPEM(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, ErrInvalidPublicKey
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, ErrInvalidPublicKey
	}

	return rsaPub, nil
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return LicenseExpired
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}

	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}

	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	if len(parts) == 0 {
		return "Less than 1 minute"
	}

	return strings.Join(parts, " ")
}

// Utility functions for convenience

// HasService checks if a license includes a specific service by ID or name.
func HasService(license *License, identifier string) bool {
	for _, service := range license.Services {
		if service.ID == identifier || service.Name == identifier {
			return true
		}
	}

	return false
}

// HasServiceByID checks if a license includes a specific service by ID.
func HasServiceByID(license *License, serviceID string) bool {
	for _, service := range license.Services {
		if service.ID == serviceID {
			return true
		}
	}

	return false
}

// HasServiceByName checks if a license includes a specific service by name.
func HasServiceByName(license *License, serviceName string) bool {
	for _, service := range license.Services {
		if service.Name == serviceName {
			return true
		}
	}

	return false
}

// IsExpiringSoon checks if a license is expiring within the specified duration.
func IsExpiringSoon(license *License, within time.Duration) bool {
	if license.ExpiresAt == 0 {
		return false
	}

	expiresAt := time.Unix(license.ExpiresAt, 0)

	return time.Until(expiresAt) <= within
}

// CalculateRemainingTime calculates the remaining time for a license.
func CalculateRemainingTime(expiresAt int64) time.Duration {
	if expiresAt == 0 {
		return 0 // Never expires
	}

	remaining := time.Until(time.Unix(expiresAt, 0))
	if remaining < 0 {
		return 0 // Already expired
	}

	return remaining
}

// FormatTimeUntilExpiry formats the time remaining until expiration.
func FormatTimeUntilExpiry(expiresAt int64) string {
	if expiresAt == 0 {
		return LicenseNeverExpired
	}

	remaining := CalculateRemainingTime(expiresAt)
	if remaining == 0 {
		return LicenseExpired
	}

	return formatDuration(remaining)
}

// FormatExpiry formats an expiration timestamp as a human-readable string.
func FormatExpiry(expiresAt int64) string {
	if expiresAt == 0 {
		return LicenseNeverExpired
	}

	return time.Unix(expiresAt, 0).Format("2006-01-02 15:04:05 MST")
}

// GetLicenseStatus returns the status of a license.
func GetLicenseStatus(license *License) string {
	if license.ExpiresAt == 0 {
		return StatusActive
	}

	if time.Now().Unix() > license.ExpiresAt {
		return StatusExpired
	}

	return StatusActive
}
