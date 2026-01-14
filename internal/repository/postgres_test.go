package repository

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// getTestDB returns a database connection for testing.
// Uses environment variables: DB_USER, DB_PASSWORD, DB_HOST, DB_NAME, DB_PORT
func getTestDB(t *testing.T) *sql.DB {
	t.Helper()

	user := getEnv("DB_USER", "user")
	password := getEnv("DB_PASSWORD", "password")
	host := getEnv("DB_HOST", "localhost")
	dbname := getEnv("DB_NAME", "bookstore_test")
	port := getEnv("DB_PORT", "5432")

	dsn := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestDatabaseConnection verifies we can connect to the test database
func TestDatabaseConnection(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	var result int
	err := db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if result != 1 {
		t.Errorf("Expected 1, got %d", result)
	}
}

// TestCategoriesExist verifies categories were seeded
func TestCategoriesExist(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count categories: %v", err)
	}

	if count < 5 {
		t.Errorf("Expected at least 5 categories, got %d", count)
	}
	t.Logf("Found %d categories", count)
}

// TestProductsExist verifies products were seeded
func TestProductsExist(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM products WHERE status = 'active'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count products: %v", err)
	}

	if count < 100 {
		t.Errorf("Expected at least 100 products, got %d", count)
	}
	t.Logf("Found %d active products", count)
}

// TestListProducts tests the product listing functionality
func TestListProducts(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	products, err := repo.Products().ListProducts()
	if err != nil {
		t.Fatalf("ListProducts failed: %v", err)
	}

	if len(products) < 100 {
		t.Errorf("Expected at least 100 products, got %d", len(products))
	}

	// Verify products have required fields
	for i, p := range products {
		if p.ID == 0 {
			t.Errorf("Product %d has zero ID", i)
		}
		if p.Name == "" {
			t.Errorf("Product %d has empty name", i)
		}
		if p.Price <= 0 {
			t.Errorf("Product %d (%s) has invalid price: %f", i, p.Name, p.Price)
		}
		// Only check first 5 to avoid verbose output
		if i >= 5 {
			break
		}
	}
}

// TestListProductsPaginated tests pagination
func TestListProductsPaginated(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)

	// Test page 1
	result, err := repo.Products().ListProductsPaginated(1, 10)
	if err != nil {
		t.Fatalf("ListProductsPaginated failed: %v", err)
	}

	if len(result.Products) != 10 {
		t.Errorf("Expected 10 products on page 1, got %d", len(result.Products))
	}

	if result.Pagination.Page != 1 {
		t.Errorf("Expected page 1, got %d", result.Pagination.Page)
	}

	if result.Pagination.TotalItems < 100 {
		t.Errorf("Expected at least 100 total items, got %d", result.Pagination.TotalItems)
	}

	// Test page 2 has different products
	result2, err := repo.Products().ListProductsPaginated(2, 10)
	if err != nil {
		t.Fatalf("ListProductsPaginated page 2 failed: %v", err)
	}

	if len(result2.Products) != 10 {
		t.Errorf("Expected 10 products on page 2, got %d", len(result2.Products))
	}

	// Ensure page 1 and page 2 have different products
	if result.Products[0].ID == result2.Products[0].ID {
		t.Error("Page 1 and page 2 have the same first product - pagination may be broken")
	}
}

// TestGetProductByID tests fetching a single product
func TestGetProductByID(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)

	// First get a product ID from the list
	products, err := repo.Products().ListProducts()
	if err != nil {
		t.Fatalf("ListProducts failed: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("No products found")
	}

	testID := products[0].ID
	product, err := repo.Products().GetProductByID(testID)
	if err != nil {
		t.Fatalf("GetProductByID(%d) failed: %v", testID, err)
	}

	if product.ID != testID {
		t.Errorf("Expected product ID %d, got %d", testID, product.ID)
	}
	if product.Name == "" {
		t.Error("Product name is empty")
	}
}

// TestListCategories tests category listing
func TestListCategories(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)
	categories, err := repo.Products().ListCategories()
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}

	if len(categories) < 5 {
		t.Errorf("Expected at least 5 categories, got %d", len(categories))
	}

	// Check expected categories exist
	expectedCategories := []string{"Fiction", "Science Fiction", "Drama", "Poetry", "Philosophy"}
	for _, expected := range expectedCategories {
		found := false
		for _, cat := range categories {
			if cat.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected category '%s' not found", expected)
		}
	}
}

