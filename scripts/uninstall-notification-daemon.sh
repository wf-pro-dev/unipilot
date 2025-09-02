#!/bin/bash

# UniPilot Notification Daemon Uninstallation Script

set -e

echo "��️  Uninstalling UniPilot Notification Daemon..."

USER_HOME=$(eval echo ~$USER)
PLIST_NAME="com.unipilot.notifications.plist"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Stop and unload the launch agent
print_status "Stopping daemon..."
if launchctl list | grep -q "com.unipilot.notifications"; then
    launchctl unload "$USER_HOME/Library/LaunchAgents/$PLIST_NAME"
    print_status "Daemon stopped"
else
    print_status "Daemon was not running"
fi

# Remove the launch agent file
print_status "Removing Launch Agent..."
if [[ -f "$USER_HOME/Library/LaunchAgents/$PLIST_NAME" ]]; then
    rm "$USER_HOME/Library/LaunchAgents/$PLIST_NAME"
    print_status "Launch Agent removed"
fi

# Remove logs (optional - comment out if you want to keep them)
print_status "Removing logs..."
rm -rf "$USER_HOME/Library/Logs/unipilot"

# Remove test script
print_status "Removing test script..."
rm -f "$USER_HOME/Library/Application Support/unipilot/test-notifications.sh"

print_status "✅ UniPilot Notification Daemon uninstalled successfully!"
