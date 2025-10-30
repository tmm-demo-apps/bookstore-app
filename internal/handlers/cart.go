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
	Product    models.Product
}

type CartViewData struct {
	IsAuthenticated bool
	Items           []CartItemView
	Total           float64
}

func (h *Handlers) AddToCart(w http.ResponseWriter, r *http.Request) {
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
		ts.ExecuteTemplate(w, "cart.html", nil)
		return
	}

	rows, err := h.DB.Query(`
		SELECT ci.id, p.name, p.description, p.price
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
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
		if err := rows.Scan(&item.CartItemID, &item.Product.Name, &item.Product.Description, &item.Product.Price); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		items = append(items, item)
		total += item.Product.Price
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
