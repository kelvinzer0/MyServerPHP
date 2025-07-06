#!/bin/bash

# Exit on error
set -e

# Build the application
echo "Building phpservermanager..."
go build -o phpservermanager cmd/server/main.go

# Move the binary to /usr/local/bin
echo "Installing phpservermanager to /usr/local/bin..."
sudo mv phpservermanager /usr/local/bin/

# Create config directory
echo "Creating config directory /etc/phpservermanager..."
sudo mkdir -p /etc/phpservermanager

# Move the service file
echo "Installing systemd service..."
sudo mv phpservermanager.service /etc/systemd/system/

# Reload systemd
echo "Reloading systemd..."
sudo systemctl daemon-reload

# Enable the service
echo "Enabling phpservermanager service..."
sudo systemctl enable phpservermanager

# Start the service
echo "Starting phpservermanager service..."
sudo systemctl start phpservermanager

echo "Installation complete!"
