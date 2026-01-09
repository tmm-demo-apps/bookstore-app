# Harbor Registry Setup Guide

**Last Updated**: January 9, 2026

## Overview

This guide covers Harbor registry integration for the DemoApp bookstore application. Most Harbor operations are automated via `deploy-complete.sh`, but this document provides details for troubleshooting and manual operations.

## Quick Reference

```bash
# Automated deployment (handles Harbor automatically)
./scripts/deploy-complete.sh v1.1.0 bookstore

# Manual Harbor login
docker login harbor.corp.vmbeans.com

# Check images in Harbor
curl -k https://harbor.corp.vmbeans.com/api/v2.0/projects/bookstore/repositories
```

## Prerequisites

- ✅ Harbor registry accessible at `harbor.corp.vmbeans.com`
- ✅ Docker installed on deployment machine
- ✅ Kubernetes cluster with kubectl access
- ✅ CA certificate at `/etc/docker/certs.d/harbor.corp.vmbeans.com/ca.crt`

## Step 1: Harbor Project Setup

### 1.1 Access Harbor UI

Navigate to your Harbor instance (e.g., `https://harbor.example.com`)

### 1.2 Create Project

1. Click **"Projects"** → **"New Project"**
2. Project Name: `bookstore`
3. Access Level: **Private**
4. Storage Quota: 10GB (adjust as needed)
5. Click **"OK"**

### 1.3 Create Robot Account (Recommended for CI/CD)

1. Navigate to **Projects** → **bookstore** → **Robot Accounts**
2. Click **"New Robot Account"**
3. Name: `bookstore-ci`
4. Expiration: 365 days (or never)
5. Permissions:
   - ✅ Push artifact
   - ✅ Pull artifact
   - ✅ Read artifact
