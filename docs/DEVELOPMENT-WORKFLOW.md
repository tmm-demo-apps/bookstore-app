# Development Workflow Guide

## Overview

This guide explains how to work with the application in **two environments**:

1. **Local Development** - Docker Compose on your machine (for testing)
2. **Production Deployment** - Kubernetes cluster via remote VM (for deployment)

---

## 🖥️ Local Development (Your Machine)

### Prerequisites

- ✅ Docker Desktop installed and **running**
- ✅ Go 1.25+ installed
- ✅ Git installed

### Quick Start

```bash
# Start all services
./local-dev.sh start

# Wait for services to start, then run tests
./local-dev.sh test

# View logs
./local-dev.sh logs

# Stop services
./local-dev.sh stop
```

### Local Development Commands

```bash
# Start services
./local-dev.sh start          # Start all services
./local-dev.sh stop           # Stop all services
./local-dev.sh restart        # Restart all services
./local-dev.sh status         # Show service status

# Testing
./local-dev.sh test           # Run all 25 smoke tests
./tests/smoke.sh              # Run tests directly (if services running)

# Logs
./local-dev.sh logs           # Show all logs
./local-dev.sh logs app       # Show app logs only
./local-dev.sh logs db        # Show database logs

# Database & Services
./local-dev.sh db             # Open PostgreSQL shell
./local-dev.sh redis          # Open Redis CLI
./local-dev.sh shell          # Open shell in app container

# Maintenance
./local-dev.sh rebuild        # Clean rebuild
./local-dev.sh clean          # Remove all containers and volumes
```

### Local Development Workflow

```bash
# 1. Start Docker Desktop (make sure it's running!)

# 2. Start services
./local-dev.sh start

# 3. Make code changes
# Edit files in cmd/, internal/, templates/, etc.

# 4. Restart to see changes
./local-dev.sh restart

# 5. Run tests
./local-dev.sh test

# 6. View logs if needed
./local-dev.sh logs app

# 7. Stop when done
./local-dev.sh stop
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
# Try again: ./local-dev.sh start
```

#### "Services are not running"
```bash
# Check status
./local-dev.sh status

# View logs
./local-dev.sh logs

# Restart
./local-dev.sh restart
```

#### "Tests are failing"
```bash
# Make sure services are healthy
./local-dev.sh status

# Check app logs
./local-dev.sh logs app

# Verify app is responding
curl http://localhost:8080/health
```

#### "Port already in use"
```bash
# Stop any existing services
./local-dev.sh stop

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
./local-dev.sh start
./local-dev.sh test

# Format code
go fmt ./...

# Stop local services
./local-dev.sh stop
```

#### Step 2: Push to GitHub

```bash
# Commit changes
git add -A
git commit -m "feat: your changes"
git push origin main
```

#### Step 3: Deploy from Remote VM

```bash
# SSH to jumpbox
ssh devops@cli-vm

# Clone or pull latest
git clone https://github.com/johnnyr0x/bookstore-app.git
# OR
cd bookstore-app && git pull

# Build and push to Harbor
./scripts/harbor-remote-setup.sh v1.0.1

# Deploy to Kubernetes
kubectl apply -f kubernetes/namespace.yaml
kubectl apply -f kubernetes/configmap.yaml
kubectl apply -f kubernetes/postgres.yaml
kubectl apply -f kubernetes/redis.yaml
kubectl apply -f kubernetes/elasticsearch.yaml
kubectl apply -f kubernetes/minio.yaml
kubectl apply -f kubernetes/app.yaml

# Verify
kubectl get pods -n bookstore -w
```

---

## 🔄 Complete Development Cycle

### Making Changes

```bash
# 1. Start local environment
./local-dev.sh start

# 2. Make code changes
# Edit your files...

# 3. Test locally
./local-dev.sh restart
./local-dev.sh test

# 4. Commit and push
git add -A
git commit -m "feat: your feature"
git push origin main

# 5. Deploy to K8s (from remote VM)
ssh devops@cli-vm
cd bookstore-app
git pull
./scripts/harbor-remote-setup.sh v1.0.2
kubectl apply -f kubernetes/app.yaml

# 6. Stop local environment
./local-dev.sh stop
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
2. **Use `./local-dev.sh` for all operations**
3. **Run tests before committing**: `./local-dev.sh test`
4. **Check logs when debugging**: `./local-dev.sh logs app`
5. **Stop services when done**: `./local-dev.sh stop`

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
./local-dev.sh test          # Run tests
git status                   # Review changes
```

---

## 🆘 Common Issues

### Issue: "Docker is not running"

**Solution**: Start Docker Desktop application

```bash
# macOS: Open Docker Desktop from Applications
# Wait for Docker icon in menu bar to show "running"
# Then try: ./local-dev.sh start
```

### Issue: "Port 8080 already in use"

**Solution**: Stop existing services

```bash
./local-dev.sh stop

# Or find and kill the process
lsof -i :8080
kill <PID>
```

### Issue: "Tests failing locally"

**Solution**: Check service health

```bash
# Check all services
./local-dev.sh status

# Check app health
curl http://localhost:8080/health

# View app logs
./local-dev.sh logs app

# Restart if needed
./local-dev.sh restart
```

### Issue: "Can't connect to database"

**Solution**: Verify PostgreSQL is running

```bash
# Check database
./local-dev.sh status | grep db

# Test connection
./local-dev.sh db
# Should open psql shell
```

### Issue: "Images not loading"

**Solution**: Check MinIO

```bash
# Check MinIO status
./local-dev.sh status | grep minio

# View MinIO logs
./local-dev.sh logs minio

# Access MinIO console
# http://localhost:9001 (minioadmin/minioadmin)
```

---

## 📝 Quick Reference

### Local Development

```bash
./local-dev.sh start     # Start everything
./local-dev.sh test      # Run tests
./local-dev.sh logs      # View logs
./local-dev.sh stop      # Stop everything
```

### Production Deployment

```bash
# On remote VM
./scripts/harbor-remote-setup.sh v1.0.0    # Build & push
kubectl apply -f kubernetes/               # Deploy
kubectl get pods -n bookstore -w           # Monitor
```

### Testing

```bash
# Local
./local-dev.sh test

# Check specific service
curl http://localhost:8080/health
docker compose exec db psql -U user -d bookstore -c "SELECT COUNT(*) FROM products;"
docker compose exec redis redis-cli PING
```

---

## 🔗 Related Documentation

- **local-dev.sh** - Local development manager (this workflow)
- **START-HERE.md** - Quick start for deployment
- **REMOTE-VM-DEPLOYMENT.md** - Complete K8s deployment guide
- **tests/smoke.sh** - Smoke test suite

---

## 💡 Tips

1. **Use `./local-dev.sh` for everything** - It handles Docker checks and errors
2. **Keep Docker Desktop running** - Services won't work without it
3. **Test locally before deploying** - Catch issues early
4. **Use semantic versioning** - v1.0.0, v1.0.1, v1.1.0, v2.0.0
5. **Monitor logs** - Both locally and in K8s
6. **Clean up when done** - `./local-dev.sh stop` to free resources

---

**Happy coding!** 🚀

