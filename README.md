# DemoApp - A 12-Factor Bookstore Application

This project, DemoApp, is a demonstration of a bookstore shopping cart application designed and developed with the [Twelve-Factor App methodology](https://12factor.net/). This document outlines how the application adheres to each of the twelve factors.

## Quick Start

### Local Development with Docker Compose

```bash
# Clone the repository
git clone <repository-url>
cd DemoApp

# Start the application
docker compose up --build

# Access the application
open http://localhost:8080
```

**⚠️ Security Note**: The `docker-compose.yml` file contains development-only credentials (`user`/`password`). **Never use these in production!** For production deployments, use proper secrets management (Kubernetes secrets, AWS Secrets Manager, etc.).

### Technology Stack

- **Backend**: Go 1.24
- **Frontend**: HTML templates with HTMX and Pico CSS
- **Database**: PostgreSQL 14
- **Container**: Docker & Docker Compose
- **Orchestration**: Kubernetes (optional)

## The Twelve Factors

### I. Codebase
One codebase tracked in revision control, many deploys. The entire application is stored in a single repository and deployed to various environments (development, staging, production) from this single codebase.

### II. Dependencies
Explicitly declare and isolate dependencies. All dependencies will be explicitly declared and isolated using appropriate dependency management tools for the chosen programming language(s).

### III. Config
Store config in the environment. Configuration will be stored in environment variables, separate from the codebase, and will vary per deploy.

### IV. Backing services
Treat backing services as attached resources. Databases, message queues, caching systems, and other backing services are treated as attached resources, accessible via URLs or other locator/credential stored in environment variables.

### V. Build, release, run
Strictly separate build and run stages. The build stage transforms the codebase into an executable bundle. The release stage combines the build with the deploy's current config. The run stage executes the app as one or more processes.

### VI. Processes
Execute the app as one or more stateless processes. The application will execute as one or more stateless, share-nothing processes. Any necessary state will be stored in a backing service.

### VII. Port binding
Export services via port binding. The application will be entirely self-contained and export its services via port binding. This allows it to be accessible to other services via an assigned port.

### VIII. Concurrency
Scale out via the process model. Concurrency will be managed by scaling out the number of processes, rather than adding more threads within a single process.

### IX. Disposability
Maximize robustness with fast startup and graceful shutdown. Processes will be disposable, meaning they can be started or stopped quickly. This includes fast startup times and graceful shutdowns upon termination.

### X. Dev/prod parity
Keep development, staging, and production as similar as possible. The goal is to minimize the gaps between development and production environments, including using the same backing services and dependencies.

### XI. Logs
Treat logs as event streams. Logs will be treated as continuous streams of aggregated, time-ordered events. The application will not attempt to write to or manage logfiles.

### XII. Admin processes
Run admin/management tasks as one-off processes. Administrative and management tasks (e.g., database migrations, running scripts) will be run as one-off processes in an identical environment to the regular long-running processes.
