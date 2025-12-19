package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" {
		dbUser = "user"
	}
	if dbPassword == "" {
		dbPassword = "password"
	}
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbName == "" {
		dbName = "bookstore"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Connected to database")

	// First, clean up duplicate categories
	log.Println("Cleaning up duplicate categories...")
	_, err = db.Exec(`
		-- Keep only unique categories
		DELETE FROM categories WHERE id IN (5, 6, 7, 8);
	`)
	if err != nil {
		log.Printf("Note: Could not delete duplicate categories (may not exist): %v", err)
	}

	// Define categories based on Project Gutenberg's main categories
	categories := map[string]int{
		"Fiction":       1,
		"Non-Fiction":   2,
		"Science":       3,
		"Technology":    4,
		"Philosophy":    0, // Will be created
		"History":       0,
		"Poetry":        0,
		"Drama":         0,
		"Political Science": 0,
	}

	// Create new categories if they don't exist
	log.Println("Creating new categories...")
	for name, id := range categories {
		if id == 0 {
			var newID int
			err := db.QueryRow("INSERT INTO categories (name) VALUES ($1) ON CONFLICT DO NOTHING RETURNING id", name).Scan(&newID)
			if err == nil {
				categories[name] = newID
				log.Printf("Created category: %s (ID: %d)", name, newID)
			} else {
				// Category might already exist, try to get its ID
				err = db.QueryRow("SELECT id FROM categories WHERE name = $1", name).Scan(&newID)
				if err == nil {
					categories[name] = newID
				}
			}
		}
	}

	// Categorize books based on their titles and known information
	bookCategories := map[string]string{
		// Classic Fiction
		"Pride and Prejudice":                       "Fiction",
		"Alice's Adventures in Wonderland":          "Fiction",
		"The Great Gatsby":                          "Fiction",
		"Moby-Dick; or, The Whale":                  "Fiction",
		"A Tale of Two Cities":                      "Fiction",
		"The Adventures of Sherlock Holmes":         "Fiction",
		"Frankenstein; Or, The Modern Prometheus":   "Fiction",
		"The Picture of Dorian Gray":                "Fiction",
		"Dracula":                                   "Fiction",
		"The Adventures of Tom Sawyer":              "Fiction",
		"Adventures of Huckleberry Finn":            "Fiction",
		"Jane Eyre":                                 "Fiction",
		"Wuthering Heights":                         "Fiction",
		"The Count of Monte Cristo":                 "Fiction",
		"The Three Musketeers":                      "Fiction",
		"Little Women":                              "Fiction",
		"The Scarlet Letter":                        "Fiction",
		"The Wonderful Wizard of Oz":                "Fiction",
		"The Secret Garden":                         "Fiction",
		"Treasure Island":                           "Fiction",
		"The Strange Case of Dr. Jekyll and Mr. Hyde": "Fiction",
		"Heart of Darkness":                         "Fiction",
		"The Metamorphosis":                         "Fiction",
		"Don Quixote":                               "Fiction",
		"War and Peace":                             "Fiction",
		"Anna Karenina":                             "Fiction",
		"Crime and Punishment":                      "Fiction",
		"The Brothers Karamazov":                    "Fiction",
		"Les Misérables":                            "Fiction",
		"The Hunchback of Notre-Dame":               "Fiction",
		"Madame Bovary":                             "Fiction",
		"The Time Machine":                          "Fiction",
		"The War of the Worlds":                     "Fiction",
		"Twenty Thousand Leagues Under the Sea":     "Fiction",
		"Around the World in Eighty Days":           "Fiction",
		"Journey to the Center of the Earth":        "Fiction",
		"The Jungle Book":                           "Fiction",
		"The Call of the Wild":                      "Fiction",
		"White Fang":                                "Fiction",
		"1984":                                      "Fiction",
		"To Kill a Mockingbird":                     "Fiction",
		"The Catcher in the Rye":                    "Fiction",
		
		// Poetry & Drama
		"The Iliad":                                 "Poetry",
		"The Odyssey":                               "Poetry",
		"The Importance of Being Earnest":           "Drama",
		
		// History & Biography
		"A Christmas Carol":                         "Fiction",
		"Great Expectations":                        "Fiction",
		"Oliver Twist":                              "Fiction",
		
		// Philosophy & Political Science
		"The Prince":                                "Philosophy",
		"The Republic":                              "Philosophy",
		"The Communist Manifesto":                   "Political Science",
		"The Art of War":                            "Philosophy",
		
		// Non-Fiction
		"Walden":                                    "Non-Fiction",
		"Sapiens: A Brief History of Humankind":     "Non-Fiction",
		"Educated":                                  "Non-Fiction",
		"Becoming":                                  "Non-Fiction",
		"Thinking, Fast and Slow":                   "Non-Fiction",
		"Silent Spring":                             "Non-Fiction",
		"A Brief History of Time":                   "Science",
		"Cosmos":                                    "Science",
		"The Gene: An Intimate History":             "Science",
		"Astrophysics for People in a Hurry":        "Science",
		
		// Technology
		"The Pragmatic Programmer":                  "Technology",
		"Clean Code":                                "Technology",
		"Introduction to Algorithms":                "Technology",
		"Design Patterns":                           "Technology",
		"The Phoenix Project":                       "Technology",
	}

	log.Println("Updating book categories...")
	updateCount := 0
	for bookTitle, categoryName := range bookCategories {
		categoryID, ok := categories[categoryName]
		if !ok {
			log.Printf("Warning: Category '%s' not found for book '%s'", categoryName, bookTitle)
			continue
		}

		result, err := db.Exec("UPDATE products SET category_id = $1 WHERE name = $2", categoryID, bookTitle)
		if err != nil {
			log.Printf("Error updating '%s': %v", bookTitle, err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			updateCount++
			log.Printf("✓ Updated '%s' → %s (ID: %d)", bookTitle, categoryName, categoryID)
		}
	}

	// Update sample products to Fiction
	db.Exec("UPDATE products SET category_id = $1 WHERE name LIKE 'Sample Product%'", categories["Fiction"])

	log.Printf("\n✅ Successfully categorized %d books", updateCount)
	
	// Show summary
	rows, _ := db.Query(`
		SELECT c.name, COUNT(p.id) as count 
		FROM categories c 
		LEFT JOIN products p ON c.id = p.category_id 
		GROUP BY c.id, c.name 
		ORDER BY c.name
	`)
	defer rows.Close()
	
	log.Println("\n📊 Category Summary:")
	for rows.Next() {
		var catName string
		var count int
		rows.Scan(&catName, &count)
		log.Printf("  %s: %d products", catName, count)
	}
}

