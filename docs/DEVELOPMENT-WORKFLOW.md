# Development Workflow Guide

**Last Updated**: January 9, 2026

## Overview

This guide explains how to work with the application in **two environments**:

1. **Local Development** - Docker Compose on your machine (for testing)
2. **Production Deployment** - Kubernetes cluster via `deploy-complete.sh`

---

## 🖥️ Local Development (Your Machine)

### Prerequisites

- ✅ Docker Desktop installed and **running**
- ✅ Go 1.25+ installed
- ✅ Git installed

### Quick Start

```bash
# Start all services
./scripts/local-dev.sh start

# Wait for services to start, then run tests
./scripts/local-dev.sh test

# View logs
./scripts/local-dev.sh logs

# Stop services
./scripts/local-dev.sh stop
```

### Local Development Commands

```bash
# Start services
./scripts/local-dev.sh start          # Start all services
./scripts/local-dev.sh stop           # Stop all services
./scripts/local-dev.sh restart        # Restart all services
./scripts/local-dev.sh status         # Show service status

# Testing
./scripts/local-dev.sh test           # Run all 25 smoke tests
./tests/smoke.sh              # Run tests directly (if services running)

# Logs
./scripts/local-dev.sh logs           # Show all logs
./scripts/local-dev.sh logs app       # Show app logs only
./scripts/local-dev.sh logs db        # Show database logs

# Database & Services
./scripts/local-dev.sh db             # Open PostgreSQL shell
./scripts/local-dev.sh redis          # Open Redis CLI
./scripts/local-dev.sh shell          # Open shell in app container

# Maintenance
./scripts/local-dev.sh rebuild        # Clean rebuild
./scripts/local-dev.sh clean          # Remove all containers and volumes
```

### Local Development Workflow

```bash
# 1. Start Docker Desktop (make sure it's running!)

# 2. Start services
./scripts/local-dev.sh start

# 3. Make code changes
# Edit files in cmd/, internal/, templates/, etc.

# 4. Restart to see changes
./scripts/local-dev.sh restart

# 5. Run tests
./scripts/local-dev.sh test

# 6. View logs if needed
./scripts/local-dev.sh logs app

# 7. Stop when done
./scripts/local-dev.sh stop
```

### Local URLs

When running locally:
- **Application**: http://localhost:8080
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)
- **Elasticsearch**: http://localhost:9200
- **PostgreSQL**: localhost:5432 (user/password)
- **Redis**: localhost:6379

### Troubleshooting Local Development

#### "Docker is not running"
```bash
# Start Docker Desktop application
# Wait for it to fully start
# Try again: ./scripts/local-dev.sh start
```

#### "Services are not running"
```bash
# Check status
./scripts/local-dev.sh status

# View logs
./scripts/local-dev.sh logs

# Restart
./scripts/local-dev.sh restart
```

#### "Tests are failing"
```bash
# Make sure services are healthy
./scripts/local-dev.sh status

# Check app logs
./scripts/local-dev.sh logs app

# Verify app is responding
curl http://localhost:8080/health
```

#### "Port already in use"
```bash
# Stop any existing services
./scripts/local-dev.sh stop

# Or find what's using the port
lsof -i :8080
lsof -i :5432
```

---

## ☸️ Production Deployment (Kubernetes)

### Prerequisites

- ✅ Harbor registry accessible
- ✅ Kubernetes cluster accessible
- ✅ kubectl configured
- ✅ Code pushed to GitHub

### Deployment Workflow

#### Step 1: Test Locally First

```bash
# Always test locally before deploying
./scripts/local-dev.sh start
./scripts/local-dev.sh test

# Format code
go fmt ./...

# Stop local services
./scripts/local-dev.sh stop
```

#### Step 2: Push to GitHub

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Commit changes
git add -A
git commit -m "feat: your changes"

# Push branch and create PR
git push origin feature/your-feature-name

# Then create a Pull Request on GitHub for review
# After approval, merge to main
```

#### Step 3: Deploy from Remote VM

```bash
# SSH to jumpbox
ssh devops@cli-vm

# Clone or pull latest
git clone https://github.com/johnnyr0x/bookstore-app.git
# OR
cd bookstore-app && git pull

# One-command deployment (handles everything)
./scripts/deploy-complete.sh v1.1.0 bookstore

# Or deploy to test namespace
./scripts/deploy-complete.sh v1.1.0 bookstore-test

# Verify
kubectl get pods -n bookstore -w
```

The `deploy-complete.sh` script handles:
- Harbor login, image build, and push
- NGINX Ingress Controller installation (if missing)
- Database migrations and seeding
- All Kubernetes manifests
- Dynamic hostname based on namespace

---

## 🔄 Complete Development Cycle

### Making Changes

```bash
# 1. Start local environment
./scripts/local-dev.sh start

# 2. Make code changes
# Edit your files...

