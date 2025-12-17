// Package server provides a professional-grade dual-domain logging system.
//
// Logging Architecture:
//
// The logging system implements two distinct domains for different purposes:
//
// 1. Console Logger (Logger):
//   - Human-readable, single-line format optimized for debugging
//   - Compact output: "LEVEL [request_id] message key=value key=value"
//   - Focused on critical information for quick scanning
//   - Available in all environments, controlled by LOG_LEVEL
//
// 2. File Logger (FileLogger):
//   - JSON structured logs with comprehensive information
//   - Used for metrics collection, analysis, and monitoring
//   - Always enabled with full structured data
//   - Includes all context fields, metadata, and performance metrics
//
// Logging Levels:
//
// The system supports five logging levels with clear usage guidelines:
//
//   - DEBUG: Detailed diagnostic information for development/debugging
//
//   - Use for: Detailed flow tracking, variable dumps, development diagnostics
//
//   - Avoid in production code paths unless necessary for troubleshooting
//
//   - INFO: General informational messages about normal operations
//
//   - Use for: Successful operations, business events, important state changes
//
//   - This is the default level for production
//
//   - WARN: Warning messages for recoverable issues
//
//   - Use for: Degraded functionality, recoverable errors, deprecation notices
//
//   - Also used for slow operations (>1s) in middleware
//
//   - ERROR: Error messages for failures requiring attention
//
//   - Use for: Operation failures, exceptions, errors that need investigation
//
//   - Automatically written to error.log for easy error tracking
//
//   - FATAL: Critical errors causing application shutdown
//
//   - Use for: Unrecoverable errors, critical system failures
//
//   - Should be used sparingly - causes application exit
//
// Best Practices:
//
// 1. Log at appropriate levels:
//   - Don't use DEBUG for production information
//   - Use INFO for normal operations and business events
//   - Use WARN for recoverable issues and performance concerns
//   - Use ERROR for failures requiring attention
//   - Use FATAL only for unrecoverable critical failures
//
// 2. Include relevant context:
//   - Context fields (request_id, user_id, duration) are automatically extracted
//   - Add additional context via key-value pairs when relevant
//   - Include identifiers that help trace issues (IDs, names, etc.)
//
// 3. Avoid logging sensitive data:
//   - Never log passwords, tokens, or sensitive user information
//   - Be cautious with PII (Personally Identifiable Information)
//   - Sanitize data before logging if necessary
//
// 4. Log errors with context:
//   - Always include the error object, not just error messages
//   - Add relevant context (what operation failed, why it matters)
//   - Include request IDs for distributed tracing
//
// 5. Use structured fields:
//   - Use key-value pairs, not string concatenation
//   - This enables better log analysis and filtering
//   - Example: LogInfo(ctx, "User created", "user_id", userID, "email", email)
//
// 6. Avoid log spam:
//   - Don't log in tight loops or high-frequency operations
//   - Use DEBUG level for verbose logging that's not needed in production
//   - Consider sampling for high-volume operations
//
// 7. Log important business events:
//   - User actions (login, logout, data modifications)
//   - State changes (status updates, configuration changes)
//   - External integrations (API calls, webhook processing)
//
// 8. Log performance metrics:
//   - Log slow operations (>1s) as WARN level
//   - Include duration in logs for performance analysis
//   - Track important metrics (cache hits/misses, DB query times)
//
// Environment Configuration:
//
// The logging system is configured via environment variables:
//
//   - LOG_LEVEL: Set the minimum log level (DEBUG, INFO, WARN, ERROR, FATAL)
//     Default: INFO
//
//   - LOG_DIR: Directory for log files
//     Default: "/app/logs"
//
//   - LOG_CONSOLE_ENABLED: Enable/disable console logging (true/false)
//     Default: "true"
//
// Example Usage:
//
//	// Simple info log
//	LogInfo(ctx, "User created successfully", "user_id", userID)
//
//	// Warning with error
//	LogWarn(ctx, "Cache miss", err, "key", cacheKey, "fallback", "database")
//
//	// Error with context
//	LogError(ctx, "Failed to process payment", err,
//	    "order_id", orderID, "amount", amount, "payment_method", method)
//
//	// Debug for development
//	LogDebug(ctx, "Processing request", "method", r.Method, "path", r.URL.Path)
// ... existing package documentation ...

package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"unipilot/internal/errors"
)

var Logger *zap.SugaredLogger
var FileLogger *zap.SugaredLogger

