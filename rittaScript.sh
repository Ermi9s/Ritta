#!/usr/bin/env bash

set -e

echo "Installing Ritta dependencies..."

sudo apt-get update

sudo apt-get install -y \
    git \
    nginx \
    certbot \
    python3-certbot-nginx

echo "Ritta dependencies installed."

echo
echo "Add application-specific dependencies below."
echo

# Example:
#
# sudo apt-get install -y docker.io
# sudo systemctl enable --now docker

echo "Setup complete."
