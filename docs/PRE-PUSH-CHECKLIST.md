# Pre-Push Checklist

Before pushing to GitHub for remote VM deployment, verify:

## ✅ Code Quality

- [ ] All smoke tests passing: `./test-smoke.sh`
- [ ] Code formatted: `go fmt ./...`
- [ ] No linter errors
- [ ] No sensitive data in code (passwords, tokens, etc.)

## ✅ Git Status

- [ ] `.gitignore` includes:
  - `dev_docs/`
  - `kubernetes/secret.yaml`
  - `kubernetes/secrets-generated.txt`
  - `.harbor-credentials`
  - `.env`
- [ ] No secrets in git history
- [ ] Commit message follows convention (feat:, fix:, docs:, etc.)

## ✅ Files to Push

**Should be included**:
- ✅ `Dockerfile`
- ✅ `docker-compose.yml`
- ✅ `go.mod` and `go.sum`
- ✅ All application code (`cmd/`, `internal/`, `templates/`)
- ✅ Migration files (`migrations/`)
- ✅ Scripts (`scripts/*.go`, `scripts/*.sh`)
- ✅ Kubernetes manifests (`kubernetes/*.yaml`)
- ✅ Documentation (`docs/`, `README.md`, `*.md`)

**Should NOT be included**:
- ❌ `dev_docs/` (personal notes)
- ❌ `kubernetes/secret.yaml` (secrets)
- ❌ `kubernetes/secrets-generated.txt` (generated secrets)
- ❌ `.harbor-credentials` (Harbor credentials)
- ❌ `.env` (environment variables)
- ❌ `tests/cookies.txt` (test artifacts)

## ✅ Verify Files

```bash
# Check what will be pushed
git status

# Check for sensitive data
git diff

# Search for potential secrets
grep -r "password" --exclude-dir=.git --exclude-dir=dev_docs .
grep -r "secret" --exclude-dir=.git --exclude-dir=dev_docs .
grep -r "token" --exclude-dir=.git --exclude-dir=dev_docs .
```

## ✅ Harbor Scripts Ready

- [ ] `scripts/harbor-remote-setup.sh` exists and is executable
- [ ] `scripts/build-and-push.sh` exists and is executable
- [ ] `REMOTE-VM-DEPLOYMENT.md` has correct Harbor URL
- [ ] `HARBOR-QUICKSTART.md` exists

## ✅ Kubernetes Manifests

Check if you have these files (if not, you'll create them on remote VM):

- [ ] `kubernetes/namespace.yaml`
- [ ] `kubernetes/configmap.yaml`
- [ ] `kubernetes/postgres.yaml`
- [ ] `kubernetes/redis.yaml`
- [ ] `kubernetes/elasticsearch.yaml`
- [ ] `kubernetes/minio.yaml`
- [ ] `kubernetes/app.yaml`
- [ ] `kubernetes/ingress.yaml` (optional)

**Note**: If these don't exist, see `docs/DEPLOYMENT-PLAN.md` for templates.

## ✅ Documentation

- [ ] `README.md` is up to date
- [ ] `REMOTE-VM-DEPLOYMENT.md` has correct Harbor URL
- [ ] `docs/DEPLOYMENT-PLAN.md` exists
- [ ] `docs/HARBOR-SETUP.md` exists

## 🚀 Ready to Push

```bash
# Final check
git status

# Add all files
git add -A

# Commit
git commit -m "feat: add Harbor deployment scripts and K8s manifests"

# Push to GitHub
git push origin main
```

## 📋 After Push - On Remote VM

1. SSH to jumpbox: `ssh devops@cli-vm`
2. Clone or pull repo
3. Follow `REMOTE-VM-DEPLOYMENT.md`
4. Run `./scripts/harbor-remote-setup.sh v1.0.0`
5. Deploy to Kubernetes

---

**Current Status**: Ready to push to GitHub ✅

