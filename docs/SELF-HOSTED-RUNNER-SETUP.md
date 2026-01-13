# Self-Hosted GitHub Actions Runner Setup

This guide walks through setting up a self-hosted GitHub Actions runner inside your corporate network to enable automated builds and pushes to Harbor.

## Why Self-Hosted?

GitHub's hosted runners cannot access private registries like `harbor.corp.vmbeans.com`. A self-hosted runner inside your network solves this by:

- Running builds within your corporate network
- Direct access to Harbor registry
- Access to internal Kubernetes clusters
- Better security (no credentials leave your network)

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Corporate Network                                 │
│                                                                         │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐              │
│  │  GitHub     │────▶│ Self-Hosted │────▶│   Harbor    │              │
│  │  (Internet) │     │   Runner    │     │  Registry   │              │
│  └─────────────┘     │    (VM)     │     └─────────────┘              │
│        │             └──────┬──────┘            │                      │
│        │                    │                   │                      │
│        │                    ▼                   ▼                      │
│        │             ┌─────────────┐     ┌─────────────┐              │
│        │             │ Kubernetes  │◀────│  Argo CD    │              │
│        │             │   Cluster   │     │  (GitOps)   │              │
│        │             └─────────────┘     └─────────────┘              │
│        │                                                               │
│  ┌─────▼─────┐                                                        │
│  │  Jumpbox  │  (Your RDP access point)                               │
│  └───────────┘                                                        │
└─────────────────────────────────────────────────────────────────────────┘
```

## VM Requirements

### Option A: Dedicated Runner VM (Recommended)

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| vCPUs | 2 | 4 |
| RAM | 4 GB | 8 GB |
| Root Disk | 10 GB | 10 GB |
| PVC (Data) | 50 GB | 100 GB |
| OS | Ubuntu 22.04 LTS | Ubuntu 22.04 LTS |

**Storage Layout:**
- **Root disk** (10 GB): OS, runner binaries, runner work directory
- **PVC** (70-100 GB): Mounted to `/var/lib/docker` for Docker images, Go cache, and build artifacts

### Option B: Use Existing CLI VM

If using your existing `cli-vm`:
- Ensure at least 50 GB free disk space
- Docker already installed ✅
- May compete for resources during builds

## Step 1: Prepare the VM

### 1.1 Create or Access the VM

```bash
# If creating new VM, use your standard provisioning process
# Use cloud-config from: dev_docs/cloud-config-runner.yml
# If using existing cli-vm, SSH in:
ssh devops@cli-vm
```

### 1.2 Mount PVC for Docker and Go Cache (Important!)

If using a separate PVC for storage (recommended):

```bash
# Find your PVC device (usually /dev/sdb)
lsblk

# Stop Docker, format PVC, mount to /var/lib/docker
sudo systemctl stop docker
sudo mkfs.ext4 /dev/sdb
sudo mount /dev/sdb /var/lib/docker
echo '/dev/sdb /var/lib/docker ext4 defaults 0 2' | sudo tee -a /etc/fstab
sudo systemctl start docker

# Create Go cache directories on PVC
sudo mkdir -p /var/lib/docker/go-cache /var/lib/docker/go-mod-cache
sudo chown $USER:$USER /var/lib/docker/go-cache /var/lib/docker/go-mod-cache

# Verify
df -h /var/lib/docker
ls -la /var/lib/docker/go-*
```

**Why this matters:** The CI workflow sets `GOCACHE` and `GOMODCACHE` environment variables to use `/var/lib/docker/go-cache` and `/var/lib/docker/go-mod-cache`. This keeps the root disk free for the OS and runner binaries while the large Go build cache uses the PVC.

### 1.3 Install Prerequisites

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install required packages
sudo apt-get install -y \
    ca-certificates \
    curl \
    wget \
    git \
    jq \
    zip \
    unzip

# Install Docker (if not already installed)
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker $USER

# Log out and back in for docker group to take effect
# Or run: newgrp docker

# Verify Docker
docker --version
docker run hello-world
```

### 1.3 Configure Docker for Harbor

```bash
# Create directory for Harbor CA certificate
sudo mkdir -p /etc/docker/certs.d/harbor.corp.vmbeans.com

# Copy or download the Harbor CA certificate
# Option 1: If you have the cert file
sudo cp /path/to/harbor-ca.crt /etc/docker/certs.d/harbor.corp.vmbeans.com/ca.crt

# Option 2: Download from Harbor (if available)
# curl -k https://harbor.corp.vmbeans.com/api/v2.0/systeminfo/getcert | sudo tee /etc/docker/certs.d/harbor.corp.vmbeans.com/ca.crt

# Test Docker login to Harbor
docker login harbor.corp.vmbeans.com
# Enter your Harbor credentials when prompted
```

## Step 2: Register the GitHub Actions Runner

### 2.1 Get Registration Token from GitHub

1. Go to your GitHub repository: `https://github.com/johnnyr0x/bookstore-app`
2. Navigate to: **Settings → Actions → Runners**
3. Click **"New self-hosted runner"**
4. Select **Linux** and **x64**
5. Copy the registration token (valid for 1 hour)

### 2.2 Download and Configure Runner

```bash
# Create a directory for the runner
mkdir -p ~/actions-runner && cd ~/actions-runner

# Download the latest runner package (check GitHub for current version)
# As of Jan 2026, use the version shown in GitHub UI
curl -o actions-runner-linux-x64-2.321.0.tar.gz -L \
    https://github.com/actions/runner/releases/download/v2.321.0/actions-runner-linux-x64-2.321.0.tar.gz

# Extract
tar xzf ./actions-runner-linux-x64-2.321.0.tar.gz

# Configure the runner
./config.sh --url https://github.com/johnnyr0x/bookstore-app \
    --token YOUR_REGISTRATION_TOKEN_HERE
```

