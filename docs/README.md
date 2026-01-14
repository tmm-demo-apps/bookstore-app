# Documentation Index

## 🚀 Getting Started

### Quick Start
The bookstore application is deployed using a single command:

```bash
# Deploy to Kubernetes (from remote VM)
./scripts/deploy-complete.sh v1.1.0 bookstore

# Deploy to test namespace
./scripts/deploy-complete.sh v1.1.0 bookstore-test
```

This handles everything: Harbor image build/push, NGINX Ingress auto-install, database migrations, seeding, and application deployment.

### Local Development
```bash
# Start local environment
./scripts/local-dev.sh start

# Run tests
./scripts/local-dev.sh test

# Stop
./scripts/local-dev.sh stop
```

## 📖 Documentation by Category

### Core Guides
| Document | Purpose |
|----------|---------|
| [DEVELOPMENT-WORKFLOW.md](DEVELOPMENT-WORKFLOW.md) | Local development with Docker Compose |
| [HARBOR-SETUP.md](HARBOR-SETUP.md) | Harbor registry configuration |
| [GRACEFUL-STARTUP.md](GRACEFUL-STARTUP.md) | Health checks and retry logic |

### VCF 9.1 Features
| Document | Purpose |
|----------|---------|
| [DUAL-NETWORK-VKS-DEMO.md](DUAL-NETWORK-VKS-DEMO.md) | Dual-NIC VKS cluster demo plan |

### Future Features (Phase 2+)
| Document | Purpose |
|----------|---------|
| [ADMIN-CONSOLE-PLAN.md](ADMIN-CONSOLE-PLAN.md) | Admin dashboard implementation plan |
| [AI-ASSISTANT-PLAN.md](AI-ASSISTANT-PLAN.md) | AI chat bot microservice plan |

### Architecture
| Document | Purpose |
|----------|---------|
| [architecture/ARCHITECTURE.md](architecture/ARCHITECTURE.md) | System architecture overview |

## 📊 Current Deployment

### Clusters
- **Production (vks-04)**: `http://bookstore.corp.vmbeans.com` (32.32.0.16)
- **Test (vks-03)**: `http://bookstore-test.corp.vmbeans.com` (32.32.0.17)

### Services
- **PostgreSQL**: StatefulSet with vSAN storage (10Gi)
- **Redis**: Session management and caching (5Gi)
- **Elasticsearch**: Full-text search (10Gi)
- **MinIO**: Object storage for images (20Gi)
- **Application**: 3 replicas with HPA
- **NGINX Ingress**: Auto-installed per cluster

### Key Files
```
scripts/
├── deploy-complete.sh          # One-command deployment
├── harbor-remote-setup.sh      # Harbor integration
└── k8s-diagnose.sh             # Troubleshooting

kubernetes/
├── ingress-nginx.yaml          # NGINX Ingress Controller
├── ingress.yaml                # Application ingress
├── init-db-job.yaml            # Automated migrations + seeding
├── app.yaml                    # Application deployment
├── postgres.yaml               # PostgreSQL
├── redis.yaml                  # Redis
├── elasticsearch.yaml          # Elasticsearch
└── minio.yaml                  # MinIO
```

## 🔗 External Resources

- **Main README**: [../README.md](../README.md)
- **Kubernetes Manifests**: [../kubernetes/README.md](../kubernetes/README.md)
- **Personal Dev Notes**: `../dev_docs/` (not in git)

---

**Last Updated**: January 9, 2026
