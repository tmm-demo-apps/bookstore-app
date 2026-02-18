#!/bin/bash
# Mirror infrastructure images to GHCR for the Demo Bookstore Suite
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

declare -A IMAGES=(
  ["postgres:14-alpine"]="postgres:14-alpine"
  ["redis:7-alpine"]="redis:7-alpine"
  ["docker.elastic.co/elasticsearch/elasticsearch:8.11.0"]="elasticsearch:8.11.0"
  ["minio/minio:latest"]="minio:latest"
  ["ollama/ollama:latest"]="ollama:latest"
)

echo "=== Mirroring infrastructure images to GHCR ==="
echo "Target: ${GHCR_ORG}"
echo ""

for SOURCE in "${!IMAGES[@]}"; do
  TARGET="${GHCR_ORG}/${IMAGES[$SOURCE]}"
  echo "--- ${SOURCE} -> ${TARGET} ---"

  echo "  Pulling ${SOURCE}..."
  docker pull "${SOURCE}"

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
for SOURCE in "${!IMAGES[@]}"; do
  NAME="${IMAGES[$SOURCE]%%:*}"
  echo "  - ${NAME}"
done
