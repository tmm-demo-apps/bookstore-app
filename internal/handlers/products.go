package handlers

import (
	"DemoApp/internal/models"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type ProductListViewData struct {
	IsAuthenticated  bool
	Products         []models.Product
	Categories       []models.Category
	SearchQuery      string
	SelectedCategory int
	ResultCount      int
}

type ProductDetailViewData struct {
	IsAuthenticated bool
	UserID          int
	Product         models.Product
	Reviews         []models.ReviewWithUser
	Rating          *models.ProductRating
	RatingBars      []models.RatingBar
	UserReview      *models.Review
}

func (h *Handlers) ListProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	categoryIDStr := r.URL.Query().Get("category")

	// Parse category ID
	categoryID := 0
	if categoryIDStr != "" {
		if id, err := strconv.Atoi(categoryIDStr); err == nil {
			categoryID = id
		}
	}

	var products []models.Product
	var err error

	if query != "" || categoryID > 0 {
		products, err = h.Repo.Products().SearchProducts(query, categoryID)
	} else {
		products, err = h.Repo.Products().ListProducts()
	}

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Fetch all categories for the sidebar
	categories, err := h.Repo.Products().ListCategories()
	if err != nil {
		log.Println("Error fetching categories:", err)
		categories = []models.Category{} // Continue with empty categories
	}

	data := ProductListViewData{
		IsAuthenticated:  h.IsAuthenticated(r),
		Products:         products,
		Categories:       categories,
		SearchQuery:      query,
		SelectedCategory: categoryID,
		ResultCount:      len(products),
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/products.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if err := ts.ExecuteTemplate(w, "products.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

func (h *Handlers) SearchSuggestions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	// Don't search for very short queries
	if len(query) < 2 {
		if _, err := w.Write([]byte("")); err != nil {
			log.Printf("Error writing response: %v", err)
		}
		return
	}

	// Search for products
	products, err := h.Repo.Products().SearchProducts(query, 0)
	if err != nil {
		log.Println(err)
		if _, err := w.Write([]byte("")); err != nil {
			log.Printf("Error writing response: %v", err)
		}
		return
	}

	// Limit to top 5 results
	if len(products) > 5 {
		products = products[:5]
	}

	// Return HTML list of suggestions
	if len(products) == 0 {
		if _, err := w.Write([]byte("<li><em>No results found</em></li>")); err != nil {
			log.Printf("Error writing response: %v", err)
		}
		return
	}

	for _, p := range products {
		html := `<li><a href="/products/` + strconv.Itoa(p.ID) + `">` + p.Name + `</a></li>`
		if _, err := w.Write([]byte(html)); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	}
}

func (h *Handlers) ProductDetail(w http.ResponseWriter, r *http.Request) {
	// Extract product ID from URL path
	idStr := r.PathValue("id")
	productID, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println("Invalid product ID:", err)
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Fetch product from database
	product, err := h.Repo.Products().GetProductByID(productID)
	if err != nil {
		log.Println("Product not found:", err)
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	// Fetch reviews for this product
	reviews, err := h.Repo.Reviews().GetReviewsByProductID(productID)
	if err != nil {
		log.Printf("Error fetching reviews: %v", err)
		reviews = []models.ReviewWithUser{} // Continue with empty reviews
	}

	// Fetch product rating statistics
	rating, err := h.Repo.Reviews().GetProductRating(productID)
	if err != nil {
		log.Printf("Error fetching product rating: %v", err)
		rating = nil // Continue without rating
	}

	// Calculate rating bar percentages for template
	var ratingBars []models.RatingBar
	if rating != nil && rating.TotalReviews > 0 {
		for stars := 5; stars >= 1; stars-- {
			count := rating.RatingCounts[stars]
			percentage := (float64(count) / float64(rating.TotalReviews)) * 100.0
			ratingBars = append(ratingBars, models.RatingBar{
				Stars:      stars,
				Count:      count,
				Percentage: percentage,
			})
		}
	}

	// Check if current user has already reviewed this product
	var userReview *models.Review
	userID, authenticated := h.GetUserID(r)
	if authenticated {
		userReview, err = h.Repo.Reviews().GetReviewByUserAndProduct(userID, productID)
		if err != nil {
			log.Printf("Error fetching user review: %v", err)
		}
	}

	data := ProductDetailViewData{
		IsAuthenticated: authenticated,
		UserID:          userID,
		Product:         *product,
		Reviews:         reviews,
		Rating:          rating,
		RatingBars:      ratingBars,
		UserReview:      userReview,
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/product-detail.html")
	if err != nil {
		log.Println("Template parsing error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "product-detail.html", data)
	if err != nil {
		log.Println("Template execution error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
