package models

type Product struct {
	ID              int
	Name            string
	Description     string
	Price           float64
	SKU             *string // Nullable
	StockQuantity   int
	ImageURL        *string // Nullable
	CategoryID      *int    // Nullable
	Status          string
	Author          *string // Nullable - for book products
	PopularityScore int     // Gutenberg download count, used for sorting
}

type Category struct {
	ID          int
	Name        string
	Description *string
}

// Pagination represents pagination parameters and metadata
type Pagination struct {
	Page       int // Current page (1-indexed)
	PageSize   int // Items per page
	TotalItems int // Total number of items
	TotalPages int // Total number of pages
}

// ProductsResult represents paginated products with metadata
type ProductsResult struct {
	Products   []Product
	Pagination Pagination
}

// PurchasedBook represents a book that a user has purchased (for Reader app integration)
type PurchasedBook struct {
	SKU         string `json:"sku"`
	GutenbergID int    `json:"gutenberg_id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	CoverURL    string `json:"cover_url"`
	PurchasedAt string `json:"purchased_at"` // ISO 8601 format
}
