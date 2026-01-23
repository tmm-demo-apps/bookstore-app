package repository

import (
	"DemoApp/internal/models"
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
)

type PostgresRepository struct {
	DB             *sql.DB
	ES             *ElasticsearchRepository
	CachedProducts ProductRepository
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{DB: db, ES: nil, CachedProducts: nil}
}

func (r *PostgresRepository) SetElasticsearch(es *ElasticsearchRepository) {
	r.ES = es
}

func (r *PostgresRepository) SetCachedProducts(cached ProductRepository) {
	r.CachedProducts = cached
}

func (r *PostgresRepository) Products() ProductRepository {
	// Return cached version if available, otherwise return direct DB access
	if r.CachedProducts != nil {
		return r.CachedProducts
	}
	return &postgresProductRepo{DB: r.DB, ES: r.ES}
}

func (r *PostgresRepository) Orders() OrderRepository {
	return &postgresOrderRepo{DB: r.DB}
}

func (r *PostgresRepository) Cart() CartRepository {
	return &postgresCartRepo{DB: r.DB}
}

func (r *PostgresRepository) Users() UserRepository {
	return &postgresUserRepo{DB: r.DB}
}

func (r *PostgresRepository) Reviews() ReviewRepository {
	return &postgresReviewRepo{DB: r.DB}
}

// --- Product Implementation ---

type postgresProductRepo struct {
	DB *sql.DB
	ES *ElasticsearchRepository
}

func (r *postgresProductRepo) ListProducts() ([]models.Product, error) {
	// Updated query for new schema
	query := `SELECT id, name, description, price, sku, stock_quantity, image_url, category_id, status, author, COALESCE(popularity_score, 0) 
	          FROM products WHERE status = 'active' ORDER BY name`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.StockQuantity, &p.ImageURL, &p.CategoryID, &p.Status, &p.Author, &p.PopularityScore); err != nil {
			// Handle case where columns might be NULL if left joined (though here we query products directly)
			// The pointer types in struct handle NULLs automatically via Scan if valid.
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *postgresProductRepo) ListProductsPaginated(page, pageSize int) (*models.ProductsResult, error) {
	return r.ListProductsPaginatedSorted(page, pageSize, "name")
}

func (r *postgresProductRepo) ListProductsPaginatedSorted(page, pageSize int, sortBy string) (*models.ProductsResult, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10 // Default
	}

	// Get total count
	var totalItems int
	err := r.DB.QueryRow("SELECT COUNT(*) FROM products WHERE status = 'active'").Scan(&totalItems)
	if err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Determine ORDER BY clause based on sortBy parameter
	orderClause := getOrderClause(sortBy)

	// Get paginated products
	query := fmt.Sprintf(`SELECT id, name, description, price, sku, stock_quantity, image_url, category_id, status, author, COALESCE(popularity_score, 0) 
	          FROM products WHERE status = 'active' ORDER BY %s LIMIT $1 OFFSET $2`, orderClause)
	rows, err := r.DB.Query(query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.StockQuantity, &p.ImageURL, &p.CategoryID, &p.Status, &p.Author, &p.PopularityScore); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	// Calculate total pages
	totalPages := (totalItems + pageSize - 1) / pageSize

	return &models.ProductsResult{
		Products: products,
		Pagination: models.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	}, nil
}

// getOrderClause returns a safe ORDER BY clause based on the sortBy parameter
func getOrderClause(sortBy string) string {
	switch sortBy {
	case "price_asc":
		return "price ASC, name ASC"
	case "price_desc":
		return "price DESC, name ASC"
	case "popularity":
		return "popularity_score DESC, name ASC"
	case "newest":
		return "id DESC, name ASC"
	default:
		return "name ASC"
	}
}

