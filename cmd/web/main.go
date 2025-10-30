package main

import (
	"fmt"
	"log"
	"net/http"
	"DemoApp/internal/handlers"
	"database/sql"
	"os"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.ListBooks(db))
	mux.HandleFunc("/cart/add", handlers.AddToCart(db))
	mux.HandleFunc("/cart/remove", handlers.RemoveFromCart(db))
	mux.HandleFunc("/cart", handlers.ViewCart(db))
	mux.HandleFunc("/checkout", handlers.CheckoutPage(db))
	mux.HandleFunc("/checkout/process", handlers.ProcessOrder(db))
	mux.HandleFunc("/confirmation", handlers.ConfirmationPage())
	mux.HandleFunc("/partials/cart-count", handlers.CartCount(db))
	mux.HandleFunc("/partials/cart-summary", handlers.CartSummary(db))

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}
