#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          Building Seed Binaries                                            ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

# Build for Linux (remote VM)
echo "🔨 Building seed-gutenberg-books for Linux..."
GOOS=linux GOARCH=amd64 go build -o scripts/bin/seed-gutenberg-books scripts/seed-gutenberg-books.go
echo "✅ Built: scripts/bin/seed-gutenberg-books"

echo ""
echo "🔨 Building seed-images for Linux..."
GOOS=linux GOARCH=amd64 go build -o scripts/bin/seed-images scripts/seed-images.go
echo "✅ Built: scripts/bin/seed-images"

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ BINARIES BUILT SUCCESSFULLY                                    ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""
echo "Binaries created:"
echo "  - scripts/bin/seed-gutenberg-books (Linux amd64)"
echo "  - scripts/bin/seed-images (Linux amd64)"
echo ""
echo "These binaries are ready to run on the remote VM."