func (r *postgresProductRepo) GetProductByID(id int) (*models.Product, error) {
	var p models.Product
	query := `SELECT id, name, description, price, sku, stock_quantity, image_url, category_id, status, author, COALESCE(popularity_score, 0) 
	          FROM products WHERE id = $1`
	err := r.DB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.StockQuantity, &p.ImageURL, &p.CategoryID, &p.Status, &p.Author, &p.PopularityScore)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *postgresProductRepo) SearchProducts(query string, categoryID int) ([]models.Product, error) {
	// If Elasticsearch is available, use it for search
	if r.ES != nil {
		log.Printf("Using Elasticsearch for search: query='%s', categoryID=%d", query, categoryID)
		productIDs, err := r.ES.SearchProducts(query, categoryID)
		if err != nil {
			log.Printf("Elasticsearch search failed, falling back to SQL: %v", err)
			// Fall through to SQL search
		} else {
			log.Printf("Elasticsearch returned %d product IDs", len(productIDs))
			// Fetch products from database by IDs in the order returned by Elasticsearch
			if len(productIDs) == 0 {
				return []models.Product{}, nil
			}
			products, err := r.getProductsByIDs(productIDs)
			if err != nil {
				log.Printf("Error fetching products by IDs, falling back to SQL: %v", err)
				// Fall through to SQL search
			} else {
				return products, nil
			}
		}
	}

	// Fallback to SQL-based search (original implementation)
	q := `SELECT id, name, description, price, sku, stock_quantity, image_url, category_id, status, author, COALESCE(popularity_score, 0) 
	      FROM products WHERE status = 'active'`
	var args []interface{}
	argID := 1

	if query != "" {
		q += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR author ILIKE $%d)", argID, argID, argID)
		args = append(args, "%"+query+"%")
		argID++
	}

	if categoryID > 0 {
		q += fmt.Sprintf(" AND category_id = $%d", argID)
		args = append(args, categoryID)
	}

	q += " ORDER BY name"

	rows, err := r.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.StockQuantity, &p.ImageURL, &p.CategoryID, &p.Status, &p.Author, &p.PopularityScore); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *postgresProductRepo) SearchProductsPaginated(query string, categoryID, page, pageSize int) (*models.ProductsResult, error) {
	return r.SearchProductsPaginatedSorted(query, categoryID, page, pageSize, "name")
}

func (r *postgresProductRepo) SearchProductsPaginatedSorted(query string, categoryID, page, pageSize int, sortBy string) (*models.ProductsResult, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10 // Default
	}

	// Build count query
	countQuery := `SELECT COUNT(*) FROM products WHERE status = 'active'`
	var countArgs []interface{}
	argID := 1

	if query != "" {
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR author ILIKE $%d)", argID, argID, argID)
		countArgs = append(countArgs, "%"+query+"%")
		argID++
	}

	if categoryID > 0 {
		countQuery += fmt.Sprintf(" AND category_id = $%d", argID)
		countArgs = append(countArgs, categoryID)
	}

	// Get total count
	var totalItems int
	err := r.DB.QueryRow(countQuery, countArgs...).Scan(&totalItems)
	if err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Determine ORDER BY clause based on sortBy parameter
	orderClause := getOrderClause(sortBy)

	// Build products query
	q := `SELECT id, name, description, price, sku, stock_quantity, image_url, category_id, status, author, COALESCE(popularity_score, 0) 
	      FROM products WHERE status = 'active'`
	var args []interface{}
	argID = 1

	if query != "" {
		q += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR author ILIKE $%d)", argID, argID, argID)
		args = append(args, "%"+query+"%")
		argID++
	}

	if categoryID > 0 {
		q += fmt.Sprintf(" AND category_id = $%d", argID)
		args = append(args, categoryID)
		argID++
	}

	q += fmt.Sprintf(" ORDER BY %s LIMIT $%d OFFSET $%d", orderClause, argID, argID+1)
	args = append(args, pageSize, offset)

	rows, err := r.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.StockQuantity, &p.ImageURL, &p.CategoryID, &p.Status, &p.Author, &p.PopularityScore); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	// Calculate total pages
	totalPages := (totalItems + pageSize - 1) / pageSize

	return &models.ProductsResult{
		Products: products,
		Pagination: models.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	}, nil
}

