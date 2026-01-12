# GitHub Actions CI/CD Setup

This guide explains how to configure GitHub Actions for automated builds and deployments to Harbor.

## Overview

The project uses two workflows:

| Workflow | File | Trigger | Purpose |
|----------|------|---------|---------|
| **CI** | `ci.yml` | Push/PR to main/develop | Test, lint, validate |
| **Deploy** | `deploy.yml` | After CI passes on main | Build & push to Harbor |

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Developer  │────▶│   GitHub    │────▶│   Harbor    │────▶│ Kubernetes  │
│  (git push) │     │   Actions   │     │  Registry   │     │  (Argo CD)  │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
                          │                                        ▲
                          │ 1. CI runs tests                       │
                          │ 2. Deploy builds image                 │
                          │ 3. Push to Harbor                      │
                          └────────────────────────────────────────┘
                                    4. Argo CD syncs
```

## Required GitHub Secrets

Navigate to your repository: **Settings → Secrets and variables → Actions → New repository secret**

Add these secrets:

| Secret Name | Value | Example |
|-------------|-------|---------|
| `HARBOR_URL` | Your Harbor registry URL (no https://) | `harbor.corp.vmbeans.com` |
| `HARBOR_USERNAME` | Harbor robot account name | `robot$bookstore-ci` |
| `HARBOR_PASSWORD` | Harbor robot account token | `eyJhbGciOiJS...` |

### Creating a Harbor Robot Account

1. Log into Harbor UI
2. Go to **Projects → bookstore → Robot Accounts**
3. Click **New Robot Account**
4. Configure:
   - Name: `bookstore-ci`
   - Expiration: 365 days (or never)
   - Permissions: Push, Pull, Read artifacts
5. Copy the generated token immediately

## Workflow Behavior

### CI Workflow (`ci.yml`)

Runs on:
- Every push to `main` or `develop`
- Every pull request to `main` or `develop`

Jobs:
1. **Test** - Run Go tests with PostgreSQL
2. **Lint** - Run golangci-lint
3. **Build Docker** - Validate Docker build (no push)

### Deploy Workflow (`deploy.yml`)

Runs on:
- Automatically after CI passes on `main`
- Manual trigger via GitHub UI (workflow_dispatch)

Jobs:
1. **Deploy** - Build and push to Harbor with version tags

### Version Tagging

| Trigger | Tag Format | Example |
|---------|------------|---------|
| Automatic | `v{YYYYMMDD}-{short-sha}` | `v20260112-a1b2c3d` |
| Manual | User-provided | `v1.2.0` |

Both triggers also update the `latest` tag.

## Manual Deployment

To manually trigger a deployment:

1. Go to **Actions → Deploy → Run workflow**
2. Optionally enter a version tag (e.g., `v1.3.0`)
3. Click **Run workflow**

## Viewing Deployment Status

After deployment:
1. Check the **Actions** tab for workflow status
2. View the **Summary** for pushed image tags
3. Verify in Harbor UI under **Projects → bookstore → Repositories**

## Kubernetes Deployment

After the image is pushed to Harbor, deploy to Kubernetes:

### Option A: Manual kubectl

```bash
# Get the version from GitHub Actions summary
VERSION="v20260112-a1b2c3d"

# Update deployment
kubectl set image deployment/app-deployment \
  bookstore-app=harbor.corp.vmbeans.com/bookstore/app:${VERSION} \
  -n bookstore

# Watch rollout
kubectl rollout status deployment/app-deployment -n bookstore
```

### Option B: Argo CD (Recommended)

Configure Argo CD to watch for new images:

```yaml
# In your Argo CD Application
spec:
  source:
    # ... your repo config ...
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

## Troubleshooting

### Build fails with disk space error

The workflow includes disk cleanup, but if it still fails:
- Check if the Go module cache is too large
- Consider using a self-hosted runner with more disk

### Harbor login fails

1. Verify secrets are set correctly (no extra spaces)
2. Check robot account hasn't expired
3. Ensure robot account has push permissions

### Image not appearing in Harbor

1. Check workflow logs for push errors
2. Verify Harbor project name matches (`bookstore`)
3. Check Harbor storage quota

## Security Notes

- Never commit Harbor credentials to the repository
- Use robot accounts with minimal permissions
- Rotate tokens periodically
- Consider using OIDC for keyless authentication (advanced)

## Related Documentation

- [Harbor Setup Guide](./HARBOR-SETUP.md)
- [Kubernetes Deployment](../kubernetes/README.md)
- [Development Workflow](./DEVELOPMENT-WORKFLOW.md)
