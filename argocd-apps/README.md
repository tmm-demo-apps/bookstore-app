# ArgoCD App-of-Apps

This directory contains ArgoCD Application manifests for orchestrating the multi-app demo suite.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          ArgoCD App-of-Apps                                 │
│                                                                             │
│  ┌─────────────────┐                                                        │
│  │   apps.yaml     │  ← Parent Application                                  │
│  │  (App-of-Apps)  │                                                        │
│  └────────┬────────┘                                                        │
│           │                                                                 │
│     ┌─────┴─────┬────────────┐                                              │
│     │           │            │                                              │
│     ▼           ▼            ▼                                              │
│ ┌─────────┐ ┌─────────┐ ┌─────────┐                                         │
│ │bookstore│ │ reader  │ │ chatbot │  ← Child Applications                   │
│ │  .yaml  │ │  .yaml  │ │  .yaml  │                                         │
│ └────┬────┘ └────┬────┘ └────┬────┘                                         │
│      │           │           │                                              │
│      ▼           ▼           ▼                                              │
│ ┌─────────┐ ┌─────────┐ ┌─────────┐                                         │
│ │bookstore│ │ reader  │ │ chatbot │  ← Git Repos                            │
│ │  -app   │ │  -app   │ │  -app   │                                         │
│ └─────────┘ └─────────┘ └─────────┘                                         │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Usage

### Deploy All Apps

```bash
# Apply the parent app-of-apps
kubectl apply -f apps.yaml -n dev-wrcc9

# ArgoCD will automatically create and sync all child apps
```

### Deploy Individual Apps

```bash
# Apply a single app
kubectl apply -f bookstore.yaml -n dev-wrcc9
kubectl apply -f reader.yaml -n dev-wrcc9
kubectl apply -f chatbot.yaml -n dev-wrcc9
```

### View in ArgoCD UI

1. Open ArgoCD UI: `https://32.32.0.10`
2. Login with admin credentials
3. See all apps in the dashboard
4. Click on "demo-apps" to see the parent app
5. Each child app shows its own sync status

## Sync Strategies

### Automated Sync (Default)

All apps are configured with automated sync:
- `automated.prune: true` - Remove resources deleted from Git
- `automated.selfHeal: true` - Revert manual changes to match Git

### Manual Sync

To disable automated sync for testing:
```yaml
syncPolicy:
  # Remove or comment out 'automated'
  syncOptions:
    - CreateNamespace=true
```

## Upgrade Demo Script

Demonstrate app interdependency during upgrades:

1. **Show all apps healthy** in ArgoCD dashboard
2. **Push change to bookstore-app** - e.g., update version number
3. **Watch CI pipeline** trigger build and Harbor push
4. **Watch ArgoCD sync** - bookstore updates while reader/chatbot stay stable
5. **Test reader app** - still works (calls bookstore API)
6. **Test chatbot** - still responds (graceful degradation)
7. **Bookstore sync completes** - all apps interconnected and healthy

## Files

| File | Description |
|------|-------------|
| `apps.yaml` | Parent App-of-Apps application |
| `bookstore.yaml` | Bookstore application manifest |
| `reader.yaml` | Reader application manifest |
| `chatbot.yaml` | Chatbot application manifest |

## Prerequisites

- ArgoCD installed on Supervisor cluster
- VKS cluster registered with ArgoCD
- Harbor credentials secret in each namespace
- GitHub repos accessible from ArgoCD