// getProductsByIDs fetches products by IDs and maintains the order
func (r *postgresProductRepo) getProductsByIDs(ids []int) ([]models.Product, error) {
	if len(ids) == 0 {
		return []models.Product{}, nil
	}

	// Create a map to store products by ID
	productMap := make(map[int]models.Product)

	// Build query with IN clause
	q := `SELECT id, name, description, price, sku, stock_quantity, image_url, category_id, status, author, COALESCE(popularity_score, 0) 
	      FROM products WHERE id = ANY($1) AND status = 'active'`

	rows, err := r.DB.Query(q, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Fetch all products into the map
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.StockQuantity, &p.ImageURL, &p.CategoryID, &p.Status, &p.Author, &p.PopularityScore); err != nil {
			return nil, err
		}
		productMap[p.ID] = p
	}

	// Build result slice in the order of input IDs (Elasticsearch relevance order)
	products := make([]models.Product, 0, len(ids))
	for _, id := range ids {
		if p, exists := productMap[id]; exists {
			products = append(products, p)
		}
	}

	return products, nil
}

func (r *postgresProductRepo) ListCategories() ([]models.Category, error) {
	query := `SELECT id, name, description FROM categories ORDER BY name`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// --- Order Implementation ---

type postgresOrderRepo struct {
	DB *sql.DB
}

func (r *postgresOrderRepo) CreateOrder(sessionID string, userID int, items []models.CartItem) (int, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}

	var orderID int
	var errCreate error

	// Calculate total amount first?
	// For now, we insert order then items, then maybe update total?
	// Or calculate total in SQL?
	// Let's stick to basic insertion first, update total later or via trigger?
	// The prompt asks to add `total_amount` column.
	// Ideally we calculate it.

	// Let's just insert the order.

	if userID > 0 {
		errCreate = tx.QueryRow("INSERT INTO orders (session_id, user_id) VALUES ($1, $2) RETURNING id", sessionID, userID).Scan(&orderID)
	} else {
		errCreate = tx.QueryRow("INSERT INTO orders (session_id) VALUES ($1) RETURNING id", sessionID).Scan(&orderID)
	}

	if errCreate != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Error rolling back transaction: %v", rbErr)
		}
		return 0, errCreate
	}

	// Insert items
	if userID > 0 {
		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, quantity, price)
			SELECT $1, product_id, SUM(quantity), (SELECT price FROM products WHERE id = product_id)
			FROM cart_items 
			WHERE user_id = $2
			GROUP BY product_id`, orderID, userID)
	} else {
		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, quantity, price)
			SELECT $1, product_id, SUM(quantity), (SELECT price FROM products WHERE id = product_id)
			FROM cart_items 
			WHERE session_id = $2
			GROUP BY product_id`, orderID, sessionID)
	}

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Error rolling back transaction: %v", rbErr)
		}
		return 0, err
	}

	// Update Order Total
	_, err = tx.Exec(`
		UPDATE orders 
		SET total_amount = (SELECT SUM(price * quantity) FROM order_items WHERE order_id = $1)
		WHERE id = $1`, orderID)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Error rolling back transaction: %v", rbErr)
		}
		return 0, err
	}

	// Reduce Stock Quantities
	// For each order item, reduce the corresponding product stock
	_, err = tx.Exec(`
		UPDATE products p
		SET stock_quantity = stock_quantity - oi.quantity
		FROM order_items oi
		WHERE p.id = oi.product_id AND oi.order_id = $1`, orderID)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Error rolling back transaction: %v", rbErr)
		}
		return 0, err
	}

	// Clear Cart
	if userID > 0 {
		_, err = tx.Exec("DELETE FROM cart_items WHERE user_id = $1", userID)
	} else {
		_, err = tx.Exec("DELETE FROM cart_items WHERE session_id = $1", sessionID)
	}
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Error rolling back transaction: %v", rbErr)
		}
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return orderID, nil
}

