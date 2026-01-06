# Remote VM Deployment Guide

## Overview

This guide is for deploying the DemoApp bookstore to a Kubernetes cluster accessible only through a remote VM/jumpbox.

**Your Setup**:
- **Harbor**: `harbor.corp.vmbeans.com`
- **Project**: `bookstore` (needs to be created)
- **CA Cert**: `/etc/docker/certs.d/harbor.corp.vmbeans.com/ca.crt`
- **Access**: Via jumpbox VM (devops@cli-vm)

## Deployment Strategy

```
┌─────────────────┐      ┌──────────────────┐      ┌─────────────────┐
│  Local Machine  │─────▶│   GitHub Repo    │─────▶│   Remote VM     │
│  (Development)  │ push │  (Code Storage)  │ pull │  (Deployment)   │
└─────────────────┘      └──────────────────┘      └─────────────────┘
                                                            │
                                                            ▼
                                                    ┌─────────────────┐
                                                    │  Harbor + K8s   │
                                                    └─────────────────┘
```

## Step-by-Step Process

### Phase 1: Prepare on Local Machine (You Are Here)

#### 1.1 Verify Everything Works Locally

```bash
# Run smoke tests
./test-smoke.sh

# Should see: All 25 tests passing ✅
```

#### 1.2 Format Code

```bash
go fmt ./...
```

#### 1.3 Commit and Push to GitHub

```bash
git add -A
git commit -m "feat: add Harbor deployment scripts and K8s manifests"
git push origin main
```

**What gets pushed**:
- ✅ Application code
- ✅ Dockerfile
- ✅ Kubernetes manifests
- ✅ Harbor setup scripts
- ✅ Documentation
- ❌ `dev_docs/` (excluded by .gitignore)
- ❌ Secrets (never commit!)

---

### Phase 2: Deploy from Remote VM

#### 2.1 SSH to Jumpbox

```bash
ssh devops@cli-vm
```

#### 2.2 Clone Repository

```bash
# First time
git clone https://github.com/johnnyr0x/bookstore-app.git
cd bookstore-app

# Or if already cloned, update it
cd bookstore-app
git pull origin main
```

#### 2.3 Verify Prerequisites

```bash
# Check Docker
docker --version
# Expected: Docker version 28.5.2, build ecc6942 ✅

# Check kubectl
kubectl version --client
# Expected: Client Version: v1.34.1 ✅

# Check Harbor CA cert
ls -la /etc/docker/certs.d/harbor.corp.vmbeans.com/ca.crt
# Should exist ✅

# Test Harbor connectivity
curl -k https://harbor.corp.vmbeans.com/api/v2.0/systeminfo
# Should return JSON ✅
```

#### 2.4 Create Harbor Project (One-Time Setup)

Before running the script, create the `bookstore` project in Harbor:

1. Open browser: `https://harbor.corp.vmbeans.com`
2. Login with your credentials
3. Click **"Projects"** → **"New Project"**
4. Project Name: `bookstore`
5. Access Level: **Private**
6. Click **"OK"**

#### 2.5 Run Harbor Setup Script

```bash
# Run the automated setup script
./scripts/harbor-remote-setup.sh v1.0.0

# The script will:
# 1. Check Harbor connectivity ✅
# 2. Prompt for Harbor credentials
# 3. Login to Docker registry
# 4. Build the application image
# 5. Push to Harbor (v1.0.0 and latest tags)
# 6. Create Kubernetes namespace
# 7. Create image pull secret
# 8. Create application secrets (DB, MinIO)
```

**What to expect**:
- Build time: ~2-5 minutes (depending on VM resources)
- Push time: ~1-3 minutes (depending on network)
- Total time: ~5-10 minutes

#### 2.6 Update Kubernetes Manifests

The script will tell you the exact image name. Update `kubernetes/app.yaml`:

```bash
# Edit the file
nano kubernetes/app.yaml

# Change the image line to:
image: harbor.corp.vmbeans.com/bookstore/app:v1.0.0
```

