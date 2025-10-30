package handlers

import (
	"DemoApp/internal/models"
	"html/template"
	"net/http"
	"strings"
	"unicode"
)

func (h *Handlers) SignupPage(w http.ResponseWriter, r *http.Request) {
	ts, _ := template.ParseFiles("./templates/base.html", "./templates/signup.html")
	ts.Execute(w, nil)
}

func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	// --- Validation ---
	if !strings.Contains(email, "@") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	var (
		hasMinLen = len(password) >= 8
		hasNumber = false
		hasLetter = false
	)
	for _, char := range password {
		switch {
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsLetter(char):
			hasLetter = true
		}
	}
	if !hasMinLen || !hasNumber || !hasLetter {
		http.Error(w, "Password does not meet requirements", http.StatusBadRequest)
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

	// Log the user in automatically
	session, _ := h.Store.Get(r, "cart-session")
	session.Values["user_id"] = user.ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	error := r.URL.Query().Get("error")
	data := struct{ Error bool }{Error: error != ""}
	ts, _ := template.ParseFiles("./templates/base.html", "./templates/login.html")
	ts.Execute(w, data)
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "cart-session")
	email := r.FormValue("email")
	password := r.FormValue("password")

	var user models.User
	err := h.DB.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", email).Scan(&user.ID, &user.PasswordHash)
	if err != nil || !user.CheckPassword(password) {
		http.Redirect(w, r, "/login?error=true", http.StatusSeeOther)
		return
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
