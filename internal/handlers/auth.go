package handlers

import (
	"DemoApp/internal/models"
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

func (h *Handlers) SignupPage(w http.ResponseWriter, r *http.Request) {
	ts, _ := template.ParseFiles("./templates/base.html", "./templates/signup.html")
	ts.Execute(w, nil)
}

func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	var user models.User
	err := user.SetPassword(password)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	_, err = h.DB.Exec("INSERT INTO users (email, password_hash) VALUES ($1, $2)", email, user.PasswordHash)
	if err != nil {
		http.Error(w, "Could not create user", 500)
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	ts, _ := template.ParseFiles("./templates/base.html", "./templates/login.html")
	ts.Execute(w, nil)
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "cart-session")
	email := r.FormValue("email")
	password := r.FormValue("password")

	var user models.User
	err := h.DB.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", email).Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if !user.CheckPassword(password) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
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
