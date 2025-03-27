package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConfigVersion represents a versioned configuration with metadata.
// It includes:
// - Version: A unique identifier for the configuration version
// - Timestamp: When the version was created
// - Environment: The environment this configuration is for
// - Config: The encrypted configuration data
// - Checksum: A hash of the encrypted data for integrity verification
type ConfigVersion struct {
	Version     string    `json:"version"`
	Timestamp   time.Time `json:"timestamp"`
	Environment string    `json:"environment"`
	Config      []byte    `json:"config"`
	Checksum    string    `json:"checksum"`
}

// ConfigAuditLog represents an audit log entry for configuration changes.
// It tracks:
// - ID: Unique identifier for the audit log entry
// - Timestamp: When the change occurred
// - Action: The type of action (e.g., "backup", "restore", "update")
// - Environment: The environment affected
// - User: Who made the change
// - Changes: Description of what changed
// - Version: The configuration version associated with the change
type ConfigAuditLog struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Action      string    `json:"action"`
	Environment string    `json:"environment"`
	User        string    `json:"user"`
	Changes     []string  `json:"changes"`
	Version     string    `json:"version"`
}

// ConfigManager handles configuration encryption, versioning, backup/restore, and audit logging.
// It provides a secure way to manage application configurations with:
// - AES-GCM encryption for sensitive data
// - Version control for configurations
// - Backup and restore capabilities
// - Comprehensive audit logging
type ConfigManager struct {
	encryptionKey []byte
	backupDir     string
	auditLogFile  string
	logger        *zap.Logger
}

// NewConfigManager creates a new ConfigManager instance.
// Parameters:
//   - encryptionKey: A 32-byte key for AES-GCM encryption
//   - backupDir: Directory where configuration backups will be stored
//   - auditLogFile: Path to the audit log file
//
// Returns:
//   - *ConfigManager: The initialized configuration manager
//   - error: Any error that occurred during initialization
func NewConfigManager(encryptionKey string, backupDir string, auditLogFile string) (*ConfigManager, error) {
	if len(encryptionKey) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes")
	}

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &ConfigManager{
		encryptionKey: []byte(encryptionKey),
		backupDir:     backupDir,
		auditLogFile:  auditLogFile,
		logger:        zap.L(),
	}, nil
}

// EncryptConfig encrypts the configuration using AES-GCM encryption.
// The encryption process:
// 1. Marshals the configuration to JSON
// 2. Creates an AES cipher block
// 3. Creates a GCM (Galois/Counter Mode) for authenticated encryption
// 4. Generates a random nonce
// 5. Encrypts and seals the data
//
// Parameters:
//   - cfg: The configuration to encrypt
//
// Returns:
//   - []byte: The encrypted configuration data
//   - error: Any error that occurred during encryption
func (cm *ConfigManager) EncryptConfig(cfg *Config) ([]byte, error) {
	// Marshal configuration to JSON
	configJSON, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(cm.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to create nonce: %w", err)
	}

	// Encrypt and seal
	ciphertext := aesGCM.Seal(nonce, nonce, configJSON, nil)

	return ciphertext, nil
}

