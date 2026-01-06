#!/bin/bash
set -e

# Build and Push Script for Harbor Registry
# Usage: ./scripts/build-and-push.sh <harbor-url> <project-name> <version>

HARBOR_URL=${1:-"harbor.vcf.local"}
PROJECT_NAME=${2:-"bookstore"}
VERSION=${3:-"v1.0.0"}

IMAGE_NAME="${HARBOR_URL}/${PROJECT_NAME}/app"
IMAGE_TAG="${IMAGE_NAME}:${VERSION}"
IMAGE_LATEST="${IMAGE_NAME}:latest"

echo "=========================================="
echo "Building and Pushing to Harbor"
echo "=========================================="
echo "Harbor URL: ${HARBOR_URL}"
echo "Project: ${PROJECT_NAME}"
echo "Version: ${VERSION}"
echo "Full Image: ${IMAGE_TAG}"
echo "=========================================="

# Check if logged in to Harbor
echo ""
echo "Checking Harbor authentication..."
if ! docker login ${HARBOR_URL} --username test --password test 2>/dev/null; then
    echo "⚠️  Not logged in to Harbor. Please run:"
    echo "   docker login ${HARBOR_URL}"
    echo ""
    read -p "Press Enter after logging in, or Ctrl+C to exit..."
fi

# Build the image
echo ""
echo "🔨 Building Docker image..."
docker build -t ${IMAGE_TAG} .

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful!"

# Tag as latest
echo ""
echo "🏷️  Tagging as latest..."
docker tag ${IMAGE_TAG} ${IMAGE_LATEST}

# Push versioned tag
echo ""
echo "📤 Pushing ${IMAGE_TAG}..."
docker push ${IMAGE_TAG}

if [ $? -ne 0 ]; then
    echo "❌ Push failed!"
    exit 1
fi

echo "✅ Pushed ${IMAGE_TAG}"

# Push latest tag
echo ""
echo "📤 Pushing ${IMAGE_LATEST}..."
docker push ${IMAGE_LATEST}

if [ $? -ne 0 ]; then
    echo "❌ Push failed!"
    exit 1
fi

echo "✅ Pushed ${IMAGE_LATEST}"

# Summary
echo ""
echo "=========================================="
echo "✅ SUCCESS!"
echo "=========================================="
echo "Images pushed to Harbor:"
echo "  - ${IMAGE_TAG}"
echo "  - ${IMAGE_LATEST}"
echo ""
echo "Next steps:"
echo "  1. Update kubernetes/app.yaml with image: ${IMAGE_TAG}"
echo "  2. Create Kubernetes secrets"
echo "  3. Deploy to cluster"
echo "=========================================="

