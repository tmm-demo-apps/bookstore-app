//go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// GutenbergBook represents a book from Project Gutenberg
type GutenbergBook struct {
	Title         string
	Author        string
	Description   string
	Category      string
	GutenbergID   int
	DownloadCount int // 30-day download count from Gutenberg, used for popularity sorting
}

// Categories for the bookstore
var categories = []string{
	"Fiction",
	"Non-Fiction",
	"Science",
	"Technology",
	"Philosophy",
	"Science Fiction",
	"Drama",
	"Poetry",
	"History",
	"Political Science",
}

func main() {
	// Parse command line flags
	generateSQL := flag.Bool("generate-sql", false, "Generate SQL migration file instead of seeding database")
	outputFile := flag.String("output", "migrations/002_seed_books.sql", "Output file for SQL generation")
	flag.Parse()

	// Curated list of 150 popular public domain books from Project Gutenberg
	// Download counts are approximate based on Gutenberg popularity data
	books := []GutenbergBook{
		// Classic Literature - Fiction (sorted roughly by popularity)
		{Title: "Pride and Prejudice", Author: "Jane Austen", Description: "A romantic novel of manners that follows the character development of Elizabeth Bennet, who learns about the repercussions of hasty judgments and comes to appreciate the difference between superficial goodness and actual goodness.", Category: "Fiction", GutenbergID: 1342, DownloadCount: 65000},
		{Title: "Alice's Adventures in Wonderland", Author: "Lewis Carroll", Description: "A young girl named Alice falls through a rabbit hole into a fantasy world populated by peculiar, anthropomorphic creatures. A classic tale of childhood imagination and Victorian nonsense literature.", Category: "Fiction", GutenbergID: 11, DownloadCount: 45000},
		{Title: "The Great Gatsby", Author: "F. Scott Fitzgerald", Description: "A tragic love story set in the Jazz Age that explores themes of decadence, idealism, resistance to change, and excess. Chronicles Jay Gatsby's pursuit of his lost love, Daisy Buchanan.", Category: "Fiction", GutenbergID: 64317, DownloadCount: 42000},
		{Title: "Moby-Dick; or, The Whale", Author: "Herman Melville", Description: "The saga of Captain Ahab's obsessive quest to avenge himself on Moby Dick, the white whale that destroyed his ship and took his leg. A masterpiece of American literature exploring obsession, revenge, and man versus nature.", Category: "Fiction", GutenbergID: 2701, DownloadCount: 38000},
		{Title: "A Tale of Two Cities", Author: "Charles Dickens", Description: "Set in London and Paris before and during the French Revolution, this novel depicts the plight of the French peasantry under the brutal oppression of the aristocracy. Features the famous opening line 'It was the best of times, it was the worst of times.'", Category: "Fiction", GutenbergID: 98, DownloadCount: 55000},
		{Title: "The Adventures of Sherlock Holmes", Author: "Arthur Conan Doyle", Description: "A collection of twelve short stories featuring the legendary detective Sherlock Holmes and his companion Dr. Watson. Includes classics like 'A Scandal in Bohemia' and 'The Red-Headed League.'", Category: "Fiction", GutenbergID: 1661, DownloadCount: 52000},
		{Title: "Frankenstein; Or, The Modern Prometheus", Author: "Mary Wollstonecraft Shelley", Description: "The story of Victor Frankenstein, a young scientist who creates a sapient creature in an unorthodox scientific experiment. A Gothic novel that explores themes of ambition, responsibility, and the consequences of playing God.", Category: "Fiction", GutenbergID: 84, DownloadCount: 48000},
		{Title: "The Picture of Dorian Gray", Author: "Oscar Wilde", Description: "A philosophical novel about a young man who sells his soul for eternal youth and beauty. As Dorian descends into a life of sin and corruption, his portrait ages and reflects his moral decay while he remains young and beautiful.", Category: "Fiction", GutenbergID: 174, DownloadCount: 35000},
		{Title: "Dracula", Author: "Bram Stoker", Description: "The classic vampire novel told through journal entries, letters, and newspaper clippings. Follows the attempt to defeat Count Dracula, a centuries-old vampire terrorizing England. The definitive vampire story that launched a genre.", Category: "Fiction", GutenbergID: 345, DownloadCount: 47000},
		{Title: "The Adventures of Tom Sawyer", Author: "Mark Twain", Description: "The adventures of a mischievous boy growing up along the Mississippi River. Tom's escapades include witnessing a murder, getting lost in a cave, and attending his own funeral. A beloved American classic.", Category: "Fiction", GutenbergID: 74, DownloadCount: 32000},
		{Title: "Adventures of Huckleberry Finn", Author: "Mark Twain", Description: "Often called 'The Great American Novel,' this story follows Huck Finn and Jim, a runaway slave, as they journey down the Mississippi River. A powerful exploration of racism, morality, and freedom in pre-Civil War America.", Category: "Fiction", GutenbergID: 76, DownloadCount: 36000},
		{Title: "Jane Eyre", Author: "Charlotte Brontë", Description: "The story of an orphaned girl who becomes a governess and falls in love with her employer, Mr. Rochester. A groundbreaking novel that explores themes of class, sexuality, religion, and feminism.", Category: "Fiction", GutenbergID: 1260, DownloadCount: 41000},
		{Title: "Wuthering Heights", Author: "Emily Brontë", Description: "A tale of passion and revenge set on the Yorkshire moors. The turbulent relationship between Catherine Earnshaw and Heathcliff spans two generations and explores the destructive nature of obsessive love.", Category: "Fiction", GutenbergID: 768, DownloadCount: 33000},
		{Title: "The Count of Monte Cristo", Author: "Alexandre Dumas", Description: "An adventure novel of revenge and redemption. After being wrongly imprisoned, Edmond Dantès escapes and discovers a treasure that allows him to exact elaborate revenge on those who betrayed him.", Category: "Fiction", GutenbergID: 1184, DownloadCount: 29000},
		{Title: "The Three Musketeers", Author: "Alexandre Dumas", Description: "Set in 17th century France, this swashbuckling adventure follows d'Artagnan as he joins forces with three musketeers—Athos, Porthos, and Aramis. Famous for the motto 'All for one, one for all!'", Category: "Fiction", GutenbergID: 1257, DownloadCount: 27000},
		{Title: "Little Women", Author: "Louisa May Alcott", Description: "The story of the four March sisters—Meg, Jo, Beth, and Amy—growing up in Civil War-era New England. A timeless tale of family, love, loss, and the pursuit of dreams.", Category: "Fiction", GutenbergID: 514, DownloadCount: 31000},
		{Title: "The Scarlet Letter", Author: "Nathaniel Hawthorne", Description: "Set in Puritan Massachusetts, this novel tells the story of Hester Prynne, who conceives a daughter through an affair and struggles to create a new life of repentance and dignity. A powerful exploration of sin, guilt, and redemption.", Category: "Fiction", GutenbergID: 25344, DownloadCount: 24000},
		{Title: "The Wonderful Wizard of Oz", Author: "L. Frank Baum", Description: "Dorothy and her dog Toto are swept away to the magical Land of Oz, where they meet the Scarecrow, Tin Woodman, and Cowardly Lion on a journey to meet the Wizard. An American fairy tale classic.", Category: "Fiction", GutenbergID: 55, DownloadCount: 28000},
		{Title: "The Secret Garden", Author: "Frances Hodgson Burnett", Description: "A young orphan discovers a hidden, neglected garden and brings it back to life, transforming herself and those around her in the process. A beloved children's classic about healing, growth, and the power of nature.", Category: "Fiction", GutenbergID: 113, DownloadCount: 26000},
		{Title: "Treasure Island", Author: "Robert Louis Stevenson", Description: "Young Jim Hawkins discovers a treasure map and sets sail on an adventure filled with pirates, mutiny, and buried gold. Features the iconic Long John Silver and established many pirate story tropes.", Category: "Fiction", GutenbergID: 120, DownloadCount: 25000},
		{Title: "The Strange Case of Dr. Jekyll and Mr. Hyde", Author: "Robert Louis Stevenson", Description: "A London lawyer investigates strange occurrences between his friend Dr. Jekyll and the evil Mr. Hyde. A psychological thriller exploring the duality of human nature and the battle between good and evil.", Category: "Fiction", GutenbergID: 43, DownloadCount: 34000},
		{Title: "Heart of Darkness", Author: "Joseph Conrad", Description: "A voyage up the Congo River into the heart of Africa and the human psyche. Marlow's journey to find the mysterious Kurtz becomes a meditation on colonialism, civilization, and the darkness within humanity.", Category: "Fiction", GutenbergID: 219, DownloadCount: 23000},
		{Title: "The Metamorphosis", Author: "Franz Kafka", Description: "Gregor Samsa wakes one morning to find himself transformed into a giant insect. This surreal novella explores themes of alienation, identity, guilt, and absurdity in modern life.", Category: "Fiction", GutenbergID: 5200, DownloadCount: 30000},
		{Title: "The Iliad", Author: "Homer", Description: "An ancient Greek epic poem set during the Trojan War, focusing on the hero Achilles. One of the oldest works of Western literature, exploring themes of honor, glory, wrath, and mortality.", Category: "Poetry", GutenbergID: 6130, DownloadCount: 22000},
		{Title: "The Odyssey", Author: "Homer", Description: "The epic tale of Odysseus's ten-year journey home after the Trojan War. Filled with mythical creatures, gods, and adventures, this foundational work explores themes of perseverance, cunning, and homecoming.", Category: "Poetry", GutenbergID: 1727, DownloadCount: 24000},
		{Title: "Don Quixote", Author: "Miguel de Cervantes Saavedra", Description: "A Spanish nobleman reads so many chivalric romances that he loses his sanity and decides to become a knight-errant. Often considered the first modern novel and one of the greatest works of fiction ever published.", Category: "Fiction", GutenbergID: 996, DownloadCount: 21000},
		{Title: "War and Peace", Author: "Leo Tolstoy", Description: "An epic novel chronicling the French invasion of Russia and its impact on Tsarist society through the lives of five aristocratic families. A masterpiece exploring history, philosophy, love, and war.", Category: "Fiction", GutenbergID: 2600, DownloadCount: 19000},
		{Title: "Anna Karenina", Author: "Leo Tolstoy", Description: "A tragic love story of a married aristocrat who has an affair with Count Vronsky, leading to her social ostracism and downfall. Explores themes of family, faith, and Russian society.", Category: "Fiction", GutenbergID: 1399, DownloadCount: 18000},
		{Title: "Crime and Punishment", Author: "Fyodor Dostoevsky", Description: "A psychological drama following Raskolnikov, an impoverished student who murders a pawnbroker. The novel explores themes of guilt, redemption, morality, and the psychological torment of crime.", Category: "Fiction", GutenbergID: 2554, DownloadCount: 20000},
		{Title: "The Brothers Karamazov", Author: "Fyodor Dostoevsky", Description: "A philosophical novel exploring faith, doubt, free will, and morality through the story of three brothers and their father's murder. Considered one of the greatest novels ever written.", Category: "Fiction", GutenbergID: 28054, DownloadCount: 17000},
		{Title: "Les Misérables", Author: "Victor Hugo", Description: "The story of ex-convict Jean Valjean's quest for redemption in 19th century France. An epic tale of justice, love, sacrifice, and revolution that examines the nature of law and grace.", Category: "Fiction", GutenbergID: 135, DownloadCount: 16000},
		{Title: "The Hunchback of Notre-Dame", Author: "Victor Hugo", Description: "Set in medieval Paris, this Gothic novel tells the story of Quasimodo, the deformed bell-ringer of Notre-Dame Cathedral, and his love for the beautiful gypsy Esmeralda.", Category: "Fiction", GutenbergID: 2610, DownloadCount: 15000},
		{Title: "Madame Bovary", Author: "Gustave Flaubert", Description: "The story of Emma Bovary, a doctor's wife who has adulterous affairs and lives beyond her means in search of passion and fulfillment. A landmark novel of literary realism.", Category: "Fiction", GutenbergID: 2413, DownloadCount: 14000},
		{Title: "A Christmas Carol", Author: "Charles Dickens", Description: "Ebenezer Scrooge, a miserly old man, is visited by three ghosts on Christmas Eve who show him his past, present, and future. A timeless tale of redemption and the Christmas spirit.", Category: "Fiction", GutenbergID: 46, DownloadCount: 44000},
		{Title: "Great Expectations", Author: "Charles Dickens", Description: "The coming-of-age story of Pip, an orphan who dreams of becoming a gentleman. A tale of ambition, love, and the true meaning of being a gentleman in Victorian England.", Category: "Fiction", GutenbergID: 1400, DownloadCount: 28000},
		{Title: "Oliver Twist", Author: "Charles Dickens", Description: "An orphan boy escapes a workhouse and falls in with a gang of pickpockets in London. A scathing critique of poverty and child labor in Victorian England.", Category: "Fiction", GutenbergID: 730, DownloadCount: 26000},
		{Title: "David Copperfield", Author: "Charles Dickens", Description: "The life story of David Copperfield from childhood to maturity. Dickens's most autobiographical novel, exploring themes of perseverance, memory, and personal growth.", Category: "Fiction", GutenbergID: 766, DownloadCount: 18000},
		{Title: "Bleak House", Author: "Charles Dickens", Description: "A complex narrative centered around a long-running court case. A critique of the British legal system and a mystery involving hidden identities and dark secrets.", Category: "Fiction", GutenbergID: 1023, DownloadCount: 12000},
		{Title: "The Pickwick Papers", Author: "Charles Dickens", Description: "The episodic adventures of Samuel Pickwick and his club members as they travel through England. Dickens's first novel, full of humor and memorable characters.", Category: "Fiction", GutenbergID: 580, DownloadCount: 11000},

		// Science Fiction
		{Title: "The Time Machine", Author: "H. G. Wells", Description: "A scientist invents a machine that allows him to travel through time, journeying to the year 802,701 AD where he discovers a dystopian future. A pioneering work of science fiction.", Category: "Science Fiction", GutenbergID: 35, DownloadCount: 39000},
		{Title: "The War of the Worlds", Author: "H. G. Wells", Description: "Martians invade Earth with advanced technology, devastating Victorian England. One of the earliest stories depicting conflict between mankind and an extraterrestrial race, and a science fiction classic.", Category: "Science Fiction", GutenbergID: 36, DownloadCount: 37000},
		{Title: "The Invisible Man", Author: "H. G. Wells", Description: "A scientist discovers how to make himself invisible but cannot reverse the process. His descent into madness and terror makes this a classic tale of science gone wrong.", Category: "Science Fiction", GutenbergID: 5230, DownloadCount: 25000},
		{Title: "The Island of Doctor Moreau", Author: "H. G. Wells", Description: "A shipwrecked man discovers an island where a scientist creates human-like beings from animals. A disturbing exploration of ethics, identity, and the nature of humanity.", Category: "Science Fiction", GutenbergID: 159, DownloadCount: 18000},
		{Title: "The First Men in the Moon", Author: "H. G. Wells", Description: "Two men travel to the moon and discover an alien civilization living beneath its surface. An imaginative early work of science fiction exploring space travel.", Category: "Science Fiction", GutenbergID: 1013, DownloadCount: 14000},
		{Title: "Twenty Thousand Leagues Under the Sea", Author: "Jules Verne", Description: "The adventures of Captain Nemo and his submarine Nautilus as seen through the eyes of Professor Aronnax. A pioneering work of science fiction featuring underwater exploration and marine biology.", Category: "Science Fiction", GutenbergID: 164, DownloadCount: 32000},
		{Title: "Around the World in Eighty Days", Author: "Jules Verne", Description: "Phileas Fogg bets that he can circumnavigate the globe in 80 days. An adventure novel filled with exotic locations, narrow escapes, and a race against time.", Category: "Fiction", GutenbergID: 103, DownloadCount: 28000},
		{Title: "Journey to the Center of the Earth", Author: "Jules Verne", Description: "A German professor discovers a coded message that leads him on an expedition to the Earth's core. A classic adventure novel of underground exploration and prehistoric discoveries.", Category: "Science Fiction", GutenbergID: 18857, DownloadCount: 24000},
		{Title: "From the Earth to the Moon", Author: "Jules Verne", Description: "Members of a post-Civil War gun club build a cannon to shoot a projectile to the moon. A remarkably prescient tale of space exploration written in 1865.", Category: "Science Fiction", GutenbergID: 83, DownloadCount: 16000},
		{Title: "The Mysterious Island", Author: "Jules Verne", Description: "Five prisoners escape the American Civil War in a balloon and crash on a mysterious island. A tale of survival, ingenuity, and adventure with connections to Captain Nemo.", Category: "Science Fiction", GutenbergID: 1268, DownloadCount: 15000},

		// More Fiction
		{Title: "The Jungle Book", Author: "Rudyard Kipling", Description: "A collection of stories set in India, most notably about Mowgli, a boy raised by wolves in the jungle. Features beloved characters like Baloo the bear and Bagheera the panther.", Category: "Fiction", GutenbergID: 236, DownloadCount: 23000},
		{Title: "Kim", Author: "Rudyard Kipling", Description: "The story of an Irish orphan boy in India who becomes a spy for the British. A vivid portrait of colonial India and a coming-of-age adventure.", Category: "Fiction", GutenbergID: 2226, DownloadCount: 12000},
		{Title: "The Call of the Wild", Author: "Jack London", Description: "Buck, a domesticated dog, is stolen and sold as a sled dog in Alaska during the Klondike Gold Rush. He must adapt to survive in the harsh wilderness, eventually answering the call of his wild ancestors.", Category: "Fiction", GutenbergID: 215, DownloadCount: 27000},
		{Title: "White Fang", Author: "Jack London", Description: "The mirror image of The Call of the Wild—a wild wolf-dog's journey toward domestication. Set during the Klondike Gold Rush, exploring themes of survival, nature versus nurture, and redemption.", Category: "Fiction", GutenbergID: 910, DownloadCount: 19000},
		{Title: "The Sea-Wolf", Author: "Jack London", Description: "A literary critic is rescued by a seal-hunting ship captained by the brutal Wolf Larsen. A tale of survival, philosophy, and the clash between civilization and nature.", Category: "Fiction", GutenbergID: 1074, DownloadCount: 13000},
		{Title: "Sense and Sensibility", Author: "Jane Austen", Description: "The story of the Dashwood sisters, Elinor and Marianne, who represent sense and sensibility respectively. A novel about love, heartbreak, and finding balance.", Category: "Fiction", GutenbergID: 161, DownloadCount: 22000},
		{Title: "Emma", Author: "Jane Austen", Description: "The story of Emma Woodhouse, a young woman who fancies herself a matchmaker. A comedy of manners exploring self-deception, social class, and the journey to self-awareness.", Category: "Fiction", GutenbergID: 158, DownloadCount: 21000},
		{Title: "Mansfield Park", Author: "Jane Austen", Description: "Fanny Price is sent to live with wealthy relatives at Mansfield Park. A novel exploring morality, social mobility, and the contrast between country and city values.", Category: "Fiction", GutenbergID: 141, DownloadCount: 16000},
		{Title: "Persuasion", Author: "Jane Austen", Description: "Anne Elliot, persuaded years ago to reject a suitor, gets a second chance at love. Austen's last completed novel, exploring themes of constancy, regret, and mature love.", Category: "Fiction", GutenbergID: 105, DownloadCount: 18000},
		{Title: "Northanger Abbey", Author: "Jane Austen", Description: "A satire of Gothic novels following Catherine Morland's adventures and romantic misunderstandings. A witty commentary on reading, imagination, and coming of age.", Category: "Fiction", GutenbergID: 121, DownloadCount: 14000},
		{Title: "The Turn of the Screw", Author: "Henry James", Description: "A governess becomes convinced that the children in her care are being haunted by ghosts. A masterpiece of psychological horror and ambiguity.", Category: "Fiction", GutenbergID: 209, DownloadCount: 17000},
		{Title: "Daisy Miller", Author: "Henry James", Description: "A young American woman's unconventional behavior scandalizes European society. A study of cultural clash and the 'American Girl' abroad.", Category: "Fiction", GutenbergID: 208, DownloadCount: 11000},
		{Title: "The Portrait of a Lady", Author: "Henry James", Description: "Isabel Archer, a young American heiress, travels to Europe where she falls prey to scheming expatriates. A profound study of freedom, choice, and their consequences.", Category: "Fiction", GutenbergID: 2833, DownloadCount: 13000},
		{Title: "Tess of the d'Urbervilles", Author: "Thomas Hardy", Description: "The tragic story of Tess Durbeyfield, a peasant girl whose life is destroyed by the men around her. A powerful critique of Victorian morality and social hypocrisy.", Category: "Fiction", GutenbergID: 110, DownloadCount: 16000},
		{Title: "Far from the Madding Crowd", Author: "Thomas Hardy", Description: "Bathsheba Everdene, an independent young woman, attracts three very different suitors. A pastoral romance set in Hardy's fictional Wessex.", Category: "Fiction", GutenbergID: 107, DownloadCount: 12000},
		{Title: "Jude the Obscure", Author: "Thomas Hardy", Description: "Jude Fawley dreams of attending university but is thwarted by class barriers and tragic circumstances. Hardy's most controversial and pessimistic novel.", Category: "Fiction", GutenbergID: 153, DownloadCount: 11000},
		{Title: "The Mayor of Casterbridge", Author: "Thomas Hardy", Description: "A man sells his wife and daughter while drunk, then spends years trying to atone. A tragedy of pride, fate, and the consequences of past actions.", Category: "Fiction", GutenbergID: 143, DownloadCount: 10000},
		{Title: "Middlemarch", Author: "George Eliot", Description: "A study of provincial life in a Midlands town, following several interconnected stories. Often considered one of the greatest novels in the English language.", Category: "Fiction", GutenbergID: 145, DownloadCount: 14000},
		{Title: "Silas Marner", Author: "George Eliot", Description: "A weaver, betrayed and exiled, finds redemption through the love of an orphaned child. A moral tale about community, faith, and human connection.", Category: "Fiction", GutenbergID: 550, DownloadCount: 13000},
		{Title: "The Mill on the Floss", Author: "George Eliot", Description: "The story of Maggie Tulliver and her brother Tom, growing up in rural England. A tragedy exploring family loyalty, gender expectations, and social constraints.", Category: "Fiction", GutenbergID: 6688, DownloadCount: 10000},
		{Title: "Vanity Fair", Author: "William Makepeace Thackeray", Description: "A satirical panorama of English society following the contrasting fortunes of Becky Sharp and Amelia Sedley. A novel without a hero.", Category: "Fiction", GutenbergID: 599, DownloadCount: 12000},
		{Title: "The Scarlet Pimpernel", Author: "Baroness Orczy", Description: "A mysterious English nobleman rescues French aristocrats from the guillotine during the Reign of Terror. The original masked hero adventure.", Category: "Fiction", GutenbergID: 60, DownloadCount: 18000},
		{Title: "The Phantom of the Opera", Author: "Gaston Leroux", Description: "A masked musical genius haunts the Paris Opera House and becomes obsessed with a young soprano. The original Gothic romance that inspired countless adaptations.", Category: "Fiction", GutenbergID: 175, DownloadCount: 20000},
		{Title: "The Hound of the Baskervilles", Author: "Arthur Conan Doyle", Description: "Sherlock Holmes investigates the legend of a supernatural hound haunting an aristocratic family. The most famous of all Holmes novels.", Category: "Fiction", GutenbergID: 2852, DownloadCount: 35000},
		{Title: "A Study in Scarlet", Author: "Arthur Conan Doyle", Description: "The first appearance of Sherlock Holmes and Dr. Watson, as they investigate a mysterious murder with connections to America.", Category: "Fiction", GutenbergID: 244, DownloadCount: 28000},
		{Title: "The Sign of the Four", Author: "Arthur Conan Doyle", Description: "Holmes and Watson investigate a case involving a missing father, a mysterious pact, and a stolen treasure from India.", Category: "Fiction", GutenbergID: 2097, DownloadCount: 22000},
		{Title: "The Lost World", Author: "Arthur Conan Doyle", Description: "Professor Challenger leads an expedition to a South American plateau where dinosaurs still exist. A thrilling adventure that inspired countless imitators.", Category: "Science Fiction", GutenbergID: 139, DownloadCount: 16000},
		{Title: "Uncle Tom's Cabin", Author: "Harriet Beecher Stowe", Description: "The story of Uncle Tom, an enslaved man, and the people whose lives he touches. The novel that helped lay the groundwork for the American Civil War.", Category: "Fiction", GutenbergID: 203, DownloadCount: 19000},
		{Title: "The Red Badge of Courage", Author: "Stephen Crane", Description: "A young Union soldier faces the horrors of the Civil War and struggles with his own cowardice. A groundbreaking psychological study of fear and courage.", Category: "Fiction", GutenbergID: 73, DownloadCount: 15000},
		{Title: "Sister Carrie", Author: "Theodore Dreiser", Description: "A young woman moves to Chicago and rises in society while her lover falls. A naturalistic novel about ambition, desire, and the American Dream.", Category: "Fiction", GutenbergID: 5267, DownloadCount: 9000},
		{Title: "The Age of Innocence", Author: "Edith Wharton", Description: "A lawyer in 1870s New York falls in love with his fiancée's unconventional cousin. A Pulitzer Prize-winning novel about society, duty, and forbidden passion.", Category: "Fiction", GutenbergID: 541, DownloadCount: 14000},
		{Title: "Ethan Frome", Author: "Edith Wharton", Description: "A tragic love triangle in rural New England ends in disaster. A stark, powerful novella about unfulfilled desire and the constraints of duty.", Category: "Fiction", GutenbergID: 4517, DownloadCount: 13000},
		{Title: "The House of Mirth", Author: "Edith Wharton", Description: "Lily Bart, a beautiful but poor woman, navigates New York high society in search of a wealthy husband. A tragedy of social ambition and moral compromise.", Category: "Fiction", GutenbergID: 284, DownloadCount: 12000},

		// Drama
		{Title: "The Importance of Being Earnest", Author: "Oscar Wilde", Description: "A farcical comedy about two men who create fictitious personas to escape their social obligations. Wilde's masterpiece of wit and satire skewering Victorian society and its hypocrisies.", Category: "Drama", GutenbergID: 844, DownloadCount: 31000},
		{Title: "Romeo and Juliet", Author: "William Shakespeare", Description: "The tragic love story of two young star-crossed lovers whose deaths ultimately reconcile their feuding families.", Category: "Drama", GutenbergID: 1513, DownloadCount: 38000},
		{Title: "Hamlet", Author: "William Shakespeare", Description: "A tragedy about Prince Hamlet's quest to avenge his father's murder by his uncle, who has married Hamlet's mother and seized the throne.", Category: "Drama", GutenbergID: 1524, DownloadCount: 42000},
		{Title: "Macbeth", Author: "William Shakespeare", Description: "A Scottish general receives a prophecy that he will become king, leading him down a path of murder and madness. A tragedy about ambition, guilt, and the supernatural.", Category: "Drama", GutenbergID: 1533, DownloadCount: 36000},
		{Title: "A Midsummer Night's Dream", Author: "William Shakespeare", Description: "A comedy about love, magic, and mischief in an enchanted forest, involving four young lovers and a troupe of amateur actors.", Category: "Drama", GutenbergID: 1514, DownloadCount: 28000},
		{Title: "Othello", Author: "William Shakespeare", Description: "A Moorish general in the Venetian army is manipulated into believing his wife is unfaithful. A tragedy of jealousy, manipulation, and betrayal.", Category: "Drama", GutenbergID: 1531, DownloadCount: 26000},
		{Title: "The Tempest", Author: "William Shakespeare", Description: "A sorcerer stranded on an island uses magic to bring his enemies to justice. A story of magic, betrayal, forgiveness, and redemption.", Category: "Drama", GutenbergID: 1540, DownloadCount: 22000},
		{Title: "King Lear", Author: "William Shakespeare", Description: "An aging king divides his kingdom among his daughters based on their flattery, with tragic consequences. A profound exploration of family, madness, and mortality.", Category: "Drama", GutenbergID: 1532, DownloadCount: 24000},
		{Title: "Julius Caesar", Author: "William Shakespeare", Description: "The conspiracy against and assassination of Julius Caesar, and its aftermath. A political tragedy exploring power, honor, and betrayal.", Category: "Drama", GutenbergID: 1522, DownloadCount: 25000},
		{Title: "The Merchant of Venice", Author: "William Shakespeare", Description: "A merchant borrows money from a Jewish moneylender to help his friend woo a wealthy heiress. A complex play about justice, mercy, and prejudice.", Category: "Drama", GutenbergID: 1515, DownloadCount: 21000},
		{Title: "Much Ado About Nothing", Author: "William Shakespeare", Description: "Two couples navigate love and deception in Messina. A witty romantic comedy featuring the sparring lovers Beatrice and Benedick.", Category: "Drama", GutenbergID: 1519, DownloadCount: 18000},
		{Title: "Twelfth Night", Author: "William Shakespeare", Description: "A shipwrecked woman disguises herself as a man, leading to romantic complications. A festive comedy of mistaken identity and unrequited love.", Category: "Drama", GutenbergID: 1526, DownloadCount: 17000},
		{Title: "As You Like It", Author: "William Shakespeare", Description: "Rosalind, banished to the Forest of Arden, disguises herself as a man and encounters her love. A pastoral comedy celebrating love and nature.", Category: "Drama", GutenbergID: 1523, DownloadCount: 14000},
		{Title: "The Taming of the Shrew", Author: "William Shakespeare", Description: "Petruchio attempts to tame the headstrong Katherina. A controversial comedy about gender, marriage, and power.", Category: "Drama", GutenbergID: 1508, DownloadCount: 16000},
		{Title: "Richard III", Author: "William Shakespeare", Description: "The villainous Richard plots his way to the English throne. A historical tragedy featuring one of Shakespeare's most memorable villains.", Category: "Drama", GutenbergID: 1503, DownloadCount: 15000},
		{Title: "Henry V", Author: "William Shakespeare", Description: "The young King Henry V leads England to victory at Agincourt. A stirring historical drama about leadership, war, and national identity.", Category: "Drama", GutenbergID: 1521, DownloadCount: 13000},
		{Title: "Antigone", Author: "Sophocles", Description: "A young woman defies the king's order and buries her brother, facing death as punishment. A Greek tragedy about duty, honor, and civil disobedience.", Category: "Drama", GutenbergID: 31, DownloadCount: 19000},
		{Title: "Oedipus Rex", Author: "Sophocles", Description: "A king discovers he has unknowingly killed his father and married his mother. The archetypal Greek tragedy about fate and the consequences of hubris.", Category: "Drama", GutenbergID: 27673, DownloadCount: 21000},
		{Title: "A Doll's House", Author: "Henrik Ibsen", Description: "Nora Helmer discovers the hollowness of her marriage and makes a radical choice. A groundbreaking play about marriage, identity, and women's rights.", Category: "Drama", GutenbergID: 2542, DownloadCount: 17000},
		{Title: "The Cherry Orchard", Author: "Anton Chekhov", Description: "An aristocratic family faces the loss of their beloved estate. Chekhov's final play, a tragicomedy about change, memory, and the passing of an era.", Category: "Drama", GutenbergID: 7986, DownloadCount: 12000},
		{Title: "Faust", Author: "Johann Wolfgang von Goethe", Description: "A scholar makes a pact with the devil in exchange for unlimited knowledge and worldly pleasures. Germany's most famous literary work.", Category: "Drama", GutenbergID: 14591, DownloadCount: 15000},

		// Poetry
		{Title: "Leaves of Grass", Author: "Walt Whitman", Description: "A collection of poetry celebrating nature, democracy, the human body, and the American spirit. One of the most influential works in American literature.", Category: "Poetry", GutenbergID: 1322, DownloadCount: 18000},
		{Title: "The Divine Comedy", Author: "Dante Alighieri", Description: "An epic poem describing Dante's journey through Hell, Purgatory, and Paradise, guided by Virgil and Beatrice. One of the greatest works of world literature.", Category: "Poetry", GutenbergID: 8800, DownloadCount: 16000},
		{Title: "Paradise Lost", Author: "John Milton", Description: "An epic poem about the Fall of Man, the temptation of Adam and Eve, and their expulsion from the Garden of Eden. A masterpiece of English literature.", Category: "Poetry", GutenbergID: 20, DownloadCount: 19000},
		{Title: "The Raven and Other Poems", Author: "Edgar Allan Poe", Description: "A collection including the famous narrative poem of a talking raven's mysterious visit to a distraught lover. Poe's most celebrated poetry.", Category: "Poetry", GutenbergID: 17192, DownloadCount: 22000},
		{Title: "Beowulf", Author: "Unknown", Description: "An Old English epic poem about the hero Beowulf and his battles against monsters and a dragon. The oldest surviving long poem in English.", Category: "Poetry", GutenbergID: 16328, DownloadCount: 17000},
		{Title: "Songs of Innocence and Experience", Author: "William Blake", Description: "A collection of poems showing two contrary states of the human soul. Blake's illuminated poetry exploring childhood, society, and spirituality.", Category: "Poetry", GutenbergID: 1934, DownloadCount: 11000},
		{Title: "The Canterbury Tales", Author: "Geoffrey Chaucer", Description: "A collection of stories told by pilgrims on their way to Canterbury. A vivid portrait of medieval English society in all its variety.", Category: "Poetry", GutenbergID: 2383, DownloadCount: 14000},
		{Title: "Sonnets", Author: "William Shakespeare", Description: "154 sonnets exploring themes of love, beauty, mortality, and time. Some of the most famous love poetry in the English language.", Category: "Poetry", GutenbergID: 1041, DownloadCount: 20000},
		{Title: "The Complete Poems of Emily Dickinson", Author: "Emily Dickinson", Description: "The collected works of one of America's greatest poets, exploring death, immortality, nature, and the inner life.", Category: "Poetry", GutenbergID: 12242, DownloadCount: 15000},

		// Philosophy
		{Title: "The Prince", Author: "Niccolò Machiavelli", Description: "A political treatise on acquiring and maintaining political power. Famous for its pragmatic and sometimes ruthless advice, giving rise to the term 'Machiavellian.'", Category: "Philosophy", GutenbergID: 1232, DownloadCount: 28000},
		{Title: "The Republic", Author: "Plato", Description: "A Socratic dialogue concerning justice, the order and character of the just city-state, and the just man. One of the most influential works in philosophy and political theory.", Category: "Philosophy", GutenbergID: 1497, DownloadCount: 24000},
		{Title: "Meditations", Author: "Marcus Aurelius", Description: "Personal writings of the Roman Emperor on Stoic philosophy. A timeless guide to self-improvement and living a virtuous life.", Category: "Philosophy", GutenbergID: 2680, DownloadCount: 32000},
		{Title: "The Social Contract", Author: "Jean-Jacques Rousseau", Description: "A treatise on political philosophy arguing that legitimate political authority must be based on a social contract. Foundational to modern political thought.", Category: "Philosophy", GutenbergID: 46333, DownloadCount: 14000},
		{Title: "Beyond Good and Evil", Author: "Friedrich Nietzsche", Description: "A critique of traditional morality and philosophy, introducing concepts like the 'will to power.' One of Nietzsche's most important works.", Category: "Philosophy", GutenbergID: 4363, DownloadCount: 18000},
		{Title: "Thus Spoke Zarathustra", Author: "Friedrich Nietzsche", Description: "A philosophical novel about the prophet Zarathustra, introducing the concepts of the Übermensch and eternal recurrence.", Category: "Philosophy", GutenbergID: 1998, DownloadCount: 16000},
		{Title: "The Critique of Pure Reason", Author: "Immanuel Kant", Description: "A foundational work in modern philosophy examining the limits and possibilities of human knowledge. One of the most influential philosophical works ever written.", Category: "Philosophy", GutenbergID: 4280, DownloadCount: 12000},
		{Title: "Discourse on Method", Author: "René Descartes", Description: "A philosophical and autobiographical treatise introducing Descartes' method of systematic doubt. Contains the famous 'I think, therefore I am.'", Category: "Philosophy", GutenbergID: 59, DownloadCount: 15000},
		{Title: "Leviathan", Author: "Thomas Hobbes", Description: "A work of political philosophy on the structure of society and legitimate government. Argues for a social contract and rule by an absolute sovereign.", Category: "Philosophy", GutenbergID: 3207, DownloadCount: 13000},
		{Title: "The Nicomachean Ethics", Author: "Aristotle", Description: "Aristotle's best-known work on ethics, exploring virtue, happiness, and the good life. Foundational to Western moral philosophy.", Category: "Philosophy", GutenbergID: 8438, DownloadCount: 14000},

		// Political Science
		{Title: "The Federalist Papers", Author: "Hamilton, Madison, Jay", Description: "Essays promoting the ratification of the US Constitution. Essential reading for understanding American government and political philosophy.", Category: "Political Science", GutenbergID: 1404, DownloadCount: 19000},
		{Title: "Common Sense", Author: "Thomas Paine", Description: "A pamphlet advocating independence from Great Britain. One of the most influential documents in American history.", Category: "Political Science", GutenbergID: 147, DownloadCount: 21000},
		{Title: "The Rights of Man", Author: "Thomas Paine", Description: "A defense of the French Revolution and human rights against Edmund Burke's criticisms. A foundational text of liberal democracy.", Category: "Political Science", GutenbergID: 3742, DownloadCount: 12000},
		{Title: "On Liberty", Author: "John Stuart Mill", Description: "An essay on the nature and limits of state power over the individual. A classic defense of individual freedom and free speech.", Category: "Political Science", GutenbergID: 34901, DownloadCount: 16000},
		{Title: "The Wealth of Nations", Author: "Adam Smith", Description: "A foundational work on economics and free market capitalism. Introduced concepts like the 'invisible hand' and division of labor.", Category: "Political Science", GutenbergID: 3300, DownloadCount: 18000},
		{Title: "Democracy in America", Author: "Alexis de Tocqueville", Description: "An analysis of American democracy and its strengths and weaknesses. One of the most insightful studies of American society.", Category: "Political Science", GutenbergID: 815, DownloadCount: 14000},
		{Title: "The Second Treatise of Government", Author: "John Locke", Description: "A work on natural rights and the social contract. Foundational to liberal political philosophy and the American founding.", Category: "Political Science", GutenbergID: 7370, DownloadCount: 13000},
		{Title: "Utopia", Author: "Thomas More", Description: "A work of fiction describing an ideal society on an island. Coined the term 'utopia' and influenced political thought for centuries.", Category: "Political Science", GutenbergID: 2130, DownloadCount: 15000},

		// Non-Fiction
		{Title: "Walden", Author: "Henry David Thoreau", Description: "A reflection upon simple living in natural surroundings, based on Thoreau's two-year experiment living in a cabin near Walden Pond. A foundational text of American transcendentalism.", Category: "Non-Fiction", GutenbergID: 205, DownloadCount: 23000},
		{Title: "The Communist Manifesto", Author: "Karl Marx and Friedrich Engels", Description: "A political pamphlet outlining the theory of Communism and the class struggle between the bourgeoisie and proletariat. One of the most influential political documents in history.", Category: "Political Science", GutenbergID: 61, DownloadCount: 26000},
		{Title: "The Art of War", Author: "Sun Tzu", Description: "An ancient Chinese military treatise on strategy and tactics. Its principles have been applied to business, sports, and diplomacy, making it one of the most influential strategy texts.", Category: "Non-Fiction", GutenbergID: 132, DownloadCount: 35000},
		{Title: "The Souls of Black Folk", Author: "W. E. B. Du Bois", Description: "A seminal work in African American literature exploring the history and condition of Black Americans. Introduced the concept of 'double consciousness.'", Category: "Non-Fiction", GutenbergID: 408, DownloadCount: 14000},
		{Title: "Civil Disobedience", Author: "Henry David Thoreau", Description: "An essay arguing that individuals should not permit governments to overrule their consciences. Influenced Gandhi, Martin Luther King Jr., and others.", Category: "Non-Fiction", GutenbergID: 71, DownloadCount: 17000},

		// Science
		{Title: "On the Origin of Species", Author: "Charles Darwin", Description: "The foundational work of evolutionary biology, introducing the scientific theory that populations evolve through natural selection. One of the most influential books in scientific history.", Category: "Science", GutenbergID: 1228, DownloadCount: 28000},
		{Title: "The Descent of Man", Author: "Charles Darwin", Description: "Darwin's second major work on evolutionary theory, applying evolution to human origins and discussing sexual selection.", Category: "Science", GutenbergID: 2300, DownloadCount: 12000},
		{Title: "Principia Mathematica", Author: "Isaac Newton", Description: "Newton's masterwork laying the foundations of classical mechanics, including his laws of motion and universal gravitation.", Category: "Science", GutenbergID: 28233, DownloadCount: 15000},
		{Title: "The Voyage of the Beagle", Author: "Charles Darwin", Description: "Darwin's journal of his five-year voyage around the world, documenting the observations that led to his theory of evolution.", Category: "Science", GutenbergID: 944, DownloadCount: 14000},
		{Title: "Relativity: The Special and General Theory", Author: "Albert Einstein", Description: "Einstein's own explanation of his revolutionary theories of special and general relativity, written for a general audience.", Category: "Science", GutenbergID: 5001, DownloadCount: 22000},
		{Title: "The Elements", Author: "Euclid", Description: "The foundational textbook of geometry and mathematics, used for over two thousand years. One of the most influential works in the history of mathematics.", Category: "Science", GutenbergID: 21076, DownloadCount: 11000},
		{Title: "Dialogue Concerning the Two Chief World Systems", Author: "Galileo Galilei", Description: "Galileo's comparison of the Copernican and Ptolemaic systems, a landmark work in the history of science.", Category: "Science", GutenbergID: 46036, DownloadCount: 9000},
		{Title: "The Interpretation of Dreams", Author: "Sigmund Freud", Description: "Freud's groundbreaking work introducing his theory of the unconscious and dream analysis. A foundational text in psychology.", Category: "Science", GutenbergID: 38219, DownloadCount: 16000},

		// History
		{Title: "The History of the Decline and Fall of the Roman Empire", Author: "Edward Gibbon", Description: "A comprehensive history of the Roman Empire from its height to its fall. One of the greatest historical works ever written.", Category: "History", GutenbergID: 25717, DownloadCount: 11000},
		{Title: "The Histories", Author: "Herodotus", Description: "Ancient Greek historical accounts of the Greco-Persian Wars. The first great work of history in Western literature.", Category: "History", GutenbergID: 2707, DownloadCount: 12000},
		{Title: "The Peloponnesian War", Author: "Thucydides", Description: "A historical account of the war between Athens and Sparta. A masterpiece of historical analysis and political realism.", Category: "History", GutenbergID: 7142, DownloadCount: 10000},
		{Title: "The Gallic Wars", Author: "Julius Caesar", Description: "Caesar's firsthand account of his military campaigns in Gaul. A classic of military history and Latin prose.", Category: "History", GutenbergID: 10657, DownloadCount: 9000},
		{Title: "The Autobiography of Benjamin Franklin", Author: "Benjamin Franklin", Description: "The life story of one of America's Founding Fathers. A classic American autobiography full of wit and wisdom.", Category: "History", GutenbergID: 20203, DownloadCount: 16000},
		{Title: "Narrative of the Life of Frederick Douglass", Author: "Frederick Douglass", Description: "An autobiography of the famous abolitionist and former slave. A powerful firsthand account of slavery in America.", Category: "History", GutenbergID: 23, DownloadCount: 18000},
		{Title: "Up From Slavery", Author: "Booker T. Washington", Description: "An autobiography of an African American educator and leader. A story of determination and achievement against tremendous odds.", Category: "History", GutenbergID: 2376, DownloadCount: 12000},
	}

	if *generateSQL {
		generateSQLFile(books, *outputFile)
		return
	}

	// Original database seeding behavior
	seedDatabase(books)
}

