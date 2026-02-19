#!/bin/bash
# Mirror infrastructure images to GHCR for the Demo Bookstore Suite
#
# Pulls linux/amd64 images regardless of host architecture (safe on Apple Silicon).
#
# Prerequisites:
#   1. docker login ghcr.io -u YOUR_GITHUB_USERNAME -p YOUR_GITHUB_PAT
#   2. Ensure GHCR packages are set to "public" in org settings after first push
#
# Usage:
#   ./scripts/mirror-images.sh
#
# This is a one-time operation. Run again only to update image versions.

set -euo pipefail

GHCR_ORG="ghcr.io/tmm-demo-apps"
PLATFORM="linux/amd64"

SOURCES=(
  "postgres:14-alpine"
  "redis:7-alpine"
  "elasticsearch:8.11.0"
  "minio/minio:latest"
  "ollama/ollama:latest"
)

TARGETS=(
  "postgres:14-alpine"
  "redis:7-alpine"
  "elasticsearch:8.11.0"
  "minio:latest"
  "ollama:latest"
)

echo "=== Mirroring infrastructure images to GHCR ==="
echo "Target: ${GHCR_ORG}"
echo "Platform: ${PLATFORM}"
echo ""

for i in "${!SOURCES[@]}"; do
  SOURCE="${SOURCES[$i]}"
  TARGET="${GHCR_ORG}/${TARGETS[$i]}"
  echo "--- ${SOURCE} -> ${TARGET} ---"

  echo "  Pulling ${SOURCE} (${PLATFORM})..."
  docker pull --platform "${PLATFORM}" "${SOURCE}"

  echo "  Tagging as ${TARGET}..."
  docker tag "${SOURCE}" "${TARGET}"

  echo "  Pushing ${TARGET}..."
  docker push "${TARGET}"

  echo "  Done."
  echo ""
done

echo "=== All images mirrored successfully ==="
echo ""
echo "IMPORTANT: Set each package to 'public' in your GitHub org settings:"
echo "  https://github.com/orgs/tmm-demo-apps/packages"
echo ""
echo "Packages to make public:"
for TARGET in "${TARGETS[@]}"; do
  echo "  - ${TARGET%%:*}"
done
