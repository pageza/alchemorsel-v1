package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// SecretManager handles secrets management
type SecretManager struct {
	client *secretsmanager.Client
	cache  map[string]string
	mu     sync.RWMutex
}

// NewSecretManager creates a new SecretManager instance
func NewSecretManager(ctx context.Context) (*SecretManager, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	return &SecretManager{
		client: client,
		cache:  make(map[string]string),
	}, nil
}

// GetSecret retrieves a secret from AWS Secrets Manager or environment variables
func (sm *SecretManager) GetSecret(ctx context.Context, secretName string) (string, error) {
	// First check environment variables
	if value := os.Getenv(secretName); value != "" {
		return value, nil
	}

	// Then check cache
	sm.mu.RLock()
	if value, ok := sm.cache[secretName]; ok {
		sm.mu.RUnlock()
		return value, nil
	}
	sm.mu.RUnlock()

	// Finally, try AWS Secrets Manager
	value, err := sm.getSecretFromAWS(ctx, secretName)
	if err != nil {
		return "", fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	// Update cache
	sm.mu.Lock()
	sm.cache[secretName] = value
	sm.mu.Unlock()

	return value, nil
}

// getSecretFromAWS retrieves a secret from AWS Secrets Manager
func (sm *SecretManager) getSecretFromAWS(ctx context.Context, secretName string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}

	result, err := sm.client.GetSecretValue(ctx, input)
	if err != nil {
		return "", err
	}

	if result.SecretString == nil {
		return "", fmt.Errorf("secret value is nil")
	}

	return *result.SecretString, nil
}

// RefreshSecrets refreshes all cached secrets
func (sm *SecretManager) RefreshSecrets(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Clear cache
	sm.cache = make(map[string]string)

	// List all secrets
	input := &secretsmanager.ListSecretsInput{}
	paginator := secretsmanager.NewListSecretsPaginator(sm.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list secrets: %w", err)
		}

		for _, secret := range page.SecretList {
			value, err := sm.getSecretFromAWS(ctx, *secret.Name)
			if err != nil {
				return fmt.Errorf("failed to get secret %s: %w", *secret.Name, err)
			}
			sm.cache[*secret.Name] = value
		}
	}

	return nil
}

// GetSecretJSON retrieves a JSON secret and unmarshals it into the provided interface
func (sm *SecretManager) GetSecretJSON(ctx context.Context, secretName string, v interface{}) error {
	secret, err := sm.GetSecret(ctx, secretName)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(secret), v)
}
