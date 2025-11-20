package handlers

import (
	"DemoApp/internal/repository"
	"net/http"

	"github.com/gorilla/sessions"
)

type Handlers struct {
	Repo  repository.Repository
	Store *sessions.CookieStore
}

func (h *Handlers) IsAuthenticated(r *http.Request) bool {
	session, _ := h.Store.Get(r, "cart-session")
	_, ok := session.Values["user_id"].(int)
	return ok
}
