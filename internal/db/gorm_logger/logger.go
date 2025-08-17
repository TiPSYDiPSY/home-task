//nolint:revive,nosnakecase //package name is fine
package gorm_logger

import (
	"context"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

type Logger struct {
	SlowThreshold time.Duration
	LogLevel      logger.LogLevel
	SkipMigration bool // Skip migration-related queries
}

const (
	// DefaultSlowThreshold is the default threshold for slow query detection
	DefaultSlowThreshold = 200 * time.Millisecond
	// MaxSQLLength is the maximum length for SQL logging before truncation
	MaxSQLLength = 200
)

func New() *Logger {
	return &Logger{
		SlowThreshold: DefaultSlowThreshold,
		LogLevel:      logger.Info,
		SkipMigration: true, // Skip migration queries by default
	}
}

func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	l.LogLevel = level

	return l
}

func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		logrus.WithContext(ctx).Infof(msg, data...)
	}
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		logrus.WithContext(ctx).Warnf(msg, data...)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		logrus.WithContext(ctx).Errorf(msg, data...)
	}
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	if l.shouldSkipQuery(sql) {
		return
	}

	// Categorize the query type
	queryType := l.getQueryType(sql)

	fields := logrus.Fields{
		"elapsed":    elapsed,
		"query_type": queryType,
	}

	if rows >= 0 {
		fields["rows"] = rows
	}

	// Add performance indicators
	if elapsed > l.SlowThreshold {
		fields["performance"] = "slow"
	} else if elapsed < 1*time.Millisecond {
		fields["performance"] = "fast"
	}

	cleanSQL := l.cleanSQL(sql)
	logrusLogger := logrus.WithContext(ctx).WithFields(fields)

	switch {
	case err != nil && l.LogLevel >= logger.Error:
		logrusLogger.WithField("error", err).Error("❌ SQL Error: " + cleanSQL)
	case elapsed > l.SlowThreshold && l.LogLevel >= logger.Warn:
		logrusLogger.Warn("🐌 SLOW QUERY: " + cleanSQL)
	case l.LogLevel >= logger.Info:
		emoji := l.getQueryEmoji(queryType)
		logrusLogger.Info(emoji + " " + cleanSQL)
	}
}

func (l *Logger) shouldSkipQuery(sql string) bool {
	skipPatterns := []string{
		"pg_catalog",
		"information_schema",
		"SELECT description FROM",
		"SELECT constraint_name FROM",
		"SELECT CURRENT_DATABASE()",
		"SELECT count(*) FROM INFORMATION_SCHEMA",
		"SELECT count(*) FROM information_schema.tables",
		"SELECT count(*) FROM pg_indexes",
		"SELECT c.column_name, c.is_nullable",
		"SELECT a.attname as column_name",
		"LIMIT 5",
	}

	if l.SkipMigration {
		skipPatterns = append(skipPatterns, "migration_table_name") // Add your migration table name here
	}

	sqlLower := strings.ToLower(sql)
	for _, pattern := range skipPatterns {
		if strings.Contains(sqlLower, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

func (*Logger) cleanSQL(sql string) string {
	// Replace multiple spaces with single space
	cleaned := strings.Join(strings.Fields(sql), " ")

	// Remove escaped quotes around table and column names for better readability
	cleaned = strings.ReplaceAll(cleaned, "\"", "")

	// Truncate very long queries
	if len(cleaned) > MaxSQLLength {
		cleaned = cleaned[:MaxSQLLength] + "..."
	}

	return cleaned
}

// getQueryType categorizes the query based on its SQL string
func (*Logger) getQueryType(sql string) string {
	sql = strings.TrimSpace(strings.ToUpper(sql))

	switch {
	case strings.HasPrefix(sql, "SELECT"):
		return "SELECT"
	case strings.HasPrefix(sql, "INSERT"):
		return "INSERT"
	case strings.HasPrefix(sql, "UPDATE"):
		return "UPDATE"
	case strings.HasPrefix(sql, "DELETE"):
		return "DELETE"
	default:
		return "OTHER"
	}
}

// getQueryEmoji returns an emoji representing the query type
func (*Logger) getQueryEmoji(queryType string) string {
	emojiMap := map[string]string{
		"SELECT": "🔍",
		"INSERT": "➕",
		"UPDATE": "✏️",
		"DELETE": "🗑️",
		"OTHER":  "📜",
	}

	if emoji, ok := emojiMap[queryType]; ok {
		return emoji
	}

	return "❓"
}
