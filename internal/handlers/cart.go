package handlers

import (
	"DemoApp/internal/models"
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
	Book       models.Book
}

type CartViewData struct {
	IsAuthenticated bool
	Items           []CartItemView
	Total           float64
}

func (h *Handlers) AddToCart(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "cart-session")
	if session.Values["id"] == nil {
		session.Values["id"] = uuid.New().String()
	}
	session.Save(r, w)

	bookID, err := strconv.Atoi(r.FormValue("book_id"))
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec("INSERT INTO cart_items (session_id, book_id, quantity) VALUES ($1, $2, $3)",
		session.Values["id"], bookID, 1)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	w.Header().Set("HX-Trigger", "cart-updated")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
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
	session, _ := h.Store.Get(r, "cart-session")
	sessionID := session.Values["id"]
	if sessionID == nil {
		ts, _ := template.ParseFiles("./templates/base.html", "./templates/cart.html")
		ts.Execute(w, nil)
		return
	}

	rows, err := h.DB.Query(`
		SELECT ci.id, b.title, b.author, b.price
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
		if err := rows.Scan(&item.CartItemID, &item.Book.Title, &item.Book.Author, &item.Book.Price); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		items = append(items, item)
		total += item.Book.Price
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

	ts.Execute(w, data)
}
