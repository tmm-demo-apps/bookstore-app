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

	// Get category IDs
	categories := make(map[string]int)
	rows, _ := db.Query("SELECT id, name FROM categories")
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		categories[name] = id
	}
	rows.Close()

	// Additional books to balance categories (from Project Gutenberg)
	newBooks := []struct {
		Title       string
		Author      string
		Category    string
		GutenbergID string
		Description string
	}{
		// More Poetry (need ~8 more)
		{"Leaves of Grass", "Walt Whitman", "Poetry", "1322", "A collection of poetry celebrating nature, democracy, and the human spirit."},
		{"The Divine Comedy", "Dante Alighieri", "Poetry", "8800", "An epic poem describing Dante's journey through Hell, Purgatory, and Paradise."},
		{"Paradise Lost", "John Milton", "Poetry", "20", "An epic poem about the Fall of Man and the temptation of Adam and Eve."},
		{"The Raven", "Edgar Allan Poe", "Poetry", "17192", "A narrative poem of a talking raven's mysterious visit to a distraught lover."},
		{"Beowulf", "Unknown", "Poetry", "16328", "An Old English epic poem about the hero Beowulf and his battles."},
		{"The Waste Land", "T. S. Eliot", "Poetry", "1321", "A landmark modernist poem capturing post-WWI disillusionment."},
		{"Songs of Innocence and Experience", "William Blake", "Poetry", "1934", "A collection of poems showing two contrary states of the human soul."},
		{"The Canterbury Tales", "Geoffrey Chaucer", "Poetry", "2383", "A collection of stories told by pilgrims on their way to Canterbury."},

		// More Drama (need ~9 more)
		{"Romeo and Juliet", "William Shakespeare", "Drama", "1513", "The tragic love story of two young star-crossed lovers."},
		{"Hamlet", "William Shakespeare", "Drama", "1524", "A tragedy about Prince Hamlet's quest to avenge his father's murder."},
		{"Macbeth", "William Shakespeare", "Drama", "1533", "A tragedy about ambition, guilt, and the supernatural."},
		{"A Midsummer Night's Dream", "William Shakespeare", "Drama", "1514", "A comedy about love, magic, and mischief in an enchanted forest."},
		{"Othello", "William Shakespeare", "Drama", "1531", "A tragedy of jealousy, manipulation, and betrayal."},
		{"The Tempest", "William Shakespeare", "Drama", "1540", "A story of magic, betrayal, and forgiveness on a remote island."},
		{"Antigone", "Sophocles", "Drama", "31", "A Greek tragedy about duty, honor, and civil disobedience."},
		{"Oedipus Rex", "Sophocles", "Drama", "31", "A Greek tragedy about fate and the consequences of hubris."},
		{"A Doll's House", "Henrik Ibsen", "Drama", "2542", "A groundbreaking play about marriage, identity, and women's rights."},

		// More Philosophy (need ~7 more)
		{"Meditations", "Marcus Aurelius", "Philosophy", "2680", "Personal writings of the Roman Emperor on Stoic philosophy."},
		{"The Social Contract", "Jean-Jacques Rousseau", "Philosophy", "46333", "A treatise on political philosophy and the nature of society."},
		{"Beyond Good and Evil", "Friedrich Nietzsche", "Philosophy", "4363", "A critique of traditional morality and philosophy."},
		{"Thus Spoke Zarathustra", "Friedrich Nietzsche", "Philosophy", "1998", "A philosophical novel about the Übermensch and eternal recurrence."},
		{"The Critique of Pure Reason", "Immanuel Kant", "Philosophy", "4280", "A foundational work in modern philosophy examining human knowledge."},
		{"Discourse on Method", "René Descartes", "Philosophy", "59", "A philosophical and autobiographical treatise on scientific method."},
		{"Leviathan", "Thomas Hobbes", "Philosophy", "3207", "A work of political philosophy on the structure of society and government."},

		// More Political Science (need ~9 more)
		{"The Federalist Papers", "Hamilton, Madison, Jay", "Political Science", "1404", "Essays promoting the ratification of the US Constitution."},
		{"Common Sense", "Thomas Paine", "Political Science", "147", "A pamphlet advocating independence from Great Britain."},
		{"The Rights of Man", "Thomas Paine", "Political Science", "3742", "A defense of the French Revolution and human rights."},
		{"On Liberty", "John Stuart Mill", "Political Science", "34901", "An essay on the nature and limits of state power over the individual."},
		{"The Wealth of Nations", "Adam Smith", "Political Science", "3300", "A foundational work on economics and free market capitalism."},
		{"Democracy in America", "Alexis de Tocqueville", "Political Science", "815", "An analysis of American democracy and its strengths."},
		{"The Second Treatise of Government", "John Locke", "Political Science", "7370", "A work on natural rights and the social contract."},
		{"Utopia", "Thomas More", "Political Science", "2130", "A work of fiction describing an ideal society."},
		{"The Prince", "Niccolò Machiavelli", "Political Science", "1232", "A political treatise on power and leadership."},

		// More History (need ~10)
		{"The History of the Decline and Fall of the Roman Empire", "Edward Gibbon", "History", "25717", "A comprehensive history of the Roman Empire's decline."},
		{"The Histories", "Herodotus", "History", "2707", "Ancient Greek historical accounts of the Greco-Persian Wars."},
		{"The Peloponnesian War", "Thucydides", "History", "7142", "A historical account of the war between Athens and Sparta."},
		{"The Gallic Wars", "Julius Caesar", "History", "10657", "Caesar's firsthand account of his military campaigns in Gaul."},
		{"The Autobiography of Benjamin Franklin", "Benjamin Franklin", "History", "20203", "The life story of one of America's Founding Fathers."},
		{"The Diary of a Young Girl", "Anne Frank", "History", "4650", "A Jewish girl's diary during the Nazi occupation of the Netherlands."},
		{"Narrative of the Life of Frederick Douglass", "Frederick Douglass", "History", "23", "An autobiography of the famous abolitionist and former slave."},
		{"Up From Slavery", "Booker T. Washington", "History", "2376", "An autobiography of an African American educator and leader."},
		{"The History of the Ancient World", "Herodotus", "History", "2456", "Ancient historical accounts from the father of history."},
		{"Memoirs of Napoleon Bonaparte", "Louis Antoine Fauvelet de Bourrienne", "History", "3567", "Personal memoirs of Napoleon's private secretary."},
	}

	log.Printf("Adding %d new books to balance categories...\n", len(newBooks))
	addedCount := 0

	for _, book := range newBooks {
		categoryID, ok := categories[book.Category]
		if !ok {
			log.Printf("Warning: Category '%s' not found for book '%s'", book.Category, book.Title)
			continue
		}

		// Check if book already exists
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE name = $1)", book.Title).Scan(&exists)
		if err != nil {
			log.Printf("Error checking if book exists: %v", err)
			continue
		}

		if exists {
			log.Printf("⊘ Skipped '%s' (already exists)", book.Title)
			continue
		}

		// Insert new book
		_, err = db.Exec(`
			INSERT INTO products (name, description, price, sku, stock_quantity, category_id, status, author)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, book.Title, book.Description, 12.99, "BOOK-"+book.GutenbergID, 100, categoryID, "active", book.Author)

		if err != nil {
			log.Printf("Error inserting '%s': %v", book.Title, err)
			continue
		}

		addedCount++
		log.Printf("✓ Added '%s' by %s → %s", book.Title, book.Author, book.Category)
	}

	log.Printf("\n✅ Successfully added %d new books", addedCount)

	// Show updated summary
	log.Println("\n📊 Updated Category Summary:")
	rows, _ = db.Query(`
		SELECT c.name, COUNT(p.id) as count 
		FROM categories c 
		LEFT JOIN products p ON c.id = p.category_id 
		GROUP BY c.id, c.name 
		ORDER BY c.name
	`)
	defer rows.Close()

	for rows.Next() {
		var catName string
		var count int
		rows.Scan(&catName, &count)
		log.Printf("  %s: %d products", catName, count)
	}
}

