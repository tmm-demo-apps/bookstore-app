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
}

type Category struct {
	ID          int
	Name        string
	Description *string
}
