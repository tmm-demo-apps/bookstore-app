package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

type CheckoutViewData struct {
	IsAuthenticated bool
	Items           []CartItemView
	Total           float64
}

func (h *Handlers) CheckoutPage(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Redirect(w, r, "/login?next=/checkout", http.StatusFound)
		return
	}

	session, _ := h.Store.Get(r, "cart-session")
	
	// Check if user is authenticated
	userID, userOk := session.Values["user_id"].(int)
	sessionID, sessionOk := session.Values["id"].(string)

	var rows *sql.Rows
	var err error

	if userOk && userID > 0 {
		// Query by user_id for authenticated users - group by product and sum quantities
		rows, err = h.DB.Query(`
			SELECT MIN(ci.id) as id, p.name, p.description, p.price, SUM(ci.quantity) as quantity
			FROM cart_items ci
			JOIN products p ON ci.product_id = p.id
			WHERE ci.user_id = $1
			GROUP BY p.id, p.name, p.description, p.price`, userID)
	} else if sessionOk && sessionID != "" {
		// Query by session_id for anonymous users - group by product and sum quantities
		rows, err = h.DB.Query(`
			SELECT MIN(ci.id) as id, p.name, p.description, p.price, SUM(ci.quantity) as quantity
			FROM cart_items ci
			JOIN products p ON ci.product_id = p.id
			WHERE ci.session_id = $1
			GROUP BY p.id, p.name, p.description, p.price`, sessionID)
	} else {
		// No cart items
		http.Redirect(w, r, "/cart", http.StatusFound)
		return
	}

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer rows.Close()

	var items []CartItemView
	var total float64
	for rows.Next() {
		var item CartItemView
		if err := rows.Scan(&item.CartItemID, &item.Product.Name, &item.Product.Description, &item.Product.Price, &item.Quantity); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		item.Subtotal = item.Product.Price * float64(item.Quantity)
		items = append(items, item)
		total += item.Subtotal
	}

	// Redirect to cart if empty
	if len(items) == 0 {
		http.Redirect(w, r, "/cart", http.StatusFound)
		return
	}

	data := CheckoutViewData{
		IsAuthenticated: h.IsAuthenticated(r),
		Items:           items,
		Total:           total,
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/checkout.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	ts.ExecuteTemplate(w, "checkout.html", data)
}

func (h *Handlers) ProcessOrder(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Redirect(w, r, "/login?next=/checkout", http.StatusFound)
		return
	}

	session, _ := h.Store.Get(r, "cart-session")
	
	// Check if user is authenticated
	userID, userOk := session.Values["user_id"].(int)
	sessionID, sessionOk := session.Values["id"].(string)

	if !userOk && !sessionOk {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	var orderID int
	err = tx.QueryRow("INSERT INTO orders (session_id) VALUES ($1) RETURNING id", sessionID).Scan(&orderID)
	if err != nil {
		tx.Rollback()
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Insert order items based on user or session
	if userOk && userID > 0 {
		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, quantity, price)
			SELECT $1, product_id, SUM(quantity), (SELECT price FROM products WHERE id = product_id)
			FROM cart_items 
			WHERE user_id = $2
			GROUP BY product_id`, orderID, userID)
	} else if sessionOk && sessionID != "" {
		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, quantity, price)
			SELECT $1, product_id, SUM(quantity), (SELECT price FROM products WHERE id = product_id)
			FROM cart_items 
			WHERE session_id = $2
			GROUP BY product_id`, orderID, sessionID)
	}

	if err != nil {
		tx.Rollback()
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Clear cart based on user or session
	if userOk && userID > 0 {
		_, err = tx.Exec("DELETE FROM cart_items WHERE user_id = $1", userID)
	} else if sessionOk && sessionID != "" {
		_, err = tx.Exec("DELETE FROM cart_items WHERE session_id = $1", sessionID)
	}

	if err != nil {
		tx.Rollback()
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	http.Redirect(w, r, "/confirmation", http.StatusFound)
}

type ConfirmationViewData struct {
	IsAuthenticated bool
}

func (h *Handlers) ConfirmationPage(w http.ResponseWriter, r *http.Request) {
	data := ConfirmationViewData{
		IsAuthenticated: h.IsAuthenticated(r),
	}
	ts, err := template.ParseFiles("./templates/base.html", "./templates/confirmation.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	ts.ExecuteTemplate(w, "confirmation.html", data)
}
