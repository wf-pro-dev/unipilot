package daemon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Manager handles daemon installation and management
type Manager struct {
	userID     uint
	daemonPath string
	plistPath  string
	projectDir string
	ctx        context.Context
}

// NewManager creates a new daemon manager
func NewManager(userID uint, ctx context.Context) (*Manager, error) {
	// Get user's home directory
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Get the project directory (where the main app is running from)
	projectDir, err := getProjectDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to get project directory: %w", err)
	}

	// Set the daemon path to professional standard location
	daemonPath := "/usr/local/bin/unipilot-notification"
	plistPath := filepath.Join(userHome, "Library", "LaunchAgents", "com.unipilot.notifications.plist")

	return &Manager{
		userID:     userID,
		daemonPath: daemonPath,
		plistPath:  plistPath,
		projectDir: projectDir,
		ctx:        ctx,
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

	return "", fmt.Errorf("could not find project directory (no go.mod found)")
}

// BuildDaemon builds the notification daemon binary
func (m *Manager) BuildDaemon() error {
	log.Printf("[Daemon] Building notification daemon...")

	// Check if daemon source exists
	daemonSource := filepath.Join(m.projectDir, "internal", "daemon", "notifications-daemon.go")
	if _, err := os.Stat(daemonSource); os.IsNotExist(err) {
		return fmt.Errorf("daemon source file not found: %s", daemonSource)
	}

	// Build the daemon to a temporary location in /tmp
	tempDaemonPath := filepath.Join(os.TempDir(), "unipilot-notification-temp")

	cmd := exec.Command("go", "build", "-o", tempDaemonPath, daemonSource)
	cmd.Dir = m.projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("[Daemon] Running: go build -o %s %s", tempDaemonPath, daemonSource)

	if err := cmd.Run(); err != nil {
		// Clean up temp file on error
		os.Remove(tempDaemonPath)
		return fmt.Errorf("failed to build daemon: %w", err)
	}

	// Move the binary to the professional location with authorization
	if err := m.moveBinaryToSystemLocation(tempDaemonPath); err != nil {
		// Clean up temp file
		os.Remove(tempDaemonPath)
		return fmt.Errorf("failed to install daemon to system location: %w", err)
	}

	log.Printf("[Daemon] Notification daemon built and installed successfully: %s", m.daemonPath)
	return nil
}

// moveBinaryToSystemLocation moves the binary to /usr/local/bin/ with proper authorization
func (m *Manager) moveBinaryToSystemLocation(tempPath string) error {
	// Create authorization manager
	authMgr := NewAuthorizationManager(m.ctx)

	// Execute all commands in a single privileged session
	commands := []string{
		fmt.Sprintf("/bin/mv %s %s", tempPath, m.daemonPath),
		fmt.Sprintf("/bin/chmod 755 %s", m.daemonPath),
		fmt.Sprintf("/usr/sbin/chown root:wheel %s", m.daemonPath),
	}

	if err := authMgr.RequestPrivilegesAndExecute(commands); err != nil {
		return fmt.Errorf("failed to execute privileged commands: %w", err)
	}

	log.Printf("[Daemon] Binary installed successfully: %s", m.daemonPath)
	return nil
}

// InstallDaemon installs the notification daemon
func (m *Manager) InstallDaemon() error {
	log.Printf("[Daemon] Installing notification daemon for user %d", m.userID)

	// Check if daemon binary exists, build if not
	if _, err := os.Stat(m.daemonPath); os.IsNotExist(err) {
		log.Printf("[Daemon] Daemon binary not found, building...")
		if err := m.BuildDaemon(); err != nil {
			return fmt.Errorf("failed to build daemon: %w", err)
		}
	}

	// Create necessary directories
	if err := m.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Create the plist file
	if err := m.createPlistFile(); err != nil {
		return fmt.Errorf("failed to create plist file: %w", err)
	}

	// Load the launch agent
	if err := m.loadLaunchAgent(); err != nil {
		return fmt.Errorf("failed to load launch agent: %w", err)
	}

	log.Printf("[Daemon] Notification daemon installed successfully")
	return nil
}

