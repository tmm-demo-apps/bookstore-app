package main

import (
	"DemoApp/internal/handlers"
	"DemoApp/internal/repository"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	esURL := os.Getenv("ES_URL")

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := sessions.NewCookieStore([]byte("something-very-secret"))
	// Configure session options for development (allow HTTP, not just HTTPS)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	repo := repository.NewPostgresRepository(db)

	// Initialize Elasticsearch if URL is provided
	if esURL != "" {
		log.Println("Initializing Elasticsearch...")
		es, err := repository.NewElasticsearchRepository([]string{esURL})
		if err != nil {
			log.Printf("Warning: Elasticsearch initialization failed: %v", err)
			log.Println("Continuing without Elasticsearch (will use SQL search)")
		} else {
			repo.SetElasticsearch(es)
			log.Println("Elasticsearch initialized successfully")

			// Index all products on startup
			go func() {
				log.Println("Indexing products to Elasticsearch...")
				products, err := repo.Products().ListProducts()
				if err != nil {
					log.Printf("Error listing products for indexing: %v", err)
					return
				}
				if err := es.IndexProducts(products); err != nil {
					log.Printf("Error indexing products: %v", err)
					return
				}
				log.Printf("Successfully indexed %d products", len(products))
			}()
		}
	} else {
		log.Println("ES_URL not set, using SQL-based search")
	}

	h := &handlers.Handlers{
		Repo:  repo,
		Store: store,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.ListProducts)
	mux.HandleFunc("/products/{id}", h.ProductDetail)
	mux.HandleFunc("/cart/add", h.AddToCart)
	mux.HandleFunc("/cart/update", h.UpdateCartQuantity)
	mux.HandleFunc("/cart/remove", h.RemoveFromCart)
	mux.HandleFunc("/cart", h.ViewCart)
	mux.HandleFunc("/checkout", h.CheckoutPage)
	mux.HandleFunc("/checkout/process", h.ProcessOrder)
	mux.HandleFunc("/confirmation", h.ConfirmationPage)
	mux.HandleFunc("/partials/cart-count", h.CartCount)
	mux.HandleFunc("/partials/cart-summary", h.CartSummary)
	mux.HandleFunc("/partials/search-suggestions", h.SearchSuggestions)

	mux.HandleFunc("/signup", h.SignupPage)
	mux.HandleFunc("/signup/process", h.Signup)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		h.LoginPage(w, r, "")
	})
	mux.HandleFunc("/login/process", h.Login)
	mux.HandleFunc("/logout", h.Logout)
	mux.HandleFunc("/orders", h.MyOrders)

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}