func (r *postgresOrderRepo) GetOrderByID(id int) (*models.Order, error) {
	// Placeholder
	return nil, nil
}

func (r *postgresOrderRepo) GetOrdersByUserID(userID int) ([]models.Order, error) {
	rows, err := r.DB.Query("SELECT id, session_id, user_id, total_amount, status, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		// handle nullable fields
		if err := rows.Scan(&o.ID, &o.SessionID, &o.UserID, &o.TotalAmount, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}

		// Load order items with product details
		itemRows, err := r.DB.Query(`
			SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price,
			       p.id, p.name, p.description, p.price, p.image_url
			FROM order_items oi
			JOIN products p ON oi.product_id = p.id
			WHERE oi.order_id = $1
			ORDER BY p.name`, o.ID)
		if err != nil {
			return nil, err
		}

		var items []models.OrderItem
		for itemRows.Next() {
			var item models.OrderItem
			var prod models.Product
			if err := itemRows.Scan(
				&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price,
				&prod.ID, &prod.Name, &prod.Description, &prod.Price, &prod.ImageURL,
			); err != nil {
				itemRows.Close()
				return nil, err
			}
			item.Product = prod
			items = append(items, item)
		}
		itemRows.Close()

		o.Items = items
		orders = append(orders, o)
	}
	return orders, nil
}

// GetUserPurchases returns all unique books a user has purchased (for Reader app integration)
func (r *postgresOrderRepo) GetUserPurchases(userID int) ([]models.PurchasedBook, error) {
	// Get distinct products purchased by user, with earliest purchase date
	query := `
		SELECT DISTINCT ON (p.sku) 
			p.sku, p.name, COALESCE(p.author, ''), p.image_url, 
			MIN(o.created_at) as purchased_at
		FROM orders o
		JOIN order_items oi ON o.id = oi.order_id
		JOIN products p ON oi.product_id = p.id
		WHERE o.user_id = $1 AND p.sku IS NOT NULL
		GROUP BY p.sku, p.name, p.author, p.image_url
		ORDER BY p.sku, purchased_at ASC`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var purchases []models.PurchasedBook
	for rows.Next() {
		var pb models.PurchasedBook
		var imageURL sql.NullString
		var purchasedAt sql.NullTime

		if err := rows.Scan(&pb.SKU, &pb.Title, &pb.Author, &imageURL, &purchasedAt); err != nil {
			return nil, err
		}

		// Extract Gutenberg ID from SKU (format: BOOK-{GutenbergID})
		if len(pb.SKU) > 5 && pb.SKU[:5] == "BOOK-" {
			_, _ = fmt.Sscanf(pb.SKU[5:], "%d", &pb.GutenbergID)
		}

		if imageURL.Valid {
			pb.CoverURL = imageURL.String
		}

		if purchasedAt.Valid {
			pb.PurchasedAt = purchasedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}

		purchases = append(purchases, pb)
	}

	return purchases, nil
}

