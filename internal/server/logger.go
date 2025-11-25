package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.SugaredLogger
var FileLogger *zap.SugaredLogger

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

func LogInfo(message string, keysAndValues ...interface{}) {

	logMessage, err := FormatLogMessage(message, keysAndValues...)
	if err != nil {
		LogError(err.Error(), err, keysAndValues...)
		return
	}
	if Logger != nil {
		Logger.Infof(logMessage)
	}
	if FileLogger != nil {
		FileLogger.Infow(message, keysAndValues...)
	}
}

func LogWarn(message string, err error, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, "error", err.Error())
	logMessage, err := FormatLogMessage(message, keysAndValues...)
	if err != nil {
		LogError(err.Error(), err, keysAndValues...)
		return
	}
	if Logger != nil {
		Logger.Warnf(logMessage)
	}
	if FileLogger != nil {
		FileLogger.Warnw(message, keysAndValues...)
	}
}

func LogError(message string, err error, keysAndValues ...interface{}) {
	keysAndValues = append(keysAndValues, "error", err.Error())
	logMessage, err := FormatLogMessage(message, keysAndValues...)
	if err != nil {
		LogError(err.Error(), err, keysAndValues...)
		return
	}
	if Logger != nil {
		Logger.Errorf(logMessage)
	}
	if FileLogger != nil {
		FileLogger.Errorw(message, keysAndValues...)
	}
}

func ResponseError(w http.ResponseWriter, err error, code int, message string, keysAndValues ...interface{}) {
	fields := append([]interface{}{"status_code", code}, keysAndValues...)
	if code >= 500 {
		LogError(message, err, fields...)
	} else {
		LogWarn(message, err, fields...)
	}
	http.Error(w, message, code)
}

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

func PrintLOG(tags []string, message string) {
	var tagStr string
	for _, tag := range tags {
		tagStr += fmt.Sprintf("[%s] ", tag)
	}
	log.Printf("%s, %s", tagStr, message)
}

func PrintERROR(w http.ResponseWriter, code int, message string) {
	log.Printf("[ERROR] [%d] %s", code, message)
	http.Error(w, message, code)
}
