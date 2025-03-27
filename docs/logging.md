# Logging System Documentation

## Overview

The Alchemorsel logging system provides a comprehensive solution for log management, including:
- Log rotation and compression
- Log aggregation with Elasticsearch
- Request ID tracking
- Structured logging
- Log analysis and search capabilities
- Security-aware logging

## Table of Contents
1. [Configuration](#configuration)
2. [Log Rotation](#log-rotation)
3. [Log Aggregation](#log-aggregation)
4. [Request ID Tracking](#request-id-tracking)
5. [Structured Logging](#structured-logging)
6. [Log Analysis](#log-analysis)
7. [Log Search](#log-search)
8. [Log Compression](#log-compression)
9. [Security Features](#security-features)
10. [Best Practices](#best-practices)

## Configuration

The logging system is configured through `config/logging/logging.yaml`. Key configuration sections include:

```yaml
file:
  directory: "logs"
  max_size: 100    # MB
  max_backups: 5
  max_age: 30      # days
  compress: true

elasticsearch:
  enabled: true
  url: "http://localhost:9200"
  index: "alchemorsel-logs"
  bulk_size: 1000
  flush_interval: "5s"
```

## Log Rotation

The system implements automatic log rotation with the following features:
- Size-based rotation (configurable max size)
- Age-based rotation (configurable max age)
- Backup management (configurable number of backups)
- Automatic compression of old logs

Example log file structure:
```
logs/
  ├── app.log
  ├── app.log.1
  ├── app.log.2.gz
  └── app.log.3.gz
```

## Log Aggregation

Logs are aggregated in Elasticsearch for centralized storage and analysis:

1. **Bulk Indexing**
   - Logs are collected in batches
   - Configurable batch size and flush interval
   - Automatic retry on failure

2. **Index Management**
   - Daily index rotation
   - Automatic index creation with proper mappings
   - Configurable retention period

## Request ID Tracking

Each request is assigned a unique ID for tracking:

1. **Generation**
   - Automatic generation of request IDs
   - Configurable length and prefix
   - UUID-based for uniqueness

2. **Propagation**
   - Request ID is added to response headers
   - Included in all log entries
   - Passed through service calls

## Structured Logging

Logs are structured in JSON format for better parsing and analysis:

```json
{
  "timestamp": "2024-03-20T12:00:00Z",
  "level": "info",
  "request_id": "req-1234567890",
  "message": "User login successful",
  "fields": {
    "user_id": "user123",
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0..."
  },
  "service": "auth-service",
  "env": "production"
}
```

## Log Analysis

The system includes built-in log analysis capabilities:

1. **Pattern Detection**
   - Error rate monitoring
   - Latency tracking
   - API error analysis

2. **Alerting**
   - Configurable thresholds
   - Time-window based analysis
   - Alert notifications

## Log Search

Elasticsearch provides powerful search capabilities:

1. **Search Features**
   - Full-text search
   - Field-based filtering
   - Time-range queries
   - Field highlighting

2. **Performance**
   - Configurable result limits
   - Optimized index mappings
   - Caching support

## Log Compression

Old logs are automatically compressed:

1. **Compression Settings**
   - Gzip compression
   - Configurable compression level
   - Minimum size threshold
   - Age-based compression

2. **Storage Management**
   - Automatic cleanup of old compressed logs
   - Configurable retention periods
   - Disk space optimization

## Security Features

The logging system includes security-aware features:

1. **Sensitive Data Handling**
   - Automatic masking of sensitive fields
   - Configurable mask patterns
   - Audit logging for security events

2. **Access Control**
   - Role-based access to logs
   - Audit trail for log access
   - Secure storage of sensitive information

## Best Practices

1. **Logging Levels**
   - Use appropriate log levels (debug, info, warn, error, fatal)
   - Include context in log messages
   - Avoid logging sensitive information

2. **Performance**
   - Use bulk operations for log aggregation
   - Implement proper log rotation
   - Monitor log storage usage

3. **Security**
   - Mask sensitive data
   - Implement proper access controls
   - Regular security audits

4. **Monitoring**
   - Monitor log ingestion rates
   - Track error rates
   - Set up alerts for anomalies

## Usage Examples

1. **Basic Logging**
```go
logger.Info("User action completed", map[string]interface{}{
    "user_id": "user123",
    "action": "login",
})
```

2. **Error Logging**
```go
logger.Error("Database connection failed", map[string]interface{}{
    "error": err.Error(),
    "host": "db.example.com",
    "port": 5432,
})
```

3. **Request Logging**
```go
logger.Info("HTTP request", map[string]interface{}{
    "method": "POST",
    "path": "/api/v1/users",
    "status": 200,
    "duration_ms": 150,
})
```

4. **Searching Logs**
```go
entries, err := logger.SearchLogs("error", time.Now().Add(-1*time.Hour), time.Now())
if err != nil {
    // Handle error
}
```

## Troubleshooting

1. **Common Issues**
   - Log rotation failures
   - Elasticsearch connection issues
   - Disk space problems
   - Performance bottlenecks

2. **Solutions**
   - Check file permissions
   - Verify Elasticsearch connectivity
   - Monitor disk usage
   - Review log levels

## Additional Resources

- [Elasticsearch Documentation](https://www.elastic.co/guide/index.html)
- [Logging Best Practices](https://12factor.net/logs)
- [Security Logging Guidelines](https://owasp.org/www-project-cheat-sheets/cheatsheets/Logging_Cheat_Sheet.html) 