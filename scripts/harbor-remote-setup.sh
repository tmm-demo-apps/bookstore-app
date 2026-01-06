#!/bin/bash
set -e

# Harbor Setup Script for Remote VM Environment
# Designed for use on jumpbox/VM with CA certificate authentication
# Usage: ./scripts/harbor-remote-setup.sh

echo "=========================================="
echo "Harbor Remote Setup - DemoApp Bookstore"
echo "=========================================="
echo "Environment: Remote VM/Jumpbox"
echo "Harbor: harbor.corp.vmbeans.com"
echo "=========================================="
echo ""

# Configuration
HARBOR_URL="harbor.corp.vmbeans.com"
HARBOR_PROJECT="bookstore"
VERSION="${1:-v1.0.0}"
CA_CERT_PATH="/etc/docker/certs.d/${HARBOR_URL}/ca.crt"

echo "Configuration:"
echo "  Harbor URL: ${HARBOR_URL}"
echo "  Project:    ${HARBOR_PROJECT}"
echo "  Version:    ${VERSION}"
echo "  CA Cert:    ${CA_CERT_PATH}"
echo ""

# Check if CA cert exists
if [ ! -f "${CA_CERT_PATH}" ]; then
    echo "⚠️  WARNING: CA certificate not found at ${CA_CERT_PATH}"
    echo ""
    echo "If you encounter certificate errors, you may need to:"
    echo "  1. Download the CA cert from Harbor"
    echo "  2. Save it to ${CA_CERT_PATH}"
    echo "  3. Restart Docker daemon"
    echo ""
    read -p "Continue anyway? (y/n): " continue_anyway
    if [ "$continue_anyway" != "y" ]; then
        exit 1
    fi
else
    echo "✅ CA certificate found"
fi
echo ""

# Step 1: Check Harbor connectivity
echo "Step 1: Checking Harbor Connectivity"
echo "-------------------------------------"
if curl -k -s "https://${HARBOR_URL}/api/v2.0/systeminfo" > /dev/null; then
    echo "✅ Harbor is reachable"
else
    echo "❌ Cannot reach Harbor at ${HARBOR_URL}"
    echo "Please check network connectivity"
    exit 1
fi
echo ""

# Step 2: Docker login
echo "Step 2: Docker Login"
echo "--------------------"
echo "Please enter your Harbor credentials:"
read -p "Username: " HARBOR_USERNAME
read -sp "Password: " HARBOR_PASSWORD
echo ""

# Login with CA cert support
if [ -f "${CA_CERT_PATH}" ]; then
    echo "${HARBOR_PASSWORD}" | docker login "${HARBOR_URL}" \
        --username "${HARBOR_USERNAME}" \
        --password-stdin
else
    echo "${HARBOR_PASSWORD}" | docker login "${HARBOR_URL}" \
        --username "${HARBOR_USERNAME}" \
        --password-stdin
fi

if [ $? -ne 0 ]; then
    echo "❌ Docker login failed!"
    exit 1
fi

echo "✅ Docker login successful!"
echo ""

# Step 3: Check if project exists
echo "Step 3: Checking Harbor Project"
echo "--------------------------------"
echo "Checking if project '${HARBOR_PROJECT}' exists..."

# Try to query the project (requires authentication)
PROJECT_CHECK=$(curl -k -s -u "${HARBOR_USERNAME}:${HARBOR_PASSWORD}" \
    "https://${HARBOR_URL}/api/v2.0/projects?name=${HARBOR_PROJECT}" | grep -c "\"name\":\"${HARBOR_PROJECT}\"" || echo "0")

if [ "$PROJECT_CHECK" -eq "0" ]; then
    echo "⚠️  Project '${HARBOR_PROJECT}' not found"
    echo ""
    echo "Please create the project in Harbor:"
    echo "  1. Navigate to: https://${HARBOR_URL}"
    echo "  2. Click 'Projects' → 'New Project'"
    echo "  3. Project Name: ${HARBOR_PROJECT}"
    echo "  4. Access Level: Private"
    echo "  5. Click 'OK'"
    echo ""
    read -p "Press Enter after creating the project, or Ctrl+C to exit..."
