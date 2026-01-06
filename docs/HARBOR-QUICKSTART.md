# Harbor Quick Start Guide

## 🚀 Quick Start (3 Steps)

### Step 1: Run Interactive Setup Script

```bash
./scripts/harbor-init.sh
```

This script will:
- ✅ Prompt for Harbor URL, project name, and credentials
- ✅ Login to Harbor registry
- ✅ Build Docker image
- ✅ Push image to Harbor (versioned + latest tags)
- ✅ Optionally create Kubernetes image pull secret

### Step 2: Update Kubernetes Manifest

```bash
# Replace with your actual Harbor URL and version
export HARBOR_URL="harbor.example.com"
export VERSION="v1.0.0"

# Update app.yaml with Harbor image
sed -i '' "s|image:.*|image: ${HARBOR_URL}/bookstore/app:${VERSION}|" kubernetes/app.yaml
```

### Step 3: Deploy to Kubernetes

```bash
# See DEPLOYMENT-PLAN.md for full deployment steps
kubectl apply -f kubernetes/
```

---

## 📋 Manual Process (If You Prefer)

### 1. Login to Harbor

```bash
docker login <your-harbor-url>
# Enter username and password when prompted
```

### 2. Build and Push

```bash
# Set variables
export HARBOR_URL="harbor.example.com"
export PROJECT="bookstore"
export VERSION="v1.0.0"

# Build
docker build -t ${HARBOR_URL}/${PROJECT}/app:${VERSION} .

# Tag as latest
docker tag ${HARBOR_URL}/${PROJECT}/app:${VERSION} ${HARBOR_URL}/${PROJECT}/app:latest

# Push both tags
docker push ${HARBOR_URL}/${PROJECT}/app:${VERSION}
docker push ${HARBOR_URL}/${PROJECT}/app:latest
```

### 3. Create Kubernetes Secret

```bash
kubectl create namespace bookstore

kubectl create secret docker-registry harbor-registry-secret \
  --docker-server=${HARBOR_URL} \
  --docker-username=<your-username> \
  --docker-password=<your-password> \
  -n bookstore
```

---

## 🎯 What You Need

Before starting, have ready:

1. **Harbor URL** - e.g., `harbor.example.com` or `10.0.0.100`
2. **Harbor Credentials** - Username/password or robot account token
3. **Project Name** - Default: `bookstore` (create in Harbor UI first)
4. **Kubernetes Access** - `kubectl` configured and working

---

## 🔍 Verify Everything Works

```bash
# 1. Check Harbor login
docker login ${HARBOR_URL}

# 2. Check image exists locally
docker images | grep ${HARBOR_URL}

# 3. Check image in Harbor UI
# Navigate to: https://${HARBOR_URL}/harbor/projects

# 4. Check Kubernetes secret
kubectl get secret harbor-registry-secret -n bookstore

# 5. Test image pull in Kubernetes
kubectl run test-pull \
  --image=${HARBOR_URL}/bookstore/app:latest \
  --overrides='{"spec":{"imagePullSecrets":[{"name":"harbor-registry-secret"}]}}' \
  -n bookstore --rm -it --restart=Never -- /bin/sh -c "echo 'Success!'"
```

---

## 📚 Detailed Documentation

- **HARBOR-SETUP.md** - Complete Harbor setup guide with troubleshooting
- **HARBOR-CHECKLIST.md** - Step-by-step checklist
- **DEPLOYMENT-PLAN.md** - Full Kubernetes deployment plan
- **NEXT-SESSION.md** - Project roadmap and status

---

## 🆘 Quick Troubleshooting

| Issue | Solution |
|-------|----------|
| `unauthorized: authentication required` | Run `docker login ${HARBOR_URL}` again |
| `denied: requested access to resource` | Check project exists and user has push permissions |
| `ImagePullBackOff` in K8s | Verify secret exists: `kubectl get secret -n bookstore` |
| `x509: certificate signed by unknown authority` | Add Harbor CA cert to Docker (see HARBOR-SETUP.md) |

---

## 💡 Pro Tips

1. **Use Robot Accounts** - Create robot accounts in Harbor for CI/CD (more secure than personal credentials)

2. **Save Credentials** - The `harbor-init.sh` script can save credentials to `~/.harbor-credentials` for reuse

3. **Vulnerability Scanning** - Enable automatic scanning in Harbor project settings

4. **Tag Strategy** - Always push both versioned (`v1.0.0`) and `latest` tags

5. **Test Locally First** - Run `./test-smoke.sh` before building to ensure app works

---

## 🎬 Ready to Start?

```bash
# Option 1: Interactive (Recommended)
./scripts/harbor-init.sh

# Option 2: Automated (if you have credentials saved)
source ~/.harbor-credentials
./scripts/build-and-push.sh ${HARBOR_URL} ${PROJECT} ${VERSION}
```

**Next**: See `DEPLOYMENT-PLAN.md` for deploying to Kubernetes with Argo CD!

