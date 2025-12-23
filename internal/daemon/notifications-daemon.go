package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"unipilot/internal/errors"
	"unipilot/internal/services/notifications"
	"unipilot/internal/services/utils"
	"unipilot/internal/sse"

	"github.com/kardianos/service"
)

// FileServiceLogger implements service.Logger interface and writes to a file
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

// Program structure for service management
type program struct {
	userID       uint
	logFile      string
	scheduler    *notifications.Scheduler
	eventHandler *notifications.EventHandler
	sseClient    *sse.SSE
	cancel       context.CancelFunc
	fileLogger   *FileServiceLogger
}

func (p *program) Start(s service.Service) error {
	// Create and set custom file logger
	fileLogger, err := NewFileServiceLogger(p.logFile)
	if err != nil {
		// Fallback: log error to standard log
		log.Printf("Failed to create file logger: %v", err)
	} else {
		p.fileLogger = fileLogger
		// Redirect standard log to file as well
		log.SetOutput(fileLogger.file)
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	}

	if p.fileLogger != nil {
		p.fileLogger.Info("Starting notification daemon service...")
	} else {
		log.Println("Starting notification daemon service...")
	}

	// Start the service in a goroutine
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	if p.fileLogger != nil {
		p.fileLogger.Info("Stopping notification daemon service...")
	} else {
		log.Println("Stopping notification daemon service...")
	}

	if p.scheduler != nil {
		p.scheduler.StopScheduler()
	}
	if p.eventHandler != nil {
		p.eventHandler.StopEventHandler()
	}
	if p.cancel != nil {
		p.cancel()
	}

	// Close file logger
	if p.fileLogger != nil {
		p.fileLogger.Close()
	}

	return nil
}

func (p *program) run() {
	user, err := utils.GetUserFromFile()
	if err != nil {
		wrappedErr := errors.Wrap(err, errors.FSFileNotFound, "Failed to get current user")
		log.Fatalf("Failed to get current user: %v", wrappedErr)
	}

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	// Create and start scheduler
	scheduler, err := notifications.NewScheduler()
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}
	p.scheduler = scheduler

	if err := scheduler.InitializeForDaemon(user); err != nil {
		log.Fatalf("Failed to initialize scheduler: %v", err)
	}

	if err := scheduler.StartScheduler(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	// Create and start event handler
	eventHandler, err := notifications.NewEventHandler(user.ID)
	if err != nil {
		log.Fatalf("Failed to create event handler: %v", err)
	}
	p.eventHandler = eventHandler

	// Try to get authenticated HTTP client from stored credentials
	httpClient, err := sse.NewSSEClient()
	if err != nil {
		log.Printf("Warning: Could not create authenticated HTTP client: %v", err)
		log.Printf("Event handler will not receive real-time notifications")
	} else {
		// Create SSE client for the daemon with authentication
		sseClient := sse.NewSSE()
		p.sseClient = sseClient

		// Start SSE connection in background with authenticated client
		go func() {
			sseClient.Connect(httpClient)
		}()

		// Start event handler with the SSE client
		if err := eventHandler.StartEventHandler(sseClient); err != nil {
			log.Printf("Warning: Failed to start event handler: %v", err)
		}
	}

	log.Printf("Notification daemon started for user %d", p.userID)

	// Wait for context cancellation or interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
	}

	log.Println("Shutting down notification daemon...")

	// Stop both scheduler and event handler
	scheduler.StopScheduler()
	eventHandler.StopEventHandler()
}

func main() {
	var (
		userID  = flag.Uint("user", 0, "User ID to monitor")
		logFile = flag.String("log", "", "Log file path")
		svcFlag = flag.String("service", "", "Service command: install, uninstall, start, stop, restart")
	)
	flag.Parse()

	// Get user home directory for log file default
	userHome, err := os.UserHomeDir()
	if err == nil && *logFile == "" {
		*logFile = filepath.Join(userHome, "Library", "Logs", "unipilot", "unipilot-notification.log")
	}

	// Create the program instance
	prg := &program{
		userID:  *userID,
		logFile: *logFile,
	}

	// Configure service options
	svcConfig := &service.Config{
		Name:        "com.unipilot.notifications",
		DisplayName: "UniPilot Notification Service",
		Description: "Background notification service for UniPilot",
		Arguments: []string{
			"-user", strconv.FormatUint(uint64(*userID), 10),
			"-log", *logFile,
		},
	}

	// Create service
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Handle service commands
	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Fatalf("Valid service actions: %q\n%v", service.ControlAction, err)
		}
		return
	}

	// Check if running as service
	if service.Interactive() {
		// Running interactively, just run the program
		prg.run()
	} else {
		// Running as service, use service.Run()
		err = s.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}
