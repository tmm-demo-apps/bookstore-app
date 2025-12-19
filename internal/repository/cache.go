package repository

import (
	"DemoApp/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// CachedProductRepository wraps ProductRepository with Redis caching
type CachedProductRepository struct {
	repo  ProductRepository
	redis *redis.Client
	ctx   context.Context
}

func NewCachedProductRepository(repo ProductRepository, redisClient *redis.Client) *CachedProductRepository {
	return &CachedProductRepository{
		repo:  repo,
		redis: redisClient,
		ctx:   context.Background(),
	}
}

func (c *CachedProductRepository) GetProductByID(id int) (*models.Product, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("product:%d", id)
	cached, err := c.redis.Get(c.ctx, cacheKey).Result()

	if err == nil {
		// Cache hit
		var product models.Product
		if err := json.Unmarshal([]byte(cached), &product); err == nil {
			log.Printf("Cache HIT: product:%d", id)
			return &product, nil
		}
	}

	// Cache miss - fetch from database
	log.Printf("Cache MISS: product:%d", id)
	product, err := c.repo.GetProductByID(id)
	if err != nil {
		return nil, err
	}

	// Store in cache (TTL: 5 minutes)
	if data, err := json.Marshal(product); err == nil {
		c.redis.Set(c.ctx, cacheKey, data, 5*time.Minute)
	}

	return product, nil
}

func (c *CachedProductRepository) ListProducts() ([]models.Product, error) {
	// Try cache first
	cacheKey := "products:all"
	cached, err := c.redis.Get(c.ctx, cacheKey).Result()

	if err == nil {
		// Cache hit
		var products []models.Product
		if err := json.Unmarshal([]byte(cached), &products); err == nil {
			log.Printf("Cache HIT: products:all (%d products)", len(products))
			return products, nil
		}
	}

	// Cache miss - fetch from database
	log.Println("Cache MISS: products:all")
	products, err := c.repo.ListProducts()
	if err != nil {
		return nil, err
	}

	// Store in cache (TTL: 2 minutes for list)
	if data, err := json.Marshal(products); err == nil {
		c.redis.Set(c.ctx, cacheKey, data, 2*time.Minute)
	}

	return products, nil
}

func (c *CachedProductRepository) SearchProducts(query string, categoryID int) ([]models.Product, error) {
	// Search results are not cached (too many variations)
	return c.repo.SearchProducts(query, categoryID)
}

func (c *CachedProductRepository) ListCategories() ([]models.Category, error) {
	// Try cache first
	cacheKey := "categories:all"
	cached, err := c.redis.Get(c.ctx, cacheKey).Result()

	if err == nil {
		// Cache hit
		var categories []models.Category
		if err := json.Unmarshal([]byte(cached), &categories); err == nil {
			log.Printf("Cache HIT: categories:all (%d categories)", len(categories))
			return categories, nil
		}
	}

	// Cache miss - fetch from database
	log.Println("Cache MISS: categories:all")
	categories, err := c.repo.ListCategories()
	if err != nil {
		return nil, err
	}

	// Store in cache (TTL: 10 minutes for categories)
	if data, err := json.Marshal(categories); err == nil {
		c.redis.Set(c.ctx, cacheKey, data, 10*time.Minute)
	}

	return categories, nil
}

// InvalidateProduct removes a product from cache
func (c *CachedProductRepository) InvalidateProduct(id int) {
	cacheKey := fmt.Sprintf("product:%d", id)
	c.redis.Del(c.ctx, cacheKey)
	// Also invalidate the products list
	c.redis.Del(c.ctx, "products:all")
	log.Printf("Cache INVALIDATED: product:%d", id)
}

// InvalidateAllProducts clears all product caches
func (c *CachedProductRepository) InvalidateAllProducts() {
	c.redis.Del(c.ctx, "products:all")
	log.Println("Cache INVALIDATED: products:all")
}
