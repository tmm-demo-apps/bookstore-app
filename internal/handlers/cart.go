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

	if userOk {
		_, err := h.DB.Exec("INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, 1)", userID, productID)
		if err != nil { http.Error(w, "Internal Server Error", 500); return }
	} else {
		_, err = h.DB.Exec("INSERT INTO cart_items (session_id, product_id, quantity) VALUES ($1, $2, 1)", sessionID, productID)
		if err != nil { http.Error(w, "Internal Server Error", 500); return }
	}

	w.Header().Set("HX-Trigger", "cart-updated")
	w.WriteHeader(http.StatusNoContent)
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

	_, err = h.DB.Exec("DELETE FROM cart_items WHERE id = $1", cartItemID)
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
