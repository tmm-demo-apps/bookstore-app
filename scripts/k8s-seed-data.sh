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

# Seed books
echo "📚 Seeding Gutenberg books..."
DB_HOST=localhost:5432 \
DB_USER=bookstore_user \
DB_PASSWORD="$DB_PASSWORD" \
DB_NAME=bookstore \
go run scripts/seed-gutenberg-books.go

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
go run scripts/seed-images.go

echo ""
echo "✅ Images seeded successfully!"
echo ""

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ DATABASE SEEDING COMPLETE                                      ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"

