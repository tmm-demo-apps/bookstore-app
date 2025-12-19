# Application Architecture & User Flow

## Overview

This document describes the architecture and user flow for the DemoApp e-commerce platform, designed to showcase VMware Cloud Foundation (VCF) 9.0 capabilities.

## Architecture Diagram

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

### Current Implementation
- **Application**: Go 1.24 web server
- **Database**: PostgreSQL (standalone)
- **Frontend**: Pico.css + htmx
- **CI/CD**: GitHub Actions
- **Container Registry**: Docker Hub (moving to Harbor)
- **Orchestration**: Docker Compose (local), Kubernetes (production)

### Phase 2 Additions (In Progress)
- **Search**: Elasticsearch
- **Caching**: Redis
- **Storage**: MinIO (S3-compatible)
- **Microservices**: Python/FastAPI chatbot
- **Registry**: Harbor with vulnerability scanning

### Phase 3 Additions (Planned)
- **GitOps**: Argo CD
- **Service Mesh**: Istio
- **Monitoring**: Prometheus + Grafana
- **DNS**: ExternalDNS
- **Certificates**: Cert-Manager
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

### Phase 2 Implementation
1. Deploy Elasticsearch for full-text search
2. Add Redis for caching and sessions
3. Integrate MinIO for image storage
4. Build Python chatbot microservice
5. Implement Harbor registry

### Phase 3 Implementation
1. Set up Argo CD for GitOps
2. Deploy Istio service mesh
3. Configure ExternalDNS
4. Install Cert-Manager
5. Set up observability stack (Prometheus, Grafana, Fluentbit)

## References

- [VCF 9.0 Documentation](https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vcf-9-0-and-later/9-0.html)
- [Supervisor Services](https://vsphere-tmm.github.io/Supervisor-Services/)
- [VKS Add-ons](https://techdocs.broadcom.com/us/en/vmware-cis/vcf/vsphere-supervisor-services-and-standalone-components/latest/managing-vsphere-kuberenetes-service-clusters-and-workloads/managing-add-ons-in-vks-clusters.html)
- [CNCF Projects](https://www.cncf.io/projects/)

---

**Last Updated**: December 11, 2025  
**Diagram Source**: User flow diagram showing CI/CD pipeline, VKS cluster, and infrastructure components

