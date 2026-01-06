# Kubernetes Deployment & GitOps Implementation Plan

## Overview

Deploy the bookstore application to Kubernetes with full GitOps workflow using Argo CD, demonstrating VCF 9.0 capabilities with the **current feature set** before adding admin console and AI assistant.

## Current Application Status ✅

**Ready for Production Deployment**:
- ✅ 112 products with real book covers
- ✅ Full e-commerce functionality (cart, checkout, orders)
- ✅ User authentication and profiles
- ✅ Review system with ratings
- ✅ Elasticsearch search with autocomplete
- ✅ Redis caching and sessions
- ✅ MinIO object storage
- ✅ PostgreSQL with 10 migrations
- ✅ Health check endpoints (`/health`, `/health/ready`)
- ✅ Graceful startup with retry logic
- ✅ 25 passing smoke tests

## Phase 1: Container Registry & Image Build

### 1.1 Harbor Setup (VCF Integration)

**Goal**: Store container images in VMware Harbor registry

**Steps**:
```bash
# 1. Access Harbor (VCF provides this)
# URL: https://harbor.vcf.local

# 2. Create project in Harbor
# Project name: bookstore
# Access level: Private

# 3. Create robot account for CI/CD
# Name: bookstore-ci
# Permissions: Push, Pull
# Save the token!

# 4. Login to Harbor from local machine
docker login harbor.vcf.local
# Username: robot$bookstore-ci
# Password: <token from step 3>
```

### 1.2 Build Multi-Architecture Images

**Dockerfile optimization** (already good, but verify):
```dockerfile
# Multi-stage build for minimal image size
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/web

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY templates ./templates
EXPOSE 8080
CMD ["./main"]
```

**Build and push**:
```bash
# Build for AMD64 (most K8s clusters)
docker build -t harbor.vcf.local/bookstore/app:v1.0.0 .
docker push harbor.vcf.local/bookstore/app:v1.0.0

# Tag as latest
docker tag harbor.vcf.local/bookstore/app:v1.0.0 harbor.vcf.local/bookstore/app:latest
docker push harbor.vcf.local/bookstore/app:latest
```

### 1.3 Image Pull Secrets

**Create secret in Kubernetes**:
```bash
kubectl create namespace bookstore

kubectl create secret docker-registry harbor-registry-secret \
  --docker-server=harbor.vcf.local \
  --docker-username=robot\$bookstore-ci \
  --docker-password=<token> \
  --docker-email=admin@bookstore.local \
  -n bookstore
```

## Phase 2: Kubernetes Manifests

### 2.1 Update Existing Manifests

**Current structure**:
```
kubernetes/
├── app.yaml          # Application deployment
├── postgres.yaml     # PostgreSQL StatefulSet
└── secret.yaml       # Secrets (gitignored)
```

**Need to add**:
```
kubernetes/
├── namespace.yaml           # NEW
├── configmap.yaml          # NEW
├── postgres.yaml           # UPDATE
├── redis.yaml              # NEW
├── elasticsearch.yaml      # NEW
├── minio.yaml              # NEW
├── app.yaml                # UPDATE
├── ingress.yaml            # NEW
└── kustomization.yaml      # NEW (for Kustomize)
```

### 2.2 Namespace

**kubernetes/namespace.yaml**:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: bookstore
  labels:
    app: bookstore
    environment: production
```

### 2.3 ConfigMap

**kubernetes/configmap.yaml**:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: bookstore
data:
  DB_HOST: "postgres-service"
  DB_NAME: "bookstore"
  REDIS_URL: "redis-service:6379"
  ES_URL: "http://elasticsearch-service:9200"
  MINIO_ENDPOINT: "minio-service:9000"
  MINIO_USE_SSL: "false"
```

### 2.4 Secrets (Production)

**kubernetes/secrets.yaml** (DO NOT COMMIT):
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: bookstore
type: Opaque
stringData:
  DB_USER: "bookstore_user"
  DB_PASSWORD: "CHANGE_ME_PRODUCTION_PASSWORD"
  MINIO_ACCESS_KEY: "CHANGE_ME_MINIO_KEY"
  MINIO_SECRET_KEY: "CHANGE_ME_MINIO_SECRET"
