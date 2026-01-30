package daemon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"unipilot/internal/errors"
	"unipilot/internal/services/utils"

	"github.com/kardianos/service"
)

// Manager handles daemon installation and management
type Manager struct {
	userID       uint
	daemonPath   string
	logPath      string
	errorLogPath string
	projectDir   string
	ctx          context.Context
	svc          service.Service
}

// dummyProgram is a placeholder that implements service.Interface
// This is required by service.New() but will never actually be called
// because we set Executable in the config, which causes the service
// to run the daemon binary directly. The actual program implementation
// is in internal/services/notifications/daemon.go
type dummyProgram struct{}

func (p *dummyProgram) Start(s service.Service) error { return nil }
func (p *dummyProgram) Stop(s service.Service) error  { return nil }

// NewManager creates a new daemon manager
func NewManager(userID uint, ctx context.Context) (*Manager, error) {
	// Get the project directory (where the main app is running from)
	projectDir, err := getProjectDirectory()
	if err != nil {
		return nil, errors.Wrap(err, errors.DaemonProjectNotFound, "Failed to get project directory")
	}

	// Get user home directory
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, errors.FSDirFailed, "Failed to get user home directory")
	}

	// Set paths - use user-accessible locations (no sudo needed)
	userDir, err := utils.GetUserDir()
	if err != nil {
		return nil, errors.Wrap(err, errors.FSDirFailed, "Failed to get user directory")
	}

	logsDir, err := utils.GetLogsDir()
	if err != nil {
		return nil, errors.Wrap(err, errors.FSDirFailed, "Failed to get logs directory")
	}
	daemonPath := filepath.Join(userDir, fmt.Sprintf("unipilot-notification_%d", userID))
	logPath := filepath.Join(logsDir, fmt.Sprintf("unipilot-notification_%d.log", userID))
	errorLogPath := filepath.Join(logsDir, fmt.Sprintf("unipilot-notification_%d-error.log", userID))

	// Create a dummy program for service configuration
	prg := &dummyProgram{}

	// Configure service with logging
	svcConfig := &service.Config{
		Name:        fmt.Sprintf("com.unipilot.notifications.%d", userID),
		DisplayName: fmt.Sprintf("UniPilot Notification Service for User %d", userID),
		Description: fmt.Sprintf("Background notification service for UniPilot for user %d", userID),
		Arguments: []string{
			"-user", strconv.FormatUint(uint64(userID), 10),
			"-log", logPath,
		},
		Executable: daemonPath,
	}

	// For macOS, configure LaunchAgent options
	// Set UserService to true to ensure it's user-specific
	opts := make(service.KeyValue)
	opts["UserService"] = true
	opts["StandardOutPath"] = logPath
	opts["StandardErrorPath"] = errorLogPath
	opts["RunAtLoad"] = true
	opts["KeepAlive"] = true
	opts["ProcessType"] = "Background"
	opts["WorkingDirectory"] = userHome

	// Set environment variables
	envVars := map[string]string{
		"PATH":     "/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin",
		"BASE_URL": "https://wwwill.xyz",
	}
	opts["EnvironmentVariables"] = envVars

	svcConfig.Option = opts

	// Create service instance
	svc, err := service.New(prg, svcConfig)
	if err != nil {
		return nil, errors.Wrap(err, errors.DaemonInstallFailed, "Failed to create service instance")
	}

	return &Manager{
		userID:       userID,
		daemonPath:   daemonPath,
		logPath:      logPath,
		errorLogPath: errorLogPath,
		projectDir:   projectDir,
		ctx:          ctx,
		svc:          svc,
	}, nil
}

// getProjectDirectory finds the project root directory
func getProjectDirectory() (string, error) {
	// Try to find the project directory by looking for go.mod
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir, nil
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// Reached root directory
			break
		}
		currentDir = parent
	}

	// Fallback: try to get the executable directory
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	execDir := filepath.Dir(execPath)

	// Look for go.mod in the executable directory or its parent
	for _, dir := range []string{execDir, filepath.Dir(execDir)} {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
	}

	return "", errors.NewAppError(errors.DaemonProjectNotFound, "Could not find project directory (no go.mod found)", nil)
}

