#!/bin/bash
set -e

# Usage: ./scripts/deploy-complete.sh [VERSION] [OPTIONS]
#
# Options:
#   --skip-build, -s    Skip building and pushing to Harbor (use existing image)
#   --build, -b         Full build and push to Harbor
#   --no-prompt, -y     Skip configuration prompts (use defaults)
#
# Examples:
#   ./scripts/deploy-complete.sh                     # Interactive mode (prompts)
#   ./scripts/deploy-complete.sh v1.1.0              # Build, push, and deploy
#   ./scripts/deploy-complete.sh v1.1.0 --skip-build # Deploy existing image only
#   ./scripts/deploy-complete.sh v1.1.0 -s           # Same as above (short flag)
#   ./scripts/deploy-complete.sh v1.1.0 -y           # Non-interactive with defaults

# ============================================================================
# DEFAULT CONFIGURATION - These can be changed interactively at runtime
# ============================================================================
HARBOR_URL="harbor.corp.vmbeans.com"
HARBOR_PROJECT="bookstore"
K8S_NAMESPACE="bookstore"

# Infrastructure image versions (update when you mirror new versions to Harbor)
POSTGRES_TAG="14-alpine"
REDIS_TAG="7-alpine"
ELASTICSEARCH_TAG="8.11.0"
MINIO_TAG="latest"

# NGINX Ingress Controller versions (kubernetes/ingress-nginx)
# Check latest at: https://github.com/kubernetes/ingress-nginx/releases
NGINX_INGRESS_TAG="v1.14.1"
NGINX_WEBHOOK_TAG="v1.6.5"

# Domain for ingress hostnames (namespace.DOMAIN)
INGRESS_DOMAIN="corp.vmbeans.com"
# ============================================================================

# Show help if requested
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    echo "Usage: ./scripts/deploy-complete.sh [VERSION] [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --skip-build, -s    Skip building and pushing to Harbor (use existing image)"
    echo "  --build, -b         Full build and push to Harbor"
    echo "  --no-prompt, -y     Skip configuration prompts (use defaults)"
    echo "  --help, -h          Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./scripts/deploy-complete.sh                     # Interactive mode"
    echo "  ./scripts/deploy-complete.sh v1.1.0              # Build, push, and deploy"
    echo "  ./scripts/deploy-complete.sh v1.1.0 --skip-build # Deploy existing image only"
    echo "  ./scripts/deploy-complete.sh v1.1.0 -y           # Non-interactive with defaults"
    exit 0
fi

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          Complete Kubernetes Deployment (Kustomize)                        ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

# Check for kustomize
if ! command -v kustomize &> /dev/null; then
    echo "⚠️  kustomize not found. Attempting to use 'kubectl kustomize' instead..."
    KUSTOMIZE_CMD="kubectl kustomize"
else
    KUSTOMIZE_CMD="kustomize"
fi

# Check for --no-prompt flag
NO_PROMPT=false
for arg in "$@"; do
    if [[ "$arg" == "--no-prompt" || "$arg" == "-y" ]]; then
        NO_PROMPT=true
    fi
done

