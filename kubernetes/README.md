# Kubernetes Manifests

This directory contains all Kubernetes manifests for deploying the DemoApp bookstore application.

## Files Overview

| File | Description | Required |
|------|-------------|----------|
| `namespace.yaml` | Creates bookstore namespace | ✅ Yes |
| `configmap.yaml` | Environment variables for app | ✅ Yes |
| `postgres.yaml` | PostgreSQL StatefulSet + Service + PVC | ✅ Yes |
| `redis.yaml` | Redis Deployment + Service + PVC | ✅ Yes |
| `elasticsearch.yaml` | Elasticsearch StatefulSet + Service + PVC | ✅ Yes |
| `minio.yaml` | MinIO Deployment + Service + PVC | ✅ Yes |
| `app.yaml` | Application Deployment + Service | ✅ Yes |
| `ingress.yaml` | Ingress for external access | ⚠️ Optional |
| `hpa.yaml` | HorizontalPodAutoscaler | ⚠️ Optional |
| `secret.yaml` | Secrets (gitignored) | ⚠️ Created by script |

## Deployment Order

**IMPORTANT**: Deploy in this order to ensure dependencies are met.

### Step 1: Namespace and Secrets

```bash
# Create namespace
kubectl apply -f namespace.yaml

# Create secrets (done by harbor-remote-setup.sh script)
# - harbor-registry-secret (image pull)
# - app-secrets (DB, MinIO credentials)
```

### Step 2: Infrastructure Services

Deploy in order (wait for each to be ready):

```bash
# PostgreSQL (database)
kubectl apply -f postgres.yaml
kubectl wait --for=condition=Ready pod -l app=postgres -n bookstore --timeout=300s

# Redis (caching & sessions)
kubectl apply -f redis.yaml
kubectl wait --for=condition=Ready pod -l app=redis -n bookstore --timeout=300s

# Elasticsearch (search)
kubectl apply -f elasticsearch.yaml
kubectl wait --for=condition=Ready pod -l app=elasticsearch -n bookstore --timeout=300s

# MinIO (object storage)
kubectl apply -f minio.yaml
kubectl wait --for=condition=Ready pod -l app=minio -n bookstore --timeout=300s
```

### Step 3: Application Configuration

```bash
# ConfigMap (environment variables)
kubectl apply -f configmap.yaml
```

### Step 4: Application

```bash
# Deploy application
kubectl apply -f app.yaml

# Wait for deployment
kubectl rollout status deployment/app-deployment -n bookstore
```

### Step 5: Optional Components

```bash
# Ingress (if you have ingress controller)
kubectl apply -f ingress.yaml

# HPA (if you have metrics-server)
kubectl apply -f hpa.yaml
```

## Quick Deploy (All at Once)

If you're confident everything is ready:

```bash
# Deploy everything except optional components
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
kubectl apply -f elasticsearch.yaml
kubectl apply -f minio.yaml
kubectl apply -f app.yaml

# Watch all pods
kubectl get pods -n bookstore -w
```

## Verification

### Check All Resources

```bash
# View all resources
kubectl get all -n bookstore

# Check persistent volumes
kubectl get pvc -n bookstore

# Check secrets
kubectl get secrets -n bookstore

# Check configmaps
kubectl get configmap -n bookstore
```

### Check Individual Services

```bash
# PostgreSQL
kubectl exec -n bookstore statefulset/postgres -- psql -U bookstore_user -d bookstore -c "SELECT COUNT(*) FROM products;"

# Redis
kubectl exec -n bookstore deployment/redis -- redis-cli PING

# Elasticsearch
kubectl exec -n bookstore statefulset/elasticsearch -- curl -s http://localhost:9200/_cluster/health

# MinIO
kubectl exec -n bookstore deployment/minio -- curl -s http://localhost:9000/minio/health/live

# Application
kubectl exec -n bookstore deployment/app-deployment -- wget -qO- http://localhost:8080/health
```

### Check Logs

```bash
# Application logs
kubectl logs -n bookstore deployment/app-deployment -f

# PostgreSQL logs
kubectl logs -n bookstore statefulset/postgres

# All pods logs
kubectl logs -n bookstore -l app=bookstore-app --tail=50
```

## Accessing the Application

### Port Forward (Development/Testing)

```bash
# Forward application port
kubectl port-forward -n bookstore svc/app-service 8080:80

# Access: http://localhost:8080
```

### Ingress (Production)

If ingress is deployed:

