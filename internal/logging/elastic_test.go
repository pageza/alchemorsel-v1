package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestElasticLogger(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			// Handle index creation
			w.WriteHeader(http.StatusOK)
		case "POST":
			if r.URL.Path == "/test-index/_doc" {
				// Handle log document creation
				w.WriteHeader(http.StatusCreated)
			} else if r.URL.Path == "/test-index/_search" {
				// Handle search request
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"hits": {
						"hits": [
							{
								"_source": {
									"@timestamp": "2024-03-20T12:00:00Z",
									"level": "info",
									"request_id": "test-request-id",
									"message": "test message",
									"fields": {"key": "value"},
									"service": "test-service",
									"env": "test"
								}
							}
						]
					}
				}`))
			}
		}
	}))
	defer server.Close()

	// Create ElasticLogger instance
	logger, err := NewElasticLogger(server.URL, "test-index")
	if err != nil {
		t.Fatalf("Failed to create ElasticLogger: %v", err)
	}

	// Test index creation
	t.Run("CreateIndex", func(t *testing.T) {
		err := logger.CreateIndex()
		if err != nil {
			t.Errorf("CreateIndex failed: %v", err)
		}
	})

	// Test log sending
	t.Run("SendLog", func(t *testing.T) {
		entry := LogEntry{
			Timestamp:   time.Now(),
			Level:       "info",
			RequestID:   "test-request-id",
			Message:     "test message",
			Fields:      map[string]interface{}{"key": "value"},
			Service:     "test-service",
			Environment: "test",
		}

		err := logger.SendLog(entry)
		if err != nil {
			t.Errorf("SendLog failed: %v", err)
		}
	})

	// Test log searching
	t.Run("SearchLogs", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		endTime := time.Now()

		entries, err := logger.SearchLogs("test", startTime, endTime)
		if err != nil {
			t.Errorf("SearchLogs failed: %v", err)
		}

		if len(entries) != 1 {
			t.Errorf("Expected 1 log entry, got %d", len(entries))
		}

		entry := entries[0]
		if entry.Message != "test message" {
			t.Errorf("Expected message 'test message', got '%s'", entry.Message)
		}
		if entry.Level != "info" {
			t.Errorf("Expected level 'info', got '%s'", entry.Level)
		}
		if entry.RequestID != "test-request-id" {
			t.Errorf("Expected request ID 'test-request-id', got '%s'", entry.RequestID)
		}
	})
}

func TestElasticLoggerErrorHandling(t *testing.T) {
	// Create a test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	// Create ElasticLogger instance
	logger, err := NewElasticLogger(server.URL, "test-index")
	if err != nil {
		t.Fatalf("Failed to create ElasticLogger: %v", err)
	}

	// Test error handling for index creation
	t.Run("CreateIndexError", func(t *testing.T) {
		err := logger.CreateIndex()
		if err == nil {
			t.Error("Expected error for CreateIndex, got nil")
		}
	})

	// Test error handling for log sending
	t.Run("SendLogError", func(t *testing.T) {
		entry := LogEntry{
			Timestamp:   time.Now(),
			Level:       "info",
			RequestID:   "test-request-id",
			Message:     "test message",
			Fields:      map[string]interface{}{"key": "value"},
			Service:     "test-service",
			Environment: "test",
		}

		err := logger.SendLog(entry)
		if err == nil {
			t.Error("Expected error for SendLog, got nil")
		}
	})

	// Test error handling for log searching
	t.Run("SearchLogsError", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		endTime := time.Now()

		entries, err := logger.SearchLogs("test", startTime, endTime)
		if err == nil {
			t.Error("Expected error for SearchLogs, got nil")
		}
		if entries != nil {
			t.Error("Expected nil entries for SearchLogs error, got non-nil")
		}
	})
}
