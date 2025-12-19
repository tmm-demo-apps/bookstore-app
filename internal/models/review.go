package models

import (
	"strings"
	"time"
)

// Review represents a product review with rating and comment
type Review struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	UserID    int       `json:"user_id"`
	Rating    int       `json:"rating"`  // 1-5 stars
	Title     string    `json:"title"`   // Optional short title
	Comment   string    `json:"comment"` // Optional detailed review
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ReviewWithUser extends Review with user information for display
type ReviewWithUser struct {
	Review
	UserName string `json:"user_name"` // User's full name or email
}

// ProductRating contains aggregated rating statistics for a product
type ProductRating struct {
	ProductID     int         `json:"product_id"`
	AverageRating float64     `json:"average_rating"` // Average of all ratings
	TotalReviews  int         `json:"total_reviews"`  // Count of reviews
	RatingCounts  map[int]int `json:"rating_counts"`  // Count per star (1-5)
}

// RatingBar represents a single star rating bar for display
type RatingBar struct {
	Stars      int
	Count      int
	Percentage float64
}

// FormatDisplayName formats a user's name for review display
// If full_name exists: "FirstName L." (first name + last initial)
// Otherwise: falls back to email
func FormatDisplayName(fullName, email string) string {
	if fullName == "" {
		return email
	}

	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return email
	}

	if len(parts) == 1 {
		// Only first name provided
		return parts[0]
	}

	// Multiple parts: use first name + last initial
	firstName := parts[0]
	lastName := parts[len(parts)-1]
	lastInitial := string([]rune(lastName)[0]) // Get first character (handles unicode)

	return firstName + " " + lastInitial + "."
}
