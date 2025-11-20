package repository

import (
	"DemoApp/internal/models"
	"database/sql"
)

type ProductRepository interface {
	ListProducts() ([]models.Product, error)
	GetProductByID(id int) (*models.Product, error)
	SearchProducts(query string, categoryID int) ([]models.Product, error)
}

type OrderRepository interface {
	CreateOrder(sessionID string, userID int, items []models.CartItem) (int, error)
	GetOrderByID(id int) (*models.Order, error)
	GetOrdersByUserID(userID int) ([]models.Order, error)
}

type CartRepository interface {
	GetCartItems(userID int, sessionID string) ([]models.CartItem, float64, error)
	GetCartItem(id int) (*models.CartItem, error)
	AddToCart(userID int, sessionID string, productID, quantity int) error
	UpdateQuantity(userID int, sessionID string, productID, quantity int) error
	RemoveItem(userID int, sessionID string, productID int) error
	ClearCart(userID int, sessionID string) error
	MergeCart(sessionID string, userID int) error
}

type UserRepository interface {
	CreateUser(email, passwordHash, fullName string) (int, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
}

type Repository interface {
	Products() ProductRepository
	Orders() OrderRepository
	Cart() CartRepository
	Users() UserRepository
}

