package handlers

import (
	"log"
	"net/http"
	"strconv"
)

// SubmitReview handles review submission (POST /products/{id}/review)
func (h *Handlers) SubmitReview(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	userID, ok := h.GetUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse product ID from URL
	productIDStr := r.PathValue("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get rating (required)
	ratingStr := r.FormValue("rating")
	rating, err := strconv.Atoi(ratingStr)
	if err != nil || rating < 1 || rating > 5 {
		http.Error(w, "Invalid rating (must be 1-5)", http.StatusBadRequest)
		return
	}

	// Get optional title and comment
	title := r.FormValue("title")
	comment := r.FormValue("comment")

	// Create or update review
	err = h.Repo.Reviews().CreateReview(productID, userID, rating, title, comment)
	if err != nil {
		log.Printf("Error creating review: %v", err)
		http.Error(w, "Failed to submit review", http.StatusInternalServerError)
		return
	}

	// Redirect back to product detail page
	http.Redirect(w, r, "/products/"+productIDStr, http.StatusSeeOther)
}

// DeleteReview handles review deletion (POST /reviews/{id}/delete)
func (h *Handlers) DeleteReview(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	userID, ok := h.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse review ID from URL
	reviewIDStr := r.PathValue("id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		http.Error(w, "Invalid review ID", http.StatusBadRequest)
		return
	}

	// Delete review (repository ensures user owns the review)
	err = h.Repo.Reviews().DeleteReview(reviewID, userID)
	if err != nil {
		log.Printf("Error deleting review: %v", err)
		http.Error(w, "Failed to delete review", http.StatusInternalServerError)
		return
	}

	// Get product ID from form to redirect back
	productIDStr := r.FormValue("product_id")
	if productIDStr == "" {
		// Fallback to home if no product ID
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/products/"+productIDStr, http.StatusSeeOther)
}
