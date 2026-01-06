# 🚀 START HERE - Harbor & Kubernetes Deployment

**Date**: January 5, 2026  
**Status**: ✅ Ready for Remote VM Deployment  
**Harbor**: `harbor.corp.vmbeans.com`  
**Project**: `bookstore`

---

## 📋 Quick Overview

You have a **production-ready e-commerce application** with:
- ✅ 112 products (Project Gutenberg books)
- ✅ Full e-commerce features (cart, checkout, orders, reviews)
- ✅ 4 infrastructure services (PostgreSQL, Redis, Elasticsearch, MinIO)
- ✅ 25 passing smoke tests
- ✅ Complete Kubernetes manifests
- ✅ Automated Harbor deployment scripts

**Your Setup**: Local development → GitHub → Remote VM → Harbor + Kubernetes

---

## 🎯 Three Simple Steps

### Step 1: Test Locally & Push to GitHub (10 minutes)

```bash
# Start local development environment
./local-dev.sh start

# Run tests (wait for services to start)
./local-dev.sh test

# Format code
go fmt ./...

# Stop local services
./local-dev.sh stop

# Review what will be pushed
git status

# Commit and push
git add -A
git commit -m "feat: add Harbor deployment and K8s manifests"
git push origin main
```

**Note**: See `DEVELOPMENT-WORKFLOW.md` for complete local development guide.

**What gets pushed**:
- ✅ Application code and Dockerfile
- ✅ Kubernetes manifests (all 9 files)
- ✅ Harbor deployment scripts (3 scripts)
- ✅ Complete documentation (6 guides)
- ❌ `dev_docs/` (personal notes, gitignored)
- ❌ Secrets (never committed)

### Step 2: Deploy from Remote VM (15-20 minutes)

```bash
# SSH to jumpbox
ssh devops@cli-vm

# Clone repository
git clone https://github.com/johnnyr0x/bookstore-app.git
cd bookstore-app

# Create Harbor project (one-time)
# Open: https://harbor.corp.vmbeans.com
# Create project named "bookstore"

# Run automated setup (builds, pushes, creates secrets)
./scripts/harbor-remote-setup.sh v1.0.0
```

**What this script does**:
1. ✅ Checks Harbor connectivity
2. ✅ Logs into Docker registry
3. ✅ Builds application image
4. ✅ Pushes to Harbor (v1.0.0 + latest)
5. ✅ Creates Kubernetes namespace
6. ✅ Creates image pull secret
7. ✅ Creates application secrets (DB, MinIO)

### Step 3: Deploy to Kubernetes (5-10 minutes)

```bash
# Deploy infrastructure (in order)
kubectl apply -f kubernetes/namespace.yaml
kubectl apply -f kubernetes/postgres.yaml
kubectl apply -f kubernetes/redis.yaml
kubectl apply -f kubernetes/elasticsearch.yaml
kubectl apply -f kubernetes/minio.yaml

# Wait for services to be ready
kubectl get pods -n bookstore -w
# Press Ctrl+C when all are Running

# Run database migrations (one-time)
kubectl cp migrations/ bookstore/postgres-0:/tmp/migrations/
kubectl exec -it -n bookstore postgres-0 -- sh -c 'cd /tmp/migrations && for file in *.sql; do echo "Running $file..."; psql -U bookstore_user -d bookstore -f "$file"; done'

# Deploy ConfigMap and Application
kubectl apply -f kubernetes/configmap.yaml
kubectl apply -f kubernetes/app.yaml

# Wait for app to be ready
kubectl get pods -n bookstore -w
# Press Ctrl+C when app pods are Running

# Seed database with books and images
./scripts/k8s-seed-data.sh

# Deploy ingress (for external access)
kubectl apply -f kubernetes/ingress.yaml

# Get the ingress IP
kubectl get ingress -n bookstore

# Access your application
# http://bookstore.corp.vmbeans.com (or use the ingress IP)
```

---

## 📚 Documentation Guide

**Read in this order**:

1. **START-HERE.md** ← You are here!
2. **REMOTE-VM-DEPLOYMENT.md** - Complete deployment guide
3. **kubernetes/README.md** - Kubernetes deployment details

**Reference guides**:
- **DEPLOYMENT-SUMMARY.md** - What's included and ready
- **HARBOR-QUICKSTART.md** - Quick Harbor reference
- **PRE-PUSH-CHECKLIST.md** - Verify before pushing
- **docs/DEPLOYMENT-PLAN.md** - Detailed deployment plan
- **docs/HARBOR-SETUP.md** - Harbor setup guide

---

## 📦 What's Included

### Scripts (3 files)
```
scripts/
├── harbor-remote-setup.sh    ← Main deployment script (USE THIS)
├── build-and-push.sh          ← Build helper (alternative)
└── harbor-init.sh             ← Interactive setup (alternative)
```

### Kubernetes Manifests (9 files)
```
kubernetes/
├── namespace.yaml             ← Creates bookstore namespace
├── configmap.yaml             ← Environment variables
├── postgres.yaml              ← PostgreSQL StatefulSet + PVC (10GB)
├── redis.yaml                 ← Redis Deployment + PVC (5GB)
├── elasticsearch.yaml         ← Elasticsearch StatefulSet + PVC (10GB)
├── minio.yaml                 ← MinIO Deployment + PVC (20GB)
├── app.yaml                   ← Application (3 replicas, Harbor image)
├── ingress.yaml               ← Ingress (optional)
└── hpa.yaml                   ← Auto-scaling (optional)
```

