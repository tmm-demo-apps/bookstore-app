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
	"strings"
	"time"

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

		// Test Redis connection with retry
		ctx := context.Background()
		err := retryWithBackoff("Redis", 10, 1*time.Second, func() error {
			return redisClient.Ping(ctx).Err()
		})

		if err != nil {
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
		var es *repository.ElasticsearchRepository
		err := retryWithBackoff("Elasticsearch", 10, 2*time.Second, func() error {
			var initErr error
			es, initErr = repository.NewElasticsearchRepository([]string{esURL})
			return initErr
		})

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
		err := retryWithBackoff("MinIO", 10, 2*time.Second, func() error {
			var initErr error
			minioStorage, initErr = storage.NewMinIOStorage(minioEndpoint, minioAccessKey, minioSecretKey, minioUseSSL)
			return initErr
		})

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

	// Health check endpoints
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		// Check database connectivity
		if err := db.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("Database not ready"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Ready"))
	})

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

	// API routes for service-to-service communication (Reader app, Chatbot app)
	mux.HandleFunc("/api/purchases/", func(w http.ResponseWriter, r *http.Request) {
		// Route to appropriate handler based on path segments
		// /api/purchases/{user_id} -> GetUserPurchases
		// /api/purchases/{user_id}/{sku} -> VerifyPurchase
		path := r.URL.Path[len("/api/purchases/"):]
		parts := strings.Split(path, "/")
		if len(parts) >= 2 && parts[1] != "" {
			h.VerifyPurchase(w, r)
		} else {
			h.GetUserPurchases(w, r)
		}
	})
	mux.HandleFunc("/api/products", h.APIProducts)
	mux.HandleFunc("/api/products/", h.APIProducts)
	mux.HandleFunc("/api/categories", h.APICategories)

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}

// retryWithBackoff retries a function with exponential backoff
func retryWithBackoff(operation string, maxRetries int, initialDelay time.Duration, fn func() error) error {
	var err error
	delay := initialDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = fn()
		if err == nil {
			if attempt > 1 {
				log.Printf("%s: Connected successfully after %d attempt(s)", operation, attempt)
			}
			return nil
		}

		if attempt < maxRetries {
			log.Printf("%s: Connection attempt %d/%d failed: %v. Retrying in %v...",
				operation, attempt, maxRetries, err, delay)
			time.Sleep(delay)
			delay *= 2 // Exponential backoff
			if delay > 30*time.Second {
				delay = 30 * time.Second // Cap at 30 seconds
			}
		}
	}

	return fmt.Errorf("%s: failed after %d attempts: %w", operation, maxRetries, err)
}
