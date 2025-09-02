#!/bin/bash
# Create launch agent for macOS

USER_HOME=$(eval echo ~$USER)
PLIST_NAME="com.unipilot.notifications.plist"

cat > /tmp/$PLIST_NAME << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.unipilot.notifications</string>
    <key>ProgramArguments</key>
    <array>
        <string>/path/to/unipilot-daemon</string>
        <string>-user</string>
        <string>1</string>
        <string>-log</string>
        <string>$USER_HOME/Library/Logs/unipilot-notifications.log</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$USER_HOME/Library/Logs/unipilot-notifications.log</string>
    <key>StandardErrorPath</key>
    <string>$USER_HOME/Library/Logs/unipilot-notifications-error.log</string>
</dict>
</plist>
EOF

cp /tmp/$PLIST_NAME "$USER_HOME/Library/LaunchAgents/"
launchctl load "$USER_HOME/Library/LaunchAgents/$PLIST_NAME"
