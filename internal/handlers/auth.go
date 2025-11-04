package handlers

import (
	"DemoApp/internal/models"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"unicode"
)

type LoginPageData struct {
	IsAuthenticated bool
	Error           string
	Next            string
}

type SignupPageData struct {
	IsAuthenticated bool
	PasswordHelp    string
	Error           string
	Next            string
}

func (h *Handlers) SignupPage(w http.ResponseWriter, r *http.Request) {
	data := SignupPageData{
		IsAuthenticated: h.IsAuthenticated(r),
		PasswordHelp:    "Password must be at least 8 characters long and contain at least one letter and one number.",
		Next:            r.URL.Query().Get("next"),
	}
	ts, _ := template.ParseFiles("./templates/base.html", "./templates/signup.html")
	ts.ExecuteTemplate(w, "signup.html", data)
}

func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	// --- Validation ---
	if !strings.Contains(email, "@") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if err := validatePassword(password); err != nil {
		// In a real app, you'd render the page again with the error
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// --- End Validation ---

	var user models.User
	err := user.SetPassword(password)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = h.DB.QueryRow("INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id", email, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		http.Error(w, "Could not create user", 500)
		return
	}

	// Get the ID of the new user
	var userID int
	err = h.DB.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		// User was created, but we can't log them in. Redirect to login.
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Log the user in by setting the session
	session, _ := h.Store.Get(r, "cart-session")
	
	// Merge anonymous cart with new user account
	if sessionID, ok := session.Values["id"].(string); ok && sessionID != "" {
		err := h.mergeAnonymousCart(sessionID, userID)
		if err != nil {
			log.Printf("Error merging cart during signup: %v", err)
			// Continue with signup even if cart merge fails
		}
		delete(session.Values, "id") // Remove anonymous session ID
	}
	
	session.Values["user_id"] = userID
	session.Save(r, w)

	nextURL := r.URL.Query().Get("next")
	if nextURL != "" {
		http.Redirect(w, r, nextURL, http.StatusFound)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// mergeAnonymousCart merges items from an anonymous session cart into the user's cart
// This handles the case where a user adds items while not logged in, then logs in
func (h *Handlers) mergeAnonymousCart(sessionID string, userID int) error {
	// First, get all products from anonymous cart (outside transaction)
	type cartProduct struct {
		ProductID int
		Quantity  int
	}
	
	rows, err := h.DB.Query(`
		SELECT product_id, SUM(quantity) as total_quantity
		FROM cart_items
		WHERE session_id = $1
		GROUP BY product_id`, sessionID)
	if err != nil {
		return err
	}
	
	var anonymousProducts []cartProduct
	for rows.Next() {
		var cp cartProduct
		if err := rows.Scan(&cp.ProductID, &cp.Quantity); err != nil {
			rows.Close()
			return err
		}
		anonymousProducts = append(anonymousProducts, cp)
	}
	rows.Close()
	
	// If no anonymous cart items, nothing to merge
	if len(anonymousProducts) == 0 {
		return nil
	}

	// Now process each product in a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, ap := range anonymousProducts {
		// Check if user already has this product in their cart
		var existingQty int
		err = tx.QueryRow(`
			SELECT COALESCE(SUM(quantity), 0) 
			FROM cart_items 
			WHERE user_id = $1 AND product_id = $2`, userID, ap.ProductID).Scan(&existingQty)
		if err != nil {
			return err
		}

		// Delete all existing rows for this product+user combination
		_, err = tx.Exec("DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2", userID, ap.ProductID)
		if err != nil {
			return err
		}

		// Calculate new quantity (merge quantities, cap at 99)
		newQty := existingQty + ap.Quantity
		if newQty > 99 {
			newQty = 99
		}

		// Insert single consolidated row
		_, err = tx.Exec("INSERT INTO cart_items (user_id, product_id, quantity) VALUES ($1, $2, $3)", 
			userID, ap.ProductID, newQty)
		if err != nil {
			return err
		}
	}

	// Delete all items from the anonymous cart
	_, err = tx.Exec("DELETE FROM cart_items WHERE session_id = $1", sessionID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func validatePassword(password string) error {
	var hasLetter, hasNumber bool
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	for _, char := range password {
		switch {
		case unicode.IsLetter(char):
			hasLetter = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}
	if !hasLetter || !hasNumber {
		return fmt.Errorf("password must contain at least one letter and one number")
	}
	return nil
}

func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request, errorMsg string) {
	data := LoginPageData{
		IsAuthenticated: h.IsAuthenticated(r),
		Error:           errorMsg,
		Next:            r.URL.Query().Get("next"),
	}
	ts, _ := template.ParseFiles("./templates/base.html", "./templates/login.html")
	ts.ExecuteTemplate(w, "login.html", data)
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "cart-session")
	email := r.FormValue("email")
	password := r.FormValue("password")

	var user models.User
	err := h.DB.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", email).Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		h.LoginPage(w, r, "Incorrect email address or password. Please verify they are correct or create an account if a new customer.")
		return
	}

	if !user.CheckPassword(password) {
		h.LoginPage(w, r, "Incorrect email address or password. Please verify they are correct or create an account if a new customer.")
		return
	}

	// Merge anonymous cart with user account
	if sessionID, ok := session.Values["id"].(string); ok && sessionID != "" {
		err := h.mergeAnonymousCart(sessionID, user.ID)
		if err != nil {
			log.Printf("Error merging cart: %v", err)
			// Continue with login even if cart merge fails
		}
		delete(session.Values, "id") // Remove anonymous session ID
	}

	session.Values["user_id"] = user.ID
	session.Save(r, w)

	nextURL := r.URL.Query().Get("next")
	if nextURL != "" {
		http.Redirect(w, r, nextURL, http.StatusFound)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "cart-session")
	delete(session.Values, "user_id")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}
