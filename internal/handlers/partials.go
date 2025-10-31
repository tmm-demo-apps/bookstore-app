package handlers

import (
	"DemoApp/internal/models"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func (h *Handlers) CartCount(w http.ResponseWriter, r *http.Request) {
	// Prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	session, _ := h.Store.Get(r, "cart-session")
	
	// Check if user is authenticated
	userID, userOk := session.Values["user_id"].(int)
	sessionID, sessionOk := session.Values["id"].(string)

	var count int
	var err error

	if userOk && userID > 0 {
		// Query by user_id for authenticated users
		err = h.DB.QueryRow("SELECT COUNT(id) FROM cart_items WHERE user_id = $1", userID).Scan(&count)
	} else if sessionOk && sessionID != "" {
		// Query by session_id for anonymous users
		err = h.DB.QueryRow("SELECT COUNT(id) FROM cart_items WHERE session_id = $1", sessionID).Scan(&count)
	} else {
		fmt.Fprint(w, "(0)")
		return
	}

	if err != nil {
		fmt.Fprint(w, "(0)") // Gracefully degrade
		return
	}

	fmt.Fprintf(w, "(%d)", count)
}

func (h *Handlers) CartSummary(w http.ResponseWriter, r *http.Request) {
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
		// Query by user_id for authenticated users
		rows, err = h.DB.Query(`
			SELECT p.name, p.price
			FROM cart_items ci
			JOIN products p ON ci.product_id = p.id
			WHERE ci.user_id = $1`, userID)
	} else if sessionOk && sessionID != "" {
		// Query by session_id for anonymous users
		rows, err = h.DB.Query(`
			SELECT p.name, p.price
			FROM cart_items ci
			JOIN products p ON ci.product_id = p.id
			WHERE ci.session_id = $1`, sessionID)
	} else {
		// No cart items for this user/session
		ts, err := template.ParseFiles("./templates/partials/cart-summary.html")
		if err != nil {
			log.Println(err)
			return
		}
		ts.Execute(w, nil)
		return
	}

	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	var items []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.Name, &p.Price); err != nil {
			log.Println(err)
			return
		}
		items = append(items, p)
	}

	ts, err := template.ParseFiles("./templates/partials/cart-summary.html")
	if err != nil {
		log.Println(err)
		return
	}
	ts.Execute(w, items)
}
