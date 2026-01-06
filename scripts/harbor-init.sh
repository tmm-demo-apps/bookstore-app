#!/bin/bash
set -e

# Interactive Harbor Setup Script
# This script guides you through the initial Harbor setup and image push

echo "=========================================="
echo "Harbor Registry Setup - DemoApp Bookstore"
echo "=========================================="
echo ""

# Function to prompt for input with default
prompt_with_default() {
    local prompt="$1"
    local default="$2"
    local value
    
    if [ -n "$default" ]; then
        read -p "$prompt [$default]: " value
        echo "${value:-$default}"
    else
        read -p "$prompt: " value
        echo "$value"
    fi
}

# Check if credentials file exists
CREDS_FILE="$HOME/.harbor-credentials"
if [ -f "$CREDS_FILE" ]; then
    echo "📄 Found existing credentials file: $CREDS_FILE"
    read -p "Load credentials from file? (y/n): " load_creds
    if [ "$load_creds" = "y" ]; then
        source "$CREDS_FILE"
        echo "✅ Loaded credentials from file"
        echo ""
    fi
fi

# Gather Harbor details
echo "Step 1: Harbor Registry Details"
echo "--------------------------------"
HARBOR_URL=$(prompt_with_default "Harbor URL (e.g., harbor.example.com)" "${HARBOR_URL}")
HARBOR_PROJECT=$(prompt_with_default "Harbor Project Name" "${HARBOR_PROJECT:-bookstore}")
VERSION=$(prompt_with_default "Image Version Tag" "${VERSION:-v1.0.0}")
echo ""

# Gather authentication
echo "Step 2: Harbor Authentication"
echo "------------------------------"
echo "You can use either:"
echo "  1. Robot account (recommended for automation)"
echo "  2. Your personal Harbor credentials"
echo ""
read -p "Use robot account? (y/n): " use_robot

if [ "$use_robot" = "y" ]; then
    HARBOR_USERNAME=$(prompt_with_default "Robot Account Name (e.g., robot\$bookstore-ci)" "${HARBOR_ROBOT_NAME}")
    read -sp "Robot Account Token: " HARBOR_PASSWORD
    echo ""
else
    HARBOR_USERNAME=$(prompt_with_default "Harbor Username" "${HARBOR_USERNAME}")
    read -sp "Harbor Password: " HARBOR_PASSWORD
    echo ""
fi
echo ""

# Save credentials option
read -p "Save credentials to $CREDS_FILE? (y/n): " save_creds
if [ "$save_creds" = "y" ]; then
    cat > "$CREDS_FILE" << EOF
# Harbor Credentials for DemoApp
# Generated: $(date)
HARBOR_URL=$HARBOR_URL
HARBOR_PROJECT=$HARBOR_PROJECT
HARBOR_ROBOT_NAME=$HARBOR_USERNAME
HARBOR_ROBOT_TOKEN=$HARBOR_PASSWORD
VERSION=$VERSION
EOF
    chmod 600 "$CREDS_FILE"
    echo "✅ Credentials saved to $CREDS_FILE"
    echo ""
fi

# Summary
echo "=========================================="
echo "Configuration Summary"
echo "=========================================="
echo "Harbor URL:     $HARBOR_URL"
echo "Project:        $HARBOR_PROJECT"
echo "Version:        $VERSION"
echo "Username:       $HARBOR_USERNAME"
echo "Image Name:     $HARBOR_URL/$HARBOR_PROJECT/app:$VERSION"
echo "=========================================="
echo ""

# Confirm
read -p "Proceed with build and push? (y/n): " proceed
if [ "$proceed" != "y" ]; then
    echo "❌ Aborted by user"
    exit 0
fi

# Step 3: Docker login
echo ""
echo "Step 3: Docker Login"
echo "--------------------"
echo "Logging in to Harbor..."
echo "$HARBOR_PASSWORD" | docker login "$HARBOR_URL" --username "$HARBOR_USERNAME" --password-stdin

if [ $? -ne 0 ]; then
    echo "❌ Docker login failed!"
    echo ""
    echo "Troubleshooting:"
    echo "  1. Verify Harbor URL is correct and accessible"
    echo "  2. Check username/password or robot token"
    echo "  3. Ensure Harbor project exists"
    echo "  4. If using self-signed cert, add CA to Docker"
    exit 1