// VerifyPurchase checks if a user has purchased a specific book by SKU
func (r *postgresOrderRepo) VerifyPurchase(userID int, sku string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM orders o
			JOIN order_items oi ON o.id = oi.order_id
			JOIN products p ON oi.product_id = p.id
			WHERE o.user_id = $1 AND p.sku = $2
		)`

	err := r.DB.QueryRow(query, userID, sku).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// --- Cart Implementation ---

type postgresCartRepo struct {
	DB *sql.DB
}

func (r *postgresCartRepo) GetCartItems(userID int, sessionID string) ([]models.CartItem, float64, error) {
	// Need to update this query if product fields changed?
	// Just ensure we select compatible fields.
	var rows *sql.Rows
	var err error

	if userID > 0 {
		rows, err = r.DB.Query(`
			SELECT ci.id, ci.product_id, p.name, p.description, p.price, p.image_url, ci.quantity
			FROM cart_items ci
			JOIN products p ON ci.product_id = p.id
			WHERE ci.user_id = $1
			ORDER BY p.name`, userID)
	} else {
		rows, err = r.DB.Query(`
			SELECT ci.id, ci.product_id, p.name, p.description, p.price, p.image_url, ci.quantity
			FROM cart_items ci
			JOIN products p ON ci.product_id = p.id
			WHERE ci.session_id = $1
			ORDER BY p.name`, sessionID)
	}

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.CartItem
	var total float64

	for rows.Next() {
		var item models.CartItem
		var p models.Product
		var imageURL sql.NullString
		if err := rows.Scan(&item.ID, &item.ProductID, &p.Name, &p.Description, &p.Price, &imageURL, &item.Quantity); err != nil {
			return nil, 0, err
		}
		if imageURL.Valid {
			imgURL := imageURL.String
			p.ImageURL = &imgURL
		}
		item.Product = p
		item.Subtotal = p.Price * float64(item.Quantity)
		total += item.Subtotal
		items = append(items, item)
	}
	return items, total, nil
}

func (r *postgresCartRepo) GetCartItem(id int) (*models.CartItem, error) {
	var item models.CartItem
	var userID sql.NullInt64
	var sessionID sql.NullString

	err := r.DB.QueryRow("SELECT id, product_id, user_id, session_id, quantity FROM cart_items WHERE id = $1", id).
		Scan(&item.ID, &item.ProductID, &userID, &sessionID, &item.Quantity)

	if err != nil {
		return nil, err
	}

	if userID.Valid {
		uid := int(userID.Int64)
		item.UserID = &uid
	}
	if sessionID.Valid {
		sid := sessionID.String
		item.SessionID = &sid
	}

	return &item, nil
}

func (r *postgresCartRepo) AddToCart(userID int, sessionID string, productID, quantity int) error {
	// First, check available stock
	var stockQty int
	err := r.DB.QueryRow("SELECT stock_quantity FROM products WHERE id = $1", productID).Scan(&stockQty)
	if err != nil {
		return err // Product doesn't exist or other error
	}

	// Get existing quantity in cart BEFORE deleting
	var existingQty int
	if userID > 0 {
		err = r.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, productID).Scan(&existingQty)
	} else {
		err = r.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE session_id = $1 AND product_id = $2", sessionID, productID).Scan(&existingQty)
	}
	if err != nil {
		existingQty = 0
	}

	// Calculate new quantity with stock limit
	newQty := existingQty + quantity

	// Enforce stock limit FIRST (most important)
	if newQty > stockQty {
		newQty = stockQty
	}
	// Then enforce system limit
	if newQty > 99 {
		newQty = 99
	}
	if newQty < 1 {
		newQty = 1
	}

	// Delete ALL existing rows for this product (to handle duplicates)
	if userID > 0 {
		_, err = r.DB.Exec("DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, productID)
	} else {
		_, err = r.DB.Exec("DELETE FROM cart_items WHERE session_id = $1 AND product_id = $2", sessionID, productID)
	}
	if err != nil {
		return err
	}

	// Insert new consolidated row with combined quantity
	if userID > 0 {
		_, err = r.DB.Exec("INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3)", userID, productID, newQty)
	} else {
		_, err = r.DB.Exec("INSERT INTO cart_items (session_id, product_id, quantity) VALUES ($1, $2, $3)", sessionID, productID, newQty)
	}
	return err
}

func (r *postgresCartRepo) UpdateQuantity(userID int, sessionID string, productID, quantity int) error {
	// Check available stock
	var stockQty int
	err := r.DB.QueryRow("SELECT stock_quantity FROM products WHERE id = $1", productID).Scan(&stockQty)
	if err != nil {
		return err
	}

	// Enforce stock limit
	if quantity > stockQty {
		quantity = stockQty
	}
	if quantity > 99 {
		quantity = 99
	}
	if quantity < 1 {
		quantity = 1
	}

	if userID > 0 {
		_, err = r.DB.Exec("DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, productID)
	} else {
		_, err = r.DB.Exec("DELETE FROM cart_items WHERE session_id = $1 AND product_id = $2", sessionID, productID)
	}
	if err != nil {
		return err
	}

	if userID > 0 {
		_, err = r.DB.Exec("INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3)", userID, productID, quantity)
	} else {
		_, err = r.DB.Exec("INSERT INTO cart_items (session_id, product_id, quantity) VALUES ($1, $2, $3)", sessionID, productID, quantity)
	}
	return err
}

func (r *postgresCartRepo) RemoveItem(userID int, sessionID string, productID int) error {
	var err error
	if userID > 0 {
		_, err = r.DB.Exec("DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, productID)
	} else {
		_, err = r.DB.Exec("DELETE FROM cart_items WHERE session_id = $1 AND product_id = $2", sessionID, productID)
	}
	return err
}

func (r *postgresCartRepo) ClearCart(userID int, sessionID string) error {
	var err error
	if userID > 0 {
		_, err = r.DB.Exec("DELETE FROM cart_items WHERE user_id = $1", userID)
	} else {
		_, err = r.DB.Exec("DELETE FROM cart_items WHERE session_id = $1", sessionID)
	}
	return err
}

func (r *postgresCartRepo) MergeCart(sessionID string, userID int) error {
	rows, err := r.DB.Query(`
		SELECT product_id, SUM(quantity) 
		FROM cart_items 
		WHERE session_id = $1 
		GROUP BY product_id`, sessionID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type item struct {
		PID int
		Qty int
	}
	var items []item
	for rows.Next() {
		var i item
		if err := rows.Scan(&i.PID, &i.Qty); err != nil {
			return err
		}
		items = append(items, i)
	}
	rows.Close()

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	for _, i := range items {
		var userQty int
		err = tx.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, i.PID).Scan(&userQty)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Error rolling back transaction: %v", rbErr)
			}
			return err
		}

		_, err = tx.Exec("DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, i.PID)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Error rolling back transaction: %v", rbErr)
			}
			return err
		}

		newQty := userQty + i.Qty
		if newQty > 99 {
			newQty = 99
		}
		_, err = tx.Exec("INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3)", userID, i.PID, newQty)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Error rolling back transaction: %v", rbErr)
			}
			return err
		}
	}

	_, err = tx.Exec("DELETE FROM cart_items WHERE session_id = $1", sessionID)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Error rolling back transaction: %v", rbErr)
		}
		return err
	}

	return tx.Commit()
}

// --- User Implementation ---

type postgresUserRepo struct {
	DB *sql.DB
}

func (r *postgresUserRepo) CreateUser(email, passwordHash, fullName string) (int, error) {
	var id int
	err := r.DB.QueryRow("INSERT INTO users (email, password_hash, full_name) VALUES ($1, $2, $3) RETURNING id", email, passwordHash, fullName).Scan(&id)
	return id, err
}

func (r *postgresUserRepo) GetUserByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.DB.QueryRow("SELECT id, email, password_hash, full_name, role FROM users WHERE email = $1", email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *postgresUserRepo) GetUserByID(id int) (*models.User, error) {
	var u models.User
	err := r.DB.QueryRow("SELECT id, email, password_hash, full_name, role, created_at FROM users WHERE id = $1", id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *postgresUserRepo) UpdateUserProfile(userID int, email, fullName string) error {
	query := `UPDATE users SET email = $1, full_name = $2 WHERE id = $3`
	_, err := r.DB.Exec(query, email, fullName, userID)
	return err
}

func (r *postgresUserRepo) UpdateUserPassword(userID int, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`
	_, err := r.DB.Exec(query, passwordHash, userID)
	return err
}

