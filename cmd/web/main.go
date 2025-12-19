package main

import (
	"DemoApp/internal/handlers"
	"DemoApp/internal/repository"
	"DemoApp/internal/storage"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/rbcervilla/redisstore/v9"
	"github.com/redis/go-redis/v9"
)

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	esURL := os.Getenv("ES_URL")
	redisURL := os.Getenv("REDIS_URL")
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioUseSSL := os.Getenv("MINIO_USE_SSL") == "true"

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize session store (Redis if available, fallback to cookie store)
	var store sessions.Store
	var redisClient *redis.Client

	if redisURL != "" {
		log.Println("Initializing Redis...")
		redisClient = redis.NewClient(&redis.Options{
			Addr: redisURL,
		})

		// Test Redis connection
		ctx := context.Background()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Printf("Warning: Redis connection failed: %v", err)
			log.Println("Falling back to cookie-based sessions and no caching")
			store = sessions.NewCookieStore([]byte("something-very-secret"))
			redisClient = nil
		} else {
			log.Println("Redis connected successfully")
			redisStore, err := redisstore.NewRedisStore(ctx, redisClient)
			if err != nil {
				log.Printf("Warning: Redis store initialization failed: %v", err)
				log.Println("Falling back to cookie-based sessions")
				store = sessions.NewCookieStore([]byte("something-very-secret"))
			} else {
				store = redisStore
				log.Println("Using Redis for session storage and caching")
			}
		}
	} else {
		log.Println("REDIS_URL not set, using cookie-based sessions and no caching")
		store = sessions.NewCookieStore([]byte("something-very-secret"))
	}

	// Configure session options for development (allow HTTP, not just HTTPS)
	// Note: Options() method is only available on specific store implementations
	if cookieStore, ok := store.(*sessions.CookieStore); ok {
		cookieStore.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 30, // 30 days
			HttpOnly: true,
			Secure:   false, // Set to true in production with HTTPS
			SameSite: http.SameSiteLaxMode,
		}
	} else if redisStore, ok := store.(*redisstore.RedisStore); ok {
		redisStore.Options(sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 30, // 30 days
			HttpOnly: true,
			Secure:   false, // Set to true in production with HTTPS
			SameSite: http.SameSiteLaxMode,
		})
	}

	repo := repository.NewPostgresRepository(db)

	// Wrap product repository with caching if Redis is available
	if redisClient != nil {
		log.Println("Enabling product caching with Redis")
		repo.SetCachedProducts(repository.NewCachedProductRepository(repo.Products(), redisClient))
	}

	// Initialize Elasticsearch if URL is provided
	if esURL != "" {
		log.Println("Initializing Elasticsearch...")
		es, err := repository.NewElasticsearchRepository([]string{esURL})
		if err != nil {
			log.Printf("Warning: Elasticsearch initialization failed: %v", err)
			log.Println("Continuing without Elasticsearch (will use SQL search)")
		} else {
			repo.SetElasticsearch(es)
			log.Println("Elasticsearch initialized successfully")

			// Index all products on startup
			go func() {
				log.Println("Indexing products to Elasticsearch...")
				products, err := repo.Products().ListProducts()
				if err != nil {
					log.Printf("Error listing products for indexing: %v", err)
					return
				}
				if err := es.IndexProducts(products); err != nil {
					log.Printf("Error indexing products: %v", err)
					return
				}
				log.Printf("Successfully indexed %d products", len(products))
			}()
		}
	} else {
		log.Println("ES_URL not set, using SQL-based search")
	}

	// Initialize MinIO storage
	var minioStorage *storage.MinIOStorage
	var imageHandlers *handlers.ImageHandlers
	if minioEndpoint != "" {
		log.Println("Initializing MinIO...")
		minioStorage, err = storage.NewMinIOStorage(minioEndpoint, minioAccessKey, minioSecretKey, minioUseSSL)
		if err != nil {
			log.Printf("Warning: MinIO initialization failed: %v", err)
			log.Println("Continuing without MinIO storage")
		} else {
			log.Println("MinIO storage initialized successfully")
			imageHandlers = &handlers.ImageHandlers{
				Storage: minioStorage,
			}
		}
	} else {
		log.Println("MINIO_ENDPOINT not set, MinIO storage disabled")
	}

	h := &handlers.Handlers{
		Repo:  repo,
		Store: store,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.ListProducts)
	mux.HandleFunc("/products/{id}", h.ProductDetail)
	mux.HandleFunc("/cart/add", h.AddToCart)
	mux.HandleFunc("/cart/update", h.UpdateCartQuantity)
	mux.HandleFunc("/cart/remove", h.RemoveFromCart)
	mux.HandleFunc("/cart", h.ViewCart)
	mux.HandleFunc("/checkout", h.CheckoutPage)
	mux.HandleFunc("/checkout/process", h.ProcessOrder)
	mux.HandleFunc("/confirmation", h.ConfirmationPage)
	mux.HandleFunc("/partials/cart-count", h.CartCount)
	mux.HandleFunc("/partials/cart-summary", h.CartSummary)
	mux.HandleFunc("/partials/search-suggestions", h.SearchSuggestions)

	mux.HandleFunc("/signup", h.SignupPage)
	mux.HandleFunc("/signup/process", h.Signup)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		h.LoginPage(w, r, "")
	})
	mux.HandleFunc("/login/process", h.Login)
	mux.HandleFunc("/logout", h.Logout)
	mux.HandleFunc("/orders", h.MyOrders)

	// Profile routes
	mux.HandleFunc("/profile", h.ProfilePage)
	mux.HandleFunc("/profile/edit", h.ProfileEditPage)
	mux.HandleFunc("/profile/update", h.UpdateProfile)
	mux.HandleFunc("/profile/password", h.ProfilePasswordPage)
	mux.HandleFunc("/profile/password/update", h.UpdatePassword)

	// Review routes
	mux.HandleFunc("/products/{id}/review", h.SubmitReview)
	mux.HandleFunc("/reviews/{id}/delete", h.DeleteReview)

	// Image routes (MinIO)
	if imageHandlers != nil {
		mux.HandleFunc("/images/", imageHandlers.ServeImage)
		mux.HandleFunc("/admin/upload-image", imageHandlers.UploadImage)
	}

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}
