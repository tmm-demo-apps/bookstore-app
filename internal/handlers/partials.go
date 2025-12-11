package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type CartSummaryItem struct {
	Name     string
	Price    float64
	Quantity int
}

func (h *Handlers) CartCount(w http.ResponseWriter, r *http.Request) {
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
		fmt.Fprint(w, "(0)")
		return
	}

	items, _, err := h.Repo.Cart().GetCartItems(userID, sessionID)
	if err != nil {
		fmt.Fprint(w, "(0)")
		return
	}

	count := 0
	for _, item := range items {
		count += item.Quantity
	}

	fmt.Fprintf(w, "(%d)", count)
}

func (h *Handlers) CartSummary(w http.ResponseWriter, r *http.Request) {
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
		ts, err := template.ParseFiles("./templates/partials/cart-summary.html")
		if err != nil {
			log.Println(err)
			return
		}
		if err := ts.Execute(w, nil); err != nil {
			log.Printf("Error executing template: %v", err)
		}
		return
	}

	items, _, err := h.Repo.Cart().GetCartItems(userID, sessionID)
	if err != nil {
		log.Println(err)
		return
	}

	var summaryItems []CartSummaryItem
	for _, item := range items {
		summaryItems = append(summaryItems, CartSummaryItem{
			Name:     item.Product.Name,
			Price:    item.Product.Price,
			Quantity: item.Quantity,
		})
	}

	ts, err := template.ParseFiles("./templates/partials/cart-summary.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err := ts.Execute(w, summaryItems); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}
