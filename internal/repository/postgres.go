package repository

import (
	"DemoApp/internal/models"
	"database/sql"
	"fmt"
)

type PostgresRepository struct {
	DB *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{DB: db}
}

func (r *PostgresRepository) Products() ProductRepository {
	return &postgresProductRepo{DB: r.DB}
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

// --- Product Implementation ---

type postgresProductRepo struct {
	DB *sql.DB
}

func (r *postgresProductRepo) ListProducts() ([]models.Product, error) {
	// Updated query for new schema
	query := `SELECT id, name, description, price, sku, stock_quantity, image_url, category_id, status 
	          FROM products WHERE status = 'active' ORDER BY name`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.StockQuantity, &p.ImageURL, &p.CategoryID, &p.Status); err != nil {
			// Handle case where columns might be NULL if left joined (though here we query products directly)
			// The pointer types in struct handle NULLs automatically via Scan if valid.
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *postgresProductRepo) GetProductByID(id int) (*models.Product, error) {
	var p models.Product
	query := `SELECT id, name, description, price, sku, stock_quantity, image_url, category_id, status 
	          FROM products WHERE id = $1`
	err := r.DB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.StockQuantity, &p.ImageURL, &p.CategoryID, &p.Status)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *postgresProductRepo) SearchProducts(query string, categoryID int) ([]models.Product, error) {
	q := `SELECT id, name, description, price, sku, stock_quantity, image_url, category_id, status 
	      FROM products WHERE status = 'active'`
	var args []interface{}
	argID := 1

	if query != "" {
		q += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argID, argID)
		args = append(args, "%"+query+"%")
		argID++
	}

	if categoryID > 0 {
		q += fmt.Sprintf(" AND category_id = $%d", argID)
		args = append(args, categoryID)
		argID++
	}

	rows, err := r.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.SKU, &p.StockQuantity, &p.ImageURL, &p.CategoryID, &p.Status); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
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
		tx.Rollback()
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
		tx.Rollback()
		return 0, err
	}

	// Update Order Total
	_, err = tx.Exec(`
		UPDATE orders 
		SET total_amount = (SELECT SUM(price * quantity) FROM order_items WHERE order_id = $1)
		WHERE id = $1`, orderID)
	if err != nil {
		tx.Rollback()
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
		tx.Rollback()
		return 0, err
	}

	// Clear Cart
	if userID > 0 {
		_, err = tx.Exec("DELETE FROM cart_items WHERE user_id = $1", userID)
	} else {
		_, err = tx.Exec("DELETE FROM cart_items WHERE session_id = $1", sessionID)
	}
	if err != nil {
		tx.Rollback()
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
			SELECT ci.id, ci.product_id, p.name, p.description, p.price, ci.quantity
			FROM cart_items ci
			JOIN products p ON ci.product_id = p.id
			WHERE ci.user_id = $1
			ORDER BY p.name`, userID)
	} else {
		rows, err = r.DB.Query(`
			SELECT ci.id, ci.product_id, p.name, p.description, p.price, ci.quantity
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
		if err := rows.Scan(&item.ID, &item.ProductID, &p.Name, &p.Description, &p.Price, &item.Quantity); err != nil {
			return nil, 0, err
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
	if err != nil { return err }

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
			tx.Rollback()
			return err
		}

		_, err = tx.Exec("DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, i.PID)
		if err != nil {
			tx.Rollback()
			return err
		}

		newQty := userQty + i.Qty
		if newQty > 99 { newQty = 99 }
		_, err = tx.Exec("INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3)", userID, i.PID, newQty)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	_, err = tx.Exec("DELETE FROM cart_items WHERE session_id = $1", sessionID)
	if err != nil {
		tx.Rollback()
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
	err := r.DB.QueryRow("SELECT id, email, password_hash, full_name, role FROM users WHERE id = $1", id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