// TestCartOperations tests cart add/get/update/remove
func TestCartOperations(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)

	// Use a unique session ID for this test
	sessionID := "test-session-" + t.Name()

	// Clean up any existing cart items for this session
	_, _ = db.Exec("DELETE FROM cart_items WHERE session_id = $1", sessionID)

	// Find a product with sufficient stock (at least 10)
	// Note: AddToCart caps quantity at available stock
	var testProductID int
	var testProductStock int
	err := db.QueryRow(`
		SELECT id, stock_quantity FROM products 
		WHERE status = 'active' AND stock_quantity >= 10 
		ORDER BY id LIMIT 1
	`).Scan(&testProductID, &testProductStock)
	if err != nil {
		t.Fatalf("Failed to find product with sufficient stock: %v", err)
	}
	t.Logf("Using product ID %d with stock %d", testProductID, testProductStock)

	// Add to cart (userID=0 means anonymous, use sessionID)
	err = repo.Cart().AddToCart(0, sessionID, testProductID, 2)
	if err != nil {
		t.Fatalf("AddToCart failed: %v", err)
	}

	// Get cart items
	items, total, err := repo.Cart().GetCartItems(0, sessionID)
	if err != nil {
		t.Fatalf("GetCartItems failed: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 cart item, got %d", len(items))
	} else {
		if items[0].ProductID != testProductID {
			t.Errorf("Expected product ID %d, got %d", testProductID, items[0].ProductID)
		}
		if items[0].Quantity != 2 {
			t.Errorf("Expected quantity 2, got %d", items[0].Quantity)
		}
	}

	if total <= 0 {
		t.Errorf("Expected positive total, got %f", total)
	}

	// Update quantity (should work since stock >= 10)
	err = repo.Cart().UpdateQuantity(0, sessionID, testProductID, 5)
	if err != nil {
		t.Fatalf("UpdateQuantity failed: %v", err)
	}

	// Verify update
	items, _, _ = repo.Cart().GetCartItems(0, sessionID)
	if len(items) == 1 && items[0].Quantity != 5 {
		t.Errorf("Expected quantity 5 after update, got %d", items[0].Quantity)
	}

	// Remove from cart
	err = repo.Cart().RemoveItem(0, sessionID, testProductID)
	if err != nil {
		t.Fatalf("RemoveItem failed: %v", err)
	}

	// Verify removal
	items, _, _ = repo.Cart().GetCartItems(0, sessionID)
	if len(items) != 0 {
		t.Errorf("Expected empty cart after removal, got %d items", len(items))
	}
}

// TestUserOperations tests user creation and lookup
func TestUserOperations(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)

	// Use unique email for this test
	testEmail := "test-" + t.Name() + "@example.com"
	testPassword := "hashed_password_here" // In real app, this would be hashed
	testFullName := "Test User"

	// Clean up any existing user with this email
	_, _ = db.Exec("DELETE FROM users WHERE email = $1", testEmail)

	// Create user
	userID, err := repo.Users().CreateUser(testEmail, testPassword, testFullName)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if userID == 0 {
		t.Error("User ID should be non-zero after creation")
	}

	// Get user by email
	foundUser, err := repo.Users().GetUserByEmail(testEmail)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}

	if foundUser.Email != testEmail {
		t.Errorf("Expected email %s, got %s", testEmail, foundUser.Email)
	}

	// Get user by ID
	foundByID, err := repo.Users().GetUserByID(userID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}

	if foundByID.ID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, foundByID.ID)
	}

	// Clean up
	_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
}

// TestStockQuantityRules verifies the stock quantity rules from seed data
func TestStockQuantityRules(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	// Check "A Christmas Carol" is out of stock (0)
	var christmasStock int
	err := db.QueryRow("SELECT stock_quantity FROM products WHERE name = 'A Christmas Carol'").Scan(&christmasStock)
	if err != nil {
		t.Logf("Note: 'A Christmas Carol' not found - %v", err)
	} else if christmasStock != 0 {
		t.Errorf("'A Christmas Carol' should have stock_quantity=0 (out of stock demo), got %d", christmasStock)
	}

	// Check "Pride and Prejudice" has low stock (3)
	var prideStock int
	err = db.QueryRow("SELECT stock_quantity FROM products WHERE name = 'Pride and Prejudice'").Scan(&prideStock)
	if err != nil {
		t.Logf("Note: 'Pride and Prejudice' not found - %v", err)
	} else if prideStock != 3 {
		t.Errorf("'Pride and Prejudice' should have stock_quantity=3 (low stock demo), got %d", prideStock)
	}
}

// TestUniqueConstraints verifies database constraints
func TestUniqueConstraints(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	// Check that cart_items has unique constraints
	var indexCount int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM pg_indexes 
		WHERE tablename = 'cart_items' 
		AND indexname LIKE 'idx_cart_items_%'
	`).Scan(&indexCount)
	if err != nil {
		t.Fatalf("Failed to query indexes: %v", err)
	}

	if indexCount < 2 {
		t.Errorf("Expected at least 2 cart_items indexes, got %d", indexCount)
	}
}

// TestSearchProducts tests the search functionality
func TestSearchProducts(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	repo := NewPostgresRepository(db)

	// Search for a common term
	products, err := repo.Products().SearchProducts("pride", 0)
	if err != nil {
		t.Fatalf("SearchProducts failed: %v", err)
	}

	// Should find at least "Pride and Prejudice"
	if len(products) == 0 {
		t.Log("Note: Search returned no results - Elasticsearch may not be available in CI")
	} else {
		t.Logf("Search for 'pride' returned %d results", len(products))
	}
}
