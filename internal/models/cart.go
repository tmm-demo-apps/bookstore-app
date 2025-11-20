package models

type CartItem struct {
	ID        int
	ProductID int
	UserID    *int    // Nullable
	SessionID *string // Nullable
	Quantity  int
	// Joined fields for display
	Product  Product
	Subtotal float64
}

