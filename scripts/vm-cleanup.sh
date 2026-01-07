#!/bin/bash
# VM Disk Space Cleanup Script
# Run this on your remote VM to free up space

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          VM Disk Space Cleanup                                             ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "Current disk usage:"
df -h /
echo ""

echo "Cleaning up Docker..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Stop all containers
echo "Stopping all containers..."
docker stop $(docker ps -aq) 2>/dev/null || echo "No containers to stop"

# Remove all containers
echo "Removing all containers..."
docker rm $(docker ps -aq) 2>/dev/null || echo "No containers to remove"

# Remove all images
echo "Removing unused images..."
docker image prune -af

# Remove all volumes
echo "Removing unused volumes..."
docker volume prune -f

# Remove all build cache
echo "Removing build cache..."
docker builder prune -af

# Clean up system
echo "Cleaning up Docker system..."
docker system prune -af --volumes

echo ""
echo "Cleaning up Go cache..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
rm -rf ~/.cache/go-build
rm -rf /root/.cache/go-build 2>/dev/null || true
rm -rf /tmp/go-build* 2>/dev/null || true

echo ""
echo "Cleaning up apt cache..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
sudo apt-get clean
sudo apt-get autoremove -y

echo ""
echo "Cleaning up logs..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
sudo journalctl --vacuum-time=3d

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ CLEANUP COMPLETE                                               ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "Disk usage after cleanup:"
df -h /
echo ""

echo "Docker disk usage:"
docker system df
echo ""

