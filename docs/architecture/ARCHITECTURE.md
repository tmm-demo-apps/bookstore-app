# Application Architecture & User Flow

## Overview

This document describes the architecture and user flow for the DemoApp e-commerce platform, designed to showcase VMware Cloud Foundation (VCF) 9.0 capabilities.

## Bookstore Application Architecture

The Bookstore is a Go web application following clean architecture principles with the Repository Pattern.

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              BOOKSTORE APPLICATION                                  │
│                                  (Go 1.25)                                          │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│  ┌──────────────────────────────────────────────────────────────────────────────┐   │
│  │                              HTTP Layer                                      │   │
│  │  ┌─────────────────────────────────────────────────────────────────────────┐ │   │
│  │  │                      Go Standard http.ServeMux                          │ │   │
│  │  │    /products  /cart  /checkout  /orders  /profile  /auth  /api/*        │ │   │
│  │  └─────────────────────────────────────────────────────────────────────────┘ │   │
│  │                                     │                                        │   │
│  │  ┌─────────────────────────────────────────────────────────────────────────┐ │   │
│  │  │                         Session Middleware                              │ │   │
│  │  │                   (Redis-backed via redisstore/v9)                      │ │   │
│  │  └─────────────────────────────────────────────────────────────────────────┘ │   │
│  └──────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                            │
│  ┌──────────────────────────────────────────────────────────────────────────────┐   │
│  │                           Handlers Layer                                     │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐            │   │
│  │  │ Products │ │   Cart   │ │ Checkout │ │  Orders  │ │ Profile  │            │   │
│  │  │ Handler  │ │ Handler  │ │ Handler  │ │ Handler  │ │ Handler  │            │   │
│  │  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘            │   │
│  │       │            │            │            │            │                  │   │
│  │  ┌────┴─────┐ ┌────┴─────┐ ┌────┴─────┐ ┌────┴─────┐ ┌────┴─────┐            │   │
│  │  │  Auth    │ │  Images  │ │ Reviews  │ │ Partials │ │  Base    │            │   │
│  │  │ Handler  │ │ Handler  │ │ Handler  │ │ Handler  │ │ Handler  │            │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘            │   │
│  └──────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                            │
│  ┌──────────────────────────────────────────────────────────────────────────────┐   │
│  │                         Repository Layer                                     │   │
│  │  ┌─────────────────────────────────────────────────────────────────────────┐ │   │
│  │  │                    ProductRepository Interface                          │ │   │
│  │  │    GetByID() | List() | Search() | GetCategories() | GetByCategory()    │ │   │
│  │  └─────────────────────────────────────────────────────────────────────────┘ │   │
│  │                          │                    │                              │   │
│  │           ┌──────────────┴──────┐    ┌───────┴────────┐                      │   │
│  │           ▼                     ▼    ▼                                       │   │
│  │  ┌─────────────────┐    ┌─────────────────┐                                  │   │
│  │  │   PostgresRepo  │    │ CachedProdRepo  │◄── Decorator Pattern             │   │
│  │  │   (Primary DB)  │    │ (Redis Cache)   │    (wraps PostgresRepo)          │   │
│  │  └────────┬────────┘    └────────┬────────┘                                  │   │
│  │           │                      │                                           │   │
│  │  ┌────────┴──────────────────────┴────────┐                                  │   │
│  │  │           ElasticsearchRepo             │◄── Full-text Search             │   │
│  │  │   Search() | Autocomplete() | Index()  │    (5-tier strategy)             │   │
│  │  └────────────────────────────────────────┘                                  │   │
│  └──────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                            │
│  ┌──────────────────────────────────────────────────────────────────────────────┐   │
│  │                           Models Layer                                       │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐            │   │
│  │  │ Product  │ │   User   │ │   Cart   │ │  Order   │ │  Review  │            │   │
│  │  │  Model   │ │  Model   │ │  Model   │ │  Model   │ │  Model   │            │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘            │   │
│  └──────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              EXTERNAL SERVICES                                      │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌───────────────┐   │
│  │   PostgreSQL    │  │     Redis       │  │  Elasticsearch  │  │     MinIO     │   │
│  │                 │  │                 │  │                 │  │   (S3-compat) │   │
│  ├─────────────────┤  ├─────────────────┤  ├─────────────────┤  ├───────────────┤   │
│  │ • Users         │  │ • Sessions      │  │ • Product Index │  │ • Book Covers │   │
│  │ • Products      │  │ • Product Cache │  │ • Autocomplete  │  │ • 1-year TTL  │   │
│  │ • Orders        │  │ • Cart Data     │  │ • 5-tier Search │  │ • ETags       │   │
│  │ • Reviews       │  │ • Rate Limits   │  │ • Author Search │  │ • Immutable   │   │
│  │ • Cart Items    │  │                 │  │                 │  │               │   │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘  └───────┬───────┘   │
│           │                    │                    │                   │           │
│           │                Port 5432          Port 9200           Port 9000         │
│           │                    │                    │                   │           │
└───────────┼────────────────────┼────────────────────┼───────────────────┼───────────┘
            │                    │                    │                   │
            └────────────────────┴────────────────────┴───────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                FRONTEND LAYER                                       │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │                          HTML Templates (Go html/template)                  │    │
│  │  base.html → layout with header, nav, footer                                │    │
│  │  ├── products.html, product.html (catalog views)                            │    │
│  │  ├── cart.html, checkout.html (shopping flow)                               │    │
│  │  ├── orders.html, order.html (order history)                                │    │
│  │  ├── profile.html, login.html, register.html (auth)                         │    │
│  │  └── partials/*.html (HTMX fragments)                                       │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
│                                         │                                           │
│  ┌──────────────────────────────────────┴──────────────────────────────────────┐    │
│  │                              Client-Side Tech                               │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │    │
│  │  │  Pico.css   │  │    HTMX     │  │ Dark Mode   │  │  Lazy Loading       │ │    │
│  │  │ (minimal)   │  │ (dynamic)   │  │ (toggle)    │  │  (images)           │ │    │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
│                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### Component Summary

| Layer | Components | Responsibilities |
|-------|------------|------------------|
| **HTTP** | http.ServeMux, redisstore Sessions | Routing, request handling, session management |
| **Handlers** | Products, Cart, Checkout, Orders, Profile, Auth, Images, Reviews | Business logic, request/response mapping |
| **Repository** | PostgresRepo, CachedProductRepo, ElasticsearchRepo | Data access, caching, search |
| **Models** | Product, User, Cart, Order, Review | Domain entities |
| **Storage** | MinIO Client | Object storage for images |
| **Frontend** | Go Templates, Pico.css, HTMX | Server-side rendering, progressive enhancement |

### Data Flow Example: Product Search

```
User types "Pride"
       │
       ▼
┌──────────────┐    HTMX hx-get="/search/autocomplete?q=Pride"
│   Browser    │───────────────────────────────────────────────►
└──────────────┘
       │
       ▼
┌──────────────┐    Route to ProductsHandler.Autocomplete()
│ http.ServeMux│───────────────────────────────────────────────►
└──────────────┘
       │
       ▼
┌──────────────┐    Call SearchProducts("Pride")
│   Handler    │───────────────────────────────────────────────►
└──────────────┘
       │
       ▼
┌──────────────┐    5-tier search: exact → prefix → fuzzy
│Elasticsearch │    Returns top 5 matches with relevance scores
└──────────────┘
       │
       ▼
┌──────────────┐    Render partials/autocomplete.html
│   Template   │    Return HTML fragment
└──────────────┘
       │
       ▼
┌──────────────┐    HTMX swaps in results dropdown
│   Browser    │    User sees "Pride and Prejudice" suggestion
└──────────────┘
```

## CI/CD Pipeline Architecture

The user flow diagram (see `user-flow-diagram.png` in this directory) illustrates the complete CI/CD pipeline and infrastructure components:

### Key Components Shown in Diagram

#### 1. Development & CI/CD Pipeline (Left Side)
- **App Source Code (GitHub)**
  - Version control
  - Build source
  
- **Build Pipeline**
  - Run tests
  - Build image
  - Publish image
  - Package application (Helm, etc.)

- **CI/CD Tools**
  - GitHub Actions, ADO, Actions, Argo Workflows, Tekton
  - Harbor artifact registry
  - registry.com, etc.

#### 2. Infrastructure Platform (Bottom)
- **VCF Infrastructure Layer**
  - **VCSA**: vCenter Server Appliance
  - **Supervisor NS**: Supervisor Namespace
  - **ArgoCD Instance**: GitOps deployment automation
  - **IAAS Policy**: Infrastructure as a Service policies
  - **Secrets**: Secret management
  - **generate gitops files**: Automated GitOps file generation

#### 3. Application Deployment (Center/Right)
- **app.arya.cloudx (Regional)**
  - **ArgoCD**: Continuous delivery
  - **App CD**: Application deployment
  - **Dev owned**: Developer-managed environments

- **App CD Details**
  - ArgoCD deployment
  - App helm charts
  - App CD management

#### 4. VKS Cluster (Right Side - Dashed Box)
- **Supervisor Namespace Infra**
  - VKS cluster
  - **Postgres Cluster (DSM)**: Database with Data Services Manager
  - **Apps**: Application pods
  
- **Infrastructure Services**
  - hello: Sample service
  - telegraf/prometheus: Monitoring and metrics
  - fluentbit: Log forwarding
  - external-dns: DNS automation
  - cert-manager: Certificate management
  - rabbitmq: Message queue
  - secret operator: Secret management
  - security agents: Security monitoring
  - RBAC: Role-based access control

- **Observability Stack**
  - clusters: Cluster monitoring
  - lease software: License management
  - observability: Monitoring stack
  - certs: Certificate management
  - DNS: DNS services
  - Ingress/Mesh: Traffic management
  - Security: Security policies
  - RBAC: Access control
  - secret operators: Secret management

#### 5. Additional Services (Bottom Right)
- **Infra Resource Auth**: Infrastructure resource authentication
- **Argo Scaffolding**: ArgoCD configuration templates
- **Registry**: Container registry
- **secret store service**: Centralized secret storage
- **Virtual Machines (VM Service)**: VM management in K8s

#### 6. External Connections
- **author.argoflow, etc.**: External ArgoCD instances
- **PR process**: Pull request workflow
- **update repo**: Repository updates

## User Flow Through the System

### 1. Developer Workflow
1. **Code Development**: Developer pushes code to GitHub
2. **CI Trigger**: GitHub Actions (or other CI) triggers on push/PR
3. **Build & Test**: Application is built, tested, and linted
4. **Image Creation**: Docker image is built and pushed to Harbor registry
5. **GitOps Update**: CI updates GitOps manifests (Helm charts, K8s YAML)

### 2. Deployment Workflow
1. **ArgoCD Detection**: ArgoCD monitors Git repo for manifest changes
2. **Sync & Deploy**: ArgoCD syncs changes to VKS cluster
3. **Infrastructure Setup**: Supervisor Namespace provisions required resources
4. **Service Deployment**: Application pods deployed to VKS cluster
5. **Database Connection**: Apps connect to Postgres Cluster (DSM)

### 3. Runtime Services
- **Ingress/Mesh**: Routes external traffic to application
- **DNS**: ExternalDNS automatically manages DNS records
- **Certificates**: Cert-Manager provisions and renews TLS certificates
- **Secrets**: Secret operators inject secrets into pods
- **Monitoring**: Prometheus/Telegraf collect metrics
- **Logging**: Fluentbit forwards logs to observability stack
- **Security**: RBAC and security agents enforce policies

### 4. User Access Flow
1. User accesses `app.arya.cloudx` via browser
2. DNS resolves to LoadBalancer/Ingress
3. TLS termination at Ingress (cert from Cert-Manager)
4. Request routed to application pod
5. Application queries Postgres for data
6. Response rendered and returned to user

## Technology Stack Mapping

### Current Implementation (Phase 2 Complete)
- **Application**: Go 1.25 web server with Repository Pattern
- **Database**: PostgreSQL with consolidated migrations
- **Search**: Elasticsearch (5-tier search with author support)
- **Caching**: Redis (sessions + product cache via CachedProductRepository)
- **Storage**: MinIO (S3-compatible, 1-year cache headers, ETags)
- **Frontend**: Pico.css + HTMX (server-side rendering)
- **CI/CD**: GitHub Actions with self-hosted runner
- **Container Registry**: Harbor with vulnerability scanning
- **Orchestration**: Docker Compose (local), Kubernetes/VKS (production)
- **GitOps**: ArgoCD with Kustomize

### Phase 3 Additions (In Progress)
- **Multi-App Architecture**: Reader (Go) + Chatbot (Python/FastAPI)
- **AI Integration**: Ollama LLM (local) → VCF Private AI (production)
- **App-of-Apps**: ArgoCD managing multiple microservices

### Future Considerations
- **Service Mesh**: Istio for traffic management
- **Monitoring**: Prometheus + Grafana dashboards
- **DNS**: ExternalDNS automation
- **Certificates**: Cert-Manager for TLS
- **Log Aggregation**: Fluentbit + Observability stack

## VCF 9.0 Integration Points

### Supervisor Services
- **Harbor**: Container registry with scanning
- **Argo CD**: GitOps deployment automation
- **VM Service**: Virtual machine management in K8s

### VKS Add-ons
- **Cert-Manager**: Automated certificate management
- **ExternalDNS**: DNS automation for services
- **Prometheus**: Metrics collection and alerting
- **Telegraf**: Time-series data collection
- **Istio**: Service mesh for traffic management

### CNCF Graduated Projects
- **Kubernetes**: Container orchestration (VKS)
- **Prometheus**: Monitoring and alerting
- **Envoy**: Proxy (via Istio)
- **CoreDNS**: DNS services
- **Helm**: Package management
- **Harbor**: Registry and artifact management

## Infrastructure Services in VKS Cluster

As shown in the diagram, the VKS cluster includes:

1. **Data Services**
   - Postgres Cluster (DSM) - Database with Data Services Manager
   - RabbitMQ - Message queue

2. **Observability**
   - Prometheus - Metrics collection
   - Telegraf - Time-series data
   - Fluentbit - Log forwarding

3. **Security**
   - RBAC - Role-based access control
   - Security agents - Security monitoring
   - Secret operator - Secret management

4. **Networking**
   - Ingress/Mesh - Traffic routing
   - ExternalDNS - DNS automation
   - Cert-Manager - Certificate management

5. **Application Services**
   - Application pods - Our Go application
   - Chatbot service - Python/FastAPI (Phase 2)
   - Admin panel - Management interface (Phase 2)

## Next Steps

### Phase 3 Implementation (Current)
1. Create GitHub repos for reader-app and chatbot-app
2. Test all three apps locally together
3. Add Bookstore API endpoints for Reader/Chatbot integration
4. Deploy Reader + Chatbot via ArgoCD App-of-Apps
5. Integration testing across all apps

### Future Infrastructure
1. Deploy Istio service mesh for traffic management
2. Configure ExternalDNS for automatic DNS records
3. Install Cert-Manager for automated TLS certificates
4. Set up observability stack (Prometheus, Grafana, Fluentbit)
5. VCF Private AI integration for production chatbot

## References

- [VCF 9.0 Documentation](https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vcf-9-0-and-later/9-0.html)
- [Supervisor Services](https://vsphere-tmm.github.io/Supervisor-Services/)
- [VKS Add-ons](https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vsphere-supervisor-services-and-standalone-components/latest/managing-vsphere-kuberenetes-service-clusters-and-workloads/managing-add-ons-in-vks-clusters.html)
- [CNCF Projects](https://www.cncf.io/projects/)

---

**Last Updated**: January 23, 2026  
**Diagram Source**: Application architecture diagram (ASCII), user flow diagram showing CI/CD pipeline, VKS cluster, and infrastructure components