# Interactive configuration
if [ "$NO_PROMPT" = false ]; then
    echo "📋 Current Configuration:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Harbor URL:      ${HARBOR_URL}"
    echo "  Harbor Project:  ${HARBOR_PROJECT}"
    echo "  K8s Namespace:   ${K8S_NAMESPACE}"
    echo "  Ingress Host:    ${K8S_NAMESPACE}.${INGRESS_DOMAIN}"
    echo ""
    echo "  Infrastructure Images:"
    echo "    Postgres:      ${HARBOR_URL}/library/postgres:${POSTGRES_TAG}"
    echo "    Redis:         ${HARBOR_URL}/library/redis:${REDIS_TAG}"
    echo "    Elasticsearch: ${HARBOR_URL}/library/elasticsearch:${ELASTICSEARCH_TAG}"
    echo "    MinIO:         ${HARBOR_URL}/library/minio/minio:${MINIO_TAG}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    read -p "Do you want to change any configuration? (y/N): " CHANGE_CONFIG
    
    if [[ "$CHANGE_CONFIG" =~ ^[Yy]$ ]]; then
        echo ""
        echo "Press Enter to keep the default value shown in brackets."
        echo ""
        
        read -p "Harbor URL [${HARBOR_URL}]: " NEW_HARBOR_URL
        HARBOR_URL="${NEW_HARBOR_URL:-$HARBOR_URL}"
        
        read -p "Harbor Project [${HARBOR_PROJECT}]: " NEW_HARBOR_PROJECT
        HARBOR_PROJECT="${NEW_HARBOR_PROJECT:-$HARBOR_PROJECT}"
        
        read -p "Kubernetes Namespace [${K8S_NAMESPACE}]: " NEW_K8S_NAMESPACE
        K8S_NAMESPACE="${NEW_K8S_NAMESPACE:-$K8S_NAMESPACE}"
        
        echo ""
        echo "Infrastructure Image Tags:"
        
        read -p "Postgres tag [${POSTGRES_TAG}]: " NEW_POSTGRES_TAG
        POSTGRES_TAG="${NEW_POSTGRES_TAG:-$POSTGRES_TAG}"
        
        read -p "Redis tag [${REDIS_TAG}]: " NEW_REDIS_TAG
        REDIS_TAG="${NEW_REDIS_TAG:-$REDIS_TAG}"
        
        read -p "Elasticsearch tag [${ELASTICSEARCH_TAG}]: " NEW_ELASTICSEARCH_TAG
        ELASTICSEARCH_TAG="${NEW_ELASTICSEARCH_TAG:-$ELASTICSEARCH_TAG}"
        
        read -p "MinIO tag [${MINIO_TAG}]: " NEW_MINIO_TAG
        MINIO_TAG="${NEW_MINIO_TAG:-$MINIO_TAG}"
        
        echo ""
        echo "✅ Configuration updated!"
        echo ""
    fi
fi

# Recalculate INGRESS_HOST in case namespace was changed interactively
INGRESS_HOST="${K8S_NAMESPACE}.${INGRESS_DOMAIN}"

# Build full image paths from configuration
POSTGRES_IMAGE="${HARBOR_URL}/library/postgres:${POSTGRES_TAG}"
REDIS_IMAGE="${HARBOR_URL}/library/redis:${REDIS_TAG}"
ELASTICSEARCH_IMAGE="${HARBOR_URL}/library/elasticsearch:${ELASTICSEARCH_TAG}"
MINIO_IMAGE="${HARBOR_URL}/library/minio/minio:${MINIO_TAG}"
NGINX_INGRESS_IMAGE="${HARBOR_URL}/library/ingress-nginx/controller:${NGINX_INGRESS_TAG}"
NGINX_WEBHOOK_IMAGE="${HARBOR_URL}/library/ingress-nginx/kube-webhook-certgen:${NGINX_WEBHOOK_TAG}"

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

# Function to configure kustomize for deployment
configure_kustomize() {
    local app_image="${HARBOR_URL}/${HARBOR_PROJECT}/app:${VERSION}"
    
    echo "  Configuring Kustomize..."
    pushd kubernetes > /dev/null
    
    # Backup original kustomization.yaml
    cp kustomization.yaml kustomization.yaml.bak
    
    # Set namespace
    if command -v kustomize &> /dev/null; then
        kustomize edit set namespace "${K8S_NAMESPACE}"
    else
        # Fallback: use sed to update namespace
        sed -i.tmp "s/^namespace: .*/namespace: ${K8S_NAMESPACE}/" kustomization.yaml
        rm -f kustomization.yaml.tmp
    fi
    
    # Set all images using kustomize edit or sed fallback
    if command -v kustomize &> /dev/null; then
        kustomize edit set image \
            "harbor.corp.vmbeans.com/bookstore/app=${app_image}" \
            "harbor.corp.vmbeans.com/library/postgres=${POSTGRES_IMAGE}" \
            "harbor.corp.vmbeans.com/library/redis=${REDIS_IMAGE}" \
            "harbor.corp.vmbeans.com/library/elasticsearch=${ELASTICSEARCH_IMAGE}" \
            "harbor.corp.vmbeans.com/library/minio/minio=${MINIO_IMAGE}"
    else
        echo "  ⚠️  kustomize not available, images will use defaults from manifests"
    fi
    
    popd > /dev/null
}

