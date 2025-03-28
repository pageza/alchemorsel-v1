package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger implements GORM's logger.Interface using Zap
type GormLogger struct {
	logger *zap.Logger
	level  gormlogger.LogLevel
}

// NewGormLogger creates a new GORM logger that uses Zap
func NewGormLogger(zapLogger *zap.Logger) *GormLogger {
	return &GormLogger{
		logger: zapLogger,
		level:  gormlogger.Info,
	}
}

// LogMode sets the log level for the logger
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

// Info logs info messages
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Info {
		l.logger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn logs warn messages
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Warn {
		l.logger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error logs error messages
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Error {
		l.logger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace logs SQL statements with their execution time
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("elapsed", elapsed),
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fields = append(fields, zap.Error(err))
		l.logger.Error("gorm query error", fields...)
		return
	}

	if l.level >= gormlogger.Info {
		l.logger.Debug("gorm query", fields...)
	}
}
