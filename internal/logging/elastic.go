package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ElasticLogger handles logging to Elasticsearch
type ElasticLogger struct {
	client  *http.Client
	baseURL string
	index   string
}

// NewElasticLogger creates a new Elasticsearch logger
func NewElasticLogger(baseURL, index string) (*ElasticLogger, error) {
	return &ElasticLogger{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: baseURL,
		index:   index,
	}, nil
}

// SendLog sends a log entry to Elasticsearch
func (el *ElasticLogger) SendLog(entry LogEntry) error {
	// Create the document
	doc := map[string]interface{}{
		"@timestamp": entry.Timestamp,
		"level":      entry.Level,
		"request_id": entry.RequestID,
		"message":    entry.Message,
		"fields":     entry.Fields,
		"service":    entry.Service,
		"env":        entry.Environment,
	}

	if entry.Stacktrace != "" {
		doc["stacktrace"] = entry.Stacktrace
	}

	// Marshal the document
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Create the request
	url := fmt.Sprintf("%s/%s/_doc", el.baseURL, el.index)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := el.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("elasticsearch returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SearchLogs searches logs in Elasticsearch
func (el *ElasticLogger) SearchLogs(query string, startTime, endTime time.Time) ([]LogEntry, error) {
	// Create the search query
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"range": map[string]interface{}{
							"@timestamp": map[string]interface{}{
								"gte": startTime.Format(time.RFC3339),
								"lte": endTime.Format(time.RFC3339),
							},
						},
					},
					{
						"multi_match": map[string]interface{}{
							"query":  query,
							"fields": []string{"message", "fields.*"},
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{
				"@timestamp": map[string]interface{}{
					"order": "desc",
				},
			},
		},
	}

	// Marshal the query
	data, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	// Create the request
	url := fmt.Sprintf("%s/%s/_search", el.baseURL, el.index)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := el.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("elasticsearch returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var result struct {
		Hits struct {
			Hits []struct {
				Source LogEntry `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract log entries
	var entries []LogEntry
	for _, hit := range result.Hits.Hits {
		entries = append(entries, hit.Source)
	}

	return entries, nil
}

// CreateIndex creates the Elasticsearch index if it doesn't exist
func (el *ElasticLogger) CreateIndex() error {
	// Create the index mapping
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"@timestamp": map[string]interface{}{
					"type": "date",
				},
				"level": map[string]interface{}{
					"type": "keyword",
				},
				"request_id": map[string]interface{}{
					"type": "keyword",
				},
				"message": map[string]interface{}{
					"type": "text",
				},
				"fields": map[string]interface{}{
					"type":    "object",
					"dynamic": true,
				},
				"stacktrace": map[string]interface{}{
					"type": "text",
				},
				"service": map[string]interface{}{
					"type": "keyword",
				},
				"env": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	// Marshal the mapping
	data, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal index mapping: %w", err)
	}

	// Create the request
	url := fmt.Sprintf("%s/%s", el.baseURL, el.index)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := el.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("elasticsearch returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