// BuildDaemon builds the notification daemon binary
func (m *Manager) BuildDaemon() error {
	// Check if daemon source exists
	daemonSource := filepath.Join(m.projectDir, "scripts", "notifications", "main.go")
	if _, err := os.Stat(daemonSource); os.IsNotExist(err) {
		return errors.NewAppError(errors.DaemonSourceNotFound, "Daemon source file not found", err)
	}

	// Ensure target directory exists
	daemonDir := filepath.Dir(m.daemonPath)
	if err := os.MkdirAll(daemonDir, 0755); err != nil {
		return errors.Wrap(err, errors.FSDirCreateFailed, "Failed to create daemon directory")
	}

	// Build the daemon
	cmd := exec.Command("go", "build", "-o", m.daemonPath, daemonSource)
	cmd.Dir = m.projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.DaemonBuildFailed, "Failed to build daemon")
	}

	// Make executable
	if err := os.Chmod(m.daemonPath, 0755); err != nil {
		return errors.Wrap(err, errors.DaemonBuildFailed, "Failed to set executable permissions")
	}

	return nil
}

// InstallDaemon installs the notification daemon using the service library
func (m *Manager) InstallDaemon() error {
	// Check if daemon binary exists, build if not
	if _, err := os.Stat(m.daemonPath); os.IsNotExist(err) {
		if err := m.BuildDaemon(); err != nil {
			return errors.Wrap(err, errors.DaemonInstallFailed, "Failed to build daemon")
		}
	}

	// Create necessary directories
	if err := m.createDirectories(); err != nil {
		return errors.Wrap(err, errors.DaemonInstallFailed, "Failed to create directories")
	}

	// Use service library to install
	// On macOS, this will create the LaunchAgent plist automatically
	if err := m.svc.Install(); err != nil {
		return errors.Wrap(err, errors.DaemonInstallFailed, "Failed to install service")
	}

	return nil
}

// UninstallDaemon removes the notification daemon
func (m *Manager) UninstallDaemon() error {
	// Stop the service first

	if m.IsDaemonRunning() {
		if err := m.svc.Stop(); err != nil {
			log.Printf("Warning: Failed to stop notification daemon: %v", err)
		}
	}

	// Uninstall using service library
	if err := m.svc.Uninstall(); err != nil {
		return errors.Wrap(err, errors.DaemonUninstallFailed, "Failed to uninstall service")
	}

	// Clean up
	if err := m.CleanUp(); err != nil {
		return errors.Wrap(err, errors.DaemonUninstallFailed, "Failed to clean up")
	}

	return nil
}

// IsDaemonInstalled checks if the daemon is installed
func (m *Manager) IsDaemonInstalled() bool {
	// Check if daemon binary exists
	if _, err := os.Stat(m.daemonPath); os.IsNotExist(err) {
		return false
	}

	// Check if service is installed using service library
	status, err := m.svc.Status()
	if err != nil {
		return false
	}

	// Service is installed if status is not "Unknown"
	return status != service.StatusUnknown
}

// IsDaemonRunning checks if the daemon is currently running
func (m *Manager) IsDaemonRunning() bool {
	status, err := m.svc.Status()
	if err != nil {
		return false
	}
	return status == service.StatusRunning
}

// StartDaemon starts the daemon if it's not running
func (m *Manager) StartDaemon() error {
	if m.IsDaemonRunning() {
		return nil
	}

	return m.svc.Start()
}

// StopDaemon stops the daemon
func (m *Manager) StopDaemon() error {

	if !m.IsDaemonRunning() {
		return nil
	}

	if err := m.svc.Stop(); err != nil {
		return err
	}

	return nil
}

// RebuildDaemon rebuilds the daemon binary
func (m *Manager) RebuildDaemon() error {
	// Stop the service first
	if m.IsDaemonRunning() {
		if err := m.svc.Stop(); err != nil {
			// Log warning but continue
		}
	}

	// Remove existing binary
	if err := os.Remove(m.daemonPath); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, errors.DaemonBuildFailed, "Failed to remove existing binary")
	}

	// Build new binary
	return m.BuildDaemon()
}

// Helper methods
func (m *Manager) createDirectories() error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dirs := []string{
		filepath.Join(userHome, "Library", "Application Support", "unipilot"),
		filepath.Join(userHome, "Library", "Logs", "unipilot"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrap(err, errors.FSDirCreateFailed, "Failed to create directory")
		}
	}

	return nil
}

func (m *Manager) CleanUp() error {

	// remove the daemon binary
	if err := os.Remove(m.daemonPath); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, errors.DaemonBuildFailed, "Failed to remove existing binary")
	}

	// remove log files
	if err := os.Remove(m.logPath); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, errors.DaemonBuildFailed, "Failed to remove log file")
	}
	if err := os.Remove(m.errorLogPath); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, errors.DaemonBuildFailed, "Failed to remove error log file")
	}

	return nil
}
