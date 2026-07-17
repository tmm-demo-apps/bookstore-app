# Documentation Index

## 🚀 Getting Started

### Deploy on Any Kubernetes Cluster (Helm)

```bash
# Clone and deploy with a single Helm command
git clone https://github.com/tmm-demo-apps/bookstore-app.git
cd bookstore-app
helm dependency update ./helm/demo-suite
helm install demo ./helm/demo-suite --set global.domain=<your-domain>
```

See the [main README](../README.md) for the full list of Helm options (small clusters, Harbor, restricted networks, etc.).

### GitOps Deployment (ArgoCD)

```bash
# Push to main — CI handles build, push to GHCR/Harbor, and kustomize update
git push

# ArgoCD detects the change and syncs automatically
argocd app get bookstore
```

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
| [GITHUB-ACTIONS-SETUP.md](GITHUB-ACTIONS-SETUP.md) | CI/CD pipeline configuration |
| [SELF-HOSTED-RUNNER-SETUP.md](SELF-HOSTED-RUNNER-SETUP.md) | GitHub Actions runner setup |

### App Specifications
| Document | Purpose |
|----------|---------|
| [READER-APP-SPEC.md](READER-APP-SPEC.md) | Reader app specification |
| [AI-ASSISTANT-PLAN.md](AI-ASSISTANT-PLAN.md) | Chatbot architecture (Ollama/VCF Private AI) |
| [ADMIN-CONSOLE-PLAN.md](ADMIN-CONSOLE-PLAN.md) | Admin dashboard implementation plan |

### Architecture
| Document | Purpose |
|----------|---------|
| [architecture/ARCHITECTURE.md](architecture/ARCHITECTURE.md) | System architecture overview |

## 📁 Key Files

```
helm/demo-suite/              # Umbrella Helm chart (primary deployment method)
├── values.yaml               # Defaults: GHCR images, Gateway API, TLS
├── values-small.yaml         # Lightweight profile for 1-2 node clusters
└── values-harbor.yaml        # Override for Harbor enterprise registry

kubernetes/                   # Kustomize manifests (GitOps/ArgoCD)
├── base/                     # Base manifests (app, postgres, redis, minio, gateway, etc.)
├── overlays/
│   ├── prod/                 # Production environment patches
│   ├── dev/                  # Development environment patches
│   └── lab/                  # Air-gapped lab environment patches
└── components/init-db/       # Database migration + seed job component

scripts/
├── local-dev.sh              # Local development helper
├── install-prerequisites.sh  # Install Istio + cert-manager (non-VKS clusters)
└── seed-gutenberg-books.go   # Book data source (regenerates 002_seed_books.sql)
```

## 🔗 External Resources

- **Main README**: [../README.md](../README.md)
- **Kubernetes Manifests**: [../kubernetes/README.md](../kubernetes/README.md)

---

**Last Updated**: July 17, 2026
