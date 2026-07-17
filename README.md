# Bookstore App - E-commerce Platform for VCF 9 Demonstrations

[![CI](https://github.com/tmm-demo-apps/bookstore-app/workflows/CI/badge.svg)](https://github.com/tmm-demo-apps/bookstore-app/actions)
[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14-336791?logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis)](https://redis.io/)
[![Elasticsearch](https://img.shields.io/badge/Elasticsearch-8.11-005571?logo=elasticsearch)](https://www.elastic.co/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A production-ready e-commerce platform built to demonstrate **VMware Cloud Foundation (VCF) 9.0** capabilities. Features enterprise-grade infrastructure including Elasticsearch search, Redis caching, MinIO object storage, and real-world content from Project Gutenberg.

**🎯 Purpose**: Showcase VCF 9.0 Supervisor Services, VKS (vSphere Kubernetes Service), VKS Add-ons, dual-network support, and CNCF graduated projects through a realistic e-commerce application.

## Multi-App Demo Suite

This Bookstore is part of a 3-app demo suite:

| App | Description | Endpoint | Repo |
|-----|-------------|----------|------|
| **Bookstore** | E-commerce platform (this repo) | `bookstore.<your-domain>` | [bookstore-app](https://github.com/tmm-demo-apps/bookstore-app) |
| **Reader** | EPUB library reader | `reader.<your-domain>` | [reader-app](https://github.com/tmm-demo-apps/reader-app) |
| **Chatbot** | AI customer support | `chatbot.<your-domain>` | [chatbot-app](https://github.com/tmm-demo-apps/chatbot-app) |

All apps can be deployed via **Helm** (any Kubernetes cluster) or **ArgoCD GitOps** (automated CI/CD), and share services (MinIO, Redis) where appropriate.

## 🚀 Deploy on a New Kubernetes Cluster

Deploy the entire demo suite on **any Kubernetes cluster** with a single Helm command. All images are public on GHCR -- no registry credentials, no rate limits, no Docker Hub dependency.

### Prerequisites

1. **Kubernetes cluster** with `kubectl` access from a jumpbox/workstation
2. **Helm 3** installed on the jumpbox (`brew install helm` on macOS, or [install guide](https://helm.sh/docs/intro/install/))
3. **Gateway API prerequisites** (Istio + cert-manager + Gateway API CRDs):
   - **VKS clusters**: Install as VKS standard packages via `AddonInstall` resources on the Supervisor
   - **Other clusters**: Run `./scripts/install-prerequisites.sh` (installs via Helm)
   - **Legacy fallback**: Set `global.gatewayAPI.enabled=false` and use ingress-nginx instead
4. **DNS or /etc/hosts** pointing your chosen domain to the gateway/ingress IP:
   - `bookstore.<your-domain>` -> gateway IP
   - `reader.<your-domain>` -> gateway IP (if deploying reader)
   - `chatbot.<your-domain>` -> gateway IP (if deploying chatbot)

### Step-by-Step

```bash
# 1. Clone the repo to your jumpbox
git clone https://github.com/tmm-demo-apps/bookstore-app.git
cd bookstore-app

# 2. Build Helm dependencies
helm dependency update ./helm/demo-suite

# 3. Install prerequisites (one-time, non-VKS clusters only)
./scripts/install-prerequisites.sh

# 4. Deploy (pick one)

# Full suite (bookstore + reader + chatbot) with Gateway API + TLS:
helm install demo ./helm/demo-suite \
  --set global.domain=<your-domain>

# Bookstore + Reader only (no chatbot):
helm install demo ./helm/demo-suite \
  --set global.domain=<your-domain> \
  --set chatbot.enabled=false

# Bookstore only:
helm install demo ./helm/demo-suite \
  --set global.domain=<your-domain> \
  --set reader.enabled=false \
  --set chatbot.enabled=false

# With a specific storage class (if your cluster doesn't have a default):
helm install demo ./helm/demo-suite \
  --set global.domain=<your-domain> \
  --set global.storageClassName=my-storage-policy

# Legacy mode (ingress-nginx instead of Gateway API):
helm install demo ./helm/demo-suite \
  --set global.domain=<your-domain> \
  --set global.gatewayAPI.enabled=false \
  --set global.tls.enabled=false \
  --set ingress-nginx.enabled=true

# Small cluster (1-2 worker nodes)? Use the lightweight profile:
helm install demo ./helm/demo-suite \
  -f ./helm/demo-suite/values-small.yaml \
  --set global.domain=<your-domain> \
  --set global.storageClassName=<your-storage-class>

# Restricted network (pods can't reach external mirrors)?
# Pre-seed EPUBs into MinIO via init container:
helm install demo ./helm/demo-suite \
  --set global.domain=<your-domain> \
  --set reader.epubSeed.enabled=true
```

> **Gateway API (default)**: The chart uses Kubernetes Gateway API with Istio and cert-manager for TLS. Prerequisites must be installed before deploying -- use VKS AddonInstall for VKS clusters or `scripts/install-prerequisites.sh` for others. Set `global.gatewayAPI.enabled=false` to fall back to legacy Ingress + ingress-nginx.

> **Small Clusters**: The `values-small.yaml` profile reduces all replicas to 1, lowers CPU/memory requests (~550m total CPU vs ~1900m default), and disables Elasticsearch (search falls back to SQL). Use this for clusters with 1-2 worker nodes.

> **EPUB Pre-Seeding**: By default, the Reader app downloads EPUBs on-demand from a Gutenberg mirror. If your cluster's pods cannot reach external hosts (corporate firewall, no egress), enable `reader.epubSeed.enabled=true` to pre-load all 150 EPUBs into MinIO via an init container. The init container image (`ghcr.io/tmm-demo-apps/reader-epubs:v1`) is public on GHCR.

### What happens

Helm creates three namespaces (`bookstore`, `reader`, `chatbot`) and deploys:

| Component | What it creates |
|-----------|----------------|
| **bookstore** | App (3 replicas), PostgreSQL, Redis, Elasticsearch, MinIO, Gateway, HTTPRoute, TLS Certificate, HPA, init-job (migrations + seed data) |
| **reader** | App (2 replicas), PostgreSQL, HTTPRoute, optional EPUB seed init container |
| **chatbot** | App (1 replica), Ollama (disabled by default), HTTPRoute |

The `global.domain` value automatically configures:
- HTTPRoute hostnames: `bookstore.<your-domain>`, `reader.<your-domain>`, `chatbot.<your-domain>`
- TLS certificate covering all three hostnames (via cert-manager)
- Cross-app browser URLs (e.g. "Back to shop" link in reader points to `bookstore.<your-domain>`)
- Internal service-to-service API calls use Kubernetes DNS (automatic, no config needed)

### Verify

```bash
# Check all pods are running
kubectl get pods -n bookstore
kubectl get pods -n reader
kubectl get pods -n chatbot

# Check gateway and routes
kubectl get gateway,httproute -A

# Check TLS certificate
kubectl get certificate -n bookstore

# Open in browser
# https://bookstore.<your-domain>
```

### Upgrade or Uninstall

```bash
# Upgrade (e.g. after pulling new chart version)
helm upgrade demo ./helm/demo-suite --set global.domain=<your-domain>

# Uninstall (removes all resources)
helm uninstall demo
```

### For Harbor Environments (VMware VCF)

If you have a Harbor registry and want to use private images instead of GHCR:

```bash
helm install demo ./helm/demo-suite -f ./helm/demo-suite/values-harbor.yaml
```

See `helm/demo-suite/values-harbor.yaml` for the full Harbor override configuration.

### Image Registry Summary

| Registry | Images | Auth Required | Rate Limits |
|----------|--------|--------------|-------------|
| **GHCR** (default) | All app + infra images (including `reader-epubs:v1` for EPUB seeding) | No | None |
| **Harbor** (override) | All images via `values-harbor.yaml` | Yes | None |

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
- ☸️ **Kubernetes Ready** - VKS deployment with Gateway API (Istio), cert-manager TLS, and HPA
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
| **Registry** | GHCR + Harbor | Public images (GHCR) + enterprise registry (Harbor) |
| **GitOps** | ArgoCD | Automated deployments from git |
| **CI/CD** | GitHub Actions | Build, test, push to GHCR + Harbor |
| **Packaging** | Helm | Portable deployment on any K8s cluster |

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

### Production Deployment (Helm - Any Cluster)

```bash
# Deploy on any K8s cluster with public GHCR images
helm install demo ./helm/demo-suite --set global.domain=apps.your-env.com

# Deploy on Harbor environment (VCF)
helm install demo ./helm/demo-suite -f ./helm/demo-suite/values-harbor.yaml

# Upgrade an existing deployment
helm upgrade demo ./helm/demo-suite --set global.domain=apps.your-env.com
```

### Production Deployment (GitOps with ArgoCD)

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
                                                   VKS Cluster
```

```bash
# Just push to main - CI/CD handles the rest
git add -A && git commit -m "feat: your feature"
git push

# Check ArgoCD status
argocd app get bookstore

# View in ArgoCD UI
# https://<your-argocd-host>
```

The CI workflow automatically:
1. Runs linting and tests
2. Builds Docker image
3. Pushes to **GHCR** (public, for Helm deployments) and **Harbor** (enterprise)
4. Updates `kubernetes/base/kustomization.yaml` with new image tag
5. ArgoCD detects the change and syncs to the target cluster

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
│   ├── install-prerequisites.sh  # Install Istio + cert-manager (non-VKS)
│   ├── local-dev.sh              # Local development helper
│   ├── mirror-images.sh          # Mirror infra images to GHCR (one-time)
│   ├── setup-secrets.sh          # Multi-app secret management
│   ├── seed-gutenberg-books.go   # Book data source
│   └── seed-images.go            # Image seeding from Gutenberg
├── helm/                 # Helm chart for portable deployment
│   └── demo-suite/               # Umbrella chart (bookstore + reader + chatbot)
│       ├── Chart.yaml            # Chart metadata and dependencies
│       ├── values.yaml           # Defaults: all GHCR images, zero config
│       ├── values-small.yaml     # Small cluster profile (1-2 nodes)
│       ├── values-harbor.yaml    # Override: Harbor images + vSAN storage
│       └── charts/               # Subcharts (bookstore, reader, chatbot)
├── kubernetes/           # Kustomize manifests (ArgoCD/GitOps)
│   ├── base/                     # Base manifests with image overrides
│   ├── overlays/                 # Environment-specific patches (dev, prod)
│   ├── ingress-nginx.yaml        # NGINX Ingress Controller
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
- **Gateway API + Istio**: Modern Kubernetes networking with automatic TLS via cert-manager
- **VKS Standard Packages**: Istio and cert-manager deployed as VKS add-ons via AddonInstall
- **CNCF Graduated Projects**: Elasticsearch, Redis with StatefulSet/Deployment
- **Horizontal Pod Autoscaling**: Scale based on CPU/Memory
- **Persistent Storage**: PostgreSQL, MinIO, Elasticsearch with vSAN PVCs
- **Harbor Registry**: Enterprise container image management with vulnerability scanning
- **ArgoCD GitOps**: Automated deployments via Supervisor Service
- **VKS (vSphere Kubernetes Service)**: Native Kubernetes on VCF
- **Multi-App Architecture**: Microservices with shared services (MinIO, Redis)


## 📚 Documentation

| Document | Purpose |
|----------|---------|
| [docs/README.md](docs/README.md) | Documentation index |
| [docs/DEVELOPMENT-WORKFLOW.md](docs/DEVELOPMENT-WORKFLOW.md) | Local development guide |
| [docs/GITHUB-ACTIONS-SETUP.md](docs/GITHUB-ACTIONS-SETUP.md) | CI/CD pipeline configuration |
| [docs/SELF-HOSTED-RUNNER-SETUP.md](docs/SELF-HOSTED-RUNNER-SETUP.md) | GitHub Actions runner setup |
| [docs/HARBOR-SETUP.md](docs/HARBOR-SETUP.md) | Harbor registry configuration |
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

### ✅ Phase 4: Portable Helm Deployment (Complete)
- Helm umbrella chart for one-command deployment on any K8s cluster
- All images mirrored to GHCR (public, no rate limits)
- CI pushes to both GHCR and Harbor
- Configurable domain, storage class, and registry via `values.yaml`
- `values-harbor.yaml` override for enterprise Harbor environments

### ✅ Phase 5: Gateway API Migration (Complete)
- Gateway API with Istio replacing Ingress NGINX
- Automatic TLS via cert-manager (self-signed ClusterIssuer)
- VKS standard package support (AddonInstall for Istio + cert-manager)
- Legacy ingress-nginx fallback toggle (`global.gatewayAPI.enabled=false`)
- Prerequisite install script for non-VKS clusters

### 🎯 Phase 6: Observability & Enhancements (Next)
- Prometheus & Grafana for metrics
- VCF Private AI integration for chatbot
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

---

**Last Updated**: July 10, 2026
