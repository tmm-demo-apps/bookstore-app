package handlers

import (
	"DemoApp/internal/repository"
	"net/http"

	"github.com/gorilla/sessions"
)

type Handlers struct {
	Repo  repository.Repository
	Store sessions.Store
}

func (h *Handlers) IsAuthenticated(r *http.Request) bool {
	session, _ := h.Store.Get(r, "cart-session")
	_, ok := session.Values["user_id"].(int)
	return ok
}

func (h *Handlers) GetUserID(r *http.Request) (int, bool) {
	session, _ := h.Store.Get(r, "cart-session")
	userID, ok := session.Values["user_id"].(int)
	return userID, ok
}
