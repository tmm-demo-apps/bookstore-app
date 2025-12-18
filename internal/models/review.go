package models

import "time"

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