# Function to restore kustomization.yaml after deployment
restore_kustomization() {
    if [ -f kubernetes/kustomization.yaml.bak ]; then
        mv kubernetes/kustomization.yaml.bak kubernetes/kustomization.yaml
        echo "  Restored kustomization.yaml to original state"
    fi
}

# Trap to restore kustomization.yaml on exit
trap restore_kustomization EXIT

# Function to deploy using kustomize
deploy_with_kustomize() {
    echo "  Applying manifests with Kustomize..."
    
    if command -v kustomize &> /dev/null; then
        kustomize build kubernetes | kubectl apply -f -
    else
        kubectl apply -k kubernetes
    fi
}

# Function to update ingress host in manifest
update_ingress_host() {
    echo "  Updating ingress host to: ${INGRESS_HOST}"
    sed -i.tmp "s/host: .*/host: ${INGRESS_HOST}/" kubernetes/ingress.yaml
    rm -f kubernetes/ingress.yaml.tmp
}

# Function to restore ingress host after deployment
restore_ingress_host() {
    # Restore to default
    sed -i.tmp "s/host: .*/host: bookstore.corp.vmbeans.com/" kubernetes/ingress.yaml
    rm -f kubernetes/ingress.yaml.tmp
}

# Function to check if NGINX Ingress Controller is installed
check_ingress_controller() {
    if kubectl get namespace ingress-nginx &>/dev/null && \
       kubectl get deployment ingress-nginx-controller -n ingress-nginx &>/dev/null; then
        return 0  # Installed
    else
        return 1  # Not installed
    fi
}

