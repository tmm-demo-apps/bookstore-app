#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          Installing Go on Remote VM                                        ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

GO_VERSION="1.25.5"
GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
GO_URL="https://go.dev/dl/${GO_TARBALL}"

echo "📥 Downloading Go ${GO_VERSION}..."
wget -q --show-progress "$GO_URL"

echo ""
echo "🗑️  Removing old Go installation (if exists)..."
sudo rm -rf /usr/local/go

echo ""
echo "📦 Extracting Go..."
sudo tar -C /usr/local -xzf "$GO_TARBALL"

echo ""
echo "🧹 Cleaning up tarball..."
rm "$GO_TARBALL"

echo ""
echo "🔧 Setting up PATH..."
if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
fi

# Also set for current session
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$HOME/go/bin

echo ""
echo "✅ Go installed successfully!"
echo ""
go version

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ GO INSTALLATION COMPLETE                                       ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""
echo "Note: You may need to run 'source ~/.bashrc' or start a new shell session"
echo "      for the PATH changes to take effect."

