# Kubernetes Manifests

This directory contains all Kubernetes manifests for deploying the DemoApp bookstore application.

## Deployment Methods

### 1. GitOps with ArgoCD (Recommended for Production)

ArgoCD watches this directory and automatically syncs changes to the cluster.

```bash
# Create the ArgoCD Application
kubectl apply -f kubernetes/argocd-application.yaml -n dev-wrcc9

# Or via ArgoCD CLI
argocd app create bookstore \
  --repo https://github.com/tmm-demo-apps/bookstore-app.git \
  --path kubernetes \
  --dest-server https://32.32.0.6:443 \
  --dest-namespace bookstore \
  --sync-policy automated \
  --auto-prune \
  --self-heal
```

### 2. Manual Deployment with Kustomize

Use the deploy script which handles Kustomize configuration:

```bash
# Full build and deploy
./scripts/deploy-complete.sh v1.0.0

# Deploy existing image from Harbor
./scripts/deploy-complete.sh v1.0.0 --skip-build

# Non-interactive mode
./scripts/deploy-complete.sh v1.0.0 -y
```

### 3. Direct Kustomize Commands

```bash
# Preview what will be deployed
kubectl kustomize kubernetes/

# Apply directly
kubectl apply -k kubernetes/

# Apply with custom namespace
cd kubernetes && kustomize edit set namespace my-namespace && cd ..
kubectl apply -k kubernetes/
```

## Files Overview

| File | Description | Managed By |
|------|-------------|------------|
| `kustomization.yaml` | Kustomize configuration | CI/CD updates image tags |
| `namespace.yaml` | Creates bookstore namespace | Kustomize |
| `configmap.yaml` | Environment variables for app | Kustomize |
| `secret.yaml` | Secrets template | Manual (not in git) |
| `postgres.yaml` | PostgreSQL StatefulSet + Service + PVC | Kustomize |
| `redis.yaml` | Redis Deployment + Service + PVC | Kustomize |
| `elasticsearch.yaml` | Elasticsearch StatefulSet + Service + PVC | Kustomize |
| `minio.yaml` | MinIO Deployment + Service + PVC | Kustomize |
| `app.yaml` | Application Deployment + Service | Kustomize |
| `ingress.yaml` | Ingress for external access | Kustomize |
| `hpa.yaml` | HorizontalPodAutoscaler | Kustomize |
| `migrations-configmap.yaml` | SQL migration files | Kustomize |
| `init-db-job.yaml` | Database initialization job | Manual (one-time) |
| `seed-job.yaml` | Database seeding jobs | Manual (one-time) |
| `ingress-nginx.yaml` | NGINX Ingress Controller | Manual (separate namespace) |
| `argocd-application.yaml` | ArgoCD Application definition | Manual |

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        GitHub Repository                         │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  kubernetes/                                                 ││
│  │  ├── kustomization.yaml  ← CI updates image tags            ││
│  │  ├── app.yaml                                                ││
│  │  ├── postgres.yaml                                           ││
│  │  └── ...                                                     ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ watches
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                          ArgoCD                                  │
│  - Detects changes in kubernetes/                               │
│  - Renders manifests with Kustomize                             │
│  - Syncs to target cluster                                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ deploys
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    VKS-04 Cluster (bookstore namespace)          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │PostgreSQL│  │  Redis   │  │Elastic   │  │  MinIO   │        │
│  └──────────┘  └──────────┘  │ search   │  └──────────┘        │
│                              └──────────┘                        │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              Bookstore App (3 replicas)                   │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## CI/CD Flow

1. **Code Push** → CI runs tests and builds Docker image
2. **Image Push** → Image pushed to Harbor with commit SHA tag
3. **Kustomize Update** → CI updates `kustomization.yaml` with new image tag
4. **Git Push** → CI commits and pushes the kustomization change
5. **ArgoCD Sync** → ArgoCD detects change and deploys new image

## Kustomize Configuration

The `kustomization.yaml` file controls:

