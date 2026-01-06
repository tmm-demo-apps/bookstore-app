//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// GutenbergBook represents a book from Project Gutenberg
type GutenbergBook struct {
	Title       string
	Author      string
	Description string
	Category    string
	GutenbergID int
}

func main() {
	// Get environment variables
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbHost := getEnv("DB_HOST", "localhost")
	dbName := getEnv("DB_NAME", "bookstore")

	// Connect to database
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	// Curated list of ~50 popular public domain books from Project Gutenberg
	books := []GutenbergBook{
		// Classic Literature
		{
			Title:       "Pride and Prejudice",
			Author:      "Jane Austen",
			Description: "A romantic novel of manners that follows the character development of Elizabeth Bennet, who learns about the repercussions of hasty judgments and comes to appreciate the difference between superficial goodness and actual goodness.",
			Category:    "Fiction",
			GutenbergID: 1342,
		},
		{
			Title:       "Alice's Adventures in Wonderland",
			Author:      "Lewis Carroll",
			Description: "A young girl named Alice falls through a rabbit hole into a fantasy world populated by peculiar, anthropomorphic creatures. A classic tale of childhood imagination and Victorian nonsense literature.",
			Category:    "Fiction",
			GutenbergID: 11,
		},
		{
			Title:       "The Great Gatsby",
			Author:      "F. Scott Fitzgerald",
			Description: "A tragic love story set in the Jazz Age that explores themes of decadence, idealism, resistance to change, and excess. Chronicles Jay Gatsby's pursuit of his lost love, Daisy Buchanan.",
			Category:    "Fiction",
			GutenbergID: 64317,
		},
		{
			Title:       "Moby-Dick; or, The Whale",
			Author:      "Herman Melville",
			Description: "The saga of Captain Ahab's obsessive quest to avenge himself on Moby Dick, the white whale that destroyed his ship and took his leg. A masterpiece of American literature exploring obsession, revenge, and man versus nature.",
			Category:    "Fiction",
			GutenbergID: 2701,
		},
		{
			Title:       "A Tale of Two Cities",
			Author:      "Charles Dickens",
			Description: "Set in London and Paris before and during the French Revolution, this novel depicts the plight of the French peasantry under the brutal oppression of the aristocracy. Features the famous opening line 'It was the best of times, it was the worst of times.'",
			Category:    "Fiction",
			GutenbergID: 98,
		},
		{
			Title:       "The Adventures of Sherlock Holmes",
			Author:      "Arthur Conan Doyle",
			Description: "A collection of twelve short stories featuring the legendary detective Sherlock Holmes and his companion Dr. Watson. Includes classics like 'A Scandal in Bohemia' and 'The Red-Headed League.'",
			Category:    "Fiction",
			GutenbergID: 1661,
		},
		{
			Title:       "Frankenstein; Or, The Modern Prometheus",
			Author:      "Mary Wollstonecraft Shelley",
			Description: "The story of Victor Frankenstein, a young scientist who creates a sapient creature in an unorthodox scientific experiment. A Gothic novel that explores themes of ambition, responsibility, and the consequences of playing God.",
			Category:    "Fiction",
			GutenbergID: 84,
		},
		{
			Title:       "The Picture of Dorian Gray",
			Author:      "Oscar Wilde",
			Description: "A philosophical novel about a young man who sells his soul for eternal youth and beauty. As Dorian descends into a life of sin and corruption, his portrait ages and reflects his moral decay while he remains young and beautiful.",
			Category:    "Fiction",
			GutenbergID: 174,
		},
		{
			Title:       "Dracula",
			Author:      "Bram Stoker",
			Description: "The classic vampire novel told through journal entries, letters, and newspaper clippings. Follows the attempt to defeat Count Dracula, a centuries-old vampire terrorizing England. The definitive vampire story that launched a genre.",
			Category:    "Fiction",
			GutenbergID: 345,
		},
		{
			Title:       "The Adventures of Tom Sawyer",
			Author:      "Mark Twain",
			Description: "The adventures of a mischievous boy growing up along the Mississippi River. Tom's escapades include witnessing a murder, getting lost in a cave, and attending his own funeral. A beloved American classic.",
			Category:    "Fiction",
			GutenbergID: 74,
		},
		{
			Title:       "Adventures of Huckleberry Finn",
			Author:      "Mark Twain",
			Description: "Often called 'The Great American Novel,' this story follows Huck Finn and Jim, a runaway slave, as they journey down the Mississippi River. A powerful exploration of racism, morality, and freedom in pre-Civil War America.",
			Category:    "Fiction",
			GutenbergID: 76,
		},
		{
			Title:       "Jane Eyre",
			Author:      "Charlotte Brontë",
			Description: "The story of an orphaned girl who becomes a governess and falls in love with her employer, Mr. Rochester. A groundbreaking novel that explores themes of class, sexuality, religion, and feminism.",
			Category:    "Fiction",
			GutenbergID: 1260,
		},
		{
			Title:       "Wuthering Heights",
			Author:      "Emily Brontë",
			Description: "A tale of passion and revenge set on the Yorkshire moors. The turbulent relationship between Catherine Earnshaw and Heathcliff spans two generations and explores the destructive nature of obsessive love.",
			Category:    "Fiction",
			GutenbergID: 768,
		},
		{
			Title:       "The Count of Monte Cristo",
			Author:      "Alexandre Dumas",
			Description: "An adventure novel of revenge and redemption. After being wrongly imprisoned, Edmond Dantès escapes and discovers a treasure that allows him to exact elaborate revenge on those who betrayed him.",
			Category:    "Fiction",
			GutenbergID: 1184,
		},
		{
			Title:       "The Three Musketeers",
			Author:      "Alexandre Dumas",
			Description: "Set in 17th century France, this swashbuckling adventure follows d'Artagnan as he joins forces with three musketeers—Athos, Porthos, and Aramis. Famous for the motto 'All for one, one for all!'",
			Category:    "Fiction",
			GutenbergID: 1257,
		},
		{
			Title:       "Little Women",
			Author:      "Louisa May Alcott",
			Description: "The story of the four March sisters—Meg, Jo, Beth, and Amy—growing up in Civil War-era New England. A timeless tale of family, love, loss, and the pursuit of dreams.",
			Category:    "Fiction",
			GutenbergID: 514,
		},
		{
			Title:       "The Scarlet Letter",
			Author:      "Nathaniel Hawthorne",
			Description: "Set in Puritan Massachusetts, this novel tells the story of Hester Prynne, who conceives a daughter through an affair and struggles to create a new life of repentance and dignity. A powerful exploration of sin, guilt, and redemption.",
			Category:    "Fiction",
			GutenbergID: 25344,
		},
		{
			Title:       "The Wonderful Wizard of Oz",
			Author:      "L. Frank Baum",
			Description: "Dorothy and her dog Toto are swept away to the magical Land of Oz, where they meet the Scarecrow, Tin Woodman, and Cowardly Lion on a journey to meet the Wizard. An American fairy tale classic.",
			Category:    "Fiction",
			GutenbergID: 55,
		},
		{
			Title:       "The Secret Garden",
			Author:      "Frances Hodgson Burnett",
			Description: "A young orphan discovers a hidden, neglected garden and brings it back to life, transforming herself and those around her in the process. A beloved children's classic about healing, growth, and the power of nature.",
			Category:    "Fiction",
			GutenbergID: 113,
		},
		{
			Title:       "Treasure Island",
			Author:      "Robert Louis Stevenson",
			Description: "Young Jim Hawkins discovers a treasure map and sets sail on an adventure filled with pirates, mutiny, and buried gold. Features the iconic Long John Silver and established many pirate story tropes.",
			Category:    "Fiction",
			GutenbergID: 120,
		},
		{
			Title:       "The Strange Case of Dr. Jekyll and Mr. Hyde",
			Author:      "Robert Louis Stevenson",
			Description: "A London lawyer investigates strange occurrences between his friend Dr. Jekyll and the evil Mr. Hyde. A psychological thriller exploring the duality of human nature and the battle between good and evil.",
			Category:    "Fiction",
			GutenbergID: 43,
		},
		{
			Title:       "Heart of Darkness",
			Author:      "Joseph Conrad",
			Description: "A voyage up the Congo River into the heart of Africa and the human psyche. Marlow's journey to find the mysterious Kurtz becomes a meditation on colonialism, civilization, and the darkness within humanity.",
			Category:    "Fiction",
			GutenbergID: 219,
		},
		{
			Title:       "The Metamorphosis",
			Author:      "Franz Kafka",
			Description: "Gregor Samsa wakes one morning to find himself transformed into a giant insect. This surreal novella explores themes of alienation, identity, guilt, and absurdity in modern life.",
			Category:    "Fiction",
			GutenbergID: 5200,
		},
		{
			Title:       "The Iliad",
			Author:      "Homer",
			Description: "An ancient Greek epic poem set during the Trojan War, focusing on the hero Achilles. One of the oldest works of Western literature, exploring themes of honor, glory, wrath, and mortality.",
			Category:    "Fiction",
			GutenbergID: 6130,
		},
		{
			Title:       "The Odyssey",
			Author:      "Homer",
			Description: "The epic tale of Odysseus's ten-year journey home after the Trojan War. Filled with mythical creatures, gods, and adventures, this foundational work explores themes of perseverance, cunning, and homecoming.",
			Category:    "Fiction",
			GutenbergID: 1727,
		},
		{
			Title:       "Don Quixote",
			Author:      "Miguel de Cervantes Saavedra",
			Description: "A Spanish nobleman reads so many chivalric romances that he loses his sanity and decides to become a knight-errant. Often considered the first modern novel and one of the greatest works of fiction ever published.",
			Category:    "Fiction",
			GutenbergID: 996,
		},
		{
			Title:       "War and Peace",
			Author:      "Leo Tolstoy",
			Description: "An epic novel chronicling the French invasion of Russia and its impact on Tsarist society through the lives of five aristocratic families. A masterpiece exploring history, philosophy, love, and war.",
			Category:    "Fiction",
			GutenbergID: 2600,
		},
		{
			Title:       "Anna Karenina",
			Author:      "Leo Tolstoy",
			Description: "A tragic love story of a married aristocrat who has an affair with Count Vronsky, leading to her social ostracism and downfall. Explores themes of family, faith, and Russian society.",
			Category:    "Fiction",
			GutenbergID: 1399,
		},
		{
			Title:       "Crime and Punishment",
			Author:      "Fyodor Dostoevsky",
			Description: "A psychological drama following Raskolnikov, an impoverished student who murders a pawnbroker. The novel explores themes of guilt, redemption, morality, and the psychological torment of crime.",
			Category:    "Fiction",
			GutenbergID: 2554,
		},
		{
			Title:       "The Brothers Karamazov",
			Author:      "Fyodor Dostoevsky",
			Description: "A philosophical novel exploring faith, doubt, free will, and morality through the story of three brothers and their father's murder. Considered one of the greatest novels ever written.",
			Category:    "Fiction",
			GutenbergID: 28054,
		},
		{
			Title:       "Les Misérables",
			Author:      "Victor Hugo",
			Description: "The story of ex-convict Jean Valjean's quest for redemption in 19th century France. An epic tale of justice, love, sacrifice, and revolution that examines the nature of law and grace.",
			Category:    "Fiction",
			GutenbergID: 135,
		},
		{
			Title:       "The Hunchback of Notre-Dame",
			Author:      "Victor Hugo",
			Description: "Set in medieval Paris, this Gothic novel tells the story of Quasimodo, the deformed bell-ringer of Notre-Dame Cathedral, and his love for the beautiful gypsy Esmeralda.",
			Category:    "Fiction",
			GutenbergID: 2610,
		},
		{
			Title:       "Madame Bovary",
			Author:      "Gustave Flaubert",
			Description: "The story of Emma Bovary, a doctor's wife who has adulterous affairs and lives beyond her means in search of passion and fulfillment. A landmark novel of literary realism.",
			Category:    "Fiction",
			GutenbergID: 2413,
		},
		{
			Title:       "The Time Machine",
			Author:      "H. G. Wells",
			Description: "A scientist invents a machine that allows him to travel through time, journeying to the year 802,701 AD where he discovers a dystopian future. A pioneering work of science fiction.",
			Category:    "Science Fiction",
			GutenbergID: 35,
		},
		{
			Title:       "The War of the Worlds",
			Author:      "H. G. Wells",
			Description: "Martians invade Earth with advanced technology, devastating Victorian England. One of the earliest stories depicting conflict between mankind and an extraterrestrial race, and a science fiction classic.",
			Category:    "Science Fiction",
			GutenbergID: 36,
		},
		{
			Title:       "Twenty Thousand Leagues Under the Sea",
			Author:      "Jules Verne",
			Description: "The adventures of Captain Nemo and his submarine Nautilus as seen through the eyes of Professor Aronnax. A pioneering work of science fiction featuring underwater exploration and marine biology.",
			Category:    "Science Fiction",
			GutenbergID: 164,
		},
		{
			Title:       "Around the World in Eighty Days",
			Author:      "Jules Verne",
			Description: "Phileas Fogg bets that he can circumnavigate the globe in 80 days. An adventure novel filled with exotic locations, narrow escapes, and a race against time.",
			Category:    "Fiction",
			GutenbergID: 103,
		},
		{
			Title:       "Journey to the Center of the Earth",
			Author:      "Jules Verne",
			Description: "A German professor discovers a coded message that leads him on an expedition to the Earth's core. A classic adventure novel of underground exploration and prehistoric discoveries.",
			Category:    "Science Fiction",
			GutenbergID: 18857,
		},
		{
			Title:       "The Jungle Book",
			Author:      "Rudyard Kipling",
			Description: "A collection of stories set in India, most notably about Mowgli, a boy raised by wolves in the jungle. Features beloved characters like Baloo the bear and Bagheera the panther.",
			Category:    "Fiction",
			GutenbergID: 236,
		},
		{
			Title:       "The Call of the Wild",
			Author:      "Jack London",
			Description: "Buck, a domesticated dog, is stolen and sold as a sled dog in Alaska during the Klondike Gold Rush. He must adapt to survive in the harsh wilderness, eventually answering the call of his wild ancestors.",
			Category:    "Fiction",
			GutenbergID: 215,
		},
		{
			Title:       "White Fang",
			Author:      "Jack London",
			Description: "The mirror image of The Call of the Wild—a wild wolf-dog's journey toward domestication. Set during the Klondike Gold Rush, exploring themes of survival, nature versus nurture, and redemption.",
			Category:    "Fiction",
			GutenbergID: 910,
		},
		{
			Title:       "The Importance of Being Earnest",
			Author:      "Oscar Wilde",
			Description: "A farcical comedy about two men who create fictitious personas to escape their social obligations. Wilde's masterpiece of wit and satire skewering Victorian society and its hypocrisies.",
			Category:    "Drama",
			GutenbergID: 844,
		},
		{
			Title:       "A Christmas Carol",
			Author:      "Charles Dickens",
			Description: "Ebenezer Scrooge, a miserly old man, is visited by three ghosts on Christmas Eve who show him his past, present, and future. A timeless tale of redemption and the Christmas spirit.",
			Category:    "Fiction",
			GutenbergID: 46,
		},
		{
			Title:       "Great Expectations",
			Author:      "Charles Dickens",
			Description: "The coming-of-age story of Pip, an orphan who dreams of becoming a gentleman. A tale of ambition, love, and the true meaning of being a gentleman in Victorian England.",
			Category:    "Fiction",
			GutenbergID: 1400,
		},
		{
			Title:       "Oliver Twist",
			Author:      "Charles Dickens",
			Description: "An orphan boy escapes a workhouse and falls in with a gang of pickpockets in London. A scathing critique of poverty and child labor in Victorian England.",
			Category:    "Fiction",
			GutenbergID: 730,
		},
		{
			Title:       "The Prince",
			Author:      "Niccolò Machiavelli",
			Description: "A political treatise on acquiring and maintaining political power. Famous for its pragmatic and sometimes ruthless advice, giving rise to the term 'Machiavellian.'",
			Category:    "Non-Fiction",
			GutenbergID: 1232,
		},
		{
			Title:       "The Republic",
			Author:      "Plato",
			Description: "A Socratic dialogue concerning justice, the order and character of the just city-state, and the just man. One of the most influential works in philosophy and political theory.",
			Category:    "Philosophy",
			GutenbergID: 1497,
		},
		{
			Title:       "Walden",
			Author:      "Henry David Thoreau",
			Description: "A reflection upon simple living in natural surroundings, based on Thoreau's two-year experiment living in a cabin near Walden Pond. A foundational text of American transcendentalism.",
			Category:    "Non-Fiction",
			GutenbergID: 205,
		},
		{
			Title:       "The Communist Manifesto",
			Author:      "Karl Marx and Friedrich Engels",
			Description: "A political pamphlet outlining the theory of Communism and the class struggle between the bourgeoisie and proletariat. One of the most influential political documents in history.",
			Category:    "Non-Fiction",
			GutenbergID: 61,
		},
		{
			Title:       "The Art of War",
			Author:      "Sun Tzu",
			Description: "An ancient Chinese military treatise on strategy and tactics. Its principles have been applied to business, sports, and diplomacy, making it one of the most influential strategy texts.",
			Category:    "Non-Fiction",
			GutenbergID: 132,
		},
		// Additional Poetry
		{
			Title:       "Leaves of Grass",
			Author:      "Walt Whitman",
			Description: "A collection of poetry celebrating nature, democracy, and the human spirit.",
			Category:    "Poetry",
			GutenbergID: 1322,
		},
		{
			Title:       "The Divine Comedy",
			Author:      "Dante Alighieri",
			Description: "An epic poem describing Dante's journey through Hell, Purgatory, and Paradise.",
			Category:    "Poetry",
			GutenbergID: 8800,
		},
		{
			Title:       "Paradise Lost",
			Author:      "John Milton",
			Description: "An epic poem about the Fall of Man and the temptation of Adam and Eve.",
			Category:    "Poetry",
			GutenbergID: 20,
		},
		{
			Title:       "The Raven",
			Author:      "Edgar Allan Poe",
			Description: "A narrative poem of a talking raven's mysterious visit to a distraught lover.",
			Category:    "Poetry",
			GutenbergID: 17192,
		},
		{
			Title:       "Beowulf",
			Author:      "Unknown",
			Description: "An Old English epic poem about the hero Beowulf and his battles.",
			Category:    "Poetry",
			GutenbergID: 16328,
		},
		{
			Title:       "The Waste Land",
			Author:      "T. S. Eliot",
			Description: "A landmark modernist poem capturing post-WWI disillusionment.",
			Category:    "Poetry",
			GutenbergID: 1321,
		},
		{
			Title:       "Songs of Innocence and Experience",
			Author:      "William Blake",
			Description: "A collection of poems showing two contrary states of the human soul.",
			Category:    "Poetry",
			GutenbergID: 1934,
		},
		{
			Title:       "The Canterbury Tales",
			Author:      "Geoffrey Chaucer",
			Description: "A collection of stories told by pilgrims on their way to Canterbury.",
			Category:    "Poetry",
			GutenbergID: 2383,
		},
		// Additional Drama
		{
			Title:       "Romeo and Juliet",
			Author:      "William Shakespeare",
			Description: "The tragic love story of two young star-crossed lovers.",
			Category:    "Drama",
			GutenbergID: 1513,
		},
		{
			Title:       "Hamlet",
			Author:      "William Shakespeare",
			Description: "A tragedy about Prince Hamlet's quest to avenge his father's murder.",
			Category:    "Drama",
			GutenbergID: 1524,
		},
		{
			Title:       "Macbeth",
			Author:      "William Shakespeare",
			Description: "A tragedy about ambition, guilt, and the supernatural.",
			Category:    "Drama",
			GutenbergID: 1533,
		},
		{
			Title:       "A Midsummer Night's Dream",
			Author:      "William Shakespeare",
			Description: "A comedy about love, magic, and mischief in an enchanted forest.",
			Category:    "Drama",
			GutenbergID: 1514,
		},
		{
			Title:       "Othello",
			Author:      "William Shakespeare",
			Description: "A tragedy of jealousy, manipulation, and betrayal.",
			Category:    "Drama",
			GutenbergID: 1531,
		},
		{
			Title:       "The Tempest",
			Author:      "William Shakespeare",
			Description: "A story of magic, betrayal, and forgiveness on a remote island.",
			Category:    "Drama",
			GutenbergID: 1540,
		},
		{
			Title:       "Antigone",
			Author:      "Sophocles",
			Description: "A Greek tragedy about duty, honor, and civil disobedience.",
			Category:    "Drama",
			GutenbergID: 31,
		},
		{
			Title:       "Oedipus Rex",
			Author:      "Sophocles",
			Description: "A Greek tragedy about fate and the consequences of hubris.",
			Category:    "Drama",
			GutenbergID: 31,
		},
		{
			Title:       "A Doll's House",
			Author:      "Henrik Ibsen",
			Description: "A groundbreaking play about marriage, identity, and women's rights.",
			Category:    "Drama",
			GutenbergID: 2542,
		},
		// Additional Philosophy
		{
			Title:       "Meditations",
			Author:      "Marcus Aurelius",
			Description: "Personal writings of the Roman Emperor on Stoic philosophy.",
			Category:    "Philosophy",
			GutenbergID: 2680,
		},
		{
			Title:       "The Social Contract",
			Author:      "Jean-Jacques Rousseau",
			Description: "A treatise on political philosophy and the nature of society.",
			Category:    "Philosophy",
			GutenbergID: 46333,
		},
		{
			Title:       "Beyond Good and Evil",
			Author:      "Friedrich Nietzsche",
			Description: "A critique of traditional morality and philosophy.",
			Category:    "Philosophy",
			GutenbergID: 4363,
		},
		{
			Title:       "Thus Spoke Zarathustra",
			Author:      "Friedrich Nietzsche",
			Description: "A philosophical novel about the Übermensch and eternal recurrence.",
			Category:    "Philosophy",
			GutenbergID: 1998,
		},
		{
			Title:       "The Critique of Pure Reason",
			Author:      "Immanuel Kant",
			Description: "A foundational work in modern philosophy examining human knowledge.",
			Category:    "Philosophy",
			GutenbergID: 4280,
		},
		{
			Title:       "Discourse on Method",
			Author:      "René Descartes",
			Description: "A philosophical and autobiographical treatise on scientific method.",
			Category:    "Philosophy",
			GutenbergID: 59,
		},
		{
			Title:       "Leviathan",
			Author:      "Thomas Hobbes",
			Description: "A work of political philosophy on the structure of society and government.",
			Category:    "Philosophy",
			GutenbergID: 3207,
		},
		// Political Science
		{
			Title:       "The Federalist Papers",
			Author:      "Hamilton, Madison, Jay",
			Description: "Essays promoting the ratification of the US Constitution.",
			Category:    "Political Science",
			GutenbergID: 1404,
		},
		{
			Title:       "Common Sense",
			Author:      "Thomas Paine",
			Description: "A pamphlet advocating independence from Great Britain.",
			Category:    "Political Science",
			GutenbergID: 147,
		},
		{
			Title:       "The Rights of Man",
			Author:      "Thomas Paine",
			Description: "A defense of the French Revolution and human rights.",
			Category:    "Political Science",
			GutenbergID: 3742,
		},
		{
			Title:       "On Liberty",
			Author:      "John Stuart Mill",
			Description: "An essay on the nature and limits of state power over the individual.",
			Category:    "Political Science",
			GutenbergID: 34901,
		},
		{
			Title:       "The Wealth of Nations",
			Author:      "Adam Smith",
			Description: "A foundational work on economics and free market capitalism.",
			Category:    "Political Science",
			GutenbergID: 3300,
		},
		{
			Title:       "Democracy in America",
			Author:      "Alexis de Tocqueville",
			Description: "An analysis of American democracy and its strengths.",
			Category:    "Political Science",
			GutenbergID: 815,
		},
		{
			Title:       "The Second Treatise of Government",
			Author:      "John Locke",
			Description: "A work on natural rights and the social contract.",
			Category:    "Political Science",
			GutenbergID: 7370,
		},
		{
			Title:       "Utopia",
			Author:      "Thomas More",
			Description: "A work of fiction describing an ideal society.",
			Category:    "Political Science",
			GutenbergID: 2130,
		},
		// History
		{
			Title:       "The History of the Decline and Fall of the Roman Empire",
			Author:      "Edward Gibbon",
			Description: "A comprehensive history of the Roman Empire's decline.",
			Category:    "History",
			GutenbergID: 25717,
		},
		{
			Title:       "The Histories",
			Author:      "Herodotus",
			Description: "Ancient Greek historical accounts of the Greco-Persian Wars.",
			Category:    "History",
			GutenbergID: 2707,
		},
		{
			Title:       "The Peloponnesian War",
			Author:      "Thucydides",
			Description: "A historical account of the war between Athens and Sparta.",
			Category:    "History",
			GutenbergID: 7142,
		},
		{
			Title:       "The Gallic Wars",
			Author:      "Julius Caesar",
			Description: "Caesar's firsthand account of his military campaigns in Gaul.",
			Category:    "History",
			GutenbergID: 10657,
		},
		{
			Title:       "The Autobiography of Benjamin Franklin",
			Author:      "Benjamin Franklin",
			Description: "The life story of one of America's Founding Fathers.",
			Category:    "History",
			GutenbergID: 20203,
		},
		{
			Title:       "The Diary of a Young Girl",
			Author:      "Anne Frank",
			Description: "A Jewish girl's diary during the Nazi occupation of the Netherlands.",
			Category:    "History",
			GutenbergID: 4650,
		},
		{
			Title:       "Narrative of the Life of Frederick Douglass",
			Author:      "Frederick Douglass",
			Description: "An autobiography of the famous abolitionist and former slave.",
			Category:    "History",
			GutenbergID: 23,
		},
		{
			Title:       "Up From Slavery",
			Author:      "Booker T. Washington",
			Description: "An autobiography of an African American educator and leader.",
			Category:    "History",
			GutenbergID: 2376,
		},
		{
			Title:       "The History of the Ancient World",
			Author:      "Herodotus",
			Description: "Ancient historical accounts from the father of history.",
			Category:    "History",
			GutenbergID: 2456,
		},
		{
			Title:       "Memoirs of Napoleon Bonaparte",
			Author:      "Louis Antoine Fauvelet de Bourrienne",
			Description: "Personal memoirs of Napoleon's private secretary.",
			Category:    "History",
			GutenbergID: 3567,
		},
	}

	log.Printf("Preparing to seed %d books from Project Gutenberg", len(books))

	// Get existing categories
	categoryMap := make(map[string]int)
	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		log.Fatalf("Failed to query categories: %v", err)
	}
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Printf("Error scanning category: %v", err)
			continue
		}
		categoryMap[name] = id
	}
	rows.Close()

	// Check which books already exist
	existingBooks := make(map[string]int)
	rows, err = db.Query("SELECT id, name FROM products")
	if err != nil {
		log.Fatalf("Failed to query products: %v", err)
	}
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Printf("Error scanning product: %v", err)
			continue
		}
		existingBooks[name] = id
	}
	rows.Close()

	updateCount := 0
	insertCount := 0

	for _, book := range books {
		// Determine category ID
		categoryID := categoryMap["Fiction"] // Default
		if catID, ok := categoryMap[book.Category]; ok {
			categoryID = catID
		}

		// Generate price based on book length/popularity (simplified)
		price := 9.99 + float64(len(book.Description))/100.0
		if price > 19.99 {
			price = 19.99
		}

		if existingID, exists := existingBooks[book.Title]; exists {
			// Update existing book
			_, err := db.Exec(`
				UPDATE products 
				SET description = $1, author = $2, category_id = $3, price = $4
				WHERE id = $5`,
				book.Description, book.Author, categoryID, price, existingID)
			if err != nil {
				log.Printf("Error updating book '%s': %v", book.Title, err)
				continue
			}
			updateCount++
			log.Printf("Updated: %s by %s", book.Title, book.Author)
		} else {
			// Insert new book
			_, err := db.Exec(`
				INSERT INTO products (name, description, price, sku, stock_quantity, category_id, status, author)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				book.Title, book.Description, price, fmt.Sprintf("BOOK-%d", book.GutenbergID),
				50, categoryID, "active", book.Author)
			if err != nil {
				log.Printf("Error inserting book '%s': %v", book.Title, err)
				continue
			}
			insertCount++
			log.Printf("Inserted: %s by %s", book.Title, book.Author)
		}
	}

	log.Printf("Successfully seeded books: %d updated, %d inserted", updateCount, insertCount)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
