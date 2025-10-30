package handlers

import (
	"DemoApp/internal/models"
	"html/template"
	"log"
	"net/http"
)

type ProductListViewData struct {
	IsAuthenticated bool
	Products        []models.Product
}

func (h *Handlers) ListProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, name, description, price FROM products")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		products = append(products, p)
	}

	data := ProductListViewData{
		IsAuthenticated: h.IsAuthenticated(r),
		Products:        products,
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/products.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	ts.ExecuteTemplate(w, "products.html", data)
}
