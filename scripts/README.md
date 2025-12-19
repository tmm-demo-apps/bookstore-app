# Scripts Documentation

This directory contains utility scripts for managing the DemoApp.

## seed-images.go

Generates and uploads product images to MinIO storage.

### Features

- **Smart Caching**: Only generates/uploads images that don't already exist in MinIO
- **Gutenberg Integration**: Attempts to download real book covers from Project Gutenberg for books with `BOOK-*` SKUs
- **Fallback Generation**: Creates colored placeholder images with title and author text for books without real covers
- **TrueType Font Rendering**: Uses high-quality font rendering for readable text on generated images

### Usage

**Normal mode** (skip existing images):
```bash
go run scripts/seed-images.go
```

**Force mode** (regenerate all images):
```bash
go run scripts/seed-images.go -force
```

### Environment Variables

- `MINIO_ENDPOINT` - MinIO server endpoint (default: `localhost:9000`)
- `MINIO_ACCESS_KEY` - MinIO access key (default: `minioadmin`)
- `MINIO_SECRET_KEY` - MinIO secret key (default: `minioadmin`)
- `DB_USER` - Database user (default: `user`)
- `DB_PASSWORD` - Database password (default: `password`)
- `DB_HOST` - Database host (default: `localhost`)
- `DB_NAME` - Database name (default: `bookstore`)

### Performance

- **With caching** (existing images): ~2 seconds for 112 products
- **Without caching** (regenerate all): ~30-40 seconds for 112 products

### Generated Image Specifications

- **Dimensions**: 400x600 pixels (book cover aspect ratio)
- **Background**: HSL-based color generation (unique per product ID)
- **Border**: 20px white border
- **Title Font**: Go Bold, 32pt
- **Author Font**: Go Bold, 24pt
- **Text Color**: White
- **Format**: PNG for generated images, JPEG for downloaded covers

### How It Works

1. Connects to database and MinIO
2. Queries all products with their IDs, names, SKUs, and authors
3. For each product:
   - Checks if image already exists in MinIO (skips if found, unless `-force` flag is used)
   - If SKU starts with `BOOK-`, attempts to download cover from Project Gutenberg
   - If download fails or no SKU, generates a colored placeholder with text
   - Uploads image to MinIO
   - Updates product's `image_url` in database
4. Reports statistics (uploaded, skipped)

### Example Output

```
2025/12/19 14:43:05 Connected to MinIO
2025/12/19 14:43:05 Connected to database
2025/12/19 14:43:05 Image already exists for product 1: Sample Product 1, skipping
2025/12/19 14:43:05 Image already exists for product 2: Sample Product 2, skipping
...
2025/12/19 14:43:05 Successfully seeded 0 product images (112 skipped, already exist)
```

## Other Scripts

### seed-gutenberg-books.go

Seeds the database with books from Project Gutenberg's catalog.

### categorize-books.go

Categorizes existing books in the database.

### add-more-books.go

Adds additional books to the database.

