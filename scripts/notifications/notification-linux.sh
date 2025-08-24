#!/bin/bash
# Create systemd service for Linux

cat > /tmp/unipilot-notifications.service << EOF
[Unit]
Description=UniPilot Notification Daemon
After=network.target

[Service]
Type=simple
User=$USER
ExecStart=/usr/local/bin/unipilot-daemon -user 1
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
EOF

sudo cp /tmp/unipilot-notifications.service /etc/systemd/user/
systemctl --user daemon-reload
systemctl --user enable unipilot-notifications
systemctl --user start unipilot-notifications