```bash
# Get ingress details
kubectl get ingress -n bookstore

# Access: http://bookstore.corp.vmbeans.com
```

## Scaling

### Manual Scaling

```bash
# Scale application
kubectl scale deployment/app-deployment -n bookstore --replicas=5

# Verify
kubectl get pods -n bookstore -l app=bookstore-app
```

### Auto Scaling (HPA)

```bash
# Apply HPA
kubectl apply -f hpa.yaml

# Check HPA status
kubectl get hpa -n bookstore

# Watch scaling
kubectl get hpa -n bookstore -w
```

## Updating the Application

### Rolling Update

```bash
# Update image version in app.yaml
sed -i 's/:v1.0.0/:v1.0.1/' app.yaml

# Apply changes
kubectl apply -f app.yaml

# Watch rollout
kubectl rollout status deployment/app-deployment -n bookstore

# Check rollout history
kubectl rollout history deployment/app-deployment -n bookstore
```

### Rollback

```bash
# Rollback to previous version
kubectl rollout undo deployment/app-deployment -n bookstore

# Rollback to specific revision
kubectl rollout undo deployment/app-deployment -n bookstore --to-revision=2
```

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl get pods -n bookstore

# Describe pod for events
kubectl describe pod <pod-name> -n bookstore

# Check logs
kubectl logs <pod-name> -n bookstore

# Check previous logs (if pod restarted)
kubectl logs <pod-name> -n bookstore --previous
```

### ImagePullBackOff

```bash
# Check image pull secret
kubectl get secret harbor-registry-secret -n bookstore

# Verify image name in deployment
kubectl get deployment app-deployment -n bookstore -o yaml | grep image:

# Test image pull manually on a node
docker pull harbor.corp.vmbeans.com/bookstore/app:v1.0.0
```

### PVC Pending

```bash
# Check PVC status
kubectl get pvc -n bookstore

# Describe PVC
kubectl describe pvc <pvc-name> -n bookstore

# Check storage classes
kubectl get storageclass

# Check if PV is available
kubectl get pv
```

### Service Not Accessible

```bash
# Check service
kubectl get svc -n bookstore

# Check endpoints
kubectl get endpoints -n bookstore

# Test from within cluster
kubectl run test -n bookstore --rm -it --image=busybox --restart=Never -- wget -qO- http://app-service
```

## Clean Up

### Delete Application Only

```bash
kubectl delete -f app.yaml
kubectl delete -f hpa.yaml
kubectl delete -f ingress.yaml
```

### Delete Everything

```bash
# Delete all resources in namespace
kubectl delete namespace bookstore

# This will delete:
# - All deployments, statefulsets, services
# - All PVCs and PVs
# - All secrets and configmaps
# - The namespace itself
```

### Delete Specific Resources

```bash
# Delete application
kubectl delete deployment app-deployment -n bookstore

# Delete service
kubectl delete service app-service -n bookstore

# Delete PVC (will delete data!)
kubectl delete pvc redis-pvc -n bookstore
```

## Resource Requirements

### Minimum Cluster Resources

- **CPU**: 4 cores minimum (6+ recommended)
- **Memory**: 8GB minimum (16GB+ recommended)
- **Storage**: 50GB minimum for PVCs

### Per-Service Resources

| Service | CPU Request | CPU Limit | Memory Request | Memory Limit | Storage |
|---------|-------------|-----------|----------------|--------------|---------|
| PostgreSQL | 250m | 1000m | 256Mi | 1Gi | 10Gi |
| Redis | 100m | 500m | 128Mi | 512Mi | 5Gi |
| Elasticsearch | 500m | 1000m | 1Gi | 2Gi | 10Gi |
| MinIO | 250m | 1000m | 256Mi | 1Gi | 20Gi |
| App (per pod) | 100m | 500m | 128Mi | 512Mi | - |

**Total** (with 3 app replicas):
- CPU: ~1.5 cores requested, ~5 cores limit
- Memory: ~2.5GB requested, ~6.5GB limit
- Storage: ~45GB

## Notes

- All manifests use the `bookstore` namespace
- Secrets are created by the `harbor-remote-setup.sh` script
- PVCs use default storage class (modify if needed)
- Application uses Harbor registry: `harbor.corp.vmbeans.com`
- Health checks are configured for graceful startup and shutdown

## Related Documentation

- **REMOTE-VM-DEPLOYMENT.md** - Complete deployment guide
- **docs/DEPLOYMENT-PLAN.md** - Detailed deployment plan
- **docs/HARBOR-SETUP.md** - Harbor registry setup