During configuration, you'll be prompted for:
- **Runner group**: Press Enter for default
- **Runner name**: `harbor-builder` (or your preference)
- **Labels**: `self-hosted,linux,x64,harbor` (add `harbor` label!)
- **Work folder**: Press Enter for default `_work`

### 2.3 Install as a Service

```bash
# Install the runner as a systemd service
sudo ./svc.sh install

# Start the service
sudo ./svc.sh start

# Check status
sudo ./svc.sh status

# View logs
journalctl -u actions.runner.johnnyr0x-bookstore-app.harbor-builder.service -f
```

### 2.4 Verify Runner is Online

1. Go to GitHub: **Settings → Actions → Runners**
2. You should see your runner listed as **"Idle"** with a green dot

## Step 3: Configure Runner Environment

### 3.1 Set Up Docker Credentials

The runner needs to authenticate with Harbor. Create a credentials helper:

```bash
# Login to Harbor (this stores credentials in ~/.docker/config.json)
docker login harbor.corp.vmbeans.com

# Verify the config exists
cat ~/.docker/config.json
```

### 3.2 (Optional) Pre-pull Base Images

Speed up builds by pre-pulling commonly used images:

```bash
# Pull Go build image
docker pull golang:1.25-alpine

# Pull Alpine base
docker pull alpine:latest

# Pull from Harbor if mirrored
docker pull harbor.corp.vmbeans.com/library/golang:1.25-alpine
docker pull harbor.corp.vmbeans.com/library/alpine:latest
```

## Step 4: Update GitHub Actions Workflow

The workflow needs to be updated to use the self-hosted runner. See the updated `deploy.yml` in `.github/workflows/`.

Key changes:
```yaml
jobs:
  deploy:
    # Use self-hosted runner with 'harbor' label
    runs-on: [self-hosted, linux, harbor]
```

## Step 5: Test the Setup

### 5.1 Trigger a Manual Deployment

1. Go to GitHub: **Actions → Deploy**
2. Click **"Run workflow"**
3. Enter a version (e.g., `v1.2.1-test`)
4. Click **"Run workflow"**

### 5.2 Monitor the Build

On the runner VM:
```bash
# Watch runner logs
journalctl -u actions.runner.johnnyr0x-bookstore-app.harbor-builder.service -f

# Watch Docker activity
docker stats
```

### 5.3 Verify in Harbor

1. Log into Harbor UI
2. Go to **Projects → bookstore → Repositories**
3. Check for the new image tag

## Maintenance

### Updating the Runner

```bash
cd ~/actions-runner

# Stop the service
sudo ./svc.sh stop

# Download new version
curl -o actions-runner-linux-x64-NEW_VERSION.tar.gz -L \
    https://github.com/actions/runner/releases/download/vNEW_VERSION/actions-runner-linux-x64-NEW_VERSION.tar.gz

# Extract (overwrites existing)
tar xzf ./actions-runner-linux-x64-NEW_VERSION.tar.gz

# Start the service
sudo ./svc.sh start
```

### Cleaning Up Disk Space

```bash
# Remove old Docker images
docker system prune -af

# Clean runner work directory
rm -rf ~/actions-runner/_work/*

# Check disk usage
df -h
```

### Viewing Logs

```bash
# Runner service logs
journalctl -u actions.runner.johnnyr0x-bookstore-app.harbor-builder.service -f

# Docker logs
docker logs <container-id>
```

## Troubleshooting

### Runner shows "Offline"

1. Check service status: `sudo ./svc.sh status`
2. Check network connectivity to GitHub
3. Restart the service: `sudo ./svc.sh stop && sudo ./svc.sh start`

### Docker build fails with disk space error

```bash
# Clean up Docker
docker system prune -af --volumes

# Check disk
df -h

# If still full, remove old runner work
rm -rf ~/actions-runner/_work/*
```

### Cannot push to Harbor

1. Verify Docker login: `docker login harbor.corp.vmbeans.com`
2. Check robot account permissions (Repository: Push, Pull)
3. Verify CA certificate is installed

### Build takes too long

1. Pre-pull base images
2. Enable Docker BuildKit caching
3. Consider adding more vCPUs/RAM to the VM

### Root disk filling up (Go cache)

The CI workflow is configured to use `/var/lib/docker/go-cache` and `/var/lib/docker/go-mod-cache` for Go build caches. If these directories don't exist on the PVC:

```bash
# Check disk usage
df -h

# Create Go cache directories on PVC
sudo mkdir -p /var/lib/docker/go-cache /var/lib/docker/go-mod-cache
sudo chown $USER:$USER /var/lib/docker/go-cache /var/lib/docker/go-mod-cache

# Verify they're on the PVC (should show /dev/sdb or similar)
df -h /var/lib/docker/go-cache
```

If the Go cache is still on the root disk, check that the CI workflow has the environment variables set:
- `GOCACHE=/var/lib/docker/go-cache`
- `GOMODCACHE=/var/lib/docker/go-mod-cache`

## Security Considerations

- **Runner token**: Keep the registration token secure; it's only valid for 1 hour
- **Docker credentials**: Stored in `~/.docker/config.json` - restrict file permissions
- **Network access**: Runner only needs outbound access to GitHub and internal access to Harbor/K8s
- **Updates**: Keep the runner software updated for security patches

## Related Documentation

- [GitHub Actions Setup](./GITHUB-ACTIONS-SETUP.md)
- [Harbor Setup](./HARBOR-SETUP.md)
- [Kubernetes Deployment](../kubernetes/README.md)