- **Namespace**: Applied to all resources
- **Image Tags**: Overrides for all container images

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: bookstore

resources:
  - namespace.yaml
  - configmap.yaml
  # ... other resources

images:
  - name: harbor.corp.vmbeans.com/bookstore/app
    newTag: latest  # CI updates this to commit SHA
```

## Manual Deployment Steps

If not using ArgoCD or the deploy script:

### Step 1: Create Namespace and Secrets

```bash
kubectl apply -f namespace.yaml

# Create secrets manually
kubectl create secret docker-registry harbor-registry-secret \
  --docker-server=harbor.corp.vmbeans.com \
  --docker-username=<user> \
  --docker-password=<pass> \
  -n bookstore

kubectl create secret generic app-secrets \
  --from-literal=DB_USER=bookstore_user \
  --from-literal=DB_PASSWORD=$(openssl rand -hex 16) \
  --from-literal=MINIO_ACCESS_KEY=$(openssl rand -hex 10) \
  --from-literal=MINIO_SECRET_KEY=$(openssl rand -hex 16) \
  -n bookstore
```

### Step 2: Deploy with Kustomize

```bash
kubectl apply -k kubernetes/
```

### Step 3: Run Database Initialization

```bash
kubectl apply -f kubernetes/init-db-job.yaml -n bookstore
kubectl wait --for=condition=complete job/init-database -n bookstore --timeout=600s
```

### Step 4: Install Ingress Controller (if needed)

```bash
kubectl apply -f kubernetes/ingress-nginx.yaml
```

## Verification

```bash
# Check all resources
kubectl get all -n bookstore

# Check ArgoCD sync status
argocd app get bookstore

# Check application health
kubectl exec -n bookstore deployment/app-deployment -- wget -qO- http://localhost:8080/health
```

## Troubleshooting

### ArgoCD Sync Issues

```bash
# Check sync status
argocd app get bookstore

# Force sync
argocd app sync bookstore

# Check events
argocd app history bookstore
```

### Kustomize Rendering Issues

```bash
# Preview rendered manifests
kubectl kustomize kubernetes/

# Validate YAML
kubectl kustomize kubernetes/ | kubectl apply --dry-run=client -f -
```

### Image Pull Errors

```bash
# Check image pull secret
kubectl get secret harbor-registry-secret -n bookstore -o yaml

# Verify image exists in Harbor
curl -sk https://harbor.corp.vmbeans.com/api/v2.0/projects/bookstore/repositories/app/artifacts
```

## Future Multi-Environment Support

When ready for dev/staging/prod environments, restructure to:

```
kubernetes/
├── base/                    # Move current files here
│   ├── kustomization.yaml
│   └── *.yaml
└── overlays/
    ├── dev/
    │   └── kustomization.yaml  # namespace: bookstore-dev
    ├── staging/
    │   └── kustomization.yaml  # namespace: bookstore-staging
    └── prod/
        └── kustomization.yaml  # namespace: bookstore
```

## Resource Requirements

| Service | CPU Request | CPU Limit | Memory Request | Memory Limit | Storage |
|---------|-------------|-----------|----------------|--------------|---------|
| PostgreSQL | 250m | 1000m | 256Mi | 1Gi | 10Gi |
| Redis | 100m | 500m | 128Mi | 512Mi | 5Gi |
| Elasticsearch | 500m | 1000m | 1Gi | 2Gi | 10Gi |
| MinIO | 250m | 1000m | 256Mi | 1Gi | 20Gi |
| App (per pod) | 100m | 500m | 128Mi | 512Mi | - |

**Total** (with 3 app replicas): ~1.5 cores requested, ~5 cores limit, ~2.5GB memory requested, ~45GB storage.

## Related Documentation

- **docs/SELF-HOSTED-RUNNER-SETUP.md** - CI/CD runner setup
- **docs/HARBOR-SETUP.md** - Harbor registry setup
- **scripts/README.md** - Deployment scripts documentation
