#!/bin/bash
set -e

# Usage: ./scripts/deploy-complete.sh [VERSION] [OPTIONS]
#
# Options:
#   --skip-build, -s    Skip building and pushing to Harbor (use existing image)
#   --build, -b         Full build and push to Harbor
#
# Examples:
#   ./scripts/deploy-complete.sh                     # Interactive mode (prompts)
#   ./scripts/deploy-complete.sh v1.1.0              # Build, push, and deploy
#   ./scripts/deploy-complete.sh v1.1.0 --skip-build # Deploy existing image only
#   ./scripts/deploy-complete.sh v1.1.0 -s           # Same as above (short flag)

# ============================================================================
# CONFIGURATION - Edit these values for your environment
# ============================================================================
HARBOR_URL="harbor.corp.vmbeans.com"
HARBOR_PROJECT="bookstore"
K8S_NAMESPACE="bookstore"

# Infrastructure image versions (update when you mirror new versions to Harbor)
POSTGRES_IMAGE="${HARBOR_URL}/library/postgres:14-alpine"
REDIS_IMAGE="${HARBOR_URL}/library/redis:7-alpine"
ELASTICSEARCH_IMAGE="${HARBOR_URL}/library/elasticsearch:8.11.0"
MINIO_IMAGE="${HARBOR_URL}/library/minio/minio:latest"
# ============================================================================

# Show help if requested
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    echo "Usage: ./scripts/deploy-complete.sh [VERSION] [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --skip-build, -s    Skip building and pushing to Harbor (use existing image)"
    echo "  --build, -b         Full build and push to Harbor"
    echo "  --help, -h          Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./scripts/deploy-complete.sh                     # Interactive mode"
    echo "  ./scripts/deploy-complete.sh v1.1.0              # Build, push, and deploy"
    echo "  ./scripts/deploy-complete.sh v1.1.0 --skip-build # Deploy existing image only"
    exit 0
fi

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          Complete Kubernetes Deployment                                    ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

# Function to get available tags from Harbor
get_harbor_tags() {
    # Try to get tags from Harbor (requires curl and jq, falls back gracefully)
    local tags=""
    if command -v curl &> /dev/null; then
        tags=$(curl -sk "https://${HARBOR_URL}/api/v2.0/projects/${HARBOR_PROJECT}/repositories/app/artifacts?page_size=10" 2>/dev/null | \
               grep -o '"name":"[^"]*"' | sed 's/"name":"//g' | sed 's/"//g' | head -5 || echo "")
    fi
    echo "$tags"
}

# Function to apply a kubernetes manifest with image substitution
apply_with_images() {
    local file="$1"
    local app_image="${HARBOR_URL}/${HARBOR_PROJECT}/app:${VERSION}"
    
    echo "  Applying: $file"
    
    # Substitute image placeholders and apply
    sed -e "s|{{APP_IMAGE}}|${app_image}|g" \
        -e "s|{{POSTGRES_IMAGE}}|${POSTGRES_IMAGE}|g" \
        -e "s|{{REDIS_IMAGE}}|${REDIS_IMAGE}|g" \
        -e "s|{{ELASTICSEARCH_IMAGE}}|${ELASTICSEARCH_IMAGE}|g" \
        -e "s|{{MINIO_IMAGE}}|${MINIO_IMAGE}|g" \
        "$file" | kubectl apply -f -
}

# Parse command line arguments
VERSION=""
SKIP_BUILD=""

for arg in "$@"; do
    case $arg in
        --skip-build|-s)
            SKIP_BUILD=true
            ;;
        --build|-b)
            SKIP_BUILD=false
            ;;
        v*)
            VERSION="$arg"
            ;;
    esac
done