func generateSQLFile(books []GutenbergBook, outputFile string) {
	var sb strings.Builder

	sb.WriteString("-- Auto-generated seed data for DemoApp Bookstore\n")
	sb.WriteString("-- Generated from seed-gutenberg-books.go\n")
	sb.WriteString("-- Contains categories and 150 books from Project Gutenberg\n\n")

	// Insert categories
	sb.WriteString("-- Seed Categories\n")
	sb.WriteString("INSERT INTO categories (name, description) VALUES\n")
	for i, cat := range categories {
		desc := getCategoryDescription(cat)
		if i < len(categories)-1 {
			sb.WriteString(fmt.Sprintf("    ('%s', '%s'),\n", cat, desc))
		} else {
			sb.WriteString(fmt.Sprintf("    ('%s', '%s')\n", cat, desc))
		}
	}
	sb.WriteString("ON CONFLICT (name) DO NOTHING;\n\n")

	// Insert books
	sb.WriteString("-- Seed Books (150 titles from Project Gutenberg)\n")
	sb.WriteString("-- Stock rules: 'A Christmas Carol' = 0 (out of stock, first alphabetically), 'Pride and Prejudice' = 3 (low stock), others = 10-100\n\n")

	for i, book := range books {
		// Calculate stock quantity based on title for specific demos
		var stockQty int
		switch book.Title {
		case "A Christmas Carol":
			stockQty = 0 // Out of stock - first alphabetically
		case "Pride and Prejudice":
			stockQty = 3 // Low stock demo
		default:
			// Deterministic "random" based on book index for reproducibility
			stockQty = 10 + ((i * 7) % 91) // Range 10-100
		}

		// Calculate price based on description length
		price := 9.99 + float64(len(book.Description))/100.0
		if price > 19.99 {
			price = 19.99
		}

		// Escape single quotes in strings
		title := strings.ReplaceAll(book.Title, "'", "''")
		author := strings.ReplaceAll(book.Author, "'", "''")
		desc := strings.ReplaceAll(book.Description, "'", "''")

		sb.WriteString(fmt.Sprintf(`INSERT INTO products (name, description, price, sku, stock_quantity, category_id, status, author, popularity_score)
VALUES ('%s', '%s', %.2f, 'BOOK-%d', %d, (SELECT id FROM categories WHERE name = '%s'), 'active', '%s', %d)
ON CONFLICT (sku) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock_quantity = EXCLUDED.stock_quantity,
    category_id = EXCLUDED.category_id,
    author = EXCLUDED.author,
    popularity_score = EXCLUDED.popularity_score;

`, title, desc, price, book.GutenbergID, stockQty, book.Category, author, book.DownloadCount))
	}

	// Write to file
	err := os.WriteFile(outputFile, []byte(sb.String()), 0644)
	if err != nil {
		log.Fatalf("Failed to write SQL file: %v", err)
	}

	log.Printf("Generated %s with %d categories and %d books", outputFile, len(categories), len(books))
}

