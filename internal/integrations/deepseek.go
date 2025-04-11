package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/utils"
	"go.uber.org/zap"
)

/* Hardcode DEEPSEEK_API_URL and DEEPSEEK_API_KEY for testing purposes */
func GenerateRecipe(query string, attributes map[string]interface{}) (string, error) {
	deepseekURL := "https://api.deepseek.com/chat/completions"
	zap.L().Debug("Hardcoded DeepSeek URL for testing", zap.String("value", deepseekURL))
	zap.L().Debug("Hardcoded API key for testing", zap.String("apiKey", apiKey))

	promptInstructions := "You are a helpful assistant. Create a recipe based on the user's input and profile attributes. Follow the specified prompt instructions."

	var recipe string
	err := utils.Retry(3, 2*time.Second, func() error {
		model := os.Getenv("DEEPSEEK_MODEL")
		if model == "" {
			model = "deepseek-chat"
		}
		payload := map[string]interface{}{
			"model": model,
			"messages": []map[string]string{
				{"role": "system", "content": promptInstructions},
				{"role": "user", "content": query},
			},
			"attributes": attributes,
			"stream":     false,
		}
		if query != "healthcheck" {
			zap.L().Debug("Payload sent to DeepSeek", zap.Any("payload", payload))
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			zap.L().Error("Error marshaling payload", zap.Error(err))
			return err
		}
		zap.L().Debug("Sending request to DeepSeek", zap.String("url", deepseekURL))
		if query != "healthcheck" {
			zap.L().Debug("Sending request to DeepSeek", zap.String("url", deepseekURL))
		}
		req, err := http.NewRequest("POST", deepseekURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			zap.L().Error("Error creating new request", zap.Error(err))
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey)
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			zap.L().Error("Error making HTTP request", zap.Error(err))
			return err
		}
		zap.L().Debug("HTTP response status", zap.Int("status", resp.StatusCode))
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			err := fmt.Errorf("DeepSeek API returned status %d", resp.StatusCode)
			zap.L().Error("HTTP error", zap.Error(err))
			return err
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			zap.L().Error("Error reading response body", zap.Error(err))
			return err
		}
		recipe = string(data)
		zap.L().Debug("Raw API response", zap.String("response", recipe))
		return nil
	})
	return recipe, err
}
