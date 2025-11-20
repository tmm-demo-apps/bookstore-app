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
	SearchQuery     string
}

func (h *Handlers) ListProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	
	var products []models.Product
	var err error

	if query != "" {
		products, err = h.Repo.Products().SearchProducts(query, 0)
	} else {
		products, err = h.Repo.Products().ListProducts()
	}

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := ProductListViewData{
		IsAuthenticated: h.IsAuthenticated(r),
		Products:        products,
		SearchQuery:     query,
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/products.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	ts.ExecuteTemplate(w, "products.html", data)
}