# Function to install NGINX Ingress Controller
install_ingress_controller() {
    echo ""
    echo "📦 Installing NGINX Ingress Controller..."
    echo "   Using images from Harbor:"
    echo "   - Controller: ${NGINX_INGRESS_IMAGE}"
    echo "   - Webhook: ${NGINX_WEBHOOK_IMAGE}"
    echo ""
    
    # Apply the ingress-nginx manifest directly (it has its own namespace)
    kubectl apply -f kubernetes/ingress-nginx.yaml
    
    echo ""
    echo "⏳ Waiting for NGINX Ingress Controller to be ready..."
    kubectl wait --for=condition=Available deployment/ingress-nginx-controller -n ingress-nginx --timeout=300s
    
    echo ""
    echo "⏳ Waiting for LoadBalancer IP assignment..."
    local retries=30
    local lb_ip=""
    while [ $retries -gt 0 ]; do
        lb_ip=$(kubectl get svc ingress-nginx-controller -n ingress-nginx -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
        if [ -n "$lb_ip" ]; then
            echo "✅ NGINX Ingress Controller ready!"
            echo "   LoadBalancer IP: $lb_ip"
            break
        fi
        echo "   Waiting for LoadBalancer IP... ($retries attempts remaining)"
        sleep 10
        retries=$((retries - 1))
    done
    
    if [ -z "$lb_ip" ]; then
        echo "⚠️  LoadBalancer IP not yet assigned. It may take a few more minutes."
        echo "   Check with: kubectl get svc -n ingress-nginx ingress-nginx-controller"
    fi
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
        --no-prompt|-y)
            # Already handled above
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

# Step 2: Configure and deploy with Kustomize
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 2: Deploying Infrastructure and Application with Kustomize"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Using images:"
echo "  - App:           ${HARBOR_URL}/${HARBOR_PROJECT}/app:${VERSION}"
echo "  - Postgres:      ${POSTGRES_IMAGE}"
echo "  - Redis:         ${REDIS_IMAGE}"
echo "  - Elasticsearch: ${ELASTICSEARCH_IMAGE}"
echo "  - MinIO:         ${MINIO_IMAGE}"
echo ""

# Configure kustomize with our settings
configure_kustomize

# Update ingress host if namespace is not default
if [ "${K8S_NAMESPACE}" != "bookstore" ]; then
    update_ingress_host
fi

# Deploy everything with kustomize
deploy_with_kustomize

# Restore ingress host if we changed it
if [ "${K8S_NAMESPACE}" != "bookstore" ]; then
    restore_ingress_host
fi

echo ""
echo "⏳ Waiting for infrastructure to be ready..."
kubectl wait --for=condition=Ready pod -l app=postgres -n ${K8S_NAMESPACE} --timeout=300s
kubectl wait --for=condition=Ready pod -l app=redis -n ${K8S_NAMESPACE} --timeout=300s
kubectl wait --for=condition=Ready pod -l app=elasticsearch -n ${K8S_NAMESPACE} --timeout=300s
kubectl wait --for=condition=Ready pod -l app=minio -n ${K8S_NAMESPACE} --timeout=300s

echo "✅ Infrastructure ready!"

echo ""
echo "⏳ Waiting for application to be ready..."
kubectl rollout status deployment/app-deployment -n ${K8S_NAMESPACE}

echo "✅ Application deployed!"

# Step 3: Run Database Init Job (migrations + image seeding)
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 3: Database Initialization"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Delete previous job if exists (jobs are immutable)
kubectl delete job init-database -n ${K8S_NAMESPACE} --ignore-not-found=true

# Apply init job with kustomize (it will use the namespace we configured)
# We need to apply it separately since it's not in the main kustomization
echo "  Applying init-db-job.yaml..."

# Create a temporary kustomization for the init job
mkdir -p kubernetes/.tmp-init
cat > kubernetes/.tmp-init/kustomization.yaml << EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: ${K8S_NAMESPACE}
resources:
  - ../init-db-job.yaml
images:
  - name: harbor.corp.vmbeans.com/bookstore/app
    newTag: ${VERSION}
  - name: harbor.corp.vmbeans.com/library/postgres
    newTag: ${POSTGRES_TAG}
EOF

if command -v kustomize &> /dev/null; then
    kustomize build kubernetes/.tmp-init | kubectl apply -f -
else
    kubectl apply -k kubernetes/.tmp-init
fi

rm -rf kubernetes/.tmp-init

echo ""
echo "⏳ Waiting for database initialization to complete..."
echo "   (Runs migrations and downloads 150 book covers from Gutenberg)"
kubectl wait --for=condition=complete job/init-database -n ${K8S_NAMESPACE} --timeout=600s

echo "✅ Database initialized!"

# Step 4: Ensure NGINX Ingress Controller is installed
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 4: Checking NGINX Ingress Controller"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if check_ingress_controller; then
    echo "✅ NGINX Ingress Controller already installed"
    INGRESS_LB_IP=$(kubectl get svc ingress-nginx-controller -n ingress-nginx -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")
    echo "   LoadBalancer IP: $INGRESS_LB_IP"
else
    echo "⚠️  NGINX Ingress Controller not found"
    if [ "$NO_PROMPT" = true ]; then
        echo "   Installing automatically (--no-prompt mode)..."
        install_ingress_controller
    else
        read -p "   Install NGINX Ingress Controller? (Y/n): " INSTALL_INGRESS
        if [[ ! "$INSTALL_INGRESS" =~ ^[Nn]$ ]]; then
            install_ingress_controller
        else
            echo "   Skipping ingress controller installation."
            echo "   ⚠️  Ingress will not work without an ingress controller!"
        fi
    fi
fi

# Get ingress info
INGRESS_LB_IP=$(kubectl get svc ingress-nginx-controller -n ingress-nginx -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ DEPLOYMENT COMPLETE!                                           ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""
echo "📊 Deployment Summary:"
echo "  - Version: $VERSION"
echo "  - Namespace: ${K8S_NAMESPACE}"
echo "  - Ingress Host: ${INGRESS_HOST}"
echo "  - LoadBalancer IP: $INGRESS_LB_IP"
echo ""
echo "🌐 Access your application:"
echo "  - http://${INGRESS_HOST}"
echo "  - http://$INGRESS_LB_IP (direct IP access)"
echo ""
echo "📝 DNS Configuration:"
echo "  Add this DNS record: ${INGRESS_HOST} -> $INGRESS_LB_IP"
echo ""
echo "📊 Check status:"
echo "  kubectl get pods -n ${K8S_NAMESPACE}"
echo "  kubectl get svc -n ${K8S_NAMESPACE}"
echo "  kubectl get ingress -n ${K8S_NAMESPACE}"
echo "  kubectl get svc -n ingress-nginx"
echo ""
