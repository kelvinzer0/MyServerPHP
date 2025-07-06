#!/bin/bash

# Stop the service
echo "Stopping phpservermanager service..."
sudo systemctl stop phpservermanager

# Disable the service
echo "Disabling phpservermanager service..."
sudo systemctl disable phpservermanager

# Remove the service file
echo "Removing systemd service..."
sudo rm /etc/systemd/system/phpservermanager.service

# Reload systemd
echo "Reloading systemd..."
sudo systemctl daemon-reload

# Remove the binary
echo "Removing phpservermanager from /usr/local/bin..."
sudo rm /usr/local/bin/phpservermanager

# Remove the config directory
echo "Removing config directory /etc/phpservermanager..."
sudo rm -rf /etc/phpservermanager

echo "Uninstallation complete!"
