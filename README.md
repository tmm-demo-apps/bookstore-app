# Bookstore App - E-commerce Platform for VCF 9.0/9.1 Demonstrations

[![CI](https://github.com/tmm-demo-apps/bookstore-app/workflows/CI/badge.svg)](https://github.com/tmm-demo-apps/bookstore-app/actions)
[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14-336791?logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis)](https://redis.io/)
[![Elasticsearch](https://img.shields.io/badge/Elasticsearch-8.11-005571?logo=elasticsearch)](https://www.elastic.co/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A production-ready e-commerce platform built to demonstrate **VMware Cloud Foundation (VCF) 9.0/9.1** capabilities. Features enterprise-grade infrastructure including Elasticsearch search, Redis caching, MinIO object storage, and real-world content from Project Gutenberg.

**🎯 Purpose**: Showcase VCF 9.0/9.1 Supervisor Services, VKS (vSphere Kubernetes Service), VKS Add-ons, dual-network support, and CNCF graduated projects through a realistic e-commerce application.

## Multi-App Demo Suite

This Bookstore is part of a 3-app demo suite:

| App | Description | Endpoint | Repo |
|-----|-------------|----------|------|
| **Bookstore** | E-commerce platform (this repo) | http://bookstore.corp.vmbeans.com | [bookstore-app](https://github.com/tmm-demo-apps/bookstore-app) |
| **Reader** | EPUB library reader | http://reader.corp.vmbeans.com | [reader-app](https://github.com/tmm-demo-apps/reader-app) |
| **Chatbot** | AI customer support | http://chatbot.corp.vmbeans.com | [chatbot-app](https://github.com/tmm-demo-apps/chatbot-app) |

All apps are deployed via **ArgoCD** using an App-of-Apps pattern and share services (MinIO, Redis) where appropriate.

## ✨ Features

### User Features
- 📚 **150 Real Products** - Public domain classics from Project Gutenberg with authentic covers
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
- ☸️ **Kubernetes Ready** - VKS deployment with NGINX Ingress and HPA
- 🔄 **GitOps with ArgoCD** - Automated deployments from git push
- 🏗️ **CI/CD Pipeline** - GitHub Actions with self-hosted runner for Harbor access
- 📦 **Harbor Registry** - Enterprise container registry with vulnerability scanning

## 🏗️ Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Backend** | Go 1.25 | High-performance application server |
| **Frontend** | HTMX + Pico CSS | Modern, lightweight UI with dynamic updates |
| **Database** | PostgreSQL 14 | Primary data store with consolidated migrations |
| **Search** | Elasticsearch 8.11 | Full-text search with autocomplete |
| **Cache** | Redis 7 | Session management and hot data caching |
| **Storage** | MinIO | S3-compatible object storage for images |
| **Container** | Docker & Docker Compose | Local development and testing |
| **Orchestration** | Kubernetes (VKS) | Production deployment on VCF |
| **Registry** | Harbor | Enterprise container registry |
| **GitOps** | ArgoCD | Automated deployments from git |
| **CI/CD** | GitHub Actions | Build, test, and push to Harbor |

## 🚀 Quick Start

### Local Development

```bash
# Start all services
./scripts/local-dev.sh start

# Run tests (25 automated tests)
./scripts/local-dev.sh test

# View logs
./scripts/local-dev.sh logs

# Stop services
./scripts/local-dev.sh stop
```

**Local URLs**:
- **App**: http://localhost:8080
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)
- **Elasticsearch**: http://localhost:9200
- **PostgreSQL**: localhost:5432 (user/password)

### Production Deployment (Kubernetes)

```bash
# SSH to remote VM
ssh devops@cli-vm
cd bookstore-app

# One-command deployment (handles everything)
./scripts/deploy-complete.sh v1.1.0 bookstore

# Or deploy to test namespace
./scripts/deploy-complete.sh v1.1.0 bookstore-test
```

The `deploy-complete.sh` script handles:
- Harbor login, image build, and push
- NGINX Ingress Controller installation (if missing)
- Database migrations and seeding (via init-db-job)
- All Kubernetes manifests
- Dynamic hostname based on namespace (`{namespace}.corp.vmbeans.com`)

**Current Deployments**:
- **Production (vks-04)**: http://bookstore.corp.vmbeans.com
- **Test (vks-03)**: http://bookstore-test.corp.vmbeans.com

### GitOps Deployment (Recommended)

The preferred deployment method is via GitOps with GitHub Actions and ArgoCD:

```
┌─────────────────────────────────────────────────────────────────────┐
│                           CI Workflow                                │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌─────────────────┐  │
│  │   Lint   │ → │   Test   │ → │  Build   │ → │  Harbor Push    │  │
│  │ (GitHub) │   │  (self)  │   │  (self)  │   │  + kustomize    │  │
│  └──────────┘   └──────────┘   └──────────┘   └────────┬────────┘  │
└────────────────────────────────────────────────────────┼────────────┘
                                                         │
                                                         ▼
                                                   ArgoCD Syncs
                                                         │
                                                         ▼
                                                   VKS-04 Cluster
```

```bash
# Just push to main - CI/CD handles the rest
git add -A && git commit -m "feat: your feature"
git push

# Check ArgoCD status
argocd app get bookstore

# View in ArgoCD UI
# https://32.32.0.10
```

The CI workflow automatically:
1. Runs linting and tests
2. Builds Docker image
3. Pushes to Harbor registry
4. Updates `kubernetes/kustomization.yaml` with new image tag
5. ArgoCD detects the change and syncs to VKS-04

## 📊 Project Structure

```
bookstore-app/
├── .github/workflows/    # CI/CD pipelines
│   ├── ci.yml                    # Lint + Test + Build + Harbor Push + Kustomize Update
│   └── deploy.yml                # Manual deployment (special cases)
├── argocd-apps/          # App-of-Apps manifests (manages all 3 apps)
│   ├── apps.yaml                 # Parent app-of-apps
│   ├── bookstore.yaml            # Bookstore ArgoCD application
│   ├── reader.yaml               # Reader ArgoCD application
│   └── chatbot.yaml              # Chatbot ArgoCD application
├── cmd/web/              # Application entry point
│   └── main.go
├── internal/
│   ├── handlers/         # HTTP request handlers (auth, cart, products, etc.)
│   ├── models/           # Data models (Product, User, Review, Order, Cart)
│   ├── repository/       # Database layer with caching (PostgreSQL, Redis, ES)
│   └── storage/          # MinIO object storage client
├── templates/            # HTML templates (Pico CSS + HTMX)
├── migrations/           # Database migrations
│   ├── 001_schema.sql            # Table definitions
│   └── 002_seed_books.sql        # 150 books from Project Gutenberg
├── scripts/              # Deployment and utility scripts
│   ├── deploy-complete.sh        # One-command K8s deployment
│   ├── local-dev.sh              # Local development helper
│   ├── setup-secrets.sh          # Multi-app secret management
│   ├── seed-gutenberg-books.go   # Book data source
│   └── seed-images.go            # Image seeding from Gutenberg
├── kubernetes/           # Kubernetes manifests + Kustomize
│   ├── kustomization.yaml        # Kustomize config (image tags updated by CI)
│   ├── ingress-nginx.yaml        # NGINX Ingress Controller
│   ├── ingress.yaml              # Application ingress
│   ├── app.yaml                  # Application deployment
│   ├── postgres.yaml             # PostgreSQL StatefulSet
│   ├── redis.yaml                # Redis deployment
│   ├── elasticsearch.yaml        # Elasticsearch StatefulSet
│   ├── minio.yaml                # MinIO deployment
│   └── init-db-job.yaml          # Automated migrations + seeding
├── tests/                # Testing scripts
│   └── smoke.sh                  # 25 automated smoke tests
├── docs/                 # Documentation
├── docker-compose.yml    # Local development
├── Dockerfile            # Container image
└── go.mod
```

## 🧪 Testing

```bash
# Run all 25 tests
./tests/smoke.sh

# Or via local-dev.sh
./scripts/local-dev.sh test

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

## 🌐 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_USER` | PostgreSQL username | `user` |
| `DB_PASSWORD` | PostgreSQL password | `password` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_NAME` | PostgreSQL database name | `bookstore` |
| `REDIS_URL` | Redis connection string | `localhost:6379` |
| `ES_URL` | Elasticsearch URL | `http://localhost:9200` |
| `MINIO_ENDPOINT` | MinIO endpoint | `localhost:9000` |
| `MINIO_ACCESS_KEY` | MinIO access key | `minioadmin` |
| `MINIO_SECRET_KEY` | MinIO secret key | `minioadmin` |

## 📈 VCF Demo Scenarios

### VCF 9.0 Demos
- **CNCF Graduated Projects**: Elasticsearch, Redis with StatefulSet/Deployment
- **Horizontal Pod Autoscaling**: Scale based on CPU/Memory
- **Persistent Storage**: PostgreSQL, MinIO, Elasticsearch with vSAN PVCs
- **Harbor Registry**: Enterprise container image management with vulnerability scanning
- **ArgoCD GitOps**: Automated deployments via Supervisor Service
- **VKS (vSphere Kubernetes Service)**: Native Kubernetes on VCF
- **Multi-App Architecture**: Microservices with shared services (MinIO, Redis)

### VCF 9.1 Demos (Coming Soon)
- **Dual-Network Support**: Separate management and workload networks
- **Supervisor Contour Service**: Centralized ingress management
- **VCF Private AI**: GPU-accelerated LLM inference for chatbot
- **See**: [docs/DUAL-NETWORK-VKS-DEMO.md](docs/DUAL-NETWORK-VKS-DEMO.md)

## 📚 Documentation

| Document | Purpose |
|----------|---------|
| [docs/README.md](docs/README.md) | Documentation index |
| [docs/DEVELOPMENT-WORKFLOW.md](docs/DEVELOPMENT-WORKFLOW.md) | Local development guide |
| [docs/GITHUB-ACTIONS-SETUP.md](docs/GITHUB-ACTIONS-SETUP.md) | CI/CD pipeline configuration |
| [docs/SELF-HOSTED-RUNNER-SETUP.md](docs/SELF-HOSTED-RUNNER-SETUP.md) | GitHub Actions runner setup |
| [docs/HARBOR-SETUP.md](docs/HARBOR-SETUP.md) | Harbor registry configuration |
| [docs/DUAL-NETWORK-VKS-DEMO.md](docs/DUAL-NETWORK-VKS-DEMO.md) | VCF 9.1 dual-network demo |
| [docs/AI-ASSISTANT-PLAN.md](docs/AI-ASSISTANT-PLAN.md) | Chatbot architecture (Ollama/VCF Private AI) |
| [docs/READER-APP-SPEC.md](docs/READER-APP-SPEC.md) | Reader app specification |
| [docs/GRACEFUL-STARTUP.md](docs/GRACEFUL-STARTUP.md) | Health checks and retry logic |
| [argocd-apps/README.md](argocd-apps/README.md) | ArgoCD App-of-Apps documentation |
| [scripts/README.md](scripts/README.md) | Scripts documentation |

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
- Real content from Project Gutenberg (150 books)
- Automated Kubernetes deployment

### ✅ Phase 3: Multi-App Suite & GitOps (Complete)
- ArgoCD for GitOps deployments
- Reader app (EPUB library reader)
- Chatbot app (AI customer support with Ollama)
- App-of-Apps pattern for centralized management
- GitHub Actions CI/CD with self-hosted runner
- Harbor registry integration

### 🎯 Phase 4: Observability & Enhancements (Next)
- Prometheus & Grafana for metrics
- VCF Private AI integration for chatbot
- Helm/Carvel packaging
- MinIO as Supervisor Service
- Elasticsearch alternatives (Meilisearch, Typesense)
- Admin Console

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes
4. Run tests: `./scripts/local-dev.sh test`
5. Format code: `go fmt ./...`
6. Commit: `git commit -m "feat: your feature"`
7. Push: `git push origin feature/your-feature`
8. Create a Pull Request for review

## 📝 License

MIT License - See LICENSE file for details

## 🙏 Acknowledgments

- **Project Gutenberg** - Public domain book content and covers
- **Pico CSS** - Minimalist CSS framework
- **HTMX** - Modern dynamic UI without heavy JavaScript
- **VMware** - VCF 9.0/9.1 platform and documentation

---

**Built with ❤️ to demonstrate VMware Cloud Foundation 9.0/9.1 capabilities**

**Last Updated**: January 30, 2026