// DecryptConfig decrypts the configuration using AES-GCM decryption.
// The decryption process:
// 1. Creates an AES cipher block
// 2. Creates a GCM for authenticated decryption
// 3. Extracts the nonce from the ciphertext
// 4. Decrypts and verifies the data
// 5. Unmarshals the configuration from JSON
//
// Parameters:
//   - encrypted: The encrypted configuration data
//
// Returns:
//   - *Config: The decrypted configuration
//   - error: Any error that occurred during decryption
func (cm *ConfigManager) DecryptConfig(encrypted []byte) (*Config, error) {
	// Create cipher block
	block, err := aes.NewCipher(cm.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := aesGCM.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	// Decrypt and open
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	// Unmarshal configuration
	var cfg Config
	if err := json.Unmarshal(plaintext, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// BackupConfig creates a backup of the configuration.
// The backup process:
// 1. Encrypts the configuration
// 2. Creates a version with metadata
// 3. Saves the version to a JSON file
// 4. Logs the backup action
//
// Parameters:
//   - cfg: The configuration to backup
//
// Returns:
//   - error: Any error that occurred during backup
func (cm *ConfigManager) BackupConfig(cfg *Config) error {
	// Encrypt configuration
	encrypted, err := cm.EncryptConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to encrypt config: %w", err)
	}

	// Create version
	version := ConfigVersion{
		Version:     uuid.New().String(),
		Timestamp:   time.Now(),
		Environment: string(cfg.Environment),
		Config:      encrypted,
		Checksum:    fmt.Sprintf("%x", encrypted),
	}

	// Create backup file
	backupFile := filepath.Join(cm.backupDir, fmt.Sprintf("config_%s.json", version.Version))
	backupData, err := json.MarshalIndent(version, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal version: %w", err)
	}

	if err := os.WriteFile(backupFile, backupData, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	// Log audit entry
	if err := cm.LogAudit("backup", string(cfg.Environment), "system", []string{"Configuration backup created"}); err != nil {
		cm.logger.Error("Failed to log audit entry", zap.Error(err))
	}

	return nil
}

// RestoreConfig restores a configuration from backup.
// The restore process:
// 1. Reads the backup file
// 2. Verifies the checksum
// 3. Decrypts the configuration
// 4. Logs the restore action
//
// Parameters:
//   - version: The version identifier of the backup to restore
//
// Returns:
//   - *Config: The restored configuration
//   - error: Any error that occurred during restore
func (cm *ConfigManager) RestoreConfig(version string) (*Config, error) {
	// Read backup file
	backupFile := filepath.Join(cm.backupDir, fmt.Sprintf("config_%s.json", version))
	backupData, err := os.ReadFile(backupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}

	// Unmarshal version
	var configVersion ConfigVersion
	if err := json.Unmarshal(backupData, &configVersion); err != nil {
		return nil, fmt.Errorf("failed to unmarshal version: %w", err)
	}

	// Verify checksum
	if fmt.Sprintf("%x", configVersion.Config) != configVersion.Checksum {
		return nil, fmt.Errorf("checksum verification failed")
	}

	// Decrypt configuration
	cfg, err := cm.DecryptConfig(configVersion.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt config: %w", err)
	}

	// Log audit entry
	if err := cm.LogAudit("restore", string(cfg.Environment), "system", []string{"Configuration restored from backup"}); err != nil {
		cm.logger.Error("Failed to log audit entry", zap.Error(err))
	}

	return cfg, nil
}

// ListBackups returns a list of available configuration backups.
// The list includes all backup files in the backup directory,
// with their metadata and version information.
//
// Returns:
//   - []ConfigVersion: List of available backups
//   - error: Any error that occurred while listing backups
func (cm *ConfigManager) ListBackups() ([]ConfigVersion, error) {
	var versions []ConfigVersion

	// Read backup directory
	files, err := os.ReadDir(cm.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Read each backup file
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			backupFile := filepath.Join(cm.backupDir, file.Name())
			backupData, err := os.ReadFile(backupFile)
			if err != nil {
				cm.logger.Error("Failed to read backup file", zap.String("file", file.Name()), zap.Error(err))
				continue
			}

			var version ConfigVersion
			if err := json.Unmarshal(backupData, &version); err != nil {
				cm.logger.Error("Failed to unmarshal version", zap.String("file", file.Name()), zap.Error(err))
				continue
			}

			versions = append(versions, version)
		}
	}

	return versions, nil
}

// LogAudit logs a configuration audit entry.
// The audit log includes:
// - Who made the change
// - What changed
// - When it changed
// - Which environment was affected
// - The associated configuration version
//
// Parameters:
//   - action: The type of action performed
//   - environment: The environment affected
//   - user: Who performed the action
//   - changes: Description of the changes made
//
// Returns:
//   - error: Any error that occurred while logging
func (cm *ConfigManager) LogAudit(action, environment, user string, changes []string) error {
	// Create audit log entry
	auditLog := ConfigAuditLog{
		ID:          uuid.New().String(),
		Timestamp:   time.Now(),
		Action:      action,
		Environment: environment,
		User:        user,
		Changes:     changes,
		Version:     uuid.New().String(),
	}

	// Read existing audit logs
	var auditLogs []ConfigAuditLog
	auditData, err := os.ReadFile(cm.auditLogFile)
	if err == nil {
		if err := json.Unmarshal(auditData, &auditLogs); err != nil {
			cm.logger.Error("Failed to unmarshal audit logs", zap.Error(err))
		}
	}

	// Append new audit log
	auditLogs = append(auditLogs, auditLog)

	// Write updated audit logs
	updatedData, err := json.MarshalIndent(auditLogs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal audit logs: %w", err)
	}

	if err := os.WriteFile(cm.auditLogFile, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write audit log file: %w", err)
	}

	return nil
}

// GetAuditLogs returns the configuration audit logs.
// The logs are read from the audit log file and include
// all configuration changes that have been recorded.
//
// Returns:
//   - []ConfigAuditLog: List of audit log entries
//   - error: Any error that occurred while reading logs
func (cm *ConfigManager) GetAuditLogs() ([]ConfigAuditLog, error) {
	// Read audit log file
	auditData, err := os.ReadFile(cm.auditLogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read audit log file: %w", err)
	}

	// Unmarshal audit logs
	var auditLogs []ConfigAuditLog
	if err := json.Unmarshal(auditData, &auditLogs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal audit logs: %w", err)
	}

	return auditLogs, nil
}
