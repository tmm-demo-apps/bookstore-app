# Docker Hub Rate Limit Fix

## Problem

You're seeing this error:
```
429 Too Many Requests
toomanyrequests: You have reached your unauthenticated pull rate limit.
```

This happens because Docker Hub limits anonymous pulls to 100 per 6 hours per IP address.

## Solutions

### Option 1: Login to Docker Hub (Recommended)

Free Docker Hub accounts get 200 pulls per 6 hours (2x anonymous limit).

```bash
# On the remote VM
docker login

# Enter your Docker Hub credentials
# Username: your-dockerhub-username
# Password: your-dockerhub-password (or access token)
```

Then retry the build:
```bash
./scripts/harbor-remote-setup.sh v1.0.0
```

### Option 2: Wait and Retry

The rate limit resets after 6 hours. You can:

```bash
# Check when the limit resets
curl -s --head https://registry-1.docker.io/v2/ | grep -i ratelimit

# Wait and retry later
```

### Option 3: Use a Mirror Registry

If your organization has a Docker Hub mirror:

```bash
# Configure Docker to use mirror
sudo nano /etc/docker/daemon.json

# Add:
{
  "registry-mirrors": ["https://your-mirror-url"]
}

# Restart Docker
sudo systemctl restart docker
```

### Option 4: Pre-pull Base Images

Pull the base images separately with retries:

```bash
# Pull Go image
docker pull golang:1.25-alpine

# Pull Alpine image
docker pull alpine:latest

# Then run the build
./scripts/harbor-remote-setup.sh v1.0.0
```

## Prevention

### For Future Builds

1. **Login to Docker Hub** before building
2. **Use a paid Docker Hub account** (unlimited pulls)
3. **Set up a registry mirror** in your infrastructure
4. **Cache base images** in your Harbor registry

### Cache Base Images in Harbor

You can mirror common base images to Harbor:

```bash
# Pull from Docker Hub
docker pull golang:1.25-alpine
docker pull alpine:latest

# Tag for Harbor
docker tag golang:1.25-alpine harbor.corp.vmbeans.com/library/golang:1.25-alpine
docker tag alpine:latest harbor.corp.vmbeans.com/library/alpine:latest

# Push to Harbor
docker push harbor.corp.vmbeans.com/library/golang:1.25-alpine
docker push harbor.corp.vmbeans.com/library/alpine:latest

# Update Dockerfile to use Harbor
FROM harbor.corp.vmbeans.com/library/golang:1.25-alpine AS builder
# ...
FROM harbor.corp.vmbeans.com/library/alpine:latest
```

## Quick Fix for Right Now

**Just login to Docker Hub and retry:**

```bash
docker login
# Username: your-dockerhub-username
# Password: your-password

./scripts/harbor-remote-setup.sh v1.0.0
```

That's it! The build should work now.

