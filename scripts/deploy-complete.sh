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

# Configuration
HARBOR_URL="harbor.corp.vmbeans.com"
HARBOR_PROJECT="bookstore"

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
fi

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

