package handlers

import (
	"DemoApp/internal/models"
	"github.com/google/uuid"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type CartViewData struct {
	IsAuthenticated bool
	Items           []models.CartItem
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

	// Create session ID if neither user nor session exists
	if !userOk && !sessionOk {
		sessionID = uuid.New().String()
		session.Values["id"] = sessionID
		if err := session.Save(r, w); err != nil {
			log.Printf("Error saving session: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		sessionOk = true // Mark as valid now
	}

	productID, err := strconv.Atoi(r.FormValue("product_id"))
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	quantity := 1
	if qtyStr := r.FormValue("quantity"); qtyStr != "" {
		qty, err := strconv.Atoi(qtyStr)
		if err == nil && qty >= 1 && qty <= 99 {
			quantity = qty
		}
	}

	// Pass 0 for userID if not authenticated, empty string for sessionID if not set
	if !userOk {
		userID = 0
	}
	if !sessionOk {
		sessionID = ""
	}

	err = h.Repo.Cart().AddToCart(userID, sessionID, productID, quantity)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	w.Header().Set("HX-Trigger", "cart-updated")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) UpdateCartQuantity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Note: Front-end sends cart_item_id, but my new repo works with product_id.
	// I need to fetch the product_id from cart_item_id first, OR update repo to work with cart_item_id.
	// The original handler did: SELECT product_id ... FROM cart_items WHERE id = cart_item_id

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

	// Helper to get context from cart_item_id
	// Since I moved this logic to Repo, I should expose a method GetCartItemByID?
	// Or just run a quick query here?
	// The plan was to move SQL out of handlers.
	// But `UpdateQuantity` in Repo takes `productID`.
	// Let's cheating slightly: I will add `GetCartItemByID` to repo, or use a raw query if I must?
	// No, I should do it right.
	// The original handler logic was:
	// 1. Get product_id, user_id, session_id from cart_item_id
	// 2. Check consistency
	// 3. Update

	// Implementation detail: Since I don't have GetCartItemByID in interface yet,
	// and I cannot change interface easily without editing multiple files...
	// Wait, I CAN change the interface. It's in `internal/repository/repository.go`.
	// But for now, let's assume I can just query the DB via a new method or just fix the frontend to pass product_id?
	// Fixing frontend is harder (template change).
	// Let's add GetCartItem to repo interface.

	// Actually, for now, let's query the DB directly? No, `h.DB` is gone.
	// I MUST add it to the Repo.

	// Let's assume I will add `GetCartItem(id int) (*models.CartItem, error)` to CartRepository.
	// I'll update `repository.go` and `postgres.go` in a moment.

	item, err := h.Repo.Cart().GetCartItem(cartItemID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Cart item not found", http.StatusNotFound)
		return
	}

	// Use the ID from the item to call UpdateQuantity
	// Note: UpdateQuantity takes userID/sessionID.
	// The item struct has UserID/SessionID pointers.

	uID := 0
	if item.UserID != nil {
		uID = *item.UserID
	}
	sID := ""
	if item.SessionID != nil {
		sID = *item.SessionID
	}

	err = h.Repo.Cart().UpdateQuantity(uID, sID, item.ProductID, quantity)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	w.Header().Set("HX-Trigger", "cart-updated")
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	cartItemID, err := strconv.Atoi(r.FormValue("cart_item_id"))
	if err != nil {
		http.Error(w, "Invalid cart item ID", http.StatusBadRequest)
		return
	}

	item, err := h.Repo.Cart().GetCartItem(cartItemID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Cart item not found", http.StatusNotFound)
		return
	}

	uID := 0
	if item.UserID != nil {
		uID = *item.UserID
	}
	sID := ""
	if item.SessionID != nil {
		sID = *item.SessionID
	}

	err = h.Repo.Cart().RemoveItem(uID, sID, item.ProductID)
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
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	session, _ := h.Store.Get(r, "cart-session")
	userID, userOk := session.Values["user_id"].(int)
	sessionID, sessionOk := session.Values["id"].(string)

	if !userOk {
		userID = 0
	}
	if !sessionOk {
		sessionID = ""
	}

	if userID == 0 && sessionID == "" {
		// Empty
		data := CartViewData{
			IsAuthenticated: h.IsAuthenticated(r),
			Items:           nil,
			Total:           0,
		}
		ts, err := template.ParseFiles("./templates/base.html", "./templates/cart.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if err := ts.ExecuteTemplate(w, "cart.html", data); err != nil {
			log.Printf("Error executing template: %v", err)
		}
		return
	}

	items, total, err := h.Repo.Cart().GetCartItems(userID, sessionID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
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

	if err := ts.ExecuteTemplate(w, "cart.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
