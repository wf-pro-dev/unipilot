package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"unipilot/internal/models/user"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is the console logger instance. In production, it falls back to FileLogger.
// It outputs structured logs to stdout/stderr with development-friendly formatting.
var Logger *zap.SugaredLogger

// FileLogger is the file-based logger instance that writes structured JSON logs
// to log files (app.log and error.log) with automatic rotation and compression.
var FileLogger *zap.SugaredLogger

// InitLogger initializes the logging system with file and console outputs.
// Configuration is read from environment variables:
//   - LOG_DIR: Directory for log files (default: "/app/logs")
//   - LOG_LEVEL: Log level - DEBUG, WARN, ERROR (default: INFO)
//   - ENV: Environment mode - if not "production", console logging is enabled
//
// Creates two log files:
//   - app.log: All logs (max 100MB, 10 backups, 30 days retention)
//   - error.log: Error-level logs only (max 50MB, 5 backups, 60 days retention)
//
// Panics if the logs directory cannot be created.
func InitLogger() {
	// Create logs directory
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "/app/logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic("Failed to create logs directory: " + err.Error())
	}

	// Determine log level
	logLevel := zapcore.InfoLevel
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		logLevel = zapcore.DebugLevel
	case "WARN":
		logLevel = zapcore.WarnLevel
	case "ERROR":
		logLevel = zapcore.ErrorLevel
	}

	// Encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	var fileCores []zapcore.Core
	var consoleCores []zapcore.Core

	// File output (always enabled)
	fileWriter := &lumberjack.Logger{
		Filename:   logDir + "/app.log",
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
	}
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(fileWriter),
		zap.InfoLevel,
	)
	fileCores = append(fileCores, fileCore)

	// Error file (separate file for errors)
	errorWriter := &lumberjack.Logger{
		Filename:   logDir + "/error.log",
		MaxSize:    50,
		MaxBackups: 5,
		MaxAge:     60,
		Compress:   true,
	}
	errorCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(errorWriter),
		zapcore.ErrorLevel,
	)
	fileCores = append(fileCores, errorCore)

	// Console output (only in development)
	if os.Getenv("ENV") != "production" {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.AddSync(os.Stdout),
			logLevel,
		)
		consoleCores = append(consoleCores, consoleCore)
	}

	// Create logger with all cores
	fileConfig := zapcore.NewTee(fileCores...)
	fileBaseLogger := zap.New(fileConfig, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	FileLogger = fileBaseLogger.Sugar()

	if len(consoleCores) > 0 {
		consoleConfig := zapcore.NewTee(consoleCores...)
		consoleBaseLogger := zap.New(consoleConfig, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		Logger = consoleBaseLogger.Sugar()
	} else {
		// Fallback to file logger if console logger is disabled
		Logger = FileLogger
	}

}

// LogDebug logs a debug-level message with context fields and additional key-value pairs.
// Context fields (request_id, user_id, duration) are automatically extracted and included.
// The formatted message is written to the console logger only.
//
// Parameters:
//   - ctx: Context containing request metadata (request_id, user, start_time)
//   - message: The log message
//   - keysAndValues: Alternating key-value pairs (e.g., "key1", value1, "key2", value2)
func LogDebug(ctx context.Context, message string, keysAndValues ...interface{}) {
	fields := extractContextFields(ctx)
	// Prepend context fields so they can be overridden by provided fields
	allFields := append(fields, keysAndValues...)
	logMessage, err := FormatLogMessage(message, allFields...)
	if err != nil {
		LogError(ctx, err.Error(), err, allFields...)
		return
	}
	if Logger != nil {
		Logger.Debugf(logMessage)
	}
}

// LogInfo logs an info-level message with context fields and additional key-value pairs.
// Context fields (request_id, user_id, duration) are automatically extracted and included.
// The message is written to both console and file loggers.
//
// Parameters:
//   - ctx: Context containing request metadata (request_id, user, start_time)
//   - message: The log message
//   - keysAndValues: Alternating key-value pairs (e.g., "key1", value1, "key2", value2)
func LogInfo(ctx context.Context, message string, keysAndValues ...interface{}) {
	fields := extractContextFields(ctx)
	// Prepend context fields so they can be overridden by provided fields
	allFields := append(fields, keysAndValues...)
	logMessage, err := FormatLogMessage(message, allFields...)

	if err != nil {
		LogError(ctx, err.Error(), err, allFields...)
		return
	}
	if Logger != nil {
		Logger.Infof(logMessage)
	}
	if FileLogger != nil {
		FileLogger.Infow(message, allFields...)
	}
}

// LogWarn logs a warning-level message with an error and context fields.
// Context fields (request_id, user_id, duration) are automatically extracted and included.
// The error is automatically added to the log fields.
// The message is written to both console and file loggers.
//
// Parameters:
//   - ctx: Context containing request metadata (request_id, user, start_time)
//   - message: The log message
//   - err: The error to log (must not be nil)
//   - keysAndValues: Alternating key-value pairs (e.g., "key1", value1, "key2", value2)
func LogWarn(ctx context.Context, message string, err error, keysAndValues ...interface{}) {
	fields := extractContextFields(ctx)
	// Prepend context fields so they can be overridden by provided fields
	allFields := append(fields, keysAndValues...)
	allFields = append(allFields, "error", err.Error())

	logMessage, formatErr := FormatLogMessage(message, allFields...)
	if formatErr != nil {
		LogError(ctx, formatErr.Error(), formatErr, allFields...)
		return
	}
	if Logger != nil {
		Logger.Warnf(logMessage)
	}
	if FileLogger != nil {
		FileLogger.Warnw(message, allFields...)
	}
}

// LogError logs an error-level message with an error and context fields.
// Context fields (request_id, user_id, duration) are automatically extracted and included.
// The error is automatically added to the log fields.
// The message is written to both console and file loggers (including error.log).
//
// Parameters:
//   - ctx: Context containing request metadata (request_id, user, start_time)
//   - message: The log message
//   - err: The error to log (must not be nil)
//   - keysAndValues: Alternating key-value pairs (e.g., "key1", value1, "key2", value2)
func LogError(ctx context.Context, message string, err error, keysAndValues ...interface{}) {
	fields := extractContextFields(ctx)
	// Prepend context fields so they can be overridden by provided fields
	allFields := append(fields, keysAndValues...)
	allFields = append(allFields, "error", err.Error())

	logMessage, formatErr := FormatLogMessage(message, allFields...)
	if formatErr != nil {
		// Avoid infinite recursion by using a simple log if FormatLogMessage fails
		if Logger != nil {
			Logger.Errorf("Failed to format log message: %v", formatErr)
		}
		if FileLogger != nil {
			FileLogger.Errorw("Failed to format log message", "error", formatErr.Error())
		}
		return
	}
	if Logger != nil {
		Logger.Errorf(logMessage)
	}
	if FileLogger != nil {
		FileLogger.Errorw(message, allFields...)
	}
}

// ResponseError logs an error and sends an HTTP error response.
// Errors with status code >= 500 are logged as errors, others as warnings.
// The status code is automatically added to the log fields.
//
// Parameters:
//   - ctx: Context containing request metadata (request_id, user, start_time)
//   - w: HTTP response writer
//   - err: The error to log (must not be nil)
//   - code: HTTP status code
//   - message: Error message to send to client and log
//   - keysAndValues: Alternating key-value pairs (e.g., "key1", value1, "key2", value2)
func ResponseError(ctx context.Context, w http.ResponseWriter, err error, code int, message string, keysAndValues ...interface{}) {
	fields := append([]interface{}{"status_code", code}, keysAndValues...)
	if code >= 500 {
		LogError(ctx, message, err, fields...)
	} else {
		LogWarn(ctx, message, err, fields...)
	}
	http.Error(w, message, code)
}

// FormatLogMessage formats a log message by appending key-value pairs as a string.
// The output format is: "message key1=value1, key2=value2, ..."
//
// Parameters:
//   - message: The base log message
//   - keysAndValues: Alternating key-value pairs (must be even number of arguments)
//
// Returns:
//   - string: Formatted log message
//   - error: Error if keysAndValues is not an even number of arguments
func FormatLogMessage(message string, keysAndValues ...interface{}) (string, error) {
	if len(keysAndValues)%2 != 0 {
		return "", errors.New("keysAndValues must be a pair of key and value")
	}

	i := 0
	var logMessage = message + " "
	for i < len(keysAndValues)-1 {
		logMessage += fmt.Sprintf("%s=%v, ", keysAndValues[i], keysAndValues[i+1])
		i += 2
	}

	return logMessage, nil
}

// extractContextFields extracts logging fields from the context.
// Extracts request_id, user_id (from user context), and duration (from start_time).
// Returns a slice of alternating key-value pairs suitable for structured logging.
func extractContextFields(ctx context.Context) []interface{} {
	fields := []interface{}{}

	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, "request_id", requestID)
	}

	if userCtx := ctx.Value("user"); userCtx != nil {
		if u, ok := userCtx.(user.User); ok {
			fields = append(fields, "user_id", u.ID)
		}
	}

	if startTime := ctx.Value("start_time"); startTime != nil {
		if st, ok := startTime.(time.Time); ok {
			duration := time.Since(st).Milliseconds()
			fields = append(fields, "duration", duration)
		}
	}

	return fields
}

// PrintERROR is a legacy error logging function that uses the standard log package.
// It logs to stdout and sends an HTTP error response.
// Consider using ResponseError instead for structured logging with context.
//
// Parameters:
//   - w: HTTP response writer
//   - code: HTTP status code
//   - message: Error message to send to client and log
func PrintERROR(w http.ResponseWriter, code int, message string) {
	log.Printf("[ERROR] [%d] %s", code, message)
	http.Error(w, message, code)
}
