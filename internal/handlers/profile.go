package handlers

import (
	"DemoApp/internal/models"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type ProfileViewData struct {
	IsAuthenticated   bool
	ReaderBrowserURL  string
	ChatbotBrowserURL string
	User              *models.User
	OrderCount        int
	Error             string
	Success           string
}

// ProfilePage displays the user's profile
func (h *Handlers) ProfilePage(w http.ResponseWriter, r *http.Request) {
	userID, authenticated := h.GetUserID(r)
	if !authenticated {
		http.Redirect(w, r, "/login?next=/profile", http.StatusFound)
		return
	}

	user, err := h.Repo.Users().GetUserByID(userID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get order count
	orders, err := h.Repo.Orders().GetOrdersByUserID(userID)
	orderCount := 0
	if err == nil {
		orderCount = len(orders)
	}

	data := ProfileViewData{
		IsAuthenticated:   true,
		ReaderBrowserURL:  h.ReaderBrowserURL,
		ChatbotBrowserURL: h.ChatbotBrowserURL,
		User:              user,
		OrderCount:        orderCount,
		Success:           r.URL.Query().Get("success"),
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/profile.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := ts.ExecuteTemplate(w, "profile.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// ProfileEditPage displays the profile edit form
func (h *Handlers) ProfileEditPage(w http.ResponseWriter, r *http.Request) {
	userID, authenticated := h.GetUserID(r)
	if !authenticated {
		http.Redirect(w, r, "/login?next=/profile/edit", http.StatusFound)
		return
	}

	user, err := h.Repo.Users().GetUserByID(userID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	data := ProfileViewData{
		IsAuthenticated:   true,
		ReaderBrowserURL:  h.ReaderBrowserURL,
		ChatbotBrowserURL: h.ChatbotBrowserURL,
		User:              user,
		Error:             r.URL.Query().Get("error"),
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/profile-edit.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := ts.ExecuteTemplate(w, "profile-edit.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// UpdateProfile handles profile information updates
func (h *Handlers) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, authenticated := h.GetUserID(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	fullName := strings.TrimSpace(r.FormValue("full_name"))
	email := strings.TrimSpace(r.FormValue("email"))

	// Validate email
	if !strings.Contains(email, "@") {
		http.Redirect(w, r, "/profile/edit?error=Invalid+email+format", http.StatusFound)
		return
	}

	// Validate full name (optional but if provided, should have at least first name)
	if fullName != "" && len(fullName) < 2 {
		http.Redirect(w, r, "/profile/edit?error=Full+name+must+be+at+least+2+characters", http.StatusFound)
		return
	}

	// Update user profile
	err := h.Repo.Users().UpdateUserProfile(userID, email, fullName)
	if err != nil {
		log.Printf("Error updating profile: %v", err)
		http.Redirect(w, r, "/profile/edit?error=Could+not+update+profile", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/profile?success=Profile+updated+successfully", http.StatusFound)
}

// ProfilePasswordPage displays the password change form
func (h *Handlers) ProfilePasswordPage(w http.ResponseWriter, r *http.Request) {
	userID, authenticated := h.GetUserID(r)
	if !authenticated {
		http.Redirect(w, r, "/login?next=/profile/password", http.StatusFound)
		return
	}

	user, err := h.Repo.Users().GetUserByID(userID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	data := ProfileViewData{
		IsAuthenticated:   true,
		ReaderBrowserURL:  h.ReaderBrowserURL,
		ChatbotBrowserURL: h.ChatbotBrowserURL,
		User:              user,
		Error:             r.URL.Query().Get("error"),
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/profile-password.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := ts.ExecuteTemplate(w, "profile-password.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

// UpdatePassword handles password changes
func (h *Handlers) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID, authenticated := h.GetUserID(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	// Get user to verify current password
	user, err := h.Repo.Users().GetUserByID(userID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		http.Redirect(w, r, "/profile/password?error=Could+not+verify+user", http.StatusFound)
		return
	}

	// Verify current password
	if !user.CheckPassword(currentPassword) {
		http.Redirect(w, r, "/profile/password?error=Current+password+is+incorrect", http.StatusFound)
		return
	}

	// Validate new password
	if newPassword != confirmPassword {
		http.Redirect(w, r, "/profile/password?error=New+passwords+do+not+match", http.StatusFound)
		return
	}

	if err := validatePassword(newPassword); err != nil {
		http.Redirect(w, r, "/profile/password?error="+err.Error(), http.StatusFound)
		return
	}

	// Hash new password
	var tempUser models.User
	if err := tempUser.SetPassword(newPassword); err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Redirect(w, r, "/profile/password?error=Could+not+update+password", http.StatusFound)
		return
	}

	// Update password in database
	if err := h.Repo.Users().UpdateUserPassword(userID, tempUser.PasswordHash); err != nil {
		log.Printf("Error updating password: %v", err)
		http.Redirect(w, r, "/profile/password?error=Could+not+update+password", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/profile?success=Password+changed+successfully", http.StatusFound)
}