# Interactive mode if no arguments provided
if [ -z "$VERSION" ] && [ -z "$SKIP_BUILD" ]; then
    echo "🔧 Deployment Mode Selection"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "  1) Full build and push to Harbor (build new image)"
    echo "  2) Skip build, pull from Harbor (use existing image)"
    echo ""
    read -p "Select option [1/2]: " BUILD_OPTION
    
    case $BUILD_OPTION in
        1)
            SKIP_BUILD=false
            echo ""
            echo "📦 Full Build Selected"
            echo ""
            # Try to show latest tag
            echo "Checking Harbor for existing versions..."
            EXISTING_TAGS=$(get_harbor_tags)
            if [ -n "$EXISTING_TAGS" ]; then
                echo "Recent versions in Harbor:"
                echo "$EXISTING_TAGS" | while read tag; do echo "  - $tag"; done
                echo ""
            fi
            read -p "Enter new version tag (e.g., v1.2.0): " VERSION
            VERSION="${VERSION:-v1.0.0}"
            ;;
        2)
            SKIP_BUILD=true
            echo ""
            echo "⏭️  Skip Build Selected"
            echo ""
            # Try to show available tags
            echo "Checking Harbor for available versions..."
            EXISTING_TAGS=$(get_harbor_tags)
            if [ -n "$EXISTING_TAGS" ]; then
                echo "Available versions in Harbor:"
                echo "$EXISTING_TAGS" | nl -w2 -s') '
                echo ""
                echo "  L) Use 'latest' tag"
                echo ""
                read -p "Enter version tag or number from list above: " TAG_INPUT
                
                # Check if user entered a number
                if [[ "$TAG_INPUT" =~ ^[0-9]+$ ]]; then
                    VERSION=$(echo "$EXISTING_TAGS" | sed -n "${TAG_INPUT}p")
                elif [[ "$TAG_INPUT" == "L" || "$TAG_INPUT" == "l" ]]; then
                    VERSION="latest"
                else
                    VERSION="$TAG_INPUT"
                fi
            else
                echo "Could not fetch tags from Harbor (may need authentication)"
                read -p "Enter version tag (e.g., v1.1.0) or 'latest': " VERSION
            fi
            VERSION="${VERSION:-latest}"
            ;;
        *)
            echo "❌ Invalid option. Exiting."
            exit 1
            ;;
    esac
    echo ""
fi

# Default values if still not set
VERSION="${VERSION:-v1.0.0}"
SKIP_BUILD="${SKIP_BUILD:-false}"

echo "📦 Deployment Version: $VERSION"
if [ "$SKIP_BUILD" = true ]; then
    echo "⏭️  Mode: Deploy existing image from Harbor"
else
    echo "🔨 Mode: Full build and push to Harbor"
fi
echo ""

