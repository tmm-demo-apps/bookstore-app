package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	bucketName = "product-images"
)

func main() {
	// Get environment variables
	minioEndpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbHost := getEnv("DB_HOST", "localhost")
	dbName := getEnv("DB_NAME", "bookstore")

	// Initialize MinIO client
	client, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	log.Println("Connected to MinIO")

	// Connect to database
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	// Get all products
	rows, err := db.Query("SELECT id, name, sku FROM products ORDER BY id")
	if err != nil {
		log.Fatalf("Failed to query products: %v", err)
	}
	defer rows.Close()

	ctx := context.Background()
	count := 0

	for rows.Next() {
		var id int
		var name string
		var sku sql.NullString
		if err := rows.Scan(&id, &name, &sku); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		var imageData []byte
		contentType := "image/png"
		imageName := fmt.Sprintf("product-%d.png", id)

		// Try to download real cover from Project Gutenberg if it's a book
		if sku.Valid && strings.HasPrefix(sku.String, "BOOK-") {
			gutenbergID := strings.TrimPrefix(sku.String, "BOOK-")
			// Pattern: https://www.gutenberg.org/cache/epub/{ID}/pg{ID}.cover.medium.jpg
			gutenbergURL := fmt.Sprintf("https://www.gutenberg.org/cache/epub/%s/pg%s.cover.medium.jpg", gutenbergID, gutenbergID)

			log.Printf("Attempting to download real cover for %s (ID: %s)...", name, gutenbergID)
			data, err := downloadImage(gutenbergURL)
			if err == nil {
				imageData = data
				contentType = "image/jpeg"
				imageName = fmt.Sprintf("product-%d.jpg", id)
				log.Printf("Successfully downloaded real cover for %s", name)
			} else {
				log.Printf("Could not download real cover for %s, falling back to generated image: %v", name, err)
			}
		}

		// Fallback to generated image if needed
		if imageData == nil {
			imageData = generateProductImage(id, name)
		}

		// Upload to MinIO
		opts := minio.PutObjectOptions{
			ContentType:  contentType,
			CacheControl: "public, max-age=31536000, immutable",
		}

		_, err = client.PutObject(ctx, bucketName, imageName, bytes.NewReader(imageData), int64(len(imageData)), opts)
		if err != nil {
			log.Printf("Error uploading image for product %d: %v", id, err)
			continue
		}

		// Update product image URL in database
		imageURL := fmt.Sprintf("/images/%s", imageName)
		_, err = db.Exec("UPDATE products SET image_url = $1 WHERE id = $2", imageURL, id)
		if err != nil {
			log.Printf("Error updating product %d image URL: %v", id, err)
			continue
		}

		count++
		log.Printf("Uploaded and updated image for product %d: %s", id, name)
	}

	log.Printf("Successfully seeded %d product images", count)
}

func downloadImage(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// generateProductImage creates a simple colored image with product info
func generateProductImage(id int, name string) []byte {
	// Create 400x600 image (book cover aspect ratio)
	width, height := 400, 600
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Generate color based on product ID
	hue := float64(id*37) / 360.0 // Use prime number for better distribution
	r, g, b := hslToRGB(hue, 0.7, 0.5)
	bgColor := color.RGBA{r, g, b, 255}

	// Fill background
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Add border
	borderColor := color.RGBA{255, 255, 255, 255}
	borderWidth := 20
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x < borderWidth || x >= width-borderWidth || y < borderWidth || y >= height-borderWidth {
				img.Set(x, y, borderColor)
			}
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatalf("Failed to encode image: %v", err)
	}

	return buf.Bytes()
}

// hslToRGB converts HSL color to RGB
func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	var r, g, b float64

	if s == 0 {
		r, g, b = l, l, l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hueToRGB(p, q, h+1.0/3.0)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-1.0/3.0)
	}

	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
