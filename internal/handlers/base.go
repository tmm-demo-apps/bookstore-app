package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/sessions"
)

type Handlers struct {
	DB    *sql.DB
	Store *sessions.CookieStore
}

func (h *Handlers) IsAuthenticated(r *http.Request) bool {
	session, _ := h.Store.Get(r, "cart-session")
	_, ok := session.Values["user_id"].(int)
	return ok
}
