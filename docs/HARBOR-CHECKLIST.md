# Harbor Deployment Checklist

## Pre-Deployment Verification ✅

- [ ] Local application tested with `./test-smoke.sh` (all 25 tests passing)
- [ ] Docker installed and running
- [ ] Harbor registry accessible
- [ ] Kubernetes cluster accessible with `kubectl`
- [ ] Code formatted with `go fmt ./...`

## Harbor Setup 🐳

- [ ] Access Harbor UI at your registry URL
- [ ] Create project named `bookstore` (or your preferred name)
- [ ] Create robot account with push/pull permissions
- [ ] Save robot account token securely
- [ ] Login to Harbor from local machine: `docker login <harbor-url>`

## Build and Push 📦

- [ ] Set environment variables:
  ```bash
  export HARBOR_URL="your-harbor-url"
  export HARBOR_PROJECT="bookstore"
  export VERSION="v1.0.0"
  ```
- [ ] Run build script: `./scripts/build-and-push.sh ${HARBOR_URL} ${HARBOR_PROJECT} ${VERSION}`
- [ ] Verify images in Harbor UI (check for `v1.0.0` and `latest` tags)
- [ ] Check vulnerability scan results (if enabled)

## Kubernetes Secrets 🔐

- [ ] Create namespace: `kubectl create namespace bookstore`
- [ ] Create image pull secret:
  ```bash
  kubectl create secret docker-registry harbor-registry-secret \
    --docker-server=${HARBOR_URL} \
    --docker-username=robot\$bookstore-ci \
    --docker-password=<token> \
    -n bookstore
  ```
- [ ] Verify secret: `kubectl get secret harbor-registry-secret -n bookstore`

## Update Manifests 📝

- [ ] Update `kubernetes/app.yaml` with Harbor image URL
- [ ] Add `imagePullSecrets` reference to deployment
- [ ] Commit changes: `git add kubernetes/app.yaml && git commit -m "deploy: use Harbor registry"`

## Test Image Pull 🧪

- [ ] Create test pod to verify image pull
- [ ] Check pod status: `kubectl get pod test-harbor-pull -n bookstore`
- [ ] Clean up test pod

## Next Steps 🚀

- [ ] Review `DEPLOYMENT-PLAN.md` for full Kubernetes deployment
- [ ] Create remaining Kubernetes manifests (postgres, redis, elasticsearch, minio)
- [ ] Set up Argo CD for GitOps
- [ ] Deploy application to cluster

## Quick Commands Reference

```bash
# Login to Harbor
docker login ${HARBOR_URL}

# Build and push (using helper script)
./scripts/build-and-push.sh ${HARBOR_URL} bookstore v1.0.0

# Create namespace
kubectl create namespace bookstore

# Create image pull secret
kubectl create secret docker-registry harbor-registry-secret \
  --docker-server=${HARBOR_URL} \
  --docker-username=robot\$bookstore-ci \
  --docker-password=${HARBOR_TOKEN} \
  -n bookstore

# Verify secret
kubectl get secret harbor-registry-secret -n bookstore

# Test image pull
kubectl run test-pull --image=${HARBOR_URL}/bookstore/app:latest \
  --overrides='{"spec":{"imagePullSecrets":[{"name":"harbor-registry-secret"}]}}' \
  -n bookstore --rm -it --restart=Never -- /bin/sh
```

## Troubleshooting 🔧

If you encounter issues:

1. **Authentication failed**: Re-run `docker login ${HARBOR_URL}`
2. **ImagePullBackOff**: Check secret with `kubectl describe secret harbor-registry-secret -n bookstore`
3. **Build failed**: Ensure all dependencies are available and `go.mod` is up to date
4. **Push failed**: Verify robot account has push permissions in Harbor

## Documentation

- 📖 **HARBOR-SETUP.md** - Detailed Harbor setup guide
- 📖 **DEPLOYMENT-PLAN.md** - Full Kubernetes deployment plan
- 📖 **NEXT-SESSION.md** - Current project status and roadmap

---

**Status**: Ready to build and push to Harbor!

