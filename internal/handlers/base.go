package handlers

import (
	"DemoApp/internal/repository"
	"net/http"

	"github.com/gorilla/sessions"
)

type Handlers struct {
	Repo              repository.Repository
	Store             sessions.Store
	ReaderBrowserURL  string
	ChatbotBrowserURL string
}

// BaseViewData contains common data passed to all templates
type BaseViewData struct {
	IsAuthenticated   bool
	ReaderBrowserURL  string
	ChatbotBrowserURL string
}

// GetBaseViewData returns common view data for templates
func (h *Handlers) GetBaseViewData(r *http.Request) BaseViewData {
	return BaseViewData{
		IsAuthenticated:   h.IsAuthenticated(r),
		ReaderBrowserURL:  h.ReaderBrowserURL,
		ChatbotBrowserURL: h.ChatbotBrowserURL,
	}
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