```

**Create with kubectl**:
```bash
kubectl create secret generic app-secrets \
  --from-literal=DB_USER=bookstore_user \
  --from-literal=DB_PASSWORD=$(openssl rand -base64 32) \
  --from-literal=MINIO_ACCESS_KEY=$(openssl rand -base64 20) \
  --from-literal=MINIO_SECRET_KEY=$(openssl rand -base64 32) \
  -n bookstore
```

### 2.5 PostgreSQL StatefulSet

**kubernetes/postgres.yaml**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: bookstore
spec:
  selector:
    app: postgres
  ports:
    - port: 5432
      targetPort: 5432
  clusterIP: None  # Headless service for StatefulSet
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: bookstore
spec:
  serviceName: postgres-service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
        - name: postgres
          image: postgres:14-alpine
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: DB_USER
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: DB_PASSWORD
            - name: POSTGRES_DB
              value: "bookstore"
            - name: PGDATA
              value: /var/lib/postgresql/data/pgdata
          volumeMounts:
            - name: postgres-storage
              mountPath: /var/lib/postgresql/data
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "1Gi"
              cpu: "1000m"
  volumeClaimTemplates:
    - metadata:
        name: postgres-storage
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName: "vsan-default-storage-policy"  # VCF storage class
        resources:
          requests:
            storage: 10Gi
```

### 2.6 Redis Deployment

**kubernetes/redis.yaml**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: redis-service
  namespace: bookstore
spec:
  selector:
    app: redis
  ports:
    - port: 6379
      targetPort: 6379
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: bookstore
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
        - name: redis
          image: redis:7-alpine
          ports:
            - containerPort: 6379
          command: ["redis-server", "--appendonly", "yes"]
          volumeMounts:
            - name: redis-storage
              mountPath: /data
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "512Mi"
              cpu: "500m"
      volumes:
        - name: redis-storage
          persistentVolumeClaim:
            claimName: redis-pvc
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-pvc
  namespace: bookstore
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: "vsan-default-storage-policy"
  resources:
    requests:
      storage: 5Gi
```

### 2.7 Elasticsearch StatefulSet

**kubernetes/elasticsearch.yaml**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: elasticsearch-service
  namespace: bookstore
spec:
  selector:
    app: elasticsearch
  ports:
    - port: 9200
      targetPort: 9200
  clusterIP: None
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: elasticsearch
  namespace: bookstore
spec:
  serviceName: elasticsearch-service
  replicas: 1
  selector:
    matchLabels:
      app: elasticsearch
  template:
    metadata:
      labels:
        app: elasticsearch
    spec:
      containers:
        - name: elasticsearch
          image: elasticsearch:8.11.0
          ports:
            - containerPort: 9200
          env:
            - name: discovery.type
              value: "single-node"
            - name: xpack.security.enabled
              value: "false"
            - name: ES_JAVA_OPTS
              value: "-Xms512m -Xmx512m"
          volumeMounts:
            - name: es-storage
              mountPath: /usr/share/elasticsearch/data
          resources:
            requests:
              memory: "1Gi"
              cpu: "500m"
            limits:
              memory: "2Gi"
              cpu: "1000m"
  volumeClaimTemplates:
    - metadata:
        name: es-storage
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName: "vsan-default-storage-policy"
        resources:
          requests:
            storage: 10Gi
```

### 2.8 MinIO Deployment

**kubernetes/minio.yaml**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: minio-service
  namespace: bookstore
spec:
  selector:
    app: minio
  ports:
    - name: api
      port: 9000
      targetPort: 9000
    - name: console
      port: 9001
      targetPort: 9001
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
  namespace: bookstore
spec:
  replicas: 1
  selector:
    matchLabels:
      app: minio
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
        - name: minio
          image: minio/minio:latest
          args:
            - server
            - /data
            - --console-address
            - ":9001"
          ports:
            - containerPort: 9000
            - containerPort: 9001
          env:
            - name: MINIO_ROOT_USER
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: MINIO_ACCESS_KEY
            - name: MINIO_ROOT_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: MINIO_SECRET_KEY
          volumeMounts:
            - name: minio-storage
              mountPath: /data
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "1Gi"
              cpu: "1000m"
      volumes:
        - name: minio-storage
          persistentVolumeClaim:
            claimName: minio-pvc
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: minio-pvc
  namespace: bookstore
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: "vsan-default-storage-policy"
  resources:
    requests:
      storage: 20Gi