func getCategoryDescription(name string) string {
	descriptions := map[string]string{
		"Fiction":          "Novels and stories",
		"Non-Fiction":      "Factual books and biographies",
		"Science":          "Physics, Chemistry, Biology",
		"Technology":       "Computers and Programming",
		"Philosophy":       "Philosophical works and treatises",
		"Science Fiction":  "Science fiction and speculative fiction",
		"Drama":            "Plays and dramatic works",
		"Poetry":           "Poems and poetic works",
		"History":          "Historical accounts and biographies",
		"Political Science": "Political theory and governance",
	}
	if desc, ok := descriptions[name]; ok {
		return desc
	}
	return name
}

func seedDatabase(books []GutenbergBook) {
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

	for i, book := range books {
		// Determine category ID
		categoryID := categoryMap["Fiction"] // Default
		if catID, ok := categoryMap[book.Category]; ok {
			categoryID = catID
		}

		// Calculate stock quantity based on title for specific demos
		var stockQty int
		switch book.Title {
		case "A Christmas Carol":
			stockQty = 0 // Out of stock - first alphabetically
		case "Pride and Prejudice":
			stockQty = 3 // Low stock demo
		default:
			stockQty = 10 + ((i * 7) % 91) // Range 10-100
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
				SET description = $1, author = $2, category_id = $3, price = $4, stock_quantity = $5, popularity_score = $6
				WHERE id = $7`,
				book.Description, book.Author, categoryID, price, stockQty, book.DownloadCount, existingID)
			if err != nil {
				log.Printf("Error updating book '%s': %v", book.Title, err)
				continue
			}
			updateCount++
			log.Printf("Updated: %s by %s", book.Title, book.Author)
		} else {
			// Insert new book
			_, err := db.Exec(`
				INSERT INTO products (name, description, price, sku, stock_quantity, category_id, status, author, popularity_score)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
				book.Title, book.Description, price, fmt.Sprintf("BOOK-%d", book.GutenbergID),
				stockQty, categoryID, "active", book.Author, book.DownloadCount)
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
