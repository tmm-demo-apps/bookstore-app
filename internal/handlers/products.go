package handlers

import (
	"DemoApp/internal/models"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type ProductListViewData struct {
	IsAuthenticated bool
	Products        []models.Product
	SearchQuery     string
}

type ProductDetailViewData struct {
	IsAuthenticated bool
	Product         models.Product
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

func (h *Handlers) ProductDetail(w http.ResponseWriter, r *http.Request) {
	// Extract product ID from URL path
	idStr := r.PathValue("id")
	productID, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println("Invalid product ID:", err)
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Fetch product from database
	product, err := h.Repo.Products().GetProductByID(productID)
	if err != nil {
		log.Println("Product not found:", err)
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	data := ProductDetailViewData{
		IsAuthenticated: h.IsAuthenticated(r),
		Product:         *product,
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/product-detail.html")
	if err != nil {
		log.Println("Template parsing error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = ts.ExecuteTemplate(w, "product-detail.html", data)
	if err != nil {
		log.Println("Template execution error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
