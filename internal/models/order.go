package models

import "time"

type Order struct {
	ID           int
	SessionID    string
	UserID       *int // Nullable
	TotalAmount  float64
	Status       string
	ShippingInfo *string // JSONB stored as string
	CreatedAt    time.Time
	Items        []OrderItem
}

type OrderItem struct {
	ID        int
	OrderID   int
	ProductID int
	Quantity  int
	Price     float64
	Product   Product // Joined
}
