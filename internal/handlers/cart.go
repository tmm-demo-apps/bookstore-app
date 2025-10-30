package handlers

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"DemoApp/internal/models"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

// CartItemView is used to pass cart item data to the template.
type CartItemView struct {
	CartItemID int
	Book       models.Book
}

func AddToCart(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "cart-session")
		if session.Values["id"] == nil {
			session.Values["id"] = uuid.New().String()
		}
		session.Save(r, w)

		bookID, err := strconv.Atoi(r.FormValue("book_id"))
		if err != nil {
			http.Error(w, "Invalid book ID", http.StatusBadRequest)
			return
		}

		// For simplicity, we add one item at a time.
		// A real implementation would check if the item is already in the cart and update the quantity.
		_, err = db.Exec("INSERT INTO cart_items (session_id, book_id, quantity) VALUES ($1, $2, $3)",
			session.Values["id"], bookID, 1)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func RemoveFromCart(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cartItemID, err := strconv.Atoi(r.FormValue("cart_item_id"))
		if err != nil {
			http.Error(w, "Invalid cart item ID", http.StatusBadRequest)
			return
		}

		// In a real application, you'd also check if this cart item
		// belongs to the current user's session before deleting.
		_, err = db.Exec("DELETE FROM cart_items WHERE id = $1", cartItemID)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}

		http.Redirect(w, r, "/cart", http.StatusFound)
	}
}

func ViewCart(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "cart-session")
		sessionID := session.Values["id"]
		if sessionID == nil {
			// No cart yet, show empty cart
			ts, _ := template.ParseFiles("./templates/cart.html")
			ts.Execute(w, nil)
			return
		}

		rows, err := db.Query(`
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
		for rows.Next() {
			var item CartItemView
			if err := rows.Scan(&item.CartItemID, &item.Book.Title, &item.Book.Author, &item.Book.Price); err != nil {
				log.Println(err)
				http.Error(w, "Internal Server Error", 500)
				return
			}
			items = append(items, item)
		}

		ts, err := template.ParseFiles("./templates/cart.html")
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}

		ts.Execute(w, items)
	}
}
