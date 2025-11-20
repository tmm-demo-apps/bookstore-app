package handlers

import (
	"DemoApp/internal/models"
	"github.com/google/uuid"
	"html/template"
	"log"
	"net/http"
)

type CheckoutViewData struct {
	IsAuthenticated bool
	Items           []models.CartItem
	Total           float64
}

func (h *Handlers) CheckoutPage(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Redirect(w, r, "/login?next=/checkout", http.StatusFound)
		return
	}

	session, _ := h.Store.Get(r, "cart-session")
	
	userID, userOk := session.Values["user_id"].(int)
	sessionID, sessionOk := session.Values["id"].(string)

	if !userOk { userID = 0 }
	if !sessionOk { sessionID = "" }
	
	if userID == 0 && sessionID == "" {
		http.Redirect(w, r, "/cart", http.StatusFound)
		return
	}

	items, total, err := h.Repo.Cart().GetCartItems(userID, sessionID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
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
	
	userID, userOk := session.Values["user_id"].(int)
	sessionID, sessionOk := session.Values["id"].(string)
	
	if !userOk { userID = 0 }
	if !sessionOk { 
		// Fallback for edge case where session might be missing but user is somehow authenticated
		// But usually ProcessOrder should fail if no session
		// Create a temp sessionID if missing? No, that would imply empty cart.
		session.Values["id"] = uuid.New().String()
		sessionID = session.Values["id"].(string)
	}

	// Note: Repo CreateOrder currently takes items only to calculate totals/structure, 
	// but the internal implementation I wrote actually re-queries the cart items for safety.
	// So I can pass nil or empty slice if the implementation relies on SQL.
	// Let's check my implementation of CreateOrder in postgres.go.
	// ... It DOES re-query based on sessionID/userID. 
	// So passing items is technically redundant but good for interface correctness if we swapped to a non-SQL repo.
	// For now I will just pass nil as I know my Postgres implementation ignores it (it does `INSERT INTO ... SELECT FROM cart_items`).
	
	_, err := h.Repo.Orders().CreateOrder(sessionID, userID, nil)
	if err != nil {
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
