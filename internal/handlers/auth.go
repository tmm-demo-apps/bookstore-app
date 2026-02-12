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
	IsAuthenticated   bool
	ReaderBrowserURL  string
	ChatbotBrowserURL string
	Error             string
	Next              string
}

type SignupPageData struct {
	IsAuthenticated   bool
	ReaderBrowserURL  string
	ChatbotBrowserURL string
	PasswordHelp      string
	Error             string
	Next              string
}

func (h *Handlers) SignupPage(w http.ResponseWriter, r *http.Request) {
	data := SignupPageData{
		IsAuthenticated:   h.IsAuthenticated(r),
		ReaderBrowserURL:  h.ReaderBrowserURL,
		ChatbotBrowserURL: h.ChatbotBrowserURL,
		PasswordHelp:      "Password must be at least 8 characters long and contain at least one letter and one number.",
		Next:              r.URL.Query().Get("next"),
	}
	ts, err := template.ParseFiles("./templates/base.html", "./templates/signup.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := ts.ExecuteTemplate(w, "signup.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if !strings.Contains(email, "@") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if err := validatePassword(password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user models.User
	err := user.SetPassword(password)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Use Repository to Create User
	// Note: CreateUser signature is (email, passwordHash, fullName)
	// We don't have Full Name yet from the form, passing empty string or email prefix as placeholder
	// Actually the plan says to add full name to schema later.
	// For now, I will pass email as full name placeholder or empty string.

	userID, err := h.Repo.Users().CreateUser(email, user.PasswordHash, "")
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not create user", 500)
		return
	}

	session, _ := h.Store.Get(r, "cart-session")

	if sessionID, ok := session.Values["id"].(string); ok && sessionID != "" {
		err := h.Repo.Cart().MergeCart(sessionID, userID)
		if err != nil {
			log.Printf("Error merging cart during signup: %v", err)
		}
		delete(session.Values, "id")
	}

	session.Values["user_id"] = userID
	if err := session.Save(r, w); err != nil {
		log.Printf("Error saving session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	nextURL := r.URL.Query().Get("next")
	if nextURL != "" {
		http.Redirect(w, r, nextURL, http.StatusFound)
		return
	}

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
		IsAuthenticated:   h.IsAuthenticated(r),
		ReaderBrowserURL:  h.ReaderBrowserURL,
		ChatbotBrowserURL: h.ChatbotBrowserURL,
		Error:             errorMsg,
		Next:              r.URL.Query().Get("next"),
	}
	ts, err := template.ParseFiles("./templates/base.html", "./templates/login.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := ts.ExecuteTemplate(w, "login.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "cart-session")
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.Repo.Users().GetUserByEmail(email)
	if err != nil || user == nil {
		h.LoginPage(w, r, "Incorrect email address or password. Please verify they are correct or create an account if a new customer.")
		return
	}

	if !user.CheckPassword(password) {
		h.LoginPage(w, r, "Incorrect email address or password. Please verify they are correct or create an account if a new customer.")
		return
	}

	if sessionID, ok := session.Values["id"].(string); ok && sessionID != "" {
		err := h.Repo.Cart().MergeCart(sessionID, user.ID)
		if err != nil {
			log.Printf("Error merging cart: %v", err)
		}
		delete(session.Values, "id")
	}

	session.Values["user_id"] = user.ID
	if err := session.Save(r, w); err != nil {
		log.Printf("Error saving session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	nextURL := r.URL.Query().Get("next")
	if nextURL != "" {
		http.Redirect(w, r, nextURL, http.StatusFound)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := h.Store.Get(r, "cart-session")
	if err != nil {
		log.Printf("Error getting session: %v", err)
	}
	delete(session.Values, "user_id")
	if err := session.Save(r, w); err != nil {
		log.Printf("Error saving session: %v", err)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