### Documentation (6 guides)
```
├── START-HERE.md              ← Quick start (this file)
├── REMOTE-VM-DEPLOYMENT.md    ← Complete deployment guide
├── DEPLOYMENT-SUMMARY.md      ← What's ready
├── HARBOR-QUICKSTART.md       ← Quick Harbor reference
├── PRE-PUSH-CHECKLIST.md      ← Pre-push verification
├── kubernetes/README.md       ← K8s deployment guide
└── docs/
    ├── DEPLOYMENT-PLAN.md     ← Detailed plan
    ├── HARBOR-SETUP.md        ← Harbor guide
    └── HARBOR-CHECKLIST.md    ← Step-by-step
```

---

## ⚡ Quick Commands

### Local Machine
```bash
# Test application
./test-smoke.sh

# Format code
go fmt ./...

# Push to GitHub
git add -A && git commit -m "feat: deployment" && git push
```

### Remote VM
```bash
# Clone and deploy
git clone https://github.com/johnnyr0x/bookstore-app.git
cd bookstore-app
./scripts/harbor-remote-setup.sh v1.0.0

# Deploy to K8s
kubectl apply -f kubernetes/namespace.yaml
kubectl apply -f kubernetes/configmap.yaml
kubectl apply -f kubernetes/postgres.yaml
kubectl apply -f kubernetes/redis.yaml
kubectl apply -f kubernetes/elasticsearch.yaml
kubectl apply -f kubernetes/minio.yaml
kubectl apply -f kubernetes/app.yaml

# Verify
kubectl get pods -n bookstore -w
```

---

## 🎬 Timeline

| Phase | Time | Task |
|-------|------|------|
| **Local** | 5 min | Push to GitHub |
| **Remote** | 2 min | SSH and clone |
| **Harbor** | 2 min | Create project |
| **Build** | 10 min | Run harbor-remote-setup.sh |
| **Deploy** | 10 min | Deploy to Kubernetes |
| **Verify** | 5 min | Test and access |
| **Total** | **~35 min** | Complete deployment |

---

## ✅ Success Checklist

After deployment, verify:

- [ ] All pods running: `kubectl get pods -n bookstore`
- [ ] Application healthy: `kubectl logs -n bookstore deployment/app-deployment`
- [ ] Port forward works: `kubectl port-forward -n bookstore svc/app-service 8080:80`
- [ ] Application accessible: `curl http://localhost:8080/health`
- [ ] Products visible in UI: Open http://localhost:8080

---

## 🆘 If You Need Help

### Quick Troubleshooting

**ImagePullBackOff**:
```bash
kubectl get secret harbor-registry-secret -n bookstore
kubectl describe pod <pod-name> -n bookstore
```

**CrashLoopBackOff**:
```bash
kubectl logs -n bookstore <pod-name>
kubectl describe pod -n bookstore <pod-name>
```

**Service not accessible**:
```bash
kubectl get svc -n bookstore
kubectl get endpoints -n bookstore
```

### Documentation

- **REMOTE-VM-DEPLOYMENT.md** - Troubleshooting section
- **kubernetes/README.md** - Detailed troubleshooting
- **docs/DEPLOYMENT-PLAN.md** - Complete deployment guide

---

## 🚀 Next Steps

After successful deployment:

1. **Data Seeding** - Load 112 books and images (see REMOTE-VM-DEPLOYMENT.md)
2. **Argo CD** - Set up GitOps workflow
3. **Monitoring** - Add Prometheus/Grafana
4. **Admin Console** - Phase 2 feature
5. **AI Assistant** - Phase 2 microservice

---

## 📝 Important Notes

### Security
- ✅ Secrets are auto-generated (32-char passwords)
- ✅ Secrets stored in Kubernetes (encrypted at rest)
- ✅ Harbor uses CA certificate authentication
- ✅ No secrets in Git (verified by .gitignore)

### Storage
- PostgreSQL: 10GB PVC
- Redis: 5GB PVC
- Elasticsearch: 10GB PVC
- MinIO: 20GB PVC
- **Total**: ~45GB required

### Resources
- **CPU**: ~1.5 cores requested, ~5 cores limit
- **Memory**: ~2.5GB requested, ~6.5GB limit
- **Pods**: 3 app replicas (auto-scales 3-10)

---

## 🎯 Your Configuration

- **Harbor URL**: `harbor.corp.vmbeans.com`
- **Harbor Project**: `bookstore`
- **CA Cert**: `/etc/docker/certs.d/harbor.corp.vmbeans.com/ca.crt`
- **K8s Namespace**: `bookstore`
- **Image**: `harbor.corp.vmbeans.com/bookstore/app:v1.0.0`

---

## ✨ Ready to Deploy!

**Current Status**: ✅ All files created and ready

**Next Action**: 
1. Review **PRE-PUSH-CHECKLIST.md**
2. Push to GitHub
3. Follow **REMOTE-VM-DEPLOYMENT.md** on remote VM

---

**Questions?** See **REMOTE-VM-DEPLOYMENT.md** for complete guide with troubleshooting.

**Good luck!** 🚀

