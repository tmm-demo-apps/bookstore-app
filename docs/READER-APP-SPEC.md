# Library Reader App Specification

## Overview

The Library Reader is a **standalone microservice** that allows users to read books they've purchased from the Bookstore. It demonstrates:
- **Service-to-service communication**: Verifies purchases via Bookstore API
- **Shared storage**: EPUB files cached in MinIO (shared with Bookstore)
- **User data persistence**: Reading progress stored per-user
- **Multi-app dependency**: Reader depends on Bookstore for purchase verification

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Library Reader Service                             │
│                              (Go + HTMX + Pico.css)                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌────────────────┐     ┌────────────────┐     ┌────────────────────────┐   │
│  │    Browser     │────▶│   Go Backend   │────▶│   MinIO (EPUB Store)   │   │
│  │  HTMX + CSS    │     │   Port 8081    │     │   bucket: books-epub   │   │
│  └────────────────┘     └───────┬────────┘     └────────────────────────┘   │
│                                 │                                            │
│                    ┌────────────┼────────────┐                              │
│                    │            │            │                              │
│                    ▼            ▼            ▼                              │
│            ┌────────────┐ ┌──────────┐ ┌────────────────────┐               │
│            │ PostgreSQL │ │  Redis   │ │   Bookstore API    │               │
│            │  (reader)  │ │ (session)│ │ (verify purchase)  │               │
│            └────────────┘ └──────────┘ └────────────────────┘               │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Backend | Go 1.25 | API server, EPUB processing |
| Frontend | HTMX + Pico.css | Dynamic UI, consistent styling |
| EPUB Parsing | readium/go-toolkit | Parse EPUB2/EPUB3 files |
| Database | PostgreSQL | Reading progress, user library |
| Cache | Redis | Sessions, hot data caching |
| Object Storage | MinIO | EPUB file storage |
| Container | Docker | Packaging |
| Orchestration | Kubernetes | Deployment |
| Registry | Harbor | Image storage |
| GitOps | ArgoCD | Continuous deployment |

## Repository Structure

```
reader-app/
├── .github/
│   └── workflows/
│       └── ci.yml                 # CI/CD pipeline
├── cmd/
│   └── web/
│       └── main.go                # Application entrypoint
├── internal/
│   ├── epub/
│   │   ├── parser.go              # EPUB parsing with Readium
│   │   ├── fetcher.go             # Download from Gutenberg
│   │   └── cache.go               # MinIO caching logic
│   ├── handlers/
│   │   ├── base.go                # Shared handler dependencies
│   │   ├── library.go             # User library endpoints
│   │   ├── reader.go              # Reader UI endpoints
│   │   └── api.go                 # JSON API endpoints
│   ├── models/
│   │   ├── book.go                # Book metadata
│   │   └── progress.go            # Reading progress
│   ├── repository/
│   │   ├── postgres.go            # Database operations
│   │   └── bookstore_client.go    # Bookstore API client
│   └── storage/
│       └── minio.go               # MinIO operations
├── migrations/
│   └── 001_schema.sql             # Database schema
├── templates/
│   ├── base.html                  # Base layout
│   ├── library.html               # User's book library
│   ├── reader.html                # EPUB reader interface
│   └── partials/
│       ├── toc.html               # Table of contents
│       └── chapter.html           # Chapter content
├── static/
│   └── css/
│       └── reader.css             # Reader-specific styles
├── kubernetes/
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   ├── configmap.yaml
│   ├── postgres.yaml
│   └── argocd-application.yaml
├── scripts/
│   └── local-dev.sh               # Local development script
├── tests/
│   └── smoke.sh                   # Smoke tests
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Database Schema

```sql
-- migrations/001_schema.sql

-- Reading progress tracking
CREATE TABLE IF NOT EXISTS reading_progress (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    book_sku VARCHAR(50) NOT NULL,           -- e.g., BOOK-1342
    gutenberg_id INTEGER NOT NULL,           -- e.g., 1342
    chapter_index INTEGER DEFAULT 0,         -- Current chapter (0-based)
    chapter_href VARCHAR(255) DEFAULT '',    -- Chapter file path in EPUB
    position_percent DECIMAL(5,2) DEFAULT 0.00,  -- 0.00 to 100.00
    scroll_position INTEGER DEFAULT 0,       -- Pixel offset within chapter
    last_read_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, book_sku)
);

-- User library (synced from Bookstore purchases)
CREATE TABLE IF NOT EXISTS user_library (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    book_sku VARCHAR(50) NOT NULL,
    gutenberg_id INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    author VARCHAR(255) NOT NULL,
    cover_url VARCHAR(500),                  -- MinIO URL for cover image
    acquired_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, book_sku)
);

