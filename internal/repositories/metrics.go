package repositories

import (
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Connection pool metrics
	dbMaxOpenConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_max_open_connections",
		Help: "Maximum number of open connections to the database",
	})

	dbOpenConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_open_connections",
		Help: "Current number of open connections to the database",
	})

	dbInUseConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_in_use_connections",
		Help: "Number of connections currently in use",
	})

	dbIdleConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_idle_connections",
		Help: "Number of idle connections",
	})

	dbWaitCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "db_wait_count",
		Help: "Total number of connections waited for",
	})

	dbWaitDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "db_wait_duration_seconds",
		Help:    "How long in seconds connections wait to be acquired",
		Buckets: prometheus.DefBuckets,
	})

	// Connection error metrics
	dbConnectionErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "db_connection_errors_total",
		Help: "Total number of database connection errors",
	})

	dbRetryAttempts = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "db_retry_attempts_total",
		Help: "Total number of database connection retry attempts",
	})

	// Circuit breaker metrics
	dbCircuitBreakerState = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_circuit_breaker_state",
		Help: "Current state of the circuit breaker (0: closed, 1: open)",
	})

	dbCircuitBreakerFailures = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_circuit_breaker_failures",
		Help: "Current number of failures in the circuit breaker",
	})
)

func init() {
	// Register all metrics
	prometheus.MustRegister(dbMaxOpenConnections)
	prometheus.MustRegister(dbOpenConnections)
	prometheus.MustRegister(dbInUseConnections)
	prometheus.MustRegister(dbIdleConnections)
	prometheus.MustRegister(dbWaitCount)
	prometheus.MustRegister(dbWaitDuration)
	prometheus.MustRegister(dbConnectionErrors)
	prometheus.MustRegister(dbRetryAttempts)
	prometheus.MustRegister(dbCircuitBreakerState)
	prometheus.MustRegister(dbCircuitBreakerFailures)
}

// UpdatePoolMetrics updates all connection pool metrics
func UpdatePoolMetrics(stats sql.DBStats) {
	dbMaxOpenConnections.Set(float64(stats.MaxOpenConnections))
	dbOpenConnections.Set(float64(stats.OpenConnections))
	dbInUseConnections.Set(float64(stats.InUse))
	dbIdleConnections.Set(float64(stats.Idle))
	dbWaitCount.Add(float64(stats.WaitCount))
	dbWaitDuration.Observe(stats.WaitDuration.Seconds())
}

// RecordConnectionError increments the connection error counter
func RecordConnectionError() {
	dbConnectionErrors.Inc()
}

// RecordRetryAttempt increments the retry attempt counter
func RecordRetryAttempt() {
	dbRetryAttempts.Inc()
}

// UpdateCircuitBreakerMetrics updates circuit breaker metrics
func UpdateCircuitBreakerMetrics(isOpen bool, failures int32) {
	if isOpen {
		dbCircuitBreakerState.Set(1)
	} else {
		dbCircuitBreakerState.Set(0)
	}
	dbCircuitBreakerFailures.Set(float64(failures))
}

// StartMetricsCollection starts collecting metrics at regular intervals
func StartMetricsCollection(db *sql.DB, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			stats := db.Stats()
			UpdatePoolMetrics(stats)
		}
	}()
}