// --- Review Implementation ---

type postgresReviewRepo struct {
	DB *sql.DB
}

func (r *postgresReviewRepo) CreateReview(productID, userID, rating int, title, comment string) error {
	query := `
		INSERT INTO reviews (product_id, user_id, rating, title, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (product_id, user_id) 
		DO UPDATE SET rating = $3, title = $4, comment = $5, updated_at = NOW()`

	_, err := r.DB.Exec(query, productID, userID, rating, title, comment)
	return err
}

func (r *postgresReviewRepo) GetReviewsByProductID(productID int) ([]models.ReviewWithUser, error) {
	query := `
		SELECT r.id, r.product_id, r.user_id, r.rating, r.title, r.comment, 
		       r.created_at, r.updated_at, u.full_name, u.email
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.product_id = $1
		ORDER BY r.created_at DESC`

	rows, err := r.DB.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []models.ReviewWithUser
	for rows.Next() {
		var rw models.ReviewWithUser
		var fullName sql.NullString
		var email string
		err := rows.Scan(
			&rw.ID, &rw.ProductID, &rw.UserID, &rw.Rating, &rw.Title, &rw.Comment,
			&rw.CreatedAt, &rw.UpdatedAt, &fullName, &email,
		)
		if err != nil {
			return nil, err
		}

		// Format display name: "FirstName L." or email fallback
		if fullName.Valid {
			rw.UserName = models.FormatDisplayName(fullName.String, email)
		} else {
			rw.UserName = models.FormatDisplayName("", email)
		}

		reviews = append(reviews, rw)
	}

	return reviews, nil
}

