package handlers

import (
	"DemoApp/internal/models"
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

func ListBooks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, title, author, price FROM books")
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		defer rows.Close()

		books := []models.Book{}
		for rows.Next() {
			var b models.Book
			if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Price); err != nil {
				log.Println(err)
				http.Error(w, "Internal Server Error", 500)
				return
			}
			books = append(books, b)
		}

		ts, err := template.ParseFiles("./templates/books.html")
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}

		err = ts.Execute(w, books)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
		}
	}
}
