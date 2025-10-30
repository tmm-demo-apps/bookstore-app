package handlers

import (
	"DemoApp/internal/models"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"unicode"
)

type LoginPageData struct {
	IsAuthenticated bool
	Error           string
}

type SignupPageData struct {
	IsAuthenticated bool
	PasswordHelp    string
	Error           string
}

func (h *Handlers) SignupPage(w http.ResponseWriter, r *http.Request) {
	data := SignupPageData{
		IsAuthenticated: false,
		PasswordHelp:    "Password must be at least 8 characters long and contain at least one letter and one number.",
	}
	ts, _ := template.ParseFiles("./templates/base.html", "./templates/signup.html")
	ts.Execute(w, data)
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
	session.Values["user_id"] = userID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
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
		IsAuthenticated: false,
		Error:           errorMsg,
	}
	ts, _ := template.ParseFiles("./templates/base.html", "./templates/login.html")
	ts.Execute(w, data)
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

	// Associate guest cart with user account
	if sessionID, ok := session.Values["id"].(string); ok {
		h.DB.Exec("UPDATE cart_items SET user_id = $1, session_id = NULL WHERE session_id = $2", user.ID, sessionID)
		delete(session.Values, "id") // Remove anonymous session ID
	}

	session.Values["user_id"] = user.ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "cart-session")
	delete(session.Values, "user_id")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}