```

### 2.9 Application Deployment (Updated)

**kubernetes/app.yaml** (enhanced version):
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-deployment
  namespace: bookstore
  labels:
    app: bookstore-app
    version: v1.0.0
spec:
  replicas: 3
  selector:
    matchLabels:
      app: bookstore-app
  template:
    metadata:
      labels:
        app: bookstore-app
        version: v1.0.0
    spec:
      imagePullSecrets:
        - name: harbor-registry-secret
      containers:
        - name: bookstore-app
          image: harbor.vcf.local/bookstore/app:v1.0.0
          ports:
            - containerPort: 8080
              name: http
          envFrom:
            - configMapRef:
                name: app-config
          env:
            - name: DB_USER
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: DB_USER
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: DB_PASSWORD
            - name: MINIO_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: MINIO_ACCESS_KEY
            - name: MINIO_SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: MINIO_SECRET_KEY
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health/ready
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
          startupProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 0
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 30
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "512Mi"
              cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: app-service
  namespace: bookstore
spec:
  selector:
    app: bookstore-app
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: ClusterIP
```

### 2.10 Ingress

**kubernetes/ingress.yaml**:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: bookstore-ingress
  namespace: bookstore
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "false"  # Enable in prod with cert-manager
spec:
  ingressClassName: nginx
  rules:
    - host: bookstore.vcf.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: app-service
                port:
                  number: 80
```

## Phase 3: GitOps with Argo CD

### 3.1 Install Argo CD

```bash
# Create namespace
kubectl create namespace argocd

# Install Argo CD
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Wait for pods to be ready
kubectl wait --for=condition=Ready pods --all -n argocd --timeout=300s

# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Port forward to access UI
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Access: https://localhost:8080
# Username: admin
# Password: (from above command)
```

### 3.2 Create Argo CD Application

**argocd/application.yaml**:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: bookstore
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/johnnyr0x/bookstore-app.git
    targetRevision: HEAD
    path: kubernetes
  destination:
    server: https://kubernetes.default.svc
    namespace: bookstore
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
      allowEmpty: false
    syncOptions:
      - CreateNamespace=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
```

**Apply**:
```bash
kubectl apply -f argocd/application.yaml
```

### 3.3 Repository Structure for GitOps

```
bookstore-app/
├── kubernetes/
│   ├── base/                    # Base manifests
│   │   ├── namespace.yaml
│   │   ├── configmap.yaml
│   │   ├── postgres.yaml
│   │   ├── redis.yaml
│   │   ├── elasticsearch.yaml
│   │   ├── minio.yaml
│   │   ├── app.yaml
│   │   ├── ingress.yaml
│   │   └── kustomization.yaml
│   ├── overlays/
│   │   ├── dev/                 # Development environment
│   │   │   ├── kustomization.yaml
│   │   │   └── patches/
│   │   ├── staging/             # Staging environment
│   │   │   ├── kustomization.yaml
│   │   │   └── patches/
│   │   └── production/          # Production environment
│   │       ├── kustomization.yaml
│   │       └── patches/
│   └── argocd/
│       ├── application-dev.yaml
│       ├── application-staging.yaml
│       └── application-prod.yaml
```

## Phase 4: VCF Integration Demonstrations

### 4.1 vSAN Storage

**Demonstrate**:
- PostgreSQL data persistence
- Elasticsearch index storage
- MinIO object storage
- Redis AOF persistence

**Show**:
```bash
# View PVCs
kubectl get pvc -n bookstore

# Show vSAN storage policy
kubectl describe pvc postgres-storage-postgres-0 -n bookstore
```

### 4.2 NSX Load Balancing

**Demonstrate**:
- Ingress load balancing
- Service discovery
- Network policies (optional)

### 4.3 Harbor Registry

**Demonstrate**:
- Image scanning
- Vulnerability reports
- Image signing (Notary)
- Replication policies

### 4.4 Horizontal Pod Autoscaling

**kubernetes/hpa.yaml**:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: app-hpa
  namespace: bookstore
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: app-deployment
  minReplicas: 3
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

## Phase 5: Deployment Workflow

### 5.1 Initial Deployment

