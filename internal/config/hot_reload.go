package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigWatcher handles configuration hot reloading
type ConfigWatcher struct {
	config   *Config
	watcher  *fsnotify.Watcher
	onChange func(*Config)
	mu       sync.RWMutex
	stopChan chan struct{}
}

// NewConfigWatcher creates a new ConfigWatcher instance
func NewConfigWatcher(cfg *Config, onChange func(*Config)) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &ConfigWatcher{
		config:   cfg,
		watcher:  watcher,
		onChange: onChange,
		stopChan: make(chan struct{}),
	}, nil
}

// Start begins watching for configuration changes
func (cw *ConfigWatcher) Start(ctx context.Context) error {
	// Watch the .env file and environment-specific .env files
	envFiles := []string{
		".env",
		fmt.Sprintf(".env.%s", cw.config.Environment),
	}

	for _, file := range envFiles {
		if err := cw.watcher.Add(file); err != nil {
			return fmt.Errorf("failed to watch file %s: %w", file, err)
		}
	}

	go cw.watch(ctx)

	return nil
}

// Stop stops watching for configuration changes
func (cw *ConfigWatcher) Stop() {
	close(cw.stopChan)
	cw.watcher.Close()
}

// watch monitors for file changes and reloads configuration
func (cw *ConfigWatcher) watch(ctx context.Context) {
	for {
		select {
		case event := <-cw.watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Debounce configuration reload
				time.Sleep(100 * time.Millisecond)

				// Reload configuration
				if err := cw.reloadConfig(); err != nil {
					fmt.Printf("Failed to reload configuration: %v\n", err)
					continue
				}

				// Notify listeners
				if cw.onChange != nil {
					cw.onChange(cw.config)
				}
			}
		case err := <-cw.watcher.Errors:
			fmt.Printf("Error watching configuration files: %v\n", err)
		case <-ctx.Done():
			return
		case <-cw.stopChan:
			return
		}
	}
}

// reloadConfig reloads the configuration from environment variables
func (cw *ConfigWatcher) reloadConfig() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	// Reload environment variables
	if err := LoadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create new configuration
	newConfig, err := NewConfig()
	if err != nil {
		return fmt.Errorf("failed to create new configuration: %w", err)
	}

	// Update configuration
	*cw.config = *newConfig

	return nil
}

// GetConfig returns the current configuration
func (cw *ConfigWatcher) GetConfig() *Config {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.config
}

// WatchDirectory watches a directory for configuration changes
func (cw *ConfigWatcher) WatchDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return cw.watcher.Add(path)
		}

		return nil
	})
}
