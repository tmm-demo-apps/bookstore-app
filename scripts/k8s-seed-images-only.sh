#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          Seed Product Images                                                ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

# Get credentials
DB_PASSWORD=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.DB_PASSWORD}' | base64 -d)
MINIO_ACCESS_KEY=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.MINIO_ACCESS_KEY}' | base64 -d)
MINIO_SECRET_KEY=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.MINIO_SECRET_KEY}' | base64 -d)

echo "✅ Retrieved credentials"
echo ""

# Port forward
echo "🔌 Setting up port forwards..."
kubectl port-forward -n bookstore svc/postgres-service 5432:5432 &
PG_PID=$!
kubectl port-forward -n bookstore svc/minio-service 9000:9000 &
MINIO_PID=$!

sleep 3
echo "✅ Port forwards established"
echo ""

# Cleanup
cleanup() {
    echo ""
    echo "🧹 Cleaning up..."
    kill $PG_PID 2>/dev/null || true
    kill $MINIO_PID 2>/dev/null || true
    echo "✅ Cleanup complete"
}
trap cleanup EXIT

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
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ IMAGES SEEDED!                                                 ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"

