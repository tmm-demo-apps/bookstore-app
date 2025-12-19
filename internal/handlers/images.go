package handlers

import (
	"DemoApp/internal/storage"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// ImageHandlers handles image-related HTTP requests
type ImageHandlers struct {
	Storage *storage.MinIOStorage
}

// ServeImage serves an image from MinIO with proper cache headers
func (h *ImageHandlers) ServeImage(w http.ResponseWriter, r *http.Request) {
	// Extract image name from URL path
	// URL format: /images/{imageName}
	imageName := strings.TrimPrefix(r.URL.Path, "/images/")
	if imageName == "" {
		http.Error(w, "Image name required", http.StatusBadRequest)
		return
	}

	// Get object from MinIO
	ctx := r.Context()
	obj, err := h.Storage.GetObject(ctx, imageName)
	if err != nil {
		log.Printf("Error getting image %s: %v", imageName, err)
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}
	defer obj.Close()

	// Get object info for metadata
	info, err := h.Storage.StatObject(ctx, imageName)
	if err != nil {
		log.Printf("Error getting image info %s: %v", imageName, err)
		http.Error(w, "Error retrieving image", http.StatusInternalServerError)
		return
	}

	// Set content type
	contentType := info.ContentType
	if contentType == "" {
		// Guess content type from extension
		ext := strings.ToLower(filepath.Ext(imageName))
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".webp":
			contentType = "image/webp"
		default:
			contentType = "application/octet-stream"
		}
	}

	// Set cache headers for optimal browser caching
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable") // 1 year
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, info.ETag))
	w.Header().Set("Last-Modified", info.LastModified.Format(http.TimeFormat))

	// Check if client has cached version (ETag)
	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, info.ETag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Check if client has cached version (Last-Modified)
	if modifiedSince := r.Header.Get("If-Modified-Since"); modifiedSince != "" {
		if t, err := time.Parse(http.TimeFormat, modifiedSince); err == nil {
			if !info.LastModified.After(t) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	// Stream image to client
	_, err = io.Copy(w, obj)
	if err != nil {
		log.Printf("Error streaming image %s: %v", imageName, err)
	}
}

// UploadImage handles image upload (for future admin functionality)
func (h *ImageHandlers) UploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get image from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "File must be an image", http.StatusBadRequest)
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	objectName := fmt.Sprintf("uploads/%d%s", time.Now().UnixNano(), ext)

	// Upload to MinIO
	ctx := r.Context()
	err = h.Storage.UploadImage(ctx, objectName, file, header.Size, contentType)
	if err != nil {
		log.Printf("Error uploading image: %v", err)
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
		return
	}

	// Return image URL
	imageURL := h.Storage.GetImageURL(objectName)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"url": "%s", "name": "%s"}`, imageURL, objectName)
}