func (r *postgresReviewRepo) GetReviewByUserAndProduct(userID, productID int) (*models.Review, error) {
	var review models.Review
	query := `
		SELECT id, product_id, user_id, rating, title, comment, created_at, updated_at
		FROM reviews
		WHERE user_id = $1 AND product_id = $2`

	err := r.DB.QueryRow(query, userID, productID).Scan(
		&review.ID, &review.ProductID, &review.UserID, &review.Rating,
		&review.Title, &review.Comment, &review.CreatedAt, &review.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No review found is not an error
	}
	if err != nil {
		return nil, err
	}

	return &review, nil
}

func (r *postgresReviewRepo) UpdateReview(reviewID, rating int, title, comment string) error {
	query := `
		UPDATE reviews 
		SET rating = $1, title = $2, comment = $3, updated_at = NOW()
		WHERE id = $4`

	_, err := r.DB.Exec(query, rating, title, comment, reviewID)
	return err
}

func (r *postgresReviewRepo) DeleteReview(reviewID, userID int) error {
	// Ensure user can only delete their own review
	query := `DELETE FROM reviews WHERE id = $1 AND user_id = $2`
	result, err := r.DB.Exec(query, reviewID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("review not found or unauthorized")
	}

	return nil
}

func (r *postgresReviewRepo) GetProductRating(productID int) (*models.ProductRating, error) {
	rating := &models.ProductRating{
		ProductID:    productID,
		RatingCounts: make(map[int]int),
	}

	// Get average rating and total count
	query := `
		SELECT COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as total
		FROM reviews
		WHERE product_id = $1`

	err := r.DB.QueryRow(query, productID).Scan(&rating.AverageRating, &rating.TotalReviews)
	if err != nil {
		return nil, err
	}

	// Get count per star rating
	countQuery := `
		SELECT rating, COUNT(*) as count
		FROM reviews
		WHERE product_id = $1
		GROUP BY rating
		ORDER BY rating DESC`

	rows, err := r.DB.Query(countQuery, productID)
	if err != nil {
		return rating, nil // Return partial data on error
	}
	defer rows.Close()

	for rows.Next() {
		var star, count int
		if err := rows.Scan(&star, &count); err != nil {
			continue
		}
		rating.RatingCounts[star] = count
	}

	return rating, nil
}
