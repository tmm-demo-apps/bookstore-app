package handlers

import (
	"DemoApp/internal/models"
	"html/template"
	"log"
	"net/http"
)

type MyOrdersViewData struct {
	IsAuthenticated   bool
	ReaderBrowserURL  string
	ChatbotBrowserURL string
	Orders            []models.Order
}

func (h *Handlers) MyOrders(w http.ResponseWriter, r *http.Request) {
	if !h.IsAuthenticated(r) {
		http.Redirect(w, r, "/login?next=/orders", http.StatusFound)
		return
	}

	session, _ := h.Store.Get(r, "cart-session")
	userID, ok := session.Values["user_id"].(int)
	if !ok || userID == 0 {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	orders, err := h.Repo.Orders().GetOrdersByUserID(userID)
	if err != nil {
		log.Printf("Error fetching orders for user %d: %v", userID, err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	log.Printf("Found %d orders for user %d", len(orders), userID)

	data := MyOrdersViewData{
		IsAuthenticated:   h.IsAuthenticated(r),
		ReaderBrowserURL:  h.ReaderBrowserURL,
		ChatbotBrowserURL: h.ChatbotBrowserURL,
		Orders:            orders,
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/orders.html")
	if err != nil {
		log.Printf("Template parse error: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if err := ts.ExecuteTemplate(w, "orders.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
}
