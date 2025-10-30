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
	session, _ := h.Store.Get(r, "cart-session")
	sessionID := session.Values["id"]
	if sessionID == nil {
		http.Redirect(w, r, "/cart", http.StatusFound)
		return
	}

	rows, err := h.DB.Query(`
		SELECT ci.id, b.title, b.author, b.price, b.id
		FROM cart_items ci
		JOIN books b ON ci.book_id = b.id
		WHERE ci.session_id = $1`, sessionID)
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
		if err := rows.Scan(&item.CartItemID, &item.Book.Title, &item.Book.Author, &item.Book.Price, &item.Book.ID); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		items = append(items, item)
		total += item.Book.Price
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

	ts.Execute(w, data)
}

func (h *Handlers) ProcessOrder(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "cart-session")
	sessionID, ok := session.Values["id"].(string)
	if !ok || sessionID == "" {
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

	_, err = tx.Exec(`
		INSERT INTO order_items (order_id, book_id, quantity, price)
		SELECT $1, book_id, quantity, (SELECT price FROM books WHERE id = book_id)
		FROM cart_items WHERE session_id = $2`, orderID, sessionID)
	if err != nil {
		tx.Rollback()
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	_, err = tx.Exec("DELETE FROM cart_items WHERE session_id = $1", sessionID)
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
	ts.Execute(w, data)
}