6. Click **"Add"**
7. **IMPORTANT**: Copy the token immediately (you won't see it again!)

Example token format:
```
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 1.4 Save Credentials Securely

```bash
# Save to environment (for current session)
export HARBOR_URL="harbor.example.com"
export HARBOR_PROJECT="bookstore"
export HARBOR_ROBOT_NAME="robot\$bookstore-ci"
export HARBOR_ROBOT_TOKEN="<paste-token-here>"

# Or save to ~/.harbor-credentials (do NOT commit to git)
cat > ~/.harbor-credentials << EOF
HARBOR_URL=harbor.example.com
HARBOR_PROJECT=bookstore
HARBOR_ROBOT_NAME=robot\$bookstore-ci
HARBOR_ROBOT_TOKEN=<paste-token-here>
EOF
chmod 600 ~/.harbor-credentials
```

## Step 2: Docker Login to Harbor

### 2.1 Login with Robot Account

```bash
# Source credentials if saved to file
source ~/.harbor-credentials

# Login to Harbor
echo "${HARBOR_ROBOT_TOKEN}" | docker login ${HARBOR_URL} \
  --username "${HARBOR_ROBOT_NAME}" \
  --password-stdin
```

**Expected output:**
```
Login Succeeded
```

### 2.2 Verify Login

```bash
# Check Docker config
cat ~/.docker/config.json | grep ${HARBOR_URL}
```

## Step 3: Build and Push Application Image

### 3.1 Using the Helper Script (Recommended)

```bash
# Navigate to project root
cd /Users/andrechakj/Documents/Projects/DemoApp

# Run build script
./scripts/build-and-push.sh <harbor-url> <project-name> <version>

# Example:
./scripts/build-and-push.sh harbor.example.com bookstore v1.0.0
```

### 3.2 Manual Build and Push

```bash
# Set variables
HARBOR_URL="harbor.example.com"
PROJECT="bookstore"
VERSION="v1.0.0"

# Build image
docker build -t ${HARBOR_URL}/${PROJECT}/app:${VERSION} .

# Tag as latest
docker tag ${HARBOR_URL}/${PROJECT}/app:${VERSION} \
           ${HARBOR_URL}/${PROJECT}/app:latest

# Push versioned tag
docker push ${HARBOR_URL}/${PROJECT}/app:${VERSION}

# Push latest tag
docker push ${HARBOR_URL}/${PROJECT}/app:latest
```

### 3.3 Verify Push in Harbor UI

1. Navigate to **Projects** → **bookstore** → **Repositories**
2. You should see: `bookstore/app`
3. Click on it to see tags: `v1.0.0` and `latest`
4. Check vulnerability scan results (if enabled)

## Step 4: Create Kubernetes Image Pull Secret

### 4.1 Create Namespace

```bash
kubectl create namespace bookstore
```

### 4.2 Create Image Pull Secret

```bash
# Using robot account credentials
kubectl create secret docker-registry harbor-registry-secret \
  --docker-server=${HARBOR_URL} \
  --docker-username="${HARBOR_ROBOT_NAME}" \
  --docker-password="${HARBOR_ROBOT_TOKEN}" \
  --docker-email=admin@bookstore.local \
  -n bookstore
```

### 4.3 Verify Secret

```bash
kubectl get secret harbor-registry-secret -n bookstore
kubectl describe secret harbor-registry-secret -n bookstore
```

## Step 5: Update Kubernetes Manifests

### 5.1 Update app.yaml

Edit `kubernetes/app.yaml` to use your Harbor image:

```yaml
spec:
  template:
    spec:
      imagePullSecrets:
        - name: harbor-registry-secret
      containers:
        - name: bookstore-app
          image: harbor.example.com/bookstore/app:v1.0.0  # Update this line
```

### 5.2 Commit Changes

```bash
git add kubernetes/app.yaml
git commit -m "deploy: update image to use Harbor registry"
git push
```

## Step 6: Test Image Pull

### 6.1 Test Pod

Create a test pod to verify image pull works:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-harbor-pull
  namespace: bookstore
spec:
  imagePullSecrets:
    - name: harbor-registry-secret
  containers:
    - name: test
      image: ${HARBOR_URL}/${PROJECT}/app:latest
      command: ["/bin/sh", "-c", "echo 'Image pull successful!' && sleep 30"]
  restartPolicy: Never
EOF
```

### 6.2 Check Status

```bash
# Watch pod status
kubectl get pod test-harbor-pull -n bookstore -w

# Check logs
kubectl logs test-harbor-pull -n bookstore

# Clean up
kubectl delete pod test-harbor-pull -n bookstore
```

## Step 7: Harbor Security Features

### 7.1 Enable Vulnerability Scanning

1. In Harbor UI, go to **Projects** → **bookstore** → **Configuration**
2. Enable **"Automatically scan images on push"**
3. Click **"Save"**

### 7.2 View Scan Results

1. Navigate to **Repositories** → **bookstore/app**
2. Click on a tag (e.g., `v1.0.0`)
3. View **Vulnerabilities** tab
4. Review CVE details and severity

### 7.3 Content Trust (Optional)

Enable Docker Content Trust for image signing:

```bash
# Enable content trust
export DOCKER_CONTENT_TRUST=1
export DOCKER_CONTENT_TRUST_SERVER=https://${HARBOR_URL}:4443

# Push signed image
docker push ${HARBOR_URL}/${PROJECT}/app:v1.0.0
```

## Troubleshooting

### Issue: "unauthorized: authentication required"

**Solution**: Re-login to Harbor
```bash
docker logout ${HARBOR_URL}
docker login ${HARBOR_URL}
```

### Issue: "denied: requested access to the resource is denied"

**Solution**: Check robot account permissions in Harbor UI

### Issue: Kubernetes pod stuck in "ImagePullBackOff"

**Solution**: Verify image pull secret
```bash
# Check secret exists
kubectl get secret harbor-registry-secret -n bookstore

# Check pod events
kubectl describe pod <pod-name> -n bookstore

# Verify image name matches exactly
kubectl get deployment app-deployment -n bookstore -o yaml | grep image:
```

### Issue: "x509: certificate signed by unknown authority"

**Solution**: Add Harbor CA certificate to Docker
```bash
# Download Harbor CA cert
curl -k https://${HARBOR_URL}/api/v2.0/systeminfo/getcert > harbor-ca.crt

# Add to Docker
sudo mkdir -p /etc/docker/certs.d/${HARBOR_URL}
sudo cp harbor-ca.crt /etc/docker/certs.d/${HARBOR_URL}/ca.crt

# Restart Docker
sudo systemctl restart docker
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Push to Harbor

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Login to Harbor
        uses: docker/login-action@v2
        with:
          registry: ${{ secrets.HARBOR_URL }}
          username: ${{ secrets.HARBOR_ROBOT_NAME }}
          password: ${{ secrets.HARBOR_ROBOT_TOKEN }}
      
      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: .
          push: true
          tags: |
            ${{ secrets.HARBOR_URL }}/bookstore/app:${{ github.sha }}
            ${{ secrets.HARBOR_URL }}/bookstore/app:latest
```

## Quick Reference

### Common Commands

```bash
# Login to Harbor
docker login ${HARBOR_URL}

# Build image
docker build -t ${HARBOR_URL}/${PROJECT}/app:${VERSION} .

# Push image
docker push ${HARBOR_URL}/${PROJECT}/app:${VERSION}

# Pull image
docker pull ${HARBOR_URL}/${PROJECT}/app:${VERSION}

# List local images
docker images | grep ${HARBOR_URL}

# Remove local image
docker rmi ${HARBOR_URL}/${PROJECT}/app:${VERSION}
```

### Environment Variables

```bash
HARBOR_URL          # Harbor registry URL (e.g., harbor.example.com)
HARBOR_PROJECT      # Project name in Harbor (e.g., bookstore)
HARBOR_ROBOT_NAME   # Robot account username (e.g., robot$bookstore-ci)
HARBOR_ROBOT_TOKEN  # Robot account token (JWT)
```

## Automated Deployment

The `deploy-complete.sh` script handles all Harbor operations:

1. Logs into Harbor automatically
2. Builds and pushes images with proper tags
3. Creates Kubernetes image pull secrets
4. Mirrors base images (postgres, redis, elasticsearch, minio) to Harbor

```bash
# Full deployment with Harbor integration
./scripts/deploy-complete.sh v1.1.0 bookstore

# The script uses these Harbor settings:
# HARBOR_URL=harbor.corp.vmbeans.com
# HARBOR_PROJECT=bookstore
```

## Resources

- [Harbor Documentation](https://goharbor.io/docs/)
- [Docker Registry Authentication](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)

