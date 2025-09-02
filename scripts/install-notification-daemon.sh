#!/bin/bash

# UniPilot Notification Daemon Installation Script
# This script installs the notification daemon as a Launch Agent

set -e  # Exit on any error

echo "🚀 Installing UniPilot Notification Daemon..."

# Get the user's home directory
USER_HOME=$(eval echo ~$USER)
APP_NAME="unipilot-notification"
PLIST_NAME="com.unipilot.notifications.plist"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    print_error "This script is designed for macOS only"
    exit 1
fi

# Get the path to the notification daemon executable
print_status "Looking for notification daemon executable..."

# Try different possible locations
POSSIBLE_PATHS=(
    "/usr/local/bin/unipilot-notification"
    "/opt/homebrew/bin/unipilot-notification"
    "$(pwd)/unipilot-notification"
    "$(pwd)/cmd/notification-daemon/unipilot-notification"
)

EXEC_PATH=""
for path in "${POSSIBLE_PATHS[@]}"; do
    if [[ -f "$path" ]]; then
        EXEC_PATH="$path"
        print_status "Found notification daemon at: $EXEC_PATH"
        break
    fi
done

if [[ -z "$EXEC_PATH" ]]; then
    print_error "Notification daemon executable not found."
    print_status "Please build the daemon first:"
    print_status "  go build -o unipilot-notification cmd/notification-daemon/main.go"
    print_status "Tried these locations:"
    for path in "${POSSIBLE_PATHS[@]}"; do
        echo "  - $path"
    done
    exit 1
fi

# Get user ID (you'll need to implement this or get it from the app)
print_status "Getting user ID..."
USER_ID=$(id -u)
print_status "Using user ID: $USER_ID"

# Create necessary directories
print_status "Creating directories..."
mkdir -p "$USER_HOME/Library/LaunchAgents"
mkdir -p "$USER_HOME/Library/Logs/unipilot"
mkdir -p "$USER_HOME/Library/Application Support/unipilot"

# Create the launch agent plist content
print_status "Creating Launch Agent configuration..."

cat > /tmp/$PLIST_NAME << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.unipilot.notifications</string>
    
    <key>ProgramArguments</key>
    <array>
        <string>$EXEC_PATH</string>
        <string>-user</string>
        <string>$USER_ID</string>
        <string>-log</string>
        <string>$USER_HOME/Library/Logs/unipilot/unipilot-notifications.log</string>
    </array>
    
    <key>RunAtLoad</key>
    <true/>
    
    <key>KeepAlive</key>
    <true/>
    
    <key>StandardOutPath</key>
    <string>$USER_HOME/Library/Logs/unipilot/unipilot-notifications.log</string>
    
    <key>StandardErrorPath</key>
    <string>$USER_HOME/Library/Logs/unipilot/unipilot-notifications-error.log</string>
    
    <key>ProcessType</key>
    <string>Background</string>
    
    <key>WorkingDirectory</key>
    <string>$USER_HOME</string>
    
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin</string>
    </dict>
</dict>
</plist>
EOF

# Install the launch agent
print_status "Installing Launch Agent..."
cp /tmp/$PLIST_NAME "$USER_HOME/Library/LaunchAgents/"

# Set proper permissions
chmod 644 "$USER_HOME/Library/LaunchAgents/$PLIST_NAME"

# Load the launch agent
print_status "Loading Launch Agent..."
launchctl load "$USER_HOME/Library/LaunchAgents/$PLIST_NAME"

# Check if it loaded successfully
if launchctl list | grep -q "com.unipilot.notifications"; then
    print_status "Launch Agent loaded successfully!"
else
    print_error "Failed to load Launch Agent"
    exit 1
fi

# Create a simple test script
print_status "Creating test script..."

cat > "$USER_HOME/Library/Application Support/unipilot/test-notifications.sh" << 'EOF'
#!/bin/bash
# Test script to manually trigger notifications

echo "Testing UniPilot notifications..."

# Get the daemon PID
DAEMON_PID=$(pgrep -f "unipilot-notification")

if [[ -n "$DAEMON_PID" ]]; then
    echo "Notification daemon is running (PID: $DAEMON_PID)"
    
    # Send a test signal (you can implement a test endpoint)
    echo "Daemon status: Running"
else
    echo "Notification daemon is not running"
fi

echo "Check logs at: ~/Library/Logs/unipilot/notification-daemon.log"
EOF

chmod +x "$USER_HOME/Library/Application Support/unipilot/test-notifications.sh"

# Installation complete
echo ""
print_status "✅ UniPilot Notification Daemon installed successfully!"
echo ""
echo "📋 Installation Summary:"
echo "  - Launch Agent: $USER_HOME/Library/LaunchAgents/$PLIST_NAME"
echo "  - Logs: $USER_HOME/Library/Logs/unipilot/"
echo "  - Test Script: $USER_HOME/Library/Application Support/unipilot/test-notifications.sh"
echo ""
echo "🔧 Management Commands:"
echo "  - Check status: launchctl list | grep unipilot"
echo "  - View logs: tail -f $USER_HOME/Library/Logs/unipilot/notification-daemon.log"
echo "  - Stop daemon: launchctl unload $USER_HOME/Library/LaunchAgents/$PLIST_NAME"
echo "  - Start daemon: launchctl load $USER_HOME/Library/LaunchAgents/$PLIST_NAME"
echo "  - Test notifications: $USER_HOME/Library/Application Support/unipilot/test-notifications.sh"
echo ""
print_warning "The daemon will start automatically on login and restart if it crashes."