Or use sed:

```bash
sed -i 's|image:.*|image: harbor.corp.vmbeans.com/bookstore/app:v1.0.0|' kubernetes/app.yaml
```

---

### Phase 3: Deploy to Kubernetes

#### 3.1 Review Manifests

```bash
# List all Kubernetes manifests
ls -la kubernetes/

# Should see:
# - namespace.yaml (if created)
# - configmap.yaml
# - postgres.yaml
# - redis.yaml
# - elasticsearch.yaml
# - minio.yaml
# - app.yaml
# - ingress.yaml (optional)
```

#### 3.2 Deploy Infrastructure Services

```bash
# Create namespace (if not already created by script)
kubectl create namespace bookstore

# Deploy PostgreSQL
kubectl apply -f kubernetes/postgres.yaml

# Deploy Redis
kubectl apply -f kubernetes/redis.yaml

# Deploy Elasticsearch
kubectl apply -f kubernetes/elasticsearch.yaml

# Deploy MinIO
kubectl apply -f kubernetes/minio.yaml

# Wait for services to be ready
kubectl get pods -n bookstore -w
# Press Ctrl+C when all pods are Running
```

#### 3.3 Deploy Application

```bash
# Deploy ConfigMap
kubectl apply -f kubernetes/configmap.yaml

# Deploy Application
kubectl apply -f kubernetes/app.yaml

# Watch deployment
kubectl get pods -n bookstore -w
```

#### 3.4 Verify Deployment

```bash
# Check all pods are running
kubectl get pods -n bookstore

# Check services
kubectl get svc -n bookstore

# Check persistent volumes
kubectl get pvc -n bookstore

# Check application logs
kubectl logs -n bookstore deployment/app-deployment --tail=50

# Check health endpoint
kubectl exec -n bookstore deployment/app-deployment -- wget -qO- http://localhost:8080/health
```

#### 3.5 Access Application

```bash
# Option 1: Port forward (for testing)
kubectl port-forward -n bookstore svc/app-service 8080:80

# Then access: http://localhost:8080
# (Or use SSH tunnel if needed)

# Option 2: Deploy Ingress (for production)
kubectl apply -f kubernetes/ingress.yaml

# Check ingress
kubectl get ingress -n bookstore
```

---

## Kubernetes Manifests Checklist

Before deploying, ensure these files exist in `kubernetes/`:

- [ ] **namespace.yaml** - Creates bookstore namespace
- [ ] **configmap.yaml** - Environment variables
- [ ] **postgres.yaml** - PostgreSQL StatefulSet + PVC
- [ ] **redis.yaml** - Redis Deployment + PVC
- [ ] **elasticsearch.yaml** - Elasticsearch StatefulSet + PVC
- [ ] **minio.yaml** - MinIO Deployment + PVC
- [ ] **app.yaml** - Application Deployment + Service
- [ ] **ingress.yaml** - Ingress (optional)
- [ ] **hpa.yaml** - HorizontalPodAutoscaler (optional)

**Note**: If these don't exist yet, see `docs/DEPLOYMENT-PLAN.md` for complete manifest templates.

---

## Data Seeding (After Deployment)

Once the application is running, seed the database:

```bash
# Port forward to access services
kubectl port-forward -n bookstore svc/postgres-service 5432:5432 &
kubectl port-forward -n bookstore svc/minio-service 9000:9000 &
kubectl port-forward -n bookstore svc/app-service 8080:80 &

# Get database password
DB_PASSWORD=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.DB_PASSWORD}' | base64 -d)

# Get MinIO credentials
MINIO_ACCESS_KEY=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.MINIO_ACCESS_KEY}' | base64 -d)
MINIO_SECRET_KEY=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.MINIO_SECRET_KEY}' | base64 -d)

# Set environment variables
export DB_HOST=localhost
export DB_USER=bookstore_user
export DB_PASSWORD="${DB_PASSWORD}"
export DB_NAME=bookstore
export MINIO_ENDPOINT=localhost:9000
export MINIO_ACCESS_KEY="${MINIO_ACCESS_KEY}"
export MINIO_SECRET_KEY="${MINIO_SECRET_KEY}"
export MINIO_USE_SSL=false

# Run seed scripts
go run scripts/seed-gutenberg-books.go
go run scripts/seed-images.go

# Verify
curl http://localhost:8080/products
```