# 3. Test locally
./scripts/local-dev.sh restart
./scripts/local-dev.sh test

# 4. Create branch, commit and push
git checkout -b feature/your-feature
git add -A
git commit -m "feat: your feature"
git push origin feature/your-feature
# Create PR on GitHub, get review, merge to main

# 5. Deploy to K8s (from remote VM)
ssh devops@cli-vm
cd bookstore-app
git pull
./scripts/harbor-remote-setup.sh v1.0.2
kubectl apply -f kubernetes/app.yaml

# 6. Stop local environment
./scripts/local-dev.sh stop
```

---

## 📊 Environment Comparison

| Feature | Local (Docker Compose) | Production (Kubernetes) |
|---------|----------------------|------------------------|
| **Purpose** | Development & Testing | Production Deployment |
| **Access** | Your machine | Remote VM |
| **Services** | All in Docker Compose | All in K8s pods |
| **Data** | Temporary (volumes) | Persistent (PVCs) |
| **Images** | Built locally | Pulled from Harbor |
| **Scaling** | Single instance | Auto-scaling (HPA) |
| **URLs** | localhost:8080 | K8s ingress/port-forward |

---

## 🎯 Best Practices

### Local Development

1. **Always start Docker Desktop first**
2. **Use `./scripts/local-dev.sh` for all operations**
3. **Run tests before committing**: `./scripts/local-dev.sh test`
4. **Check logs when debugging**: `./scripts/local-dev.sh logs app`
5. **Stop services when done**: `./scripts/local-dev.sh stop`

### Production Deployment

1. **Test locally first** - Never deploy untested code
2. **Use semantic versioning** - v1.0.0, v1.0.1, v1.1.0
3. **Update image tag in app.yaml** - Match Harbor version
4. **Monitor deployment** - `kubectl get pods -n bookstore -w`
5. **Check logs if issues** - `kubectl logs -n bookstore deployment/app-deployment`

### Code Quality

```bash
# Before every commit
go fmt ./...                  # Format code
./scripts/local-dev.sh test          # Run tests
git status                   # Review changes
```

---

## 🆘 Common Issues

### Issue: "Docker is not running"

**Solution**: Start Docker Desktop application

```bash
# macOS: Open Docker Desktop from Applications
# Wait for Docker icon in menu bar to show "running"
# Then try: ./scripts/local-dev.sh start
```

### Issue: "Port 8080 already in use"

**Solution**: Stop existing services

```bash
./scripts/local-dev.sh stop

# Or find and kill the process
lsof -i :8080
kill <PID>
```

### Issue: "Tests failing locally"

**Solution**: Check service health

```bash
# Check all services
./scripts/local-dev.sh status

# Check app health
curl http://localhost:8080/health

# View app logs
./scripts/local-dev.sh logs app

# Restart if needed
./scripts/local-dev.sh restart
```

### Issue: "Can't connect to database"

**Solution**: Verify PostgreSQL is running

```bash
# Check database
./scripts/local-dev.sh status | grep db

# Test connection
./scripts/local-dev.sh db
# Should open psql shell
```

### Issue: "Images not loading"

**Solution**: Check MinIO

```bash
# Check MinIO status
./scripts/local-dev.sh status | grep minio

# View MinIO logs
./scripts/local-dev.sh logs minio

# Access MinIO console
# http://localhost:9001 (minioadmin/minioadmin)
```

---

## 📝 Quick Reference

### Local Development

```bash
./scripts/local-dev.sh start     # Start everything
./scripts/local-dev.sh test      # Run tests
./scripts/local-dev.sh logs      # View logs
./scripts/local-dev.sh stop      # Stop everything
```

### Production Deployment

```bash
# On remote VM - one command does everything
./scripts/deploy-complete.sh v1.1.0 bookstore

# Check status
kubectl get pods -n bookstore -w
kubectl get ingress -n bookstore
```

### Testing

```bash
# Local
./scripts/local-dev.sh test

# Check specific service
curl http://localhost:8080/health
docker compose exec db psql -U user -d bookstore -c "SELECT COUNT(*) FROM products;"
docker compose exec redis redis-cli PING
```

---

## 🔗 Related Documentation

- **local-dev.sh** - Local development manager
- **deploy-complete.sh** - One-command Kubernetes deployment
- **HARBOR-SETUP.md** - Harbor registry details
- **tests/smoke.sh** - Smoke test suite

---

## 💡 Tips

1. **Use `./scripts/local-dev.sh` for everything** - It handles Docker checks and errors
2. **Keep Docker Desktop running** - Services won't work without it
3. **Test locally before deploying** - Catch issues early
4. **Use semantic versioning** - v1.0.0, v1.0.1, v1.1.0, v2.0.0
5. **Monitor logs** - Both locally and in K8s
6. **Clean up when done** - `./scripts/local-dev.sh stop` to free resources

---

**Happy coding!** 🚀

