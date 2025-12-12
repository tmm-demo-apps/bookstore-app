package repository

import (
	"DemoApp/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

const (
	productIndex = "products"
)

type ElasticsearchRepository struct {
	client *elasticsearch.Client
}

func NewElasticsearchRepository(addresses []string) (*ElasticsearchRepository, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating Elasticsearch client: %w", err)
	}

	// Check cluster health
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("error getting Elasticsearch info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	repo := &ElasticsearchRepository{client: es}

	// Initialize index
	if err := repo.initializeIndex(); err != nil {
		return nil, fmt.Errorf("error initializing index: %w", err)
	}

	return repo, nil
}

func (r *ElasticsearchRepository) initializeIndex() error {
	// Check if index exists
	res, err := r.client.Indices.Exists([]string{productIndex})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// If index exists, return
	if res.StatusCode == 200 {
		log.Println("Elasticsearch index 'products' already exists")
		return nil
	}

	// Create index with mappings
	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"autocomplete": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": ["lowercase", "autocomplete_filter"]
					},
					"autocomplete_search": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": ["lowercase"]
					}
				},
				"filter": {
					"autocomplete_filter": {
						"type": "edge_ngram",
						"min_gram": 2,
						"max_gram": 20
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"id": { "type": "integer" },
				"name": { 
					"type": "text",
					"analyzer": "autocomplete",
					"search_analyzer": "autocomplete_search",
					"fields": {
						"keyword": { "type": "keyword" },
						"standard": { 
							"type": "text",
							"analyzer": "standard"
						}
					}
				},
				"description": { 
					"type": "text",
					"analyzer": "standard"
				},
				"price": { "type": "float" },
				"sku": { "type": "keyword" },
				"stock_quantity": { "type": "integer" },
				"image_url": { "type": "keyword" },
				"category_id": { "type": "integer" },
				"status": { "type": "keyword" }
			}
		}
	}`

	req := esapi.IndicesCreateRequest{
		Index: productIndex,
		Body:  strings.NewReader(mapping),
	}

	res, err = req.Do(context.Background(), r.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error creating index: %s", res.String())
	}

	log.Println("Elasticsearch index 'products' created successfully")
	return nil
}

// IndexProduct indexes or updates a single product
func (r *ElasticsearchRepository) IndexProduct(product models.Product) error {
	// Convert product to JSON
	data, err := json.Marshal(map[string]interface{}{
		"id":             product.ID,
		"name":           product.Name,
		"description":    product.Description,
		"price":          product.Price,
		"sku":            product.SKU,
		"stock_quantity": product.StockQuantity,
		"image_url":      product.ImageURL,
		"category_id":    product.CategoryID,
		"status":         product.Status,
	})
	if err != nil {
		return fmt.Errorf("error marshaling product: %w", err)
	}

	// Index the document with product ID as document ID
	req := esapi.IndexRequest{
		Index:      productIndex,
		DocumentID: strconv.Itoa(product.ID),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), r.client)
	if err != nil {
		return fmt.Errorf("error indexing document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	return nil
}

// IndexProducts indexes multiple products in bulk
func (r *ElasticsearchRepository) IndexProducts(products []models.Product) error {
	if len(products) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, product := range products {
		// Bulk API requires two lines per document:
		// 1. Action line (index operation)
		// 2. Document line (the actual data)
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": productIndex,
				"_id":    strconv.Itoa(product.ID),
			},
		}
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("error marshaling metadata: %w", err)
		}
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		doc := map[string]interface{}{
			"id":             product.ID,
			"name":           product.Name,
			"description":    product.Description,
			"price":          product.Price,
			"sku":            product.SKU,
			"stock_quantity": product.StockQuantity,
			"image_url":      product.ImageURL,
			"category_id":    product.CategoryID,
			"status":         product.Status,
		}
		docJSON, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("error marshaling document: %w", err)
		}
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	res, err := r.client.Bulk(bytes.NewReader(buf.Bytes()), r.client.Bulk.WithIndex(productIndex), r.client.Bulk.WithRefresh("true"))
	if err != nil {
		return fmt.Errorf("error executing bulk request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from bulk request: %s", res.String())
	}

	log.Printf("Successfully indexed %d products to Elasticsearch", len(products))
	return nil
}

// SearchProducts performs full-text search on products
func (r *ElasticsearchRepository) SearchProducts(query string, categoryID int) ([]int, error) {
	// Build search query
	var searchQuery map[string]interface{}

	if query == "" && categoryID == 0 {
		// Return all products
		searchQuery = map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
	} else {
		// Build bool query
		must := []map[string]interface{}{}
		should := []map[string]interface{}{}

		// Add text search if query provided
		if query != "" {
			// Use bool query with should clauses for better matching
			// 1. Autocomplete match on name (best for prefix matching, no fuzzy needed)
			should = append(should, map[string]interface{}{
				"match": map[string]interface{}{
					"name": map[string]interface{}{
						"query": query,
						"boost": 5,
					},
				},
			})

			// 2. Wildcard match on name for substring matching
			// This handles cases like "ast" in "Fast"
			should = append(should, map[string]interface{}{
				"wildcard": map[string]interface{}{
					"name.keyword": map[string]interface{}{
						"value":            "*" + query + "*",
						"boost":            4,
						"case_insensitive": true,
					},
				},
			})

			// 3. Query string with wildcards for name and description
			// This handles partial word matching like "dan" in "Daniel"
			should = append(should, map[string]interface{}{
				"query_string": map[string]interface{}{
					"query":            "*" + query + "*",
					"fields":           []string{"name", "description"},
					"default_operator": "AND",
					"boost":            3.5,
				},
			})

			// 4. Standard match on name.standard with conservative fuzzy for typos
			// Only apply fuzzy for queries 5+ chars to avoid false positives
			nameMatch := map[string]interface{}{
				"query": query,
				"boost": 3,
			}
			if len(query) >= 5 {
				nameMatch["fuzziness"] = "1" // Allow 1 typo for longer queries
			}
			should = append(should, map[string]interface{}{
				"match": map[string]interface{}{
					"name.standard": nameMatch,
				},
			})

			// 5. Match on description with conservative fuzzy
			descMatch := map[string]interface{}{
				"query": query,
				"boost": 1,
			}
			if len(query) >= 5 {
				descMatch["fuzziness"] = "1"
			}
			should = append(should, map[string]interface{}{
				"match": map[string]interface{}{
					"description": descMatch,
				},
			})

			// Add the should clauses with minimum_should_match
			must = append(must, map[string]interface{}{
				"bool": map[string]interface{}{
					"should":               should,
					"minimum_should_match": 1,
				},
			})
		}

		// Add category filter if provided
		if categoryID > 0 {
			must = append(must, map[string]interface{}{
				"term": map[string]interface{}{
					"category_id": categoryID,
				},
			})
		}

		searchQuery = map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": must,
				},
			},
		}
	}

	// Add size and source filtering
	searchQuery["size"] = 100
	searchQuery["_source"] = []string{"id"}

	// Convert to JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, fmt.Errorf("error encoding query: %w", err)
	}

	// Perform search
	res, err := r.client.Search(
		r.client.Search.WithContext(context.Background()),
		r.client.Search.WithIndex(productIndex),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("error performing search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error response from search: %s", res.String())
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Extract product IDs
	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	productIDs := make([]int, 0, len(hits))

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		id := int(source["id"].(float64))
		productIDs = append(productIDs, id)
	}

	return productIDs, nil
}

// DeleteProduct removes a product from the index
func (r *ElasticsearchRepository) DeleteProduct(productID int) error {
	req := esapi.DeleteRequest{
		Index:      productIndex,
		DocumentID: strconv.Itoa(productID),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), r.client)
	if err != nil {
		return fmt.Errorf("error deleting document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	return nil
}