fi

echo "✅ Docker login successful!"
echo ""

# Step 4: Build image
echo "Step 4: Building Docker Image"
echo "------------------------------"
IMAGE_TAG="$HARBOR_URL/$HARBOR_PROJECT/app:$VERSION"
IMAGE_LATEST="$HARBOR_URL/$HARBOR_PROJECT/app:latest"

echo "Building: $IMAGE_TAG"
docker build -t "$IMAGE_TAG" .

if [ $? -ne 0 ]; then
    echo "❌ Docker build failed!"
    exit 1
fi

echo "✅ Build successful!"
echo ""

# Tag as latest
echo "Tagging as latest..."
docker tag "$IMAGE_TAG" "$IMAGE_LATEST"
echo "✅ Tagged as latest"
echo ""

# Step 5: Push images
echo "Step 5: Pushing Images to Harbor"
echo "---------------------------------"
echo "Pushing: $IMAGE_TAG"
docker push "$IMAGE_TAG"

if [ $? -ne 0 ]; then
    echo "❌ Push failed!"
    exit 1
fi

echo "✅ Pushed: $IMAGE_TAG"
echo ""

echo "Pushing: $IMAGE_LATEST"
docker push "$IMAGE_LATEST"

if [ $? -ne 0 ]; then
    echo "❌ Push failed!"
    exit 1
fi

echo "✅ Pushed: $IMAGE_LATEST"
echo ""

# Step 6: Verify in Harbor
echo "Step 6: Verification"
echo "--------------------"
echo "✅ Images successfully pushed to Harbor!"
echo ""
echo "Verify in Harbor UI:"
echo "  1. Navigate to: https://$HARBOR_URL"
echo "  2. Go to Projects → $HARBOR_PROJECT → Repositories"
echo "  3. You should see: $HARBOR_PROJECT/app"
echo "  4. Check tags: $VERSION and latest"
echo ""

# Step 7: Kubernetes setup
echo "Step 7: Kubernetes Setup (Optional)"
echo "------------------------------------"
read -p "Create Kubernetes image pull secret now? (y/n): " create_secret

if [ "$create_secret" = "y" ]; then
    K8S_NAMESPACE=$(prompt_with_default "Kubernetes Namespace" "bookstore")
    
    # Check if namespace exists
    if ! kubectl get namespace "$K8S_NAMESPACE" &>/dev/null; then
        echo "Creating namespace: $K8S_NAMESPACE"
        kubectl create namespace "$K8S_NAMESPACE"
    fi
    
    # Create secret
    echo "Creating image pull secret..."
    kubectl create secret docker-registry harbor-registry-secret \
        --docker-server="$HARBOR_URL" \
        --docker-username="$HARBOR_USERNAME" \
        --docker-password="$HARBOR_PASSWORD" \
        --docker-email="admin@bookstore.local" \
        -n "$K8S_NAMESPACE" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    echo "✅ Image pull secret created!"
    echo ""
    echo "Verify with:"
    echo "  kubectl get secret harbor-registry-secret -n $K8S_NAMESPACE"
    echo ""
fi

# Final summary
echo "=========================================="
echo "✅ SETUP COMPLETE!"
echo "=========================================="
echo ""
echo "Images pushed:"
echo "  • $IMAGE_TAG"
echo "  • $IMAGE_LATEST"
echo ""
echo "Next steps:"
echo "  1. Update kubernetes/app.yaml with image: $IMAGE_TAG"
echo "  2. Review DEPLOYMENT-PLAN.md for full deployment"
echo "  3. Create remaining Kubernetes manifests"
echo "  4. Deploy to cluster with kubectl or Argo CD"
echo ""
echo "Quick commands:"
echo "  # Update app.yaml"
echo "  sed -i '' 's|image:.*|image: $IMAGE_TAG|' kubernetes/app.yaml"
echo ""
echo "  # Deploy to Kubernetes"
echo "  kubectl apply -f kubernetes/"
echo ""
echo "=========================================="

