package monitoring

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Database metrics
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	dbConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections",
			Help: "Number of active database connections",
		},
	)

	// Recipe metrics
	recipeOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "recipe_operations_total",
			Help: "Total number of recipe operations",
		},
		[]string{"operation", "status"},
	)

	// Rate limiting metrics
	rateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"endpoint"},
	)

	// Cache metrics
	cacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	cacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
	)
)

// MetricsCollector collects and stores metrics
type MetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string][]string
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string][]string),
	}
}

// ObserveHTTPRequest records metrics for an HTTP request
func ObserveHTTPRequest(method, path string, status int, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(method, path, string(status)).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// ObserveDBQuery records metrics for a database query
func ObserveDBQuery(operation string, duration time.Duration) {
	dbQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// SetDBConnections updates the number of active database connections
func SetDBConnections(count int) {
	dbConnections.Set(float64(count))
}

// ObserveRecipeOperation records metrics for a recipe operation
func ObserveRecipeOperation(operation, status string) {
	recipeOperations.WithLabelValues(operation, status).Inc()
}

// ObserveRateLimitHit records a rate limit hit
func ObserveRateLimitHit(endpoint string) {
	rateLimitHits.WithLabelValues(endpoint).Inc()
}

// ObserveCacheHit records a cache hit
func ObserveCacheHit() {
	cacheHits.Inc()
}

// ObserveCacheMiss records a cache miss
func ObserveCacheMiss() {
	cacheMisses.Inc()
}

// RecordMetric records a metric with the given name and value
func (m *MetricsCollector) RecordMetric(name string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Convert value to string representation
	var valueStr string
	switch v := value.(type) {
	case int:
		valueStr = strconv.Itoa(v)
	case int64:
		valueStr = strconv.FormatInt(v, 10)
	case float64:
		valueStr = strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		valueStr = v
	default:
		valueStr = fmt.Sprintf("%v", v)
	}

	// Store the metric
	m.metrics[name] = append(m.metrics[name], valueStr)
}
