package config

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigManager(t *testing.T) {
	// Create temporary directories for testing
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	auditLogFile := filepath.Join(tmpDir, "audit.log")

	// Create a 32-byte encryption key
	encryptionKey := "0123456789abcdef0123456789abcdef"

	// Initialize ConfigManager
	cm, err := NewConfigManager(encryptionKey, backupDir, auditLogFile)
	require.NoError(t, err)
	require.NotNil(t, cm)

	// Create a test configuration
	testConfig := &Config{
		Environment: Development,
		Database: DatabaseConfig{
			Driver:   "postgres",
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			DBName:   "testdb",
		},
	}

	t.Run("Encryption and Decryption", func(t *testing.T) {
		// Test encryption
		encrypted, err := cm.EncryptConfig(testConfig)
		require.NoError(t, err)
		require.NotNil(t, encrypted)

		// Test decryption
		decrypted, err := cm.DecryptConfig(encrypted)
		require.NoError(t, err)
		require.NotNil(t, decrypted)

		// Verify decrypted config matches original
		assert.Equal(t, testConfig.Environment, decrypted.Environment)
		assert.Equal(t, testConfig.Database.Driver, decrypted.Database.Driver)
		assert.Equal(t, testConfig.Database.Host, decrypted.Database.Host)
		assert.Equal(t, testConfig.Database.Port, decrypted.Database.Port)
		assert.Equal(t, testConfig.Database.User, decrypted.Database.User)
		assert.Equal(t, testConfig.Database.Password, decrypted.Database.Password)
		assert.Equal(t, testConfig.Database.DBName, decrypted.Database.DBName)
	})

	t.Run("Backup and Restore", func(t *testing.T) {
		// Test backup
		err := cm.BackupConfig(testConfig)
		require.NoError(t, err)

		// List backups
		backups, err := cm.ListBackups()
		require.NoError(t, err)
		require.Len(t, backups, 1)

		// Verify backup content
		backup := backups[0]
		assert.NotEmpty(t, backup.Version)
		assert.NotZero(t, backup.Timestamp)
		assert.Equal(t, string(testConfig.Environment), backup.Environment)
		assert.NotEmpty(t, backup.Checksum)

		// Test restore
		restored, err := cm.RestoreConfig(backup.Version)
		require.NoError(t, err)
		require.NotNil(t, restored)

		// Verify restored config matches original
		assert.Equal(t, testConfig.Environment, restored.Environment)
		assert.Equal(t, testConfig.Database.Driver, restored.Database.Driver)
		assert.Equal(t, testConfig.Database.Host, restored.Database.Host)
		assert.Equal(t, testConfig.Database.Port, restored.Database.Port)
		assert.Equal(t, testConfig.Database.User, restored.Database.User)
		assert.Equal(t, testConfig.Database.Password, restored.Database.Password)
		assert.Equal(t, testConfig.Database.DBName, restored.Database.DBName)
	})

	t.Run("Audit Logging", func(t *testing.T) {
		// Test audit logging
		changes := []string{"Updated database password", "Changed port number"}
		err := cm.LogAudit("update", string(testConfig.Environment), "testuser", changes)
		require.NoError(t, err)

		// Read audit logs
		logs, err := cm.GetAuditLogs()
		require.NoError(t, err)
		require.NotEmpty(t, logs)

		// Verify latest log entry
		latestLog := logs[len(logs)-1]
		assert.NotEmpty(t, latestLog.ID)
		assert.NotZero(t, latestLog.Timestamp)
		assert.Equal(t, "update", latestLog.Action)
		assert.Equal(t, string(testConfig.Environment), latestLog.Environment)
		assert.Equal(t, "testuser", latestLog.User)
		assert.Equal(t, changes, latestLog.Changes)
		assert.NotEmpty(t, latestLog.Version)
	})

	t.Run("Error Cases", func(t *testing.T) {
		// Test invalid encryption key
		_, err := NewConfigManager("invalid-key", backupDir, auditLogFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "encryption key must be 32 bytes")

		// Test invalid backup file
		_, err = cm.RestoreConfig("invalid-version")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read backup file")

		// Test invalid audit log file
		invalidCM, _ := NewConfigManager(encryptionKey, backupDir, "/invalid/path/audit.log")
		err = invalidCM.LogAudit("test", "test", "test", []string{"test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write audit log file")
	})
}

func TestConfigVersion(t *testing.T) {
	// Create a test version
	version := ConfigVersion{
		Version:     "test-version",
		Timestamp:   time.Now(),
		Environment: "test",
		Config:      []byte("test-config"),
		Checksum:    "test-checksum",
	}

	// Test JSON marshaling
	data, err := json.Marshal(version)
	require.NoError(t, err)
	require.NotNil(t, data)

	// Test JSON unmarshaling
	var unmarshaled ConfigVersion
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, version.Version, unmarshaled.Version)
	assert.Equal(t, version.Environment, unmarshaled.Environment)
	assert.Equal(t, version.Config, unmarshaled.Config)
	assert.Equal(t, version.Checksum, unmarshaled.Checksum)
}

func TestConfigAuditLog(t *testing.T) {
	// Create a test audit log
	auditLog := ConfigAuditLog{
		ID:          "test-id",
		Timestamp:   time.Now(),
		Action:      "test-action",
		Environment: "test",
		User:        "test-user",
		Changes:     []string{"test-change"},
		Version:     "test-version",
	}

	// Test JSON marshaling
	data, err := json.Marshal(auditLog)
	require.NoError(t, err)
	require.NotNil(t, data)

	// Test JSON unmarshaling
	var unmarshaled ConfigAuditLog
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, auditLog.ID, unmarshaled.ID)
	assert.Equal(t, auditLog.Action, unmarshaled.Action)
	assert.Equal(t, auditLog.Environment, unmarshaled.Environment)
	assert.Equal(t, auditLog.User, unmarshaled.User)
	assert.Equal(t, auditLog.Changes, unmarshaled.Changes)
	assert.Equal(t, auditLog.Version, unmarshaled.Version)
}
