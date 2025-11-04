package handlers

import (
	"DemoApp/internal/models"
	"database/sql"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

// CartItemView is used to pass cart item data to the template.
type CartItemView struct {
	CartItemID int
	Product    models.Product
	Quantity   int
	Subtotal   float64
}

type CartViewData struct {
	IsAuthenticated bool
	Items           []CartItemView
	Total           float64
}

func (h *Handlers) AddToCart(w http.ResponseWriter, r *http.Request) {
	// Prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	session, _ := h.Store.Get(r, "cart-session")

	userID, userOk := session.Values["user_id"].(int)
	sessionID, sessionOk := session.Values["id"].(string)

	if !userOk && !sessionOk {
		session.Values["id"] = uuid.New().String()
		session.Save(r, w)
		sessionID = session.Values["id"].(string)
	}

	productID, err := strconv.Atoi(r.FormValue("product_id"))
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Get quantity from form, default to 1 if not provided
	quantity := 1
	if qtyStr := r.FormValue("quantity"); qtyStr != "" {
		qty, err := strconv.Atoi(qtyStr)
		if err == nil && qty >= 1 && qty <= 99 {
			quantity = qty
		}
	}

	if userOk {
		// Check if item already exists for this user and product
		var existingQty int
		err := h.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, productID).Scan(&existingQty)
		
		if err == nil && existingQty > 0 {
			// Update existing item(s) - first, consolidate all duplicates into one
			_, err = h.DB.Exec("DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, productID)
			if err != nil { 
				log.Println(err)
				http.Error(w, "Internal Server Error", 500)
				return 
			}
			// Insert single consolidated row with updated quantity
			newQty := existingQty + quantity
			if newQty > 99 { newQty = 99 }
			_, err = h.DB.Exec("INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3)", userID, productID, newQty)
			if err != nil { 
				log.Println(err)
				http.Error(w, "Internal Server Error", 500)
				return 
			}
		} else {
			// Insert new item
			_, err = h.DB.Exec("INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3)", userID, productID, quantity)
			if err != nil { 
				log.Println(err)
				http.Error(w, "Internal Server Error", 500)
				return 
			}
		}
	} else {
		// Check if item already exists for this session and product
		var existingQty int
		err = h.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE session_id = $1 AND product_id = $2", sessionID, productID).Scan(&existingQty)
		
		if err == nil && existingQty > 0 {
			// Update existing item(s) - first, consolidate all duplicates into one
			_, err = h.DB.Exec("DELETE FROM cart_items WHERE session_id = $1 AND product_id = $2", sessionID, productID)
			if err != nil { 
				log.Println(err)
				http.Error(w, "Internal Server Error", 500)
				return 
			}
			// Insert single consolidated row with updated quantity
			newQty := existingQty + quantity
			if newQty > 99 { newQty = 99 }
			_, err = h.DB.Exec("INSERT INTO cart_items (session_id, product_id, quantity) VALUES ($1, $2, $3)", sessionID, productID, newQty)
			if err != nil { 
				log.Println(err)
				http.Error(w, "Internal Server Error", 500)
				return 
			}
		} else {
			// Insert new item
			_, err = h.DB.Exec("INSERT INTO cart_items (session_id, product_id, quantity) VALUES ($1, $2, $3)", sessionID, productID, quantity)
			if err != nil { 
				log.Println(err)
				http.Error(w, "Internal Server Error", 500)
				return 
			}
		}
	}

	w.Header().Set("HX-Trigger", "cart-updated")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) UpdateCartQuantity(w http.ResponseWriter, r *http.Request) {
	// Prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	cartItemID, err := strconv.Atoi(r.FormValue("cart_item_id"))
	if err != nil {
		http.Error(w, "Invalid cart item ID", http.StatusBadRequest)
		return
	}

	quantity, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil || quantity < 1 || quantity > 99 {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	// Get the product_id and user/session info for this cart item
	var productID int
	var userID sql.NullInt64
	var sessionID sql.NullString
	err = h.DB.QueryRow("SELECT product_id, user_id, session_id FROM cart_items WHERE id = $1", cartItemID).Scan(&productID, &userID, &sessionID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Cart item not found", http.StatusNotFound)
		return
	}

	// Delete all duplicate rows for this product and user/session
	if userID.Valid {
		_, err = h.DB.Exec("DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2", userID.Int64, productID)
	} else if sessionID.Valid {
		_, err = h.DB.Exec("DELETE FROM cart_items WHERE session_id = $1 AND product_id = $2", sessionID.String, productID)
	} else {
		http.Error(w, "Invalid cart item", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Insert single consolidated row with new quantity
	if userID.Valid {
		_, err = h.DB.Exec("INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3)", userID.Int64, productID, quantity)
	} else {
		_, err = h.DB.Exec("INSERT INTO cart_items (session_id, product_id, quantity) VALUES ($1, $2, $3)", sessionID.String, productID, quantity)
	}

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	w.Header().Set("HX-Trigger", "cart-updated")
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	// Prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	cartItemID, err := strconv.Atoi(r.FormValue("cart_item_id"))
	if err != nil {
		http.Error(w, "Invalid cart item ID", http.StatusBadRequest)
		return
	}

	// Get the product_id and user/session info for this cart item
	var productID int
	var userID sql.NullInt64
	var sessionID sql.NullString
	err = h.DB.QueryRow("SELECT product_id, user_id, session_id FROM cart_items WHERE id = $1", cartItemID).Scan(&productID, &userID, &sessionID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Cart item not found", http.StatusNotFound)
		return
	}

	// Delete all duplicate rows for this product and user/session
	if userID.Valid {
		_, err = h.DB.Exec("DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2", userID.Int64, productID)
	} else if sessionID.Valid {
		_, err = h.DB.Exec("DELETE FROM cart_items WHERE session_id = $1 AND product_id = $2", sessionID.String, productID)
	} else {
		http.Error(w, "Invalid cart item", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	w.Header().Set("HX-Trigger", "cart-updated")
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) ViewCart(w http.ResponseWriter, r *http.Request) {
	// Prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

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
		// No session or user - show empty cart
		data := CartViewData{
			IsAuthenticated: h.IsAuthenticated(r),
			Items:           nil,
			Total:           0,
		}
		ts, _ := template.ParseFiles("./templates/base.html", "./templates/cart.html")
		ts.ExecuteTemplate(w, "cart.html", data)
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

	data := CartViewData{
		IsAuthenticated: h.IsAuthenticated(r),
		Items:           items,
		Total:           total,
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/cart.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	ts.ExecuteTemplate(w, "cart.html", data)
}
