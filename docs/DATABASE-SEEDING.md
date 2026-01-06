# 📚 Database Seeding Guide

## Overview

After deploying the application to Kubernetes, you need to:
1. **Run migrations** - Create database schema
2. **Seed books** - Add ~50 Project Gutenberg books
3. **Seed images** - Download/generate product images

## Quick Start

```bash
# 1. Run migrations (one-time)
kubectl cp migrations/ bookstore/postgres-0:/tmp/migrations/
kubectl exec -it -n bookstore postgres-0 -- sh -c 'cd /tmp/migrations && for file in *.sql; do echo "Running $file..."; psql -U bookstore_user -d bookstore -f "$file"; done'

# 2. Seed database (run from project root)
./scripts/k8s-seed-data.sh
```

## What Gets Seeded

### Books (~50 classics)
- Pride and Prejudice
- Alice's Adventures in Wonderland
- The Great Gatsby
- Moby-Dick
- Sherlock Holmes
- Frankenstein
- Dracula
- And 40+ more...

### Images
- Real covers from Project Gutenberg (when available)
- Generated placeholder images (fallback)
- Uploaded to MinIO object storage

## How It Works

The `k8s-seed-data.sh` script:

1. **Retrieves credentials** from Kubernetes secrets
2. **Port-forwards** to PostgreSQL and MinIO services
3. **Runs seeding scripts**:
   - `scripts/seed-gutenberg-books.go` - Adds books to database
   - `scripts/seed-images.go` - Downloads/generates and uploads images
4. **Cleans up** port-forwards when done

## Manual Seeding (Alternative)

If you prefer to seed manually:

### Seed Books

```bash
# Port forward to PostgreSQL
kubectl port-forward -n bookstore svc/postgres-service 5432:5432 &

# Get database password
DB_PASSWORD=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.DB_PASSWORD}' | base64 -d)

# Run book seeding
DB_HOST=localhost:5432 \
DB_USER=bookstore_user \
DB_PASSWORD="$DB_PASSWORD" \
DB_NAME=bookstore \
go run scripts/seed-gutenberg-books.go

# Kill port forward
pkill -f "port-forward.*postgres-service"
```

### Seed Images

```bash
# Port forward to PostgreSQL and MinIO
kubectl port-forward -n bookstore svc/postgres-service 5432:5432 &
kubectl port-forward -n bookstore svc/minio-service 9000:9000 &

# Get credentials
DB_PASSWORD=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.DB_PASSWORD}' | base64 -d)
MINIO_ACCESS_KEY=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.MINIO_ACCESS_KEY}' | base64 -d)
MINIO_SECRET_KEY=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.MINIO_SECRET_KEY}' | base64 -d)

# Run image seeding
DB_HOST=localhost:5432 \
DB_USER=bookstore_user \
DB_PASSWORD="$DB_PASSWORD" \
DB_NAME=bookstore \
MINIO_ENDPOINT=localhost:9000 \
MINIO_ACCESS_KEY="$MINIO_ACCESS_KEY" \
MINIO_SECRET_KEY="$MINIO_SECRET_KEY" \
go run scripts/seed-images.go

# Kill port forwards
pkill -f "port-forward"
```

## Troubleshooting

### Port Forward Already in Use

```bash
# Kill existing port forwards
pkill -f "port-forward"

# Or specify different ports
kubectl port-forward -n bookstore svc/postgres-service 5433:5432 &
# Then use DB_HOST=localhost:5433
```

### Books Already Exist

The seed script is idempotent - it will update existing books rather than duplicate them.

### Images Already Exist

By default, the image script skips existing images. To regenerate all images:

```bash
go run scripts/seed-images.go -force
```

### Connection Timeout

Ensure services are running:

```bash
kubectl get pods -n bookstore
kubectl logs -n bookstore deployment/app-deployment
```

## Future Improvements

For production, consider:

1. **Init Container** - Run migrations automatically on app startup
2. **Migration Tool** - Use golang-migrate or similar
3. **Seed Job** - Kubernetes Job to seed on first deployment
4. **Image Caching** - Pre-build images with data included

See `kubernetes/migration-job.yaml` and `kubernetes/seed-job.yaml` for examples (not yet fully implemented).

