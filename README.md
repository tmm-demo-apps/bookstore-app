# DemoApp - E-commerce Platform for VCF 9.0 Demonstrations

[![CI](https://github.com/johnnyr0x/bookstore-app/workflows/CI/badge.svg)](https://github.com/johnnyr0x/bookstore-app/actions)
[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14-336791?logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis)](https://redis.io/)
[![Elasticsearch](https://img.shields.io/badge/Elasticsearch-8.11-005571?logo=elasticsearch)](https://www.elastic.co/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A production-ready e-commerce platform built to demonstrate **VMware Cloud Foundation (VCF) 9.0** capabilities. Features enterprise-grade infrastructure including Elasticsearch search, Redis caching, MinIO object storage, and real-world content from Project Gutenberg.

**🎯 Purpose**: Showcase VCF 9.0 Supervisor Services, VKS (vSphere Kubernetes Service), VKS Add-ons, and CNCF graduated projects through a realistic e-commerce application.

## ✨ Features

### User Features
- 📚 **71 Real Products** - 50+ public domain classics from Project Gutenberg with authentic covers
- 🔍 **Intelligent Search** - Elasticsearch 5-tier search strategy with author-aware queries and autocomplete
- ⭐ **User Reviews** - Star ratings (1-5) with privacy-protected display ("FirstName L.")
- 👤 **User Profiles** - Complete account management (view, edit, password change)
- 🛒 **Smart Shopping Cart** - Real-time updates with Redis-backed sessions
- 📦 **Order Management** - Complete checkout flow and order history
- 📄 **Pagination** - Configurable page sizes (10/20/30/40/50 items)
- 🎨 **Modern UI** - Responsive design with Pico CSS, sticky header, mobile-optimized

### Infrastructure Features
- 🚀 **Redis Integration** - Session management and product caching for horizontal scaling
- 🖼️ **MinIO Storage** - S3-compatible object storage with 1-year cache headers and ETags
- 🔎 **Elasticsearch** - Full-text search with edge n-gram tokenization and fuzzy matching
- 📊 **Repository Pattern** - Clean architecture with caching decorators
- 🧪 **25 Automated Tests** - Comprehensive smoke test suite covering all services
- 🐳 **Docker Compose** - Complete local development environment
- ☸️ **Kubernetes Ready** - Production deployment manifests included

## 🏗️ Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Backend** | Go 1.25 | High-performance application server |
| **Frontend** | HTMX + Pico CSS | Modern, lightweight UI with dynamic updates |
| **Database** | PostgreSQL 14 | Primary data store with 10 migrations |
| **Search** | Elasticsearch 8.11 | Full-text search with autocomplete |
| **Cache** | Redis 7 | Session management and hot data caching |
| **Storage** | MinIO | S3-compatible object storage for images |
| **Container** | Docker & Docker Compose | Local development and testing |
| **Orchestration** | Kubernetes | Production deployment (VKS ready) |

## 🚀 Quick Start

### Two Deployment Options

1. **Local Development** - Docker Compose on your machine (for testing)
2. **Production Deployment** - Kubernetes cluster (for demos)

### Local Development (Recommended for Testing)

```bash
# Start all services
./local-dev.sh start

# Run tests
./local-dev.sh test

# View logs
./local-dev.sh logs

# Stop services
./local-dev.sh stop
```

**See `docs/DEVELOPMENT-WORKFLOW.md` for complete guide**

### Production Deployment (Kubernetes)

```bash
# See docs/START-HERE.md for complete deployment guide
# Or docs/REMOTE-VM-DEPLOYMENT.md for remote VM workflow
```

### Prerequisites

- Docker and Docker Compose
- OR Go 1.24+ with PostgreSQL, Redis, Elasticsearch, and MinIO

### Option 1: Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/johnnyr0x/bookstore-app.git
cd bookstore-app

# Start all services
docker compose up --build -d

# Run smoke tests (25 tests)
./tests/smoke.sh

# Access the application
open http://localhost:8080
```

**Services Available**:
- **App**: http://localhost:8080
- **PostgreSQL**: localhost:5432 (user/password)
- **Redis**: localhost:6379
- **Elasticsearch**: http://localhost:9200
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)

**⚠️ Security Note**: Default credentials are for development only. **Never use in production!**

### Option 2: Local Go Development

```bash
# Set environment variables
export DB_USER=user
export DB_PASSWORD=password
export DB_HOST=localhost
export DB_NAME=bookstore
export REDIS_URL=localhost:6379
export ES_URL=http://localhost:9200
export MINIO_ENDPOINT=localhost:9000
export MINIO_ACCESS_KEY=minioadmin
export MINIO_SECRET_KEY=minioadmin

# Start infrastructure services
docker compose up -d db redis elasticsearch minio

# Run the application
go run cmd/web/main.go

# Access the application
open http://localhost:8080
```

## 📊 Project Structure

```
bookstore-app/
├── cmd/web/              # Application entry point
├── internal/
│   ├── handlers/         # HTTP request handlers
│   │   ├── products.go   # Product listing, search, pagination
│   │   ├── cart.go       # Shopping cart operations
│   │   ├── profile.go    # User profile management
│   │   ├── reviews.go    # Review submission and display
│   │   └── images.go     # MinIO image serving
│   ├── models/          # Data models (Product, User, Review, etc.)
│   ├── repository/      # Database layer with caching
│   │   ├── postgres.go   # PostgreSQL implementation
│   │   ├── elasticsearch.go # Search implementation
│   │   └── cache.go      # Redis caching decorator
│   └── storage/         # Object storage
│       └── minio.go      # MinIO client
├── templates/           # HTML templates
├── migrations/          # Database migrations (10 files)
├── scripts/             # Data seeding scripts
│   ├── seed-gutenberg-books.go  # Project Gutenberg integration
│   ├── seed-images.go           # Image download and upload
│   └── README.md                # Scripts documentation
├── tests/               # Testing scripts and documentation
│   ├── smoke.sh         # Automated test suite (25 tests)
│   ├── redis-cache.sh   # Redis caching tests
│   ├── redis-sessions.sh # Redis session tests
│   └── redis-performance.sh # Redis performance tests
├── kubernetes/          # Kubernetes manifests
├── docs/                # Documentation
│   ├── ADMIN-CONSOLE-PLAN.md  # Admin feature plan
│   ├── AI-ASSISTANT-PLAN.md   # AI chatbot plan
│   ├── GRACEFUL-STARTUP.md    # Startup retry logic
│   └── architecture/    # Architecture documentation
├── docker-compose.yml   # Local development setup
├── Dockerfile          # Container image definition
└── README.md
```

## 🧪 Testing

### Automated Smoke Tests

```bash
# Run all 25 tests
./tests/smoke.sh

# Run specific test suites
./tests/redis-cache.sh        # Redis caching functionality
./tests/redis-sessions.sh     # Redis session management
./tests/redis-performance.sh  # Redis performance benchmarks

# Tests cover:
# - Application health
# - Product listing and search
# - Cart operations (anonymous + authenticated)
# - User authentication
# - Order processing
# - Redis connectivity and caching
# - Elasticsearch indexing and search
# - MinIO image serving and caching
# - Database integrity
```

### Manual Testing

```bash
# Format code
go fmt ./...

# Run Go tests
go test ./...

# Check linter
golangci-lint run
```

## 🗄️ Database Migrations

The application automatically runs migrations on startup. Key migrations include:

1. **001** - Initial schema (products, cart, orders, users)
2. **002-008** - Schema expansions (categories, SKUs, stock, roles)
3. **009** - Reviews table with ratings
4. **010** - Author field for books

### Manual Migration

```bash
# Connect to database
docker compose exec db psql -U user -d bookstore

# Run specific migration
docker compose exec db psql -U user -d bookstore -f migrations/010_add_author_field.sql
```

## 🌐 Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_USER` | PostgreSQL username | `user` | Yes |
| `DB_PASSWORD` | PostgreSQL password | `password` | Yes |
| `DB_HOST` | PostgreSQL host | `localhost` | Yes |
| `DB_NAME` | PostgreSQL database name | `bookstore` | Yes |
| `REDIS_URL` | Redis connection string | `localhost:6379` | Yes |
| `ES_URL` | Elasticsearch URL | `http://localhost:9200` | Yes |
| `MINIO_ENDPOINT` | MinIO endpoint | `localhost:9000` | Yes |
| `MINIO_ACCESS_KEY` | MinIO access key | `minioadmin` | Yes |
| `MINIO_SECRET_KEY` | MinIO secret key | `minioadmin` | Yes |

## ☸️ Kubernetes Deployment

Deploy to VKS (vSphere Kubernetes Service) or any Kubernetes cluster.

### Prerequisites

- Kubernetes cluster (VKS, EKS, GKE, AKS, minikube, kind)
- kubectl configured
- Docker registry (Harbor recommended for VCF demos)

### Step 1: Build and Push Image

```bash
# Build the Docker image
docker build -t your-registry/bookstore-app:latest .

# Push to registry (Harbor for VCF)
docker push your-registry/bookstore-app:latest
```

### Step 2: Create Secrets

```bash
# Create production secrets
kubectl create secret generic postgres-secret \
  --from-literal=POSTGRES_USER=your-user \
  --from-literal=POSTGRES_PASSWORD=your-secure-password

kubectl create secret generic redis-secret \
  --from-literal=REDIS_PASSWORD=your-redis-password

kubectl create secret generic minio-secret \
  --from-literal=MINIO_ACCESS_KEY=your-access-key \
  --from-literal=MINIO_SECRET_KEY=your-secret-key
```

### Step 3: Deploy Services

```bash
# Deploy PostgreSQL
kubectl apply -f kubernetes/postgres.yaml

# Deploy Redis
kubectl apply -f kubernetes/redis.yaml

# Deploy Elasticsearch
kubectl apply -f kubernetes/elasticsearch.yaml

# Deploy MinIO
kubectl apply -f kubernetes/minio.yaml

# Deploy Application
kubectl apply -f kubernetes/app.yaml

# Check status
kubectl get pods
kubectl get services
```

### Step 4: Access Application

```bash
# For LoadBalancer (Cloud)
kubectl get service app-service

# For Port Forwarding
kubectl port-forward service/app-service 8080:80

# Access
open http://localhost:8080
```

## 📈 VCF 9.0 Demo Scenarios

### Scenario 1: CNCF Graduated Projects
- **Elasticsearch**: Full-text search with StatefulSet deployment
- **Redis**: Distributed caching for session management
- **Prometheus**: Custom business metrics (Phase 3)

### Scenario 2: Horizontal Pod Autoscaling (HPA)
- Scale based on CPU/Memory under search load
- Scale based on custom metrics (orders per minute)

### Scenario 3: Persistent Storage
- PostgreSQL with PersistentVolumeClaims
- MinIO for object storage
- Demonstrates VCF storage services

### Scenario 4: Service Mesh (Phase 3)
- Istio for traffic management
- mTLS between services
- Canary deployments

### Scenario 5: GitOps (Phase 3)
- Argo CD for declarative deployments
- Self-healing capabilities
- Multi-environment management

## 🔒 Production Considerations

### Security
- ✅ Change default credentials
- ✅ Use Kubernetes secrets for sensitive data
- ✅ Enable TLS/SSL for all connections
- ✅ Implement rate limiting
- ✅ Use Cert-Manager for automated certificates (Phase 3)
- ✅ Enable RBAC for admin features

### Scalability
- ✅ Redis-backed sessions enable horizontal scaling
- ✅ Stateless application design (12-factor)
- ✅ Database connection pooling
- ✅ Configure HPA based on metrics
- ✅ Use CDN for static assets (MinIO compatible)

### Monitoring (Phase 3)
- 🎯 Prometheus metrics export
- 🎯 Grafana dashboards
- 🎯 Log aggregation (Loki)
- 🎯 Distributed tracing (Jaeger)
- 🎯 Custom business metrics

## 📚 Documentation

### Architecture & Planning
- **[ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md)** - System architecture and design patterns
- **[GRACEFUL-STARTUP.md](docs/GRACEFUL-STARTUP.md)** - Startup retry logic and health checks
- **[ADMIN-CONSOLE-PLAN.md](docs/ADMIN-CONSOLE-PLAN.md)** - Admin console implementation plan
- **[AI-ASSISTANT-PLAN.md](docs/AI-ASSISTANT-PLAN.md)** - AI chatbot microservice plan

### Testing
- **[tests/README.md](tests/README.md)** - Testing guide and strategies
- **[tests/REDIS.md](tests/REDIS.md)** - Redis testing and performance guide

### Scripts
- **[scripts/README.md](scripts/README.md)** - Data seeding scripts documentation

## 📚 Documentation

### Quick Start
- **[docs/START-HERE.md](docs/START-HERE.md)** - Quick start guide for deployment
- **[docs/DEVELOPMENT-WORKFLOW.md](docs/DEVELOPMENT-WORKFLOW.md)** - Local development & K8s workflow

### Deployment Guides
- **[docs/REMOTE-VM-DEPLOYMENT.md](docs/REMOTE-VM-DEPLOYMENT.md)** - Remote VM deployment guide
- **[docs/DEPLOYMENT-PLAN.md](docs/DEPLOYMENT-PLAN.md)** - Complete Kubernetes deployment plan
- **[docs/DEPLOYMENT-SUMMARY.md](docs/DEPLOYMENT-SUMMARY.md)** - Deployment summary & checklist

### Harbor Registry
- **[docs/HARBOR-QUICKSTART.md](docs/HARBOR-QUICKSTART.md)** - Quick Harbor reference
- **[docs/HARBOR-SETUP.md](docs/HARBOR-SETUP.md)** - Detailed Harbor setup guide
- **[docs/HARBOR-CHECKLIST.md](docs/HARBOR-CHECKLIST.md)** - Step-by-step Harbor checklist

### Pre-Deployment
- **[docs/PRE-PUSH-CHECKLIST.md](docs/PRE-PUSH-CHECKLIST.md)** - Verify before pushing to GitHub

### Architecture & Planning
- **[docs/architecture/ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md)** - System architecture
- **[docs/ADMIN-CONSOLE-PLAN.md](docs/ADMIN-CONSOLE-PLAN.md)** - Admin console feature plan
- **[docs/AI-ASSISTANT-PLAN.md](docs/AI-ASSISTANT-PLAN.md)** - AI assistant feature plan
- **[docs/GRACEFUL-STARTUP.md](docs/GRACEFUL-STARTUP.md)** - Graceful startup implementation

## 🎯 Roadmap

### ✅ Phase 1: Core App & Data (Complete)
- User authentication and shopping cart
- Product catalog and order management
- Responsive UI with modern design

### ✅ Phase 2: Microservices Expansion (Complete)
- Elasticsearch search with autocomplete
- Redis caching and session management
- MinIO object storage
- User reviews and profiles
- Real content from Project Gutenberg
- Pagination system

### 🎯 Phase 3: Ops & Observability (Next)
- Argo CD for GitOps
- Prometheus & Grafana for metrics
- Istio service mesh
- ExternalDNS automation
- AI Support Chatbot (Python microservice)

## 🤝 Contributing

This is a demo platform for VCF 9.0 showcases. Contributions are welcome!

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `./tests/smoke.sh`
5. Format code: `go fmt ./...`
6. Submit a pull request

## 📝 License

MIT License - See LICENSE file for details

## 🙏 Acknowledgments

- **Project Gutenberg** - Public domain book content and covers
- **Pico CSS** - Minimalist CSS framework
- **HTMX** - Modern dynamic UI without heavy JavaScript
- **VMware** - VCF 9.0 platform and documentation

---

**Built with ❤️ to demonstrate VMware Cloud Foundation 9.0 capabilities**
