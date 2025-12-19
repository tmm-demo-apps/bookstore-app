package models

type Product struct {
	ID            int
	Name          string
	Description   string
	Price         float64
	SKU           *string // Nullable
	StockQuantity int
	ImageURL      *string // Nullable
	CategoryID    *int    // Nullable
	Status        string
	Author        *string // Nullable - for book products
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