// UninstallDaemon removes the notification daemon
func (m *Manager) UninstallDaemon() error {
	log.Printf("[Daemon] Uninstalling notification daemon")

	// Unload the launch agent
	if err := m.unloadLaunchAgent(); err != nil {
		return fmt.Errorf("failed to unload launch agent: %w", err)
	}

	// Remove the plist file
	if err := m.removePlistFile(); err != nil {
		return fmt.Errorf("failed to remove plist file: %w", err)
	}

	// Remove the daemon binary from system location
	if err := m.removeDaemonBinary(); err != nil {
		log.Printf("[Daemon] Warning: failed to remove daemon binary: %v", err)
	}

	log.Printf("[Daemon] Notification daemon uninstalled successfully")
	return nil
}

// IsDaemonInstalled checks if the daemon is installed
func (m *Manager) IsDaemonInstalled() bool {
	// Check if plist file exists
	if _, err := os.Stat(m.plistPath); os.IsNotExist(err) {
		return false
	}

	// Check if daemon binary exists in system location
	if _, err := os.Stat(m.daemonPath); os.IsNotExist(err) {
		return false
	}

	// Check if launch agent is loaded
	cmd := exec.Command("launchctl", "list")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "com.unipilot.notifications")
}

// IsDaemonRunning checks if the daemon is currently running
func (m *Manager) IsDaemonRunning() bool {
	cmd := exec.Command("pgrep", "-f", "unipilot-notification")
	err := cmd.Run()
	return err == nil
}

// StartDaemon starts the daemon if it's not running
func (m *Manager) StartDaemon() error {
	if m.IsDaemonRunning() {
		log.Printf("[Daemon] Daemon is already running")
		return nil
	}

	log.Printf("[Daemon] Starting notification daemon")
	return m.loadLaunchAgent()
}

// StopDaemon stops the daemon
func (m *Manager) StopDaemon() error {
	if !m.IsDaemonRunning() {
		log.Printf("[Daemon] Daemon is not running")
		return nil
	}

	log.Printf("[Daemon] Stopping notification daemon")
	return m.unloadLaunchAgent()
}

// RebuildDaemon rebuilds the daemon binary
func (m *Manager) RebuildDaemon() error {
	log.Printf("[Daemon] Rebuilding notification daemon...")

	// Remove existing binary from system location
	if err := m.removeDaemonBinary(); err != nil {
		log.Printf("[Daemon] Warning: failed to remove existing binary: %v", err)
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
		filepath.Join(userHome, "Library", "LaunchAgents"),
		filepath.Join(userHome, "Library", "Logs", "unipilot"),
		filepath.Join(userHome, "Library", "Application Support", "unipilot"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func (m *Manager) createPlistFile() error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.unipilot.notifications</string>
    
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>-user</string>
        <string>%d</string>
        <string>-log</string>
        <string>%s/Library/Logs/unipilot/unipilot-notification.log</string>
    </array>
    
    <key>RunAtLoad</key>
    <true/>
    
    <key>KeepAlive</key>
    <true/>
    
    <key>StandardOutPath</key>
    <string>%s/Library/Logs/unipilot/unipilot-notification.log</string>
    
    <key>StandardErrorPath</key>
    <string>%s/Library/Logs/unipilot/unipilot-notification-error.log</string>
    
    <key>ProcessType</key>
    <string>Background</string>
    
    <key>WorkingDirectory</key>
    <string>%s</string>
    
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin</string>
        <key>BASE_URL</key>
        <string>https://wwwill.xyz</string>
    </dict>
</dict>
</plist>`, m.daemonPath, m.userID, userHome, userHome, userHome, userHome)

	return os.WriteFile(m.plistPath, []byte(plistContent), 0644)
}

func (m *Manager) loadLaunchAgent() error {
	cmd := exec.Command("launchctl", "load", m.plistPath)
	return cmd.Run()
}

func (m *Manager) unloadLaunchAgent() error {
	cmd := exec.Command("launchctl", "unload", m.plistPath)
	return cmd.Run()
}

func (m *Manager) removePlistFile() error {
	return os.Remove(m.plistPath)
}

func (m *Manager) removeDaemonBinary() error {
	// Create authorization manager
	authMgr := NewAuthorizationManager(m.ctx)

	// Remove the binary in a single privileged session
	commands := []string{
		fmt.Sprintf("/bin/rm -f %s", m.daemonPath),
	}

	return authMgr.RequestPrivilegesAndExecute(commands)
}
