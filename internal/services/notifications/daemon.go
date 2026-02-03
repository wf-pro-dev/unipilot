package notifications

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"unipilot/internal/errors"
	"unipilot/internal/services/utils"
	"unipilot/internal/sse"
)

// FileServiceLogger implements a file-based logger
type FileServiceLogger struct {
	logger *log.Logger
	file   *os.File
}

// NewFileServiceLogger creates a new file-based service logger
func NewFileServiceLogger(logPath string) (*FileServiceLogger, error) {
	// Ensure log directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file with append mode
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create logger with timestamp prefix
	logger := log.New(file, "", log.Ldate|log.Ltime|log.Lmicroseconds)

	return &FileServiceLogger{
		logger: logger,
		file:   file,
	}, nil
}

// Close closes the log file
func (f *FileServiceLogger) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}

// Error logs an error message
func (f *FileServiceLogger) Error(v ...interface{}) error {
	f.logger.SetPrefix("[ERROR] ")
	f.logger.Println(v...)
	return nil
}

// Warning logs a warning message
func (f *FileServiceLogger) Warning(v ...interface{}) error {
	f.logger.SetPrefix("[WARN] ")
	f.logger.Println(v...)
	return nil
}

// Info logs an informational message
func (f *FileServiceLogger) Info(v ...interface{}) error {
	f.logger.SetPrefix("[INFO] ")
	f.logger.Println(v...)
	return nil
}

// Errorf logs a formatted error message
func (f *FileServiceLogger) Errorf(format string, a ...interface{}) error {
	f.logger.SetPrefix("[ERROR] ")
	f.logger.Printf(format, a...)
	return nil
}

// Warningf logs a formatted warning message
func (f *FileServiceLogger) Warningf(format string, a ...interface{}) error {
	f.logger.SetPrefix("[WARN] ")
	f.logger.Printf(format, a...)
	return nil
}

// Infof logs a formatted informational message
func (f *FileServiceLogger) Infof(format string, a ...interface{}) error {
	f.logger.SetPrefix("[INFO] ")
	f.logger.Printf(format, a...)
	return nil
}

func RunDaemon() {
	var (
		userID  = flag.String("user", "", "User ID to monitor")
		logFile = flag.String("log", "", "Log file path")
	)
	flag.Parse()

	// Validate user ID is provided
	if *userID == "" {
		log.Fatal("User ID must be provided via -user flag")
	}

	// Get user home directory for log file default
	userHome, err := os.UserHomeDir()
	if err == nil && *logFile == "" {
		// Use user-specific log file matching manager.go pattern
		*logFile = filepath.Join(userHome, "Library", "Logs", "unipilot", fmt.Sprintf("unipilot-notification_%d.log", *userID))
	}

	// Create and set custom file logger
	fileLogger, err := NewFileServiceLogger(*logFile)
	if err != nil {
		// Fallback: log error to standard log
		log.Printf("Failed to create file logger: %v", err)
	} else {
		// Redirect standard log to file as well
		log.SetOutput(fileLogger.file)
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
		defer fileLogger.Close()
	}
	// Get user from file
	user, err := utils.GetUserFromFile()
	if err != nil {
		wrappedErr := errors.Wrap(err, errors.FSFileNotFound, "Failed to get current user")
		log.Fatalf("Failed to get current user: %v", wrappedErr)
	}

	// Validate that the user ID from file matches the flag
	if user.ID.String() != *userID {
		log.Fatalf("User ID mismatch: flag=%d, file=%d. The daemon was started for user %d but the stored user is %d",
			*userID, user.ID, *userID, user.ID)
	}

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler, err := NewScheduler()
	if err != nil {
		log.Printf("Warning: Failed to create scheduler: %v. Cannot continue without scheduler.", err)
		log.Fatalf("Failed to create scheduler: %v. Cannot continue without scheduler.", err)
	}

	eventHandler, err := NewEventHandler()
	if err != nil {
		log.Printf("Warning: Failed to create event handler: %v. Cannot continue without event handler.", err)
		log.Fatalf("Failed to create event handler: %v. Cannot continue without event handler.", err)
	}

	// Try to get authenticated HTTP client from stored credentials
	httpClient, err := sse.NewSSEClient()
	if err != nil {
		log.Fatalf("Failed to create authenticated HTTP client: %v. Cannot continue without SSE connection.", err)
	}

	// Create SSE client for the daemon with authentication
	sseClient := sse.NewSSE()

	// Start SSE connection in background with authenticated client
	go func() {
		sseClient.Connect(httpClient)
	}()

	// Start scheduler
	if err := scheduler.StartScheduler(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	// Start event handler with the SSE client
	if err := eventHandler.StartEventHandler(sseClient); err != nil {
		log.Fatalf("Failed to start event handler: %v", err)
	}
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
	}

	// Stop components in reverse order of startup
	log.Println("Stopping notification daemon...")

	// Stop event handler first (depends on SSE)
	if eventHandler != nil {
		log.Println("Stopping event handler...")
		eventHandler.StopEventHandler()
	}

	// Stop scheduler
	if scheduler != nil {
		log.Println("Stopping scheduler...")
		scheduler.StopScheduler()
	}

	// Disconnect SSE client
	if sseClient != nil {
		log.Println("Disconnecting SSE client...")
		// Note: Add sseClient.Disconnect() or sseClient.Close() method if available
		// For now, context cancellation should handle cleanup
	}

	log.Printf("Notification daemon stopped for user %d", user.ID)
}
