package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
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
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	bucketName = "product-images"
)

func main() {
	// Parse command line flags
	forceRegenerate := flag.Bool("force", false, "Force regeneration of all images, even if they already exist")
	flag.Parse()

	// Get environment variables
	minioEndpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbHost := getEnv("DB_HOST", "localhost")
	dbName := getEnv("DB_NAME", "bookstore")

	if *forceRegenerate {
		log.Println("Force regeneration enabled - will regenerate all images")
	}

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
	rows, err := db.Query("SELECT id, name, sku, author FROM products ORDER BY id")
	if err != nil {
		log.Fatalf("Failed to query products: %v", err)
	}
	defer rows.Close()

	ctx := context.Background()
	count := 0
	skipped := 0

	for rows.Next() {
		var id int
		var name string
		var sku sql.NullString
		var author sql.NullString
		if err := rows.Scan(&id, &name, &sku, &author); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Check if image already exists in MinIO (unless force flag is set)
		imageName := fmt.Sprintf("product-%d.png", id)
		imageNameJpg := fmt.Sprintf("product-%d.jpg", id)

		if !*forceRegenerate {
			// Check if PNG exists
			_, err := client.StatObject(ctx, bucketName, imageName, minio.StatObjectOptions{})
			pngExists := err == nil

			// Check if JPG exists
			_, err = client.StatObject(ctx, bucketName, imageNameJpg, minio.StatObjectOptions{})
			jpgExists := err == nil

			if pngExists || jpgExists {
				skipped++
				log.Printf("Image already exists for product %d: %s, skipping", id, name)
				continue
			}
		}

		var imageData []byte
		contentType := "image/png"

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
			authorStr := ""
			if author.Valid {
				authorStr = author.String
			}
			imageData = generateProductImage(id, name, authorStr)
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

	log.Printf("Successfully seeded %d product images (%d skipped, already exist)", count, skipped)
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
func generateProductImage(id int, name string, author string) []byte {
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

	// Add text (title and author)
	addText(img, name, author, width, height)

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

// addText renders title and author text on the image
func addText(img *image.RGBA, title string, author string, width int, height int) {
	// Parse the TrueType font
	ttfFont, err := opentype.Parse(gobold.TTF)
	if err != nil {
		log.Printf("Error parsing font: %v", err)
		return
	}

	// Create font faces for title (larger) and author (smaller)
	titleFace, err := opentype.NewFace(ttfFont, &opentype.FaceOptions{
		Size:    32,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Printf("Error creating title font face: %v", err)
		return
	}
	defer titleFace.Close()

	authorFace, err := opentype.NewFace(ttfFont, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Printf("Error creating author font face: %v", err)
		return
	}
	defer authorFace.Close()

	textColor := color.RGBA{255, 255, 255, 255}

	// Wrap title text to fit within the image width
	maxWidth := width - 80 // Leave margin
	titleLines := wrapText(title, titleFace, maxWidth)

	// Calculate starting Y position for title (centered vertically)
	lineHeight := 45 // Approximate line height for size 32 font
	titleHeight := len(titleLines) * lineHeight
	startY := (height / 2) - titleHeight/2

	// Draw title
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: titleFace,
	}

	for i, line := range titleLines {
		// Measure line width for centering
		lineWidth := font.MeasureString(titleFace, line).Ceil()
		x := (width - lineWidth) / 2
		y := startY + (i * lineHeight)

		d.Dot = fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y),
		}
		d.DrawString(line)
	}

	// Draw author (centered, below title)
	if author != "" {
		authorText := "by " + author
		authorLines := wrapText(authorText, authorFace, maxWidth)

		d.Face = authorFace
		authorLineHeight := 35 // Approximate line height for size 24 font
		authorY := startY + titleHeight + 30

		for i, line := range authorLines {
			lineWidth := font.MeasureString(authorFace, line).Ceil()
			x := (width - lineWidth) / 2
			y := authorY + (i * authorLineHeight)

			d.Dot = fixed.Point26_6{
				X: fixed.I(x),
				Y: fixed.I(y),
			}
			d.DrawString(line)
		}
	}
}

// wrapText breaks text into lines that fit within maxWidth
func wrapText(text string, face font.Face, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		lineWidth := font.MeasureString(face, testLine).Ceil()
		if lineWidth > maxWidth && currentLine != "" {
			// Current line is full, start a new line
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine = testLine
		}
	}

	// Add the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