// InitLogger initializes the professional-grade dual-domain logging system.
func InitLogger() {
	// Step 1: Determine log directory from environment or use default
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "/app/logs"
	}
	// Create directory structure if it doesn't exist (required for file logging)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic("Failed to create logs directory: " + err.Error())
	}

	// Step 2: Parse log level from environment (defaults to INFO if invalid)
	logLevel := parseLogLevel(os.Getenv("LOG_LEVEL"))

	// Step 3: Check if console logging is enabled (default: true for development)
	consoleEnabled := true
	if consoleEnabledStr := os.Getenv("LOG_CONSOLE_ENABLED"); consoleEnabledStr != "" {
		// Parse boolean string, ignore errors and keep default
		if enabled, err := strconv.ParseBool(consoleEnabledStr); err == nil {
			consoleEnabled = enabled
		}
	}

	// Step 4: Configure file logger encoder (JSON format for structured logging)
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.TimeKey = "timestamp"
	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	fileEncoderConfig.MessageKey = "message"
	fileEncoderConfig.StacktraceKey = zapcore.OmitKey

	var fileCores []zapcore.Core

	// Step 5: Create app.log writer (all logs from INFO level and above)
	// 100MB max size, 10 backups, 30 days retention, compressed
	fileWriter := &lumberjack.Logger{
		Filename:   logDir + "/app.log",
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
	}
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(fileEncoderConfig),
		zapcore.AddSync(fileWriter),
		zapcore.InfoLevel, // DEBUG logs excluded from files (console only)
	)
	fileCores = append(fileCores, fileCore)

	// Step 6: Create error.log writer (ERROR and FATAL levels only)
	// Smaller size (50MB) and longer retention (60 days) for error tracking
	errorWriter := &lumberjack.Logger{
		Filename:   logDir + "/error.log",
		MaxSize:    50,
		MaxBackups: 5,
		MaxAge:     60,
		Compress:   true,
	}
	errorCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(fileEncoderConfig),
		zapcore.AddSync(errorWriter),
		zapcore.ErrorLevel, // Only ERROR and FATAL
	)
	fileCores = append(fileCores, errorCore)

	// Step 7: Create file logger with both cores (app.log + error.log)
	fileConfig := zapcore.NewTee(fileCores...)
	fileBaseLogger := zap.New(
		fileConfig,
		zap.AddCaller(),                       // Include file:line in logs
		zap.AddStacktrace(zapcore.ErrorLevel), // Stack traces for errors
	)
	FileLogger = fileBaseLogger.Sugar()

	// Step 8: Configure console logger (compact single-line format for debugging)
	var consoleLogger *zap.SugaredLogger
	if consoleEnabled {
		// Compact encoder config: short keys, time-only format, readable output
		consoleEncoderConfig := zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey, // Omit function name to reduce noise
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339), // HH:MM:SS format
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoderConfig),
			zapcore.AddSync(os.Stdout),
			logLevel, // Respect LOG_LEVEL for console (can be DEBUG in dev)
		)

		consoleBaseLogger := zap.New(
			consoleCore,
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
		consoleLogger = consoleBaseLogger.Sugar()
	} else {
		// Fallback to file logger if console disabled (e.g., in production containers)
		consoleLogger = FileLogger
	}

	Logger = consoleLogger
}

// parseLogLevel parses the log level from environment variable string.
// Returns zapcore.InfoLevel as default if invalid or empty.
func parseLogLevel(levelStr string) zapcore.Level {
	// Normalize input: uppercase and trim whitespace
	levelStr = strings.ToUpper(strings.TrimSpace(levelStr))
	switch levelStr {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "FATAL":
		return zapcore.FatalLevel
	default:
		// Default to INFO for production safety (avoids DEBUG spam)
		return zapcore.InfoLevel
	}
}

// logWithLevel is a helper function that handles the common logging pattern:
// extract console/file fields, merge with additional fields, remove tags from console,
// and write to both console and file loggers at the specified level.
func logWithLevel(ctx context.Context, level zapcore.Level, message string, keysAndValues ...interface{}) {
	// Step 1: Extract console fields (minimal for readability)
	consoleFields := extractConsoleFields(ctx)
	consoleAllFields := mergeFields(consoleFields, keysAndValues...)

	// Step 2: Extract file fields (comprehensive for analysis)
	fileFields := extractContextFields(ctx)
	fileAllFields := mergeFields(fileFields, keysAndValues...)

	// Step 4: Write to console logger (if enabled and level permits)
	if Logger != nil {
		switch level {
		case zapcore.DebugLevel:
			Logger.Debugw(message, consoleAllFields...)
		case zapcore.InfoLevel:
			Logger.Infow(message, consoleAllFields...)
		case zapcore.WarnLevel:
			Logger.Warnw(message, consoleAllFields...)
		case zapcore.ErrorLevel:
			Logger.Errorw(message, consoleAllFields...)
		case zapcore.FatalLevel:
			Logger.Fatalw(message, consoleAllFields...)
		}
	}

	// Step 5: Write to file logger (always, if level permits)
	if FileLogger != nil {
		switch level {
		case zapcore.DebugLevel:
			FileLogger.Debugw(message, fileAllFields...)
		case zapcore.InfoLevel:
			FileLogger.Infow(message, fileAllFields...)
		case zapcore.WarnLevel:
			FileLogger.Warnw(message, fileAllFields...)
		case zapcore.ErrorLevel:
			FileLogger.Errorw(message, fileAllFields...)
		case zapcore.FatalLevel:
			FileLogger.Fatalw(message, fileAllFields...)
		}
	}
}

func LogDebug(ctx context.Context, message string, keysAndValues ...interface{}) {
	logWithLevel(ctx, zapcore.DebugLevel, message, keysAndValues...)
}

