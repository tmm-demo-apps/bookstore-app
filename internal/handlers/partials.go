package handlers

import (
	"DemoApp/internal/models"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func (h *Handlers) CartCount(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "cart-session")
	sessionID, ok := session.Values["id"].(string)
	if !ok || sessionID == "" {
		fmt.Fprint(w, "(0)")
		return
	}

	var count int
	err := h.DB.QueryRow("SELECT COUNT(id) FROM cart_items WHERE session_id = $1", sessionID).Scan(&count)
	if err != nil {
		fmt.Fprint(w, "(0)") // Gracefully degrade
		return
	}

	fmt.Fprintf(w, "(%d)", count)
}

func (h *Handlers) CartSummary(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "cart-session")
	sessionID, ok := session.Values["id"].(string)
	if !ok || sessionID == "" {
		return
	}

	rows, err := h.DB.Query(`
		SELECT b.title, b.price
		FROM cart_items ci
		JOIN books b ON ci.book_id = b.id
		WHERE ci.session_id = $1`, sessionID)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	var items []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.Title, &book.Price); err != nil {
			log.Println(err)
			return
		}
		items = append(items, book)
	}

	ts, err := template.ParseFiles("./templates/partials/cart-summary.html")
	if err != nil {
		log.Println(err)
		return
	}
	ts.Execute(w, items)
}
