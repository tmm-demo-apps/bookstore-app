# Scripts Documentation

**Last Updated**: January 9, 2026

This directory contains utility scripts for managing the DemoApp.

## Deployment Scripts

### deploy-complete.sh (Primary)

**One-command Kubernetes deployment** that handles everything:

```bash
# Deploy to production namespace
./scripts/deploy-complete.sh v1.1.0 bookstore

# Deploy to test namespace
./scripts/deploy-complete.sh v1.1.0 bookstore-test
```

**What it does**:
1. Logs into Harbor registry
2. Builds and pushes Docker image
3. Mirrors base images (postgres, redis, elasticsearch, minio) to Harbor
4. Installs NGINX Ingress Controller if missing
5. Creates Kubernetes namespace and secrets
6. Deploys all infrastructure (postgres, redis, elasticsearch, minio)
7. Runs init-db-job (migrations + seeding)
8. Deploys application
9. Configures ingress with dynamic hostname (`{namespace}.corp.vmbeans.com`)

### harbor-remote-setup.sh

Harbor integration script called by `deploy-complete.sh`. Can be run standalone for Harbor-only operations:

```bash
./scripts/harbor-remote-setup.sh v1.1.0
```

### k8s-diagnose.sh

Diagnose Kubernetes deployment issues:

```bash
./scripts/k8s-diagnose.sh
```

### k8s-reindex-elasticsearch.sh

Reindex Elasticsearch after data changes:

```bash
./scripts/k8s-reindex-elasticsearch.sh
```

## Data Seeding Scripts

### seed-gutenberg-books.go

The **source of truth** for all book data (150 books). Seeds the database with books from Project Gutenberg's catalog including:
- Title, author, description
- Category assignment
- Stock quantity (varied 0-100, with specific demo values)
- Popularity score (Gutenberg download counts for sorting)

**Usage**:

```bash
# Seed database directly
go run scripts/seed-gutenberg-books.go

# Generate SQL migration file
go run scripts/seed-gutenberg-books.go --generate-sql
# Creates migrations/002_seed_books.sql
```

**Stock Quantity Rules** (per project rules):
- Book #1 (first in list): `stock_quantity = 3` (Low Stock demo)
- Book #5 (fifth in list): `stock_quantity = 0` (Out of Stock demo)
- All other books: Varied between 10-100

### seed-images.go

Generates and uploads product images to MinIO storage.

**Features**:
- **Smart Caching**: Only generates/uploads images that don't already exist in MinIO
- **Gutenberg Integration**: Downloads real book covers from Project Gutenberg for books with `BOOK-*` SKUs
- **Fallback Generation**: Creates colored placeholder images with title and author text
- **TrueType Font Rendering**: High-quality font rendering for readable text

**Usage**:

```bash
# Normal mode (skip existing images)
go run scripts/seed-images.go

# Force mode (regenerate all images)
go run scripts/seed-images.go -force
```

**Environment Variables**:
| Variable | Default | Description |
|----------|---------|-------------|
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO server endpoint |
| `MINIO_ACCESS_KEY` | `minioadmin` | MinIO access key |
| `MINIO_SECRET_KEY` | `minioadmin` | MinIO secret key |
| `DB_USER` | `user` | Database user |
| `DB_PASSWORD` | `password` | Database password |
| `DB_HOST` | `localhost` | Database host |
| `DB_NAME` | `bookstore` | Database name |

**Performance**:
- With caching (existing images): ~2 seconds for 150 products
- Without caching (regenerate all): ~30-40 seconds for 150 products

**Generated Image Specifications**:
- Dimensions: 400x600 pixels (book cover aspect ratio)
- Background: HSL-based color generation (unique per product ID)
- Border: 20px white border
- Title Font: Go Bold, 32pt
- Author Font: Go Bold, 24pt
- Text Color: White
- Format: PNG for generated images, JPEG for downloaded covers

## bin/ Directory

Contains pre-built Linux binaries for Kubernetes deployment:
- `seed-gutenberg-books` - Book seeding binary
- `seed-images` - Image seeding binary

These are built automatically by the Dockerfile during image creation and used by the `init-db-job.yaml` Kubernetes Job.

## Other Scripts

| Script | Purpose |
|--------|---------|
| `install-go-remote.sh` | Install Go on remote VM |
| `build-seed-binaries.sh` | Build Linux binaries locally |

## Script Summary

| Script | When to Use |
|--------|-------------|
| `deploy-complete.sh` | Full deployment to Kubernetes |
| `harbor-remote-setup.sh` | Harbor-only operations |
| `k8s-diagnose.sh` | Troubleshooting K8s issues |
| `k8s-reindex-elasticsearch.sh` | After manual data changes |
| `seed-gutenberg-books.go` | Update book data or regenerate SQL |
| `seed-images.go` | Regenerate product images |
