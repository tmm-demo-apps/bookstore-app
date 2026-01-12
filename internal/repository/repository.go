package repository

import (
	"DemoApp/internal/models"
)

type ProductRepository interface {
	ListProducts() ([]models.Product, error)
	ListProductsPaginated(page, pageSize int) (*models.ProductsResult, error)
	ListProductsPaginatedSorted(page, pageSize int, sortBy string) (*models.ProductsResult, error)
	GetProductByID(id int) (*models.Product, error)
	SearchProducts(query string, categoryID int) ([]models.Product, error)
	SearchProductsPaginated(query string, categoryID, page, pageSize int) (*models.ProductsResult, error)
	SearchProductsPaginatedSorted(query string, categoryID, page, pageSize int, sortBy string) (*models.ProductsResult, error)
	ListCategories() ([]models.Category, error)
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
	UpdateUserProfile(userID int, email, fullName string) error
	UpdateUserPassword(userID int, passwordHash string) error
}

type ReviewRepository interface {
	CreateReview(productID, userID, rating int, title, comment string) error
	GetReviewsByProductID(productID int) ([]models.ReviewWithUser, error)
	GetReviewByUserAndProduct(userID, productID int) (*models.Review, error)
	UpdateReview(reviewID, rating int, title, comment string) error
	DeleteReview(reviewID, userID int) error
	GetProductRating(productID int) (*models.ProductRating, error)
}

type Repository interface {
	Products() ProductRepository
	Orders() OrderRepository
	Cart() CartRepository
	Users() UserRepository
	Reviews() ReviewRepository
}
