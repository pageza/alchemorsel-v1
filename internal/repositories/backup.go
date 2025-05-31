package repositories

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// BackupConfig holds configuration for database backups
type BackupConfig struct {
	BackupDir          string        // Directory to store backups
	RetentionPeriod    time.Duration // How long to keep backups
	BackupInterval     time.Duration // How often to create backups
	CompressionEnabled bool          // Whether to compress backups
}

// CreateBackup creates a database backup
func CreateBackup(config *BackupConfig) error {
	// Ensure backup directory exists
	if err := os.MkdirAll(config.BackupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	backupFile := filepath.Join(config.BackupDir, fmt.Sprintf("backup-%s.sql", timestamp))

	// Create backup using pg_dump
	cmd := exec.Command("pg_dump",
		"-h", getEnvOrDefault("POSTGRES_HOST", "localhost"),
		"-U", getEnvOrDefault("POSTGRES_USER", "postgres"),
		"-d", getEnvOrDefault("POSTGRES_DB", "alchemorsel"),
		"-F", "c",
		"-f", backupFile,
	)

	// Set environment variables for authentication
	cmd.Env = append(os.Environ(),
		"PGPASSWORD="+getEnvOrDefault("POSTGRES_PASSWORD", "testpass"),
	)

	// Execute backup command
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create backup: %w\nOutput: %s", err, output)
	}

	log.Info("Database backup created successfully",
		zap.String("backup_file", backupFile),
	)

	return nil
}

// RestoreBackup restores a database from a backup file
func RestoreBackup(backupFile string) error {
	// Check if backup file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupFile)
	}

	// Restore backup using pg_restore
	cmd := exec.Command("pg_restore",
		"-h", getEnvOrDefault("POSTGRES_HOST", "localhost"),
		"-U", getEnvOrDefault("POSTGRES_USER", "postgres"),
		"-d", getEnvOrDefault("POSTGRES_DB", "alchemorsel"),
		"-c",
		backupFile,
	)

	// Set environment variables for authentication
	cmd.Env = append(os.Environ(),
		"PGPASSWORD="+getEnvOrDefault("POSTGRES_PASSWORD", "testpass"),
	)

	// Execute restore command
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to restore backup: %w\nOutput: %s", err, output)
	}

	log.Info("Database backup restored successfully",
		zap.String("backup_file", backupFile),
	)

	return nil
}

// CleanupOldBackups removes backups older than the retention period
func CleanupOldBackups(config *BackupConfig) error {
	entries, err := os.ReadDir(config.BackupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			log.Error("Failed to get file info", zap.Error(err))
			continue
		}

		// Check if file is older than retention period
		if now.Sub(info.ModTime()) > config.RetentionPeriod {
			filePath := filepath.Join(config.BackupDir, info.Name())
			if err := os.Remove(filePath); err != nil {
				log.Error("Failed to remove old backup",
					zap.String("file", filePath),
					zap.Error(err),
				)
				continue
			}
			log.Info("Removed old backup file",
				zap.String("file", filePath),
			)
		}
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// StartBackupScheduler starts the backup scheduler
func StartBackupScheduler(ctx context.Context, config *BackupConfig) {
	ticker := time.NewTicker(config.BackupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := CreateBackup(config); err != nil {
				log.Error("Failed to create scheduled backup", zap.Error(err))
				continue
			}

			if err := CleanupOldBackups(config); err != nil {
				log.Error("Failed to cleanup old backups", zap.Error(err))
			}
		}
	}
}
