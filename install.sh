#!/bin/sh
set -e

REPO="naokiiida/t42-cli"
LATEST=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep "browser_download_url" | grep "$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m)" | cut -d '"' -f 4)

if [ -z "$LATEST" ]; then
  echo "No prebuilt binary found for your OS/arch. Please build from source."
  exit 1
fi

curl -L "$LATEST" -o t42-cli
chmod +x t42-cli
sudo mv t42-cli /usr/local/bin/
echo "t42-cli installed!" 