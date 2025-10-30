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
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)
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

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}