-- EPUB cache metadata (tracks what's in MinIO)
CREATE TABLE IF NOT EXISTS epub_cache (
    id SERIAL PRIMARY KEY,
    gutenberg_id INTEGER NOT NULL UNIQUE,
    book_sku VARCHAR(50) NOT NULL,
    minio_path VARCHAR(255) NOT NULL,        -- e.g., books-epub/1342/pg1342.epub
    file_size_bytes BIGINT,
    cached_at TIMESTAMP DEFAULT NOW(),
    last_accessed_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_reading_progress_user ON reading_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_user_library_user ON user_library(user_id);
CREATE INDEX IF NOT EXISTS idx_epub_cache_gutenberg ON epub_cache(gutenberg_id);
```

## API Endpoints

### Library Endpoints (HTML)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Redirect to /library |
| GET | `/library` | User's book library |
| GET | `/read/{sku}` | Open reader for book |

### Reader Endpoints (HTML + HTMX)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/read/{sku}` | Full reader page |
| GET | `/read/{sku}/toc` | Table of contents (HTMX partial) |
| GET | `/read/{sku}/chapter/{index}` | Chapter content (HTMX partial) |
| POST | `/read/{sku}/progress` | Save reading progress (HTMX) |

### JSON API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/library` | List user's books (JSON) |
| GET | `/api/books/{sku}/metadata` | Book metadata (JSON) |
| GET | `/api/books/{sku}/progress` | Get reading progress |
| PUT | `/api/books/{sku}/progress` | Update reading progress |
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |

### Internal Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/internal/sync-library` | Sync user library from Bookstore |

## Core Components

### 1. EPUB Parser (internal/epub/parser.go)

Uses the Readium Go Toolkit for EPUB parsing:

```go
package epub

import (
    "github.com/readium/go-toolkit/pkg/epub"
    "github.com/readium/go-toolkit/pkg/manifest"
)

type Book struct {
    GutenbergID int
    SKU         string
    Title       string
    Author      string
    Language    string
    TOC         []TOCEntry
    Chapters    []Chapter
}

type TOCEntry struct {
    Title string
    Href  string
    Level int      // Nesting level (1 = top level)
}

type Chapter struct {
    Index   int
    Href    string
    Title   string
    Content string  // HTML content
}

type Parser struct {
    minioClient *minio.Client
}

func (p *Parser) Parse(gutenbergID int) (*Book, error) {
    // 1. Fetch EPUB from MinIO (or download if not cached)
    // 2. Parse container.xml to find OPF
    // 3. Parse OPF for metadata and spine
    // 4. Parse NCX or EPUB3 nav for TOC
    // 5. Return Book struct with all data
}

func (p *Parser) GetChapter(gutenbergID int, chapterIndex int) (*Chapter, error) {
    // Return specific chapter content as HTML
}
```

### 2. EPUB Fetcher (internal/epub/fetcher.go)

Downloads EPUBs from Project Gutenberg and caches in MinIO:

```go
package epub

const GutenbergEPUBURL = "https://www.gutenberg.org/ebooks/%d.epub3.images"

type Fetcher struct {
    minioClient *minio.Client
    db          *sql.DB
}

func (f *Fetcher) EnsureCached(gutenbergID int) (string, error) {
    // 1. Check if EPUB exists in epub_cache table
    // 2. If exists, verify MinIO object exists
    // 3. If not cached:
    //    a. Download from Gutenberg
    //    b. Upload to MinIO bucket "books-epub"
    //    c. Insert record into epub_cache
    // 4. Update last_accessed_at
    // 5. Return MinIO path
}

func (f *Fetcher) downloadFromGutenberg(gutenbergID int) ([]byte, error) {
    url := fmt.Sprintf(GutenbergEPUBURL, gutenbergID)
    resp, err := http.Get(url)
    // Handle redirects, errors, etc.
    return io.ReadAll(resp.Body)
}
```

### 3. Bookstore API Client (internal/repository/bookstore_client.go)

Communicates with Bookstore to verify purchases:

```go
package repository

type BookstoreClient struct {
    baseURL    string  // http://bookstore-service.bookstore:8080
    httpClient *http.Client
}

type PurchasedBook struct {
    SKU         string    `json:"sku"`
    GutenbergID int       `json:"gutenberg_id"`
    Title       string    `json:"title"`
    Author      string    `json:"author"`
    CoverURL    string    `json:"cover_url"`
    PurchasedAt time.Time `json:"purchased_at"`
}

// VerifyPurchase checks if user owns the book
func (c *BookstoreClient) VerifyPurchase(userID int, bookSKU string) (bool, error) {
    url := fmt.Sprintf("%s/api/purchases/%d/%s", c.baseURL, userID, bookSKU)
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return false, err
    }
    return resp.StatusCode == http.StatusOK, nil
}

// GetUserPurchases returns all books owned by user
func (c *BookstoreClient) GetUserPurchases(userID int) ([]PurchasedBook, error) {
    url := fmt.Sprintf("%s/api/purchases/%d", c.baseURL, userID)
    resp, err := c.httpClient.Get(url)
    // Parse JSON response
}
```

### 4. Library Handler (internal/handlers/library.go)

```go
package handlers

func (h *Handlers) Library(w http.ResponseWriter, r *http.Request) {
    userID := h.getUserID(r)
    if userID == 0 {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
    
    // Get user's library from local DB (synced from Bookstore)
    books, err := h.repo.GetUserLibrary(userID)
    if err != nil {
        // Handle error
    }
    
    // Get reading progress for each book
    for i := range books {
        progress, _ := h.repo.GetReadingProgress(userID, books[i].SKU)
        books[i].Progress = progress
    }
    
    h.render(w, "library.html", map[string]interface{}{
        "Books": books,
    })
}
```

### 5. Reader Handler (internal/handlers/reader.go)

```go
package handlers

func (h *Handlers) Reader(w http.ResponseWriter, r *http.Request) {
    userID := h.getUserID(r)
    bookSKU := chi.URLParam(r, "sku")
    
    // Verify user owns the book
    owned, err := h.bookstoreClient.VerifyPurchase(userID, bookSKU)
    if !owned {
        http.Error(w, "Book not in your library", http.StatusForbidden)
        return
    }
    
    // Get book metadata from library
    book, err := h.repo.GetLibraryBook(userID, bookSKU)
    if err != nil {
        http.Error(w, "Book not found", http.StatusNotFound)
        return
    }
    
    // Ensure EPUB is cached
    _, err = h.fetcher.EnsureCached(book.GutenbergID)
    if err != nil {
        http.Error(w, "Failed to load book", http.StatusInternalServerError)
        return
    }
    
    // Parse EPUB for TOC
    parsed, err := h.parser.Parse(book.GutenbergID)
    if err != nil {
        http.Error(w, "Failed to parse book", http.StatusInternalServerError)
        return
    }
    
    // Get reading progress
    progress, _ := h.repo.GetReadingProgress(userID, bookSKU)
    
    h.render(w, "reader.html", map[string]interface{}{
        "Book":     book,
        "TOC":      parsed.TOC,
        "Progress": progress,
    })
}

func (h *Handlers) Chapter(w http.ResponseWriter, r *http.Request) {
    bookSKU := chi.URLParam(r, "sku")
    chapterIndex, _ := strconv.Atoi(chi.URLParam(r, "index"))
    
    book, _ := h.repo.GetLibraryBookBySKU(bookSKU)
    chapter, err := h.parser.GetChapter(book.GutenbergID, chapterIndex)
    if err != nil {
        http.Error(w, "Chapter not found", http.StatusNotFound)
        return
    }
    
    // Return HTML partial for HTMX
    h.render(w, "partials/chapter.html", map[string]interface{}{
        "Chapter": chapter,
    })
}

func (h *Handlers) SaveProgress(w http.ResponseWriter, r *http.Request) {
    userID := h.getUserID(r)
    bookSKU := chi.URLParam(r, "sku")
    
    var progress struct {
        ChapterIndex    int     `json:"chapter_index"`
        ChapterHref     string  `json:"chapter_href"`
        PositionPercent float64 `json:"position_percent"`
        ScrollPosition  int     `json:"scroll_position"`
    }
    json.NewDecoder(r.Body).Decode(&progress)
    
    err := h.repo.SaveReadingProgress(userID, bookSKU, progress)
    if err != nil {
        http.Error(w, "Failed to save progress", http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}
```

## UI Design

### Library Page (templates/library.html)

```
┌─────────────────────────────────────────────────────────────────┐
│  📚 My Library                                    [Back to Shop] │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │              │  │              │  │              │          │
│  │  [Cover]     │  │  [Cover]     │  │  [Cover]     │          │
│  │              │  │              │  │              │          │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤          │
│  │ Pride and    │  │ Moby Dick    │  │ War and      │          │
│  │ Prejudice    │  │              │  │ Peace        │          │
│  │ J. Austen    │  │ H. Melville  │  │ L. Tolstoy   │          │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤          │
│  │ ████████░░   │  │ ██░░░░░░░░   │  │ Not started  │          │
│  │ 80% complete │  │ 15% complete │  │              │          │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤          │
│  │ [Continue]   │  │ [Continue]   │  │ [Start]      │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Reader Page (templates/reader.html)

```
┌─────────────────────────────────────────────────────────────────┐
│  ← Library    Pride and Prejudice    Ch 12/61    [A-] [A+] [☰] │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌───────────────────────────────────────────────┐ │
│  │ Contents│  │                                               │ │
│  ├─────────┤  │   Chapter XII                                 │ │
│  │ Ch 1    │  │                                               │ │
│  │ Ch 2    │  │   In consequence of an agreement between      │ │
│  │ ...     │  │   the sisters, Elizabeth wrote the next       │ │
│  │ Ch 11   │  │   morning to their mother, to beg that the    │ │
│  │►Ch 12   │  │   carriage might be sent for them in the      │ │
│  │ Ch 13   │  │   course of the day. But Mrs. Bennet, who     │ │
│  │ ...     │  │   had calculated on her daughters remaining   │ │
│  │ Ch 61   │  │   at Netherfield till the following Tuesday,  │ │
│  │         │  │   which would exactly finish Jane's week...   │ │
│  │         │  │                                               │ │
│  │         │  │                                               │ │
│  │         │  │   [Prev Page]          [Next Page]            │ │
│  └─────────┘  └───────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│  ████████████████████████░░░░░░░░░░░░░░░░░░  45% complete       │
└─────────────────────────────────────────────────────────────────┘
```

### Mobile Layout

```
┌─────────────────────────────┐
│  ←  Pride and Prejudice  ☰ │
├─────────────────────────────┤
│                             │
│   Chapter XII               │
│                             │
│   In consequence of an      │
│   agreement between the     │
│   sisters, Elizabeth wrote  │
│   the next morning to their │
│   mother, to beg that the   │
│   carriage might be sent    │
│   for them in the course    │
│   of the day...             │
│                             │
│   [Prev]    Ch 12/61  [Next]│
├─────────────────────────────┤
│  ████████████░░░░░  45%     │
└─────────────────────────────┘
```

## Feature Details

### Font Size Adjustment

```javascript
// Reader JavaScript
function changeFontSize(delta) {
    const content = document.getElementById('chapter-content');
    const current = parseFloat(getComputedStyle(content).fontSize);
    const newSize = Math.max(12, Math.min(32, current + delta));
    content.style.fontSize = newSize + 'px';
    localStorage.setItem('reader-font-size', newSize);
}

// Restore on page load
document.addEventListener('DOMContentLoaded', () => {
    const saved = localStorage.getItem('reader-font-size');
    if (saved) {
        document.getElementById('chapter-content').style.fontSize = saved + 'px';
    }
});
```

### Reading Progress Auto-Save

```javascript
// Save progress every 30 seconds while reading
let saveTimeout;

function scheduleProgressSave() {
    clearTimeout(saveTimeout);
    saveTimeout = setTimeout(saveProgress, 30000);
}

function saveProgress() {
    const data = {
        chapter_index: currentChapter,
        chapter_href: currentHref,
        position_percent: calculateProgressPercent(),
        scroll_position: document.getElementById('chapter-content').scrollTop
    };
    
    fetch(`/read/${bookSKU}/progress`, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(data)
    });
}

// Also save when navigating away
window.addEventListener('beforeunload', saveProgress);
document.getElementById('chapter-content').addEventListener('scroll', scheduleProgressSave);
```

### HTMX Chapter Navigation

```html
<!-- Chapter content container -->
<div id="chapter-content"
     hx-get="/read/{{ .Book.SKU }}/chapter/{{ .Progress.ChapterIndex }}"
     hx-trigger="load"
     hx-swap="innerHTML">
    Loading chapter...
</div>

<!-- Navigation buttons -->
<button hx-get="/read/{{ .Book.SKU }}/chapter/{{ sub .CurrentChapter 1 }}"
        hx-target="#chapter-content"
        hx-swap="innerHTML"
        {{ if eq .CurrentChapter 0 }}disabled{{ end }}>
    Previous
</button>

<button hx-get="/read/{{ .Book.SKU }}/chapter/{{ add .CurrentChapter 1 }}"
        hx-target="#chapter-content"
        hx-swap="innerHTML"
        {{ if eq .CurrentChapter .TotalChapters }}disabled{{ end }}>
    Next
</button>
```

## Kubernetes Deployment

### deployment.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: reader-deployment
  labels:
    app: reader
spec:
  replicas: 2
  selector:
    matchLabels:
      app: reader
  template:
    metadata:
      labels:
        app: reader
    spec:
      containers:
      - name: reader
        image: harbor.corp.vmbeans.com/bookstore/reader:latest
        ports:
        - containerPort: 8081
        env:
        - name: PORT
          value: "8081"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: reader-secrets
              key: database-url
        - name: REDIS_URL
          value: "redis://redis-service:6379"
        - name: MINIO_ENDPOINT
          value: "minio-service:9000"
        - name: MINIO_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: reader-secrets
              key: minio-access-key
        - name: MINIO_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: reader-secrets
              key: minio-secret-key
        - name: BOOKSTORE_API_URL
          value: "http://bookstore-service.bookstore:8080"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
      imagePullSecrets:
      - name: harbor-registry-secret
```

### ingress.yaml

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: reader-ingress
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
spec:
  ingressClassName: nginx
  rules:
  - host: reader.corp.vmbeans.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: reader-service
            port:
              number: 8081
```

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| PORT | Server port | 8081 |
| DATABASE_URL | PostgreSQL connection | postgres://user:pass@host:5432/reader |
| REDIS_URL | Redis connection | redis://redis-service:6379 |
| MINIO_ENDPOINT | MinIO server | minio-service:9000 |
| MINIO_ACCESS_KEY | MinIO access key | minioadmin |
| MINIO_SECRET_KEY | MinIO secret key | minioadmin |
| MINIO_BUCKET | EPUB bucket name | books-epub |
| MINIO_USE_SSL | Use HTTPS for MinIO | false |
| BOOKSTORE_API_URL | Bookstore service URL | http://bookstore-service.bookstore:8080 |
| SESSION_SECRET | Session encryption key | (32+ random chars) |

## Bookstore API Requirements

The Reader requires these endpoints to be added to the Bookstore:

### GET /api/purchases/{user_id}

Returns all books purchased by the user.

```json
{
  "purchases": [
    {
      "sku": "BOOK-1342",
      "gutenberg_id": 1342,
      "title": "Pride and Prejudice",
      "author": "Jane Austen",
      "cover_url": "http://minio:9000/product-images/BOOK-1342.jpg",
      "purchased_at": "2026-01-15T10:30:00Z"
    }
  ]
}
```

### GET /api/purchases/{user_id}/{sku}

Verifies if user owns a specific book.

- Returns `200 OK` if user owns the book
- Returns `404 Not Found` if user doesn't own the book

## Security Considerations

1. **Authentication**: Uses same session cookies as Bookstore (shared Redis)
2. **Authorization**: Always verify purchase before serving content
3. **Rate Limiting**: Limit Gutenberg downloads to prevent abuse
4. **Input Validation**: Sanitize book SKU and chapter index parameters
5. **EPUB Sandboxing**: Render HTML content in sanitized container
6. **Cross-Origin**: Only allow requests from same domain

## Testing Strategy

### Smoke Tests (tests/smoke.sh)

1. Health check endpoint responds
2. Library page loads for authenticated user
3. Reader page loads for owned book
4. Reader returns 403 for unowned book
5. Chapter content loads via HTMX
6. Progress saves successfully
7. EPUB caching works (check MinIO)

### Integration Tests

1. Bookstore API client communicates correctly
2. EPUB parser handles various book formats
3. MinIO caching stores and retrieves EPUBs
4. Reading progress persists across sessions

## Demo Script

1. **Show Bookstore**: Purchase a book (e.g., "Pride and Prejudice")
2. **Open Reader**: Click "Read" button → redirect to Reader app
3. **Show Library**: User's purchased books with progress indicators
4. **Start Reading**: Click "Start Reading" → opens reader view
5. **Navigate TOC**: Click chapter in sidebar → chapter loads
6. **Adjust Font**: Click A+/A- buttons → font size changes
7. **Navigate Pages**: Use Next/Prev buttons → chapters change
8. **Progress Saves**: Close and reopen → returns to same position
9. **Multi-App Story**: "Different teams built Bookstore and Reader"
10. **Show ArgoCD**: Both apps managed independently in GitOps

## Success Criteria

- [ ] User can see purchased books in library
- [ ] User can open and read any purchased book
- [ ] Table of contents navigation works
- [ ] Page/chapter navigation works
- [ ] Font size adjustment works
- [ ] Reading progress saves and restores
- [ ] EPUB files are cached in MinIO
- [ ] Unauthorized access is blocked
- [ ] Mobile-responsive design works
- [ ] All smoke tests pass
