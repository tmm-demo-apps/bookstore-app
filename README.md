# DemoApp - Bookstore Shopping Cart

A modern, full-stack e-commerce bookstore application built with Go, PostgreSQL, HTMX, and Pico CSS. Features include product browsing, shopping cart management, user authentication, and order processing.

## Features

- 📚 **Product Catalog** - Browse books with grid and table views
- 🛒 **Shopping Cart** - Real-time cart updates with HTMX
- 👤 **User Authentication** - Secure signup/login with session management
- 📦 **Order Management** - Complete checkout flow and order history
- 📊 **Stock Management** - Real-time inventory tracking
- 🎨 **Modern UI** - Clean, responsive design with Pico CSS
- 🚀 **Lightweight** - Fast, efficient Go backend

## Technology Stack

- **Backend**: Go 1.24
- **Frontend**: HTML templates with HTMX and Pico CSS
- **Database**: PostgreSQL 14
- **Container**: Docker & Docker Compose
- **Orchestration**: Kubernetes (optional)

## Quick Start

### Prerequisites

- Docker and Docker Compose
- OR Go 1.24+ and PostgreSQL 14+

### Option 1: Docker Compose (Recommended for Local Development)

```bash
# Clone the repository
git clone <repository-url>
cd DemoApp

# Start the application
docker compose up --build

# Access the application
open http://localhost:8080
```

**⚠️ Security Note**: The `docker-compose.yml` file contains development-only credentials (`user`/`password`). **Never use these in production!**

### Option 2: Local Go Development

```bash
# Set environment variables
export DB_USER=user
export DB_PASSWORD=password
export DB_HOST=localhost
export DB_NAME=bookstore

# Start PostgreSQL (if not already running)
docker run -d -p 5432:5432 \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=bookstore \
  postgres:14-alpine

# Run the application
go run cmd/web/main.go

# Access the application
open http://localhost:8080
```

## Kubernetes Deployment

Deploy the application to a Kubernetes cluster for production or testing.

### Prerequisites

- Kubernetes cluster (minikube, kind, EKS, GKE, AKS, VKS, etc.)
- kubectl configured to access your cluster
- Docker registry (Docker Hub, ECR, GCR, Harbor, etc.)

### Step 1: Build and Push Docker Image

```bash
# Build the Docker image
docker build -t your-registry/bookstore-app:latest .

# Push to your registry
docker push your-registry/bookstore-app:latest
```

### Step 2: Create Kubernetes Secret

Create a secret file for production credentials:

```bash
# Create kubernetes/secret.yaml with your production credentials
cat > kubernetes/secret.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secret
type: Opaque
stringData:
  POSTGRES_USER: "your-production-user"
  POSTGRES_PASSWORD: "your-secure-password"
EOF

# Apply the secret
kubectl apply -f kubernetes/secret.yaml
```

**⚠️ Important**: The `kubernetes/secret.yaml` file is in `.gitignore` and should never be committed to version control!

### Step 3: Update Deployment Configuration

Edit `kubernetes/app.yaml` and replace `your-docker-registry/bookstore-app:latest` with your actual image:

```yaml
image: your-registry/bookstore-app:latest
```

### Step 4: Deploy to Kubernetes

```bash
# Apply all Kubernetes configurations
kubectl apply -f kubernetes/postgres.yaml
kubectl apply -f kubernetes/app.yaml

# Check deployment status
kubectl get pods
kubectl get services

# Get the external IP (for LoadBalancer type)
kubectl get service app-service
```

### Step 5: Access the Application

**For LoadBalancer (Cloud Kubernetes):**
```bash
# Get the external IP
kubectl get service app-service

# Access via: http://<EXTERNAL-IP>
```

**For Minikube:**
```bash
# Get the service URL
minikube service app-service --url

# Access the application
open $(minikube service app-service --url)
```

**For Port Forwarding (any cluster):**
```bash
# Forward local port to the service
kubectl port-forward service/app-service 8080:80

# Access via localhost
open http://localhost:8080
```

### Scaling the Application

```bash
# Scale the app deployment
kubectl scale deployment app-deployment --replicas=3

# Check status
kubectl get pods
```

### Viewing Logs

```bash
# View app logs
kubectl logs -l app=bookstore-app --tail=100 -f

# View database logs
kubectl logs -l app=postgres --tail=100 -f
```

### Updating the Application

```bash
# Build and push new image
docker build -t your-registry/bookstore-app:v2 .
docker push your-registry/bookstore-app:v2

# Update deployment
kubectl set image deployment/app-deployment bookstore-app=your-registry/bookstore-app:v2

# Check rollout status
kubectl rollout status deployment/app-deployment
```

## Database Migrations

The application automatically runs migrations on startup. Migrations are located in the `migrations/` directory and include:

- Initial schema creation (products, cart, orders, users)
- Sample data seeding

## Project Structure

```
DemoApp/
├── cmd/web/              # Application entry point
├── internal/
│   ├── handlers/         # HTTP request handlers
│   ├── models/          # Data models
│   └── repository/      # Database layer
├── templates/           # HTML templates
├── migrations/          # Database migrations
├── kubernetes/          # Kubernetes manifests
├── docker-compose.yml   # Local development setup
├── Dockerfile          # Container image definition
└── README.md
```

## Development

### Running Tests

```bash
go test ./...
```

### Code Formatting

```bash
go fmt ./...
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_USER` | PostgreSQL username | `user` |
| `DB_PASSWORD` | PostgreSQL password | `password` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_NAME` | PostgreSQL database name | `bookstore` |
| `DB_PORT` | PostgreSQL port | `5432` |

## Production Considerations

### Security
- Change default database credentials
- Use Kubernetes secrets for sensitive data
- Enable TLS/SSL for database connections
- Implement rate limiting
- Add authentication middleware

### Scalability
- Use persistent volumes for database data
- Configure resource limits in Kubernetes
- Set up horizontal pod autoscaling
- Use database connection pooling

### Monitoring
- Add health check endpoints
- Configure Prometheus metrics
- Set up log aggregation (ELK, Loki)
- Configure alerts for critical errors

## License

[Add your license here]

## Contributing

[Add contributing guidelines here]