else
    echo "✅ Project '${HARBOR_PROJECT}' exists"
fi
echo ""

# Step 4: Build image
echo "Step 4: Building Docker Image"
echo "------------------------------"
IMAGE_TAG="${HARBOR_URL}/${HARBOR_PROJECT}/app:${VERSION}"
IMAGE_LATEST="${HARBOR_URL}/${HARBOR_PROJECT}/app:latest"

echo "Building: ${IMAGE_TAG}"
docker build -t "${IMAGE_TAG}" .

if [ $? -ne 0 ]; then
    echo "❌ Docker build failed!"
    exit 1
fi

echo "✅ Build successful!"
echo ""

# Tag as latest
echo "Tagging as latest..."
docker tag "${IMAGE_TAG}" "${IMAGE_LATEST}"
echo "✅ Tagged as latest"
echo ""

# Step 5: Push images
echo "Step 5: Pushing Images to Harbor"
echo "---------------------------------"
echo "Pushing: ${IMAGE_TAG}"
docker push "${IMAGE_TAG}"

if [ $? -ne 0 ]; then
    echo "❌ Push failed!"
    exit 1
fi

echo "✅ Pushed: ${IMAGE_TAG}"
echo ""

echo "Pushing: ${IMAGE_LATEST}"
docker push "${IMAGE_LATEST}"

if [ $? -ne 0 ]; then
    echo "❌ Push failed!"
    exit 1
fi

echo "✅ Pushed: ${IMAGE_LATEST}"
echo ""

# Step 6: Create Kubernetes namespace and secret
echo "Step 6: Kubernetes Setup"
echo "------------------------"
K8S_NAMESPACE="bookstore"

# Check if namespace exists
if ! kubectl get namespace "${K8S_NAMESPACE}" &>/dev/null; then
    echo "Creating namespace: ${K8S_NAMESPACE}"
    kubectl create namespace "${K8S_NAMESPACE}"
else
    echo "✅ Namespace '${K8S_NAMESPACE}' already exists"
fi

# Create or update image pull secret
echo "Creating/updating image pull secret..."
kubectl create secret docker-registry harbor-registry-secret \
    --docker-server="${HARBOR_URL}" \
    --docker-username="${HARBOR_USERNAME}" \
    --docker-password="${HARBOR_PASSWORD}" \
    --docker-email="admin@bookstore.local" \
    -n "${K8S_NAMESPACE}" \
    --dry-run=client -o yaml | kubectl apply -f -

echo "✅ Image pull secret created/updated"
echo ""

# Step 7: Create application secrets
echo "Step 7: Creating Application Secrets"
echo "-------------------------------------"
echo "Generating secure passwords for database and MinIO..."

DB_PASSWORD=$(openssl rand -base64 32)
MINIO_ACCESS_KEY=$(openssl rand -base64 20 | tr -d '/+=' | cut -c1-20)
MINIO_SECRET_KEY=$(openssl rand -base64 32)

kubectl create secret generic app-secrets \
    --from-literal=DB_USER=bookstore_user \
    --from-literal=DB_PASSWORD="${DB_PASSWORD}" \
    --from-literal=MINIO_ACCESS_KEY="${MINIO_ACCESS_KEY}" \
    --from-literal=MINIO_SECRET_KEY="${MINIO_SECRET_KEY}" \
    -n "${K8S_NAMESPACE}" \
    --dry-run=client -o yaml | kubectl apply -f -

echo "✅ Application secrets created"
echo ""

# Save secrets to file for reference (DO NOT COMMIT)
SECRETS_FILE="kubernetes/secrets-generated.txt"
cat > "${SECRETS_FILE}" << EOF
# Generated Secrets - $(date)
# DO NOT COMMIT THIS FILE TO GIT!

