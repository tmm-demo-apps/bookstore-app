package handlers

import (
	"DemoApp/internal/models"
	"html/template"
	"log"
	"net/http"
)

type BookListViewData struct {
	IsAuthenticated bool
	Books           []models.Book
}

func (h *Handlers) ListBooks(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, title, author, price FROM books")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Price); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		books = append(books, b)
	}

	data := BookListViewData{
		IsAuthenticated: h.IsAuthenticated(r),
		Books:           books,
	}

	ts, err := template.ParseFiles("./templates/base.html", "./templates/books.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	ts.Execute(w, data)
}
