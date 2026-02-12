package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// APIHandlers handles JSON API endpoints for service-to-service communication
// These endpoints are used by Reader app and Chatbot app

// AuthRequest is the JSON request for API authentication
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is the JSON response for successful API authentication
type AuthResponse struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

// APIAuth validates user credentials and returns user info
// POST /api/auth
// Used by Reader app to authenticate users against the Bookstore
func (h *Handlers) APIAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password required", http.StatusBadRequest)
		return
	}

	// Validate credentials
	user, err := h.Repo.Users().GetUserByEmail(req.Email)
	if err != nil || user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !user.CheckPassword(req.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Return user info
	response := AuthResponse{
		UserID: user.ID,
		Email:  user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding auth response: %v", err)
	}
}

// PurchasesResponse is the JSON response for user purchases
type PurchasesResponse struct {
	Purchases []PurchaseItem `json:"purchases"`
}

// PurchaseItem represents a purchased book in API responses
type PurchaseItem struct {
	SKU         string `json:"sku"`
	GutenbergID int    `json:"gutenberg_id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	CoverURL    string `json:"cover_url"`
	PurchasedAt string `json:"purchased_at"`
}

// GetUserPurchases returns all books purchased by a user
// GET /api/purchases/{user_id}
func (h *Handlers) GetUserPurchases(w http.ResponseWriter, r *http.Request) {
	// Extract user_id from path: /api/purchases/{user_id}
	path := strings.TrimPrefix(r.URL.Path, "/api/purchases/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	purchases, err := h.Repo.Orders().GetUserPurchases(userID)
	if err != nil {
		log.Printf("Error fetching purchases for user %d: %v", userID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Convert to API response format
	response := PurchasesResponse{
		Purchases: make([]PurchaseItem, 0, len(purchases)),
	}
	for _, p := range purchases {
		response.Purchases = append(response.Purchases, PurchaseItem{
			SKU:         p.SKU,
			GutenbergID: p.GutenbergID,
			Title:       p.Title,
			Author:      p.Author,
			CoverURL:    p.CoverURL,
			PurchasedAt: p.PurchasedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding purchases response: %v", err)
	}
}

// VerifyPurchase checks if a user owns a specific book
// GET /api/purchases/{user_id}/{sku}
// Returns 200 OK if owned, 404 Not Found if not owned
func (h *Handlers) VerifyPurchase(w http.ResponseWriter, r *http.Request) {
	// Extract user_id and sku from path: /api/purchases/{user_id}/{sku}
	path := strings.TrimPrefix(r.URL.Path, "/api/purchases/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "user_id and sku required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	sku := parts[1]
	if sku == "" {
		http.Error(w, "sku required", http.StatusBadRequest)
		return
	}

	owned, err := h.Repo.Orders().VerifyPurchase(userID, sku)
	if err != nil {
		log.Printf("Error verifying purchase for user %d, sku %s: %v", userID, sku, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !owned {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"owned": true}`))
}

// APIProducts returns products as JSON for chatbot integration
// GET /api/products - list all products
// GET /api/products?category=Fiction - filter by category name
// GET /api/products/search?q=shakespeare - search products
func (h *Handlers) APIProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if this is a search request
	if strings.HasPrefix(r.URL.Path, "/api/products/search") {
		h.APISearchProducts(w, r)
		return
	}

	// Get category filter if provided
	categoryName := r.URL.Query().Get("category")
	var categoryID int

	if categoryName != "" {
		// Look up category by name
		categories, err := h.Repo.Products().ListCategories()
		if err != nil {
			log.Printf("Error fetching categories: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		for _, c := range categories {
			if strings.EqualFold(c.Name, categoryName) {
				categoryID = c.ID
				break
			}
		}
	}

	var products interface{}
	var err error

	if categoryID > 0 {
		products, err = h.Repo.Products().SearchProducts("", categoryID)
	} else {
		products, err = h.Repo.Products().ListProducts()
	}

	if err != nil {
		log.Printf("Error fetching products: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(products); err != nil {
		log.Printf("Error encoding products response: %v", err)
	}
}

// APISearchProducts searches products by query string
// GET /api/products/search?q=shakespeare
func (h *Handlers) APISearchProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "query parameter 'q' required", http.StatusBadRequest)
		return
	}

	products, err := h.Repo.Products().SearchProducts(query, 0)
	if err != nil {
		log.Printf("Error searching products: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(products); err != nil {
		log.Printf("Error encoding search response: %v", err)
	}
}

// APICategories returns all categories as JSON
// GET /api/categories
func (h *Handlers) APICategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.Repo.Products().ListCategories()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(categories); err != nil {
		log.Printf("Error encoding categories response: %v", err)
	}
}
