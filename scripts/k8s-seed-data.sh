#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          Kubernetes Database Seeding Script                                ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

# Get database credentials from Kubernetes secret
DB_PASSWORD=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.DB_PASSWORD}' | base64 -d)
MINIO_ACCESS_KEY=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.MINIO_ACCESS_KEY}' | base64 -d)
MINIO_SECRET_KEY=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.MINIO_SECRET_KEY}' | base64 -d)

echo "✅ Retrieved credentials from Kubernetes secrets"
echo ""

# Port forward to services
echo "🔌 Setting up port forwards..."
kubectl port-forward -n bookstore svc/postgres-service 5432:5432 &
PG_PID=$!
kubectl port-forward -n bookstore svc/minio-service 9000:9000 &
MINIO_PID=$!

# Wait for port forwards to be ready
sleep 3
echo "✅ Port forwards established"
echo ""

# Cleanup function
cleanup() {
    echo ""
    echo "🧹 Cleaning up port forwards..."
    kill $PG_PID 2>/dev/null || true
    kill $MINIO_PID 2>/dev/null || true
    echo "✅ Cleanup complete"
}
trap cleanup EXIT

# Check if binaries exist
if [ ! -f "scripts/bin/seed-gutenberg-books" ] || [ ! -f "scripts/bin/seed-images" ]; then
    echo "❌ Seed binaries not found!"
    echo ""
    echo "Please build them first by running:"
    echo "  ./scripts/build-seed-binaries.sh"
    echo ""
    echo "Then commit and push:"
    echo "  git add scripts/bin/"
    echo "  git commit -m 'feat: add seed binaries'"
    echo "  git push"
    exit 1
fi

# Seed books
echo "📚 Seeding Gutenberg books..."
DB_HOST=localhost:5432 \
DB_USER=bookstore_user \
DB_PASSWORD="$DB_PASSWORD" \
DB_NAME=bookstore \
./scripts/bin/seed-gutenberg-books

echo ""
echo "✅ Books seeded successfully!"
echo ""

# Seed images
echo "🖼️  Seeding product images..."
DB_HOST=localhost:5432 \
DB_USER=bookstore_user \
DB_PASSWORD="$DB_PASSWORD" \
DB_NAME=bookstore \
MINIO_ENDPOINT=localhost:9000 \
MINIO_ACCESS_KEY="$MINIO_ACCESS_KEY" \
MINIO_SECRET_KEY="$MINIO_SECRET_KEY" \
./scripts/bin/seed-images

echo ""
echo "✅ Images seeded successfully!"
echo ""

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ DATABASE SEEDING COMPLETE                                      ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"

