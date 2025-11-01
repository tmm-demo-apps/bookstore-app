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

	store := sessions.NewCookieStore([]byte("something-very-secret"))

	h := &handlers.Handlers{
		DB:    db,
		Store: store,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.ListProducts)
	mux.HandleFunc("/cart/add", h.AddToCart)
	mux.HandleFunc("/cart/update", h.UpdateCartQuantity)
	mux.HandleFunc("/cart/remove", h.RemoveFromCart)
	mux.HandleFunc("/cart", h.ViewCart)
	mux.HandleFunc("/checkout", h.CheckoutPage)
	mux.HandleFunc("/checkout/process", h.ProcessOrder)
	mux.HandleFunc("/confirmation", h.ConfirmationPage)
	mux.HandleFunc("/partials/cart-count", h.CartCount)
	mux.HandleFunc("/partials/cart-summary", h.CartSummary)

	mux.HandleFunc("/signup", h.SignupPage)
	mux.HandleFunc("/signup/process", h.Signup)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		h.LoginPage(w, r, "")
	})
	mux.HandleFunc("/login/process", h.Login)
	mux.HandleFunc("/logout", h.Logout)

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}