# Step 1: Build and push to Harbor (optional)
if [ "$SKIP_BUILD" = false ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Step 1: Building and Pushing to Harbor"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    ./scripts/harbor-remote-setup.sh "$VERSION"
else
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Step 1: Skipped (using existing image: ${HARBOR_URL}/${HARBOR_PROJECT}/app:$VERSION)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Ensure namespace exists
    if ! kubectl get namespace ${K8S_NAMESPACE} &>/dev/null; then
        echo "Creating namespace: ${K8S_NAMESPACE}"
        kubectl create namespace ${K8S_NAMESPACE}
    fi
    
    # Ensure secrets exist when skipping build
    echo ""
    echo "Checking required secrets..."
    
    # Check harbor-registry-secret
    if ! kubectl get secret harbor-registry-secret -n ${K8S_NAMESPACE} &>/dev/null; then
        echo "⚠️  harbor-registry-secret not found - creating..."
        read -p "Harbor username: " HARBOR_USER
        read -sp "Harbor password: " HARBOR_PASS
        echo ""
        kubectl create secret docker-registry harbor-registry-secret \
            --docker-server="${HARBOR_URL}" \
            --docker-username="${HARBOR_USER}" \
            --docker-password="${HARBOR_PASS}" \
            --docker-email="admin@${K8S_NAMESPACE}.local" \
            -n ${K8S_NAMESPACE}
        echo "✅ harbor-registry-secret created"
    else
        echo "✅ harbor-registry-secret exists"
    fi
    
    # Check app-secrets
    if ! kubectl get secret app-secrets -n ${K8S_NAMESPACE} &>/dev/null; then
        echo "⚠️  app-secrets not found - creating with random passwords..."
        # Use hex encoding to avoid special characters that break URL parsing
        kubectl create secret generic app-secrets \
            --from-literal=DB_USER=bookstore_user \
            --from-literal=DB_PASSWORD=$(openssl rand -hex 16) \
            --from-literal=MINIO_ACCESS_KEY=$(openssl rand -hex 10) \
            --from-literal=MINIO_SECRET_KEY=$(openssl rand -hex 16) \
            -n ${K8S_NAMESPACE}
        echo "✅ app-secrets created"
    else
        echo "✅ app-secrets exists"
    fi
fi

# Step 2: Deploy infrastructure
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 2: Deploying Infrastructure"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Using images:"
echo "  - Postgres: ${POSTGRES_IMAGE}"
echo "  - Redis: ${REDIS_IMAGE}"
echo "  - Elasticsearch: ${ELASTICSEARCH_IMAGE}"
echo "  - MinIO: ${MINIO_IMAGE}"
echo ""

apply_with_images kubernetes/postgres.yaml
apply_with_images kubernetes/redis.yaml
apply_with_images kubernetes/elasticsearch.yaml
apply_with_images kubernetes/minio.yaml

echo ""
echo "⏳ Waiting for infrastructure to be ready..."
kubectl wait --for=condition=Ready pod -l app=postgres -n ${K8S_NAMESPACE} --timeout=300s
kubectl wait --for=condition=Ready pod -l app=redis -n ${K8S_NAMESPACE} --timeout=300s
kubectl wait --for=condition=Ready pod -l app=elasticsearch -n ${K8S_NAMESPACE} --timeout=300s
kubectl wait --for=condition=Ready pod -l app=minio -n ${K8S_NAMESPACE} --timeout=300s

echo "✅ Infrastructure ready!"

# Step 3: Deploy application
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 3: Deploying Application"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Using app image: ${HARBOR_URL}/${HARBOR_PROJECT}/app:${VERSION}"
echo ""

kubectl apply -f kubernetes/configmap.yaml
apply_with_images kubernetes/app.yaml

echo ""
echo "⏳ Waiting for application to be ready..."
kubectl rollout status deployment/app-deployment -n ${K8S_NAMESPACE}

echo "✅ Application deployed!"

# Step 4: Run Database Init Job (migrations + image seeding)
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 4: Database Initialization"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Delete previous job if exists (jobs are immutable)
kubectl delete job init-database -n ${K8S_NAMESPACE} --ignore-not-found=true

# Apply migrations ConfigMap first
kubectl apply -f kubernetes/migrations-configmap.yaml

# Run init job (migrations + seed images from Gutenberg)
apply_with_images kubernetes/init-db-job.yaml

echo ""
echo "⏳ Waiting for database initialization to complete..."
echo "   (Runs migrations and downloads 150 book covers from Gutenberg)"
kubectl wait --for=condition=complete job/init-database -n ${K8S_NAMESPACE} --timeout=600s

echo "✅ Database initialized!"

# Step 5: Deploy ingress
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 5: Deploying Ingress"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
kubectl apply -f kubernetes/ingress.yaml

# Get ingress info
INGRESS_IP=$(kubectl get ingress bookstore-ingress -n ${K8S_NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ DEPLOYMENT COMPLETE!                                           ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""
echo "📊 Deployment Summary:"
echo "  - Version: $VERSION"
echo "  - Namespace: ${K8S_NAMESPACE}"
echo "  - Ingress IP: $INGRESS_IP"
echo ""
echo "🌐 Access your application:"
echo "  - http://${K8S_NAMESPACE}.corp.vmbeans.com"
echo "  - http://$INGRESS_IP"
echo ""
echo "📊 Check status:"
echo "  kubectl get pods -n ${K8S_NAMESPACE}"
echo "  kubectl get svc -n ${K8S_NAMESPACE}"
echo "  kubectl get ingress -n ${K8S_NAMESPACE}"
echo ""

