package handlers

import (
	"DemoApp/internal/models"
	"html/template"
	"log"
	"net/http"
)

type MyOrdersViewData struct {
	IsAuthenticated bool
	Orders          []models.Order
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
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := MyOrdersViewData{
		IsAuthenticated: h.IsAuthenticated(r),
		Orders:          orders,
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/orders.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	ts.ExecuteTemplate(w, "orders.html", data)
}