---

## Troubleshooting

### Issue: Cannot connect to Harbor

```bash
# Check network connectivity
ping harbor.corp.vmbeans.com

# Check DNS resolution
nslookup harbor.corp.vmbeans.com

# Test HTTPS
curl -k https://harbor.corp.vmbeans.com/api/v2.0/systeminfo
```

### Issue: Certificate errors

```bash
# Verify CA cert exists
ls -la /etc/docker/certs.d/harbor.corp.vmbeans.com/ca.crt

# If missing, download it
sudo mkdir -p /etc/docker/certs.d/harbor.corp.vmbeans.com
sudo curl -k https://harbor.corp.vmbeans.com/api/v2.0/systeminfo/getcert \
  -o /etc/docker/certs.d/harbor.corp.vmbeans.com/ca.crt

# Restart Docker
sudo systemctl restart docker
```

### Issue: ImagePullBackOff

```bash
# Check secret exists
kubectl get secret harbor-registry-secret -n bookstore

# Check secret contents
kubectl get secret harbor-registry-secret -n bookstore -o yaml

# Check pod events
kubectl describe pod <pod-name> -n bookstore

# Verify image name
kubectl get deployment app-deployment -n bookstore -o yaml | grep image:

# Test image pull manually
docker pull harbor.corp.vmbeans.com/bookstore/app:v1.0.0
```

### Issue: Pods not starting

```bash
# Check pod status
kubectl get pods -n bookstore

# Check pod logs
kubectl logs -n bookstore <pod-name>

# Check pod events
kubectl describe pod -n bookstore <pod-name>

# Check resource constraints
kubectl top nodes
kubectl top pods -n bookstore
```

---

## Updating the Application

When you make changes:

### On Local Machine:

```bash
# 1. Make changes
# 2. Test locally
./test-smoke.sh

# 3. Commit and push
git add -A
git commit -m "feat: your changes"
git push origin main
```

### On Remote VM:

```bash
# 1. Pull latest code
git pull origin main

# 2. Build new version
./scripts/harbor-remote-setup.sh v1.0.1

# 3. Update manifest
sed -i 's|:v1.0.0|:v1.0.1|' kubernetes/app.yaml

# 4. Apply changes
kubectl apply -f kubernetes/app.yaml

# 5. Watch rollout
kubectl rollout status deployment/app-deployment -n bookstore

# 6. Verify
kubectl get pods -n bookstore
```

---

## Quick Reference Commands

```bash
# View all resources
kubectl get all -n bookstore

# View logs
kubectl logs -n bookstore deployment/app-deployment -f

# Shell into pod
kubectl exec -it -n bookstore deployment/app-deployment -- /bin/sh

# Restart deployment
kubectl rollout restart deployment/app-deployment -n bookstore

# Scale deployment
kubectl scale deployment/app-deployment -n bookstore --replicas=5

# Delete everything
kubectl delete namespace bookstore
```

---

## Next Steps

After successful deployment:

1. ✅ **Set up Argo CD** - GitOps automation (see `docs/DEPLOYMENT-PLAN.md`)
2. ✅ **Add monitoring** - Prometheus + Grafana
3. ✅ **Admin Console** - Phase 2 feature
4. ✅ **AI Assistant** - Phase 2 microservice

---

## Files to Ignore in Git

Make sure these are in `.gitignore`:

```
dev_docs/
kubernetes/secrets-generated.txt
kubernetes/secret.yaml
.env
*.pem
*.key
```

---

**Ready to deploy?** Start with Phase 1 on your local machine! 🚀

