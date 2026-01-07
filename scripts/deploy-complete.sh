#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          Complete Kubernetes Deployment                                    ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

VERSION="${1:-v1.0.0}"

echo "📦 Deployment Version: $VERSION"
echo ""

# Step 1: Build and push to Harbor
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 1: Building and Pushing to Harbor"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
./scripts/harbor-remote-setup.sh "$VERSION"

# Step 2: Deploy infrastructure
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 2: Deploying Infrastructure"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
kubectl apply -f kubernetes/postgres.yaml
kubectl apply -f kubernetes/redis.yaml
kubectl apply -f kubernetes/elasticsearch.yaml
kubectl apply -f kubernetes/minio.yaml

echo ""
echo "⏳ Waiting for infrastructure to be ready..."
kubectl wait --for=condition=Ready pod -l app=postgres -n bookstore --timeout=300s
kubectl wait --for=condition=Ready pod -l app=redis -n bookstore --timeout=300s
kubectl wait --for=condition=Ready pod -l app=elasticsearch -n bookstore --timeout=300s
kubectl wait --for=condition=Ready pod -l app=minio -n bookstore --timeout=300s

echo "✅ Infrastructure ready!"

# Step 3: Deploy application
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 3: Deploying Application"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
kubectl apply -f kubernetes/configmap.yaml
kubectl apply -f kubernetes/app.yaml

echo ""
echo "⏳ Waiting for application to be ready..."
kubectl rollout status deployment/app-deployment -n bookstore

echo "✅ Application deployed!"

# Step 4: Run Database Init Job (migrations + image seeding)
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 4: Database Initialization"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Delete previous job if exists (jobs are immutable)
kubectl delete job init-database -n bookstore --ignore-not-found=true

# Run init job (migrations + seed images from Gutenberg)
kubectl apply -f kubernetes/init-db-job.yaml

echo ""
echo "⏳ Waiting for database initialization to complete..."
echo "   (Runs migrations and downloads 150 book covers from Gutenberg)"
kubectl wait --for=condition=complete job/init-database -n bookstore --timeout=600s

echo "✅ Database initialized!"

# Step 5: Deploy ingress
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 5: Deploying Ingress"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
kubectl apply -f kubernetes/ingress.yaml

# Get ingress info
INGRESS_IP=$(kubectl get ingress bookstore-ingress -n bookstore -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ DEPLOYMENT COMPLETE!                                           ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""
echo "📊 Deployment Summary:"
echo "  - Version: $VERSION"
echo "  - Namespace: bookstore"
echo "  - Ingress IP: $INGRESS_IP"
echo ""
echo "🌐 Access your application:"
echo "  - http://bookstore.corp.vmbeans.com"
echo "  - http://$INGRESS_IP"
echo ""
echo "📊 Check status:"
echo "  kubectl get pods -n bookstore"
echo "  kubectl get svc -n bookstore"
echo "  kubectl get ingress -n bookstore"
echo ""