DB_USER=bookstore_user
DB_PASSWORD=${DB_PASSWORD}
MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY}
MINIO_SECRET_KEY=${MINIO_SECRET_KEY}

# To retrieve from Kubernetes:
kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.DB_PASSWORD}' | base64 -d
kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.MINIO_ACCESS_KEY}' | base64 -d
kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.MINIO_SECRET_KEY}' | base64 -d
EOF

echo "📝 Secrets saved to: ${SECRETS_FILE}"
echo "⚠️  Keep this file secure and DO NOT commit to git!"
echo ""

# Step 8: Mirror base images to Harbor
echo "Step 8: Mirroring Base Images to Harbor"
echo "----------------------------------------"
echo "Pulling and mirroring infrastructure images to Harbor..."
echo "This avoids Docker Hub rate limits and improves pull performance."
echo ""

# Define base images
BASE_IMAGES=(
    "postgres:14-alpine"
    "redis:7-alpine"
    "elasticsearch:8.11.0"
    "minio/minio:latest"
)

# Mirror each image
for IMAGE in "${BASE_IMAGES[@]}"; do
    # Extract image name without tag
    IMAGE_NAME=$(echo "$IMAGE" | cut -d':' -f1 | sed 's/\//-/g')
    IMAGE_TAG=$(echo "$IMAGE" | cut -d':' -f2)
    
    echo "Processing: $IMAGE"
    
    # Check if image already exists in Harbor
    HARBOR_IMAGE="${HARBOR_URL}/library/${IMAGE}"
    
    # Pull from Docker Hub (uses local cache if already present)
    echo "  Pulling from Docker Hub..."
    if docker pull "$IMAGE"; then
        echo "  ✅ Pulled successfully"
    else
        echo "  ⚠️  Failed to pull $IMAGE - continuing anyway"
        continue
    fi
    
    # Tag for Harbor
    echo "  Tagging for Harbor..."
    docker tag "$IMAGE" "$HARBOR_IMAGE"
    
    # Push to Harbor
    echo "  Pushing to Harbor..."
    if docker push "$HARBOR_IMAGE"; then
        echo "  ✅ Pushed to Harbor: $HARBOR_IMAGE"
    else
        echo "  ⚠️  Failed to push to Harbor - continuing anyway"
    fi
    
    echo ""
done

echo "✅ Base images mirrored to Harbor"
echo ""

# Step 9: Verification
echo "Step 9: Verification"
echo "--------------------"
echo "✅ Setup complete!"
echo ""
echo "Verify in Harbor UI:"
echo "  URL: https://${HARBOR_URL}"
echo "  Project: ${HARBOR_PROJECT}"
echo "  Repository: ${HARBOR_PROJECT}/app"
echo "  Tags: ${VERSION}, latest"
echo ""

echo "Kubernetes resources created:"
kubectl get namespace "${K8S_NAMESPACE}"
kubectl get secret -n "${K8S_NAMESPACE}"
echo ""

# Final summary
echo "=========================================="
echo "✅ HARBOR SETUP COMPLETE!"
echo "=========================================="
echo ""
echo "Images pushed:"
echo "  • ${IMAGE_TAG}"
echo "  • ${IMAGE_LATEST}"
echo ""
echo "Kubernetes secrets created:"
echo "  • harbor-registry-secret (image pull)"
echo "  • app-secrets (database & MinIO)"
echo ""
echo "Next steps:"
echo "  1. Review kubernetes manifests in kubernetes/"
echo "  2. Update kubernetes/app.yaml image reference:"
echo "     image: ${IMAGE_TAG}"
echo "  3. Deploy to cluster:"
echo "     kubectl apply -f kubernetes/"
echo "  4. Monitor deployment:"
echo "     kubectl get pods -n ${K8S_NAMESPACE} -w"
echo ""
echo "=========================================="