func LogInfo(ctx context.Context, message string, keysAndValues ...interface{}) {
	logWithLevel(ctx, zapcore.InfoLevel, message, keysAndValues...)
}

func LogWarn(ctx context.Context, err *errors.ServerError) {
	// Flatten the error to a ...interface{} slice
	fields := []interface{}{
		"error_code", err.Code,
		"root", errors.GetRootAppError(err),
	}
	ctx = context.WithValue(ctx, "status_code", err.StatusCode)
	logWithLevel(ctx, zapcore.WarnLevel, err.Message, fields...)
}

func LogError(ctx context.Context, err *errors.ServerError) error {
	fields := []interface{}{
		"error_code", err.Code,
		"root", errors.GetRootAppError(err),
	}
	ctx = context.WithValue(ctx, "status_code", err.StatusCode)
	logWithLevel(ctx, zapcore.ErrorLevel, err.Message, fields...)
	return err
}

func LogFatal(ctx context.Context, message string, err error, keysAndValues ...interface{}) {
	logWithLevel(ctx, zapcore.FatalLevel, message, keysAndValues...)
	os.Exit(1)
}

// mergeFields merges context fields with additional key-value pairs.
// Later fields override earlier ones if keys match.
func mergeFields(contextFields []interface{}, keysAndValues ...interface{}) []interface{} {
	// Validate that keysAndValues has even number of elements (key-value pairs)
	if len(keysAndValues)%2 != 0 {
		// Log warning but don't crash - return context fields only
		if Logger != nil {
			Logger.Warnw("Invalid keysAndValues: must be even number of key-value pairs",
				"context_fields", contextFields,
				"keys_and_values", keysAndValues,
			)
		}
		return contextFields
	}

	NonLoggableFields := []string{"tags", "component", "remote_addr"}

	filteredKeysAndValues := []interface{}{}
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		fieldName, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		if !slices.Contains(NonLoggableFields, fieldName) {
			filteredKeysAndValues = append(filteredKeysAndValues, strings.ToUpper(fieldName[:1]))
			filteredKeysAndValues = append(filteredKeysAndValues, keysAndValues[i+1])
		}
	}
	return append(contextFields, filteredKeysAndValues...)
}

// extractContextFields extracts all logging fields from the context for file logging.
func extractContextFields(ctx context.Context) []interface{} {
	fields := []interface{}{}

	// Extract request_id (essential for distributed tracing)
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, "request_id", requestID)
	}

	// Extract user_id from user context (if available from authentication middleware)
	if userCtx := ctx.Value("user_id"); userCtx != nil {
		if u, ok := userCtx.(uint); ok {
			fields = append(fields, "user_id", u)
		}
	}

	if duration := ctx.Value("duration"); duration != nil {
		if d, ok := duration.(int64); ok {
			fields = append(fields, "duration", d)
		} else {
			if startTime := ctx.Value("start_time"); startTime != nil {
				if st, ok := startTime.(time.Time); ok {
					duration := time.Since(st).Milliseconds()
					fields = append(fields, "duration", duration)
				}
			}
		}
	}
	// Calculate duration from start_time (performance metric)

	// Extract component (api, sse, grpc) for service identification
	if component := ctx.Value("component"); component != nil {
		if comp, ok := component.(string); ok {
			fields = append(fields, "component", comp)
		}
	}

	// Extract status_code (HTTP response code) for error analysis
	if statusCode := ctx.Value("status_code"); statusCode != nil {
		if code, ok := statusCode.(int); ok {
			fields = append(fields, "status_code", code)
		}
	}

	return fields
}

// extractConsoleFields extracts only essential debugging fields from the context for console logging.
func extractConsoleFields(ctx context.Context) []interface{} {
	fields := []interface{}{}

	// Essential for request tracing (correlate logs across services)
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, "request_id", requestID)
	}

	// Essential for user context (identify which user's request)
	if userCtx := ctx.Value("user_id"); userCtx != nil {
		if u, ok := userCtx.(uint); ok {
			fields = append(fields, "user_id", u)
		}
	}

	// Essential for performance debugging (identify slow operations)
	if duration := ctx.Value("duration"); duration != nil {
		if d, ok := duration.(int64); ok {
			fields = append(fields, "duration", d)
		} else {
			if startTime := ctx.Value("start_time"); startTime != nil {
				if st, ok := startTime.(time.Time); ok {
					duration := time.Since(st).Milliseconds()
					fields = append(fields, "duration", duration)
				}
			}
		}
	}

	// Useful for debugging HTTP errors (included in console for quick diagnosis)
	if statusCode := ctx.Value("status_code"); statusCode != nil {
		if code, ok := statusCode.(int); ok {
			fields = append(fields, "status_code", code)
		}
	}

	// Note: component and tags are excluded from console
	// - component: redundant (always "api" for most logs)
	// - tags: for metrics/analysis, not debugging

	return fields
}

func PrintERROR(w http.ResponseWriter, code int, message string) {
	log.Printf("[ERROR] [%d] %s", code, message)
	http.Error(w, message, code)
}