```bash
# 1. Build and push image
docker build -t harbor.vcf.local/bookstore/app:v1.0.0 .
docker push harbor.vcf.local/bookstore/app:v1.0.0

# 2. Create secrets
kubectl create namespace bookstore
kubectl create secret generic app-secrets \
  --from-literal=DB_USER=bookstore_user \
  --from-literal=DB_PASSWORD=$(openssl rand -base64 32) \
  --from-literal=MINIO_ACCESS_KEY=$(openssl rand -base64 20) \
  --from-literal=MINIO_SECRET_KEY=$(openssl rand -base64 32) \
  -n bookstore

kubectl create secret docker-registry harbor-registry-secret \
  --docker-server=harbor.vcf.local \
  --docker-username=robot\$bookstore-ci \
  --docker-password=<token> \
  -n bookstore

# 3. Deploy with Argo CD
kubectl apply -f argocd/application.yaml

# 4. Watch deployment
kubectl get pods -n bookstore -w

# 5. Verify health
kubectl exec -it deploy/app-deployment -n bookstore -- wget -qO- http://localhost:8080/health
```

### 5.2 Data Seeding

```bash
# Port forward to app
kubectl port-forward svc/app-service -n bookstore 8080:80

# Run seed scripts (from local machine)
export MINIO_ENDPOINT=localhost:9000
export DB_HOST=localhost
export DB_USER=bookstore_user
export DB_PASSWORD=<from secret>
export DB_NAME=bookstore

# Seed books
go run scripts/seed-gutenberg-books.go

# Seed images
go run scripts/seed-images.go
```

### 5.3 GitOps Workflow

```bash
# 1. Make code changes locally
# 2. Test locally with docker-compose
./tests/smoke.sh

# 3. Commit and push
git add -A
git commit -m "feat: add new feature"
git push

# 4. Build new image
docker build -t harbor.vcf.local/bookstore/app:v1.1.0 .
docker push harbor.vcf.local/bookstore/app:v1.1.0

# 5. Update kubernetes/app.yaml
# Change image tag to v1.1.0

# 6. Commit and push manifest change
git add kubernetes/app.yaml
git commit -m "deploy: update to v1.1.0"
git push

# 7. Argo CD auto-syncs (or manual sync)
# Watch in Argo CD UI
```

## Phase 6: Demo Script

### Demo Flow (15 minutes)

1. **Show Application** (2 min)
   - Browse products
   - Add to cart
   - Complete checkout
   - Write review

2. **Show Kubernetes** (3 min)
   - `kubectl get pods -n bookstore`
   - `kubectl get pvc -n bookstore`
   - Show vSAN storage in vCenter

3. **Show Harbor** (2 min)
   - Image repository
   - Vulnerability scan results
   - Image layers

4. **Show Argo CD** (3 min)
   - Application health
   - Sync status
   - Resource tree

5. **Demonstrate Self-Healing** (2 min)
   - Delete a pod
   - Watch it recreate
   - Show Argo CD detecting drift

6. **Show Scaling** (2 min)
   - Generate load
   - Watch HPA scale up
   - Show metrics

7. **Rollback Demo** (1 min)
   - Show previous versions in Argo CD
   - Rollback to previous version
   - Verify application still works

## Success Criteria

- ✅ Application deploys successfully to K8s
- ✅ All 5 services running (app, postgres, redis, elasticsearch, minio)
- ✅ Data persists across pod restarts
- ✅ Argo CD syncs from Git repository
- ✅ Self-healing works (pod deletion)
- ✅ HPA scales based on load
- ✅ Images stored in Harbor
- ✅ Health checks working
- ✅ Ingress accessible
- ✅ All 25 smoke tests pass

## Timeline

**Week 1: Kubernetes Deployment**
- Day 1: Update manifests, test locally
- Day 2: Deploy to K8s cluster
- Day 3: Troubleshoot and fix issues
- Day 4: Data seeding and verification
- Day 5: Documentation

**Week 2: GitOps & VCF Integration**
- Day 1: Install and configure Argo CD
- Day 2: Set up GitOps workflow
- Day 3: Harbor integration
- Day 4: HPA and monitoring
- Day 5: Demo preparation and practice

## Next Steps After Deployment

Once deployed and stable:
1. ✅ **Admin Console** - Easier to develop with live K8s environment
2. ✅ **AI Assistant** - Second microservice to demonstrate service mesh
3. ✅ **Prometheus/Grafana** - Metrics and monitoring
4. ✅ **Istio** - Service mesh for AI assistant communication



