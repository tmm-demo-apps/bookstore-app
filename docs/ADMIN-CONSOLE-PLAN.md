# Admin Console Implementation Plan

## Overview

Build a comprehensive admin interface for managing the bookstore, demonstrating enterprise-grade features and Kubernetes/VCF integration patterns.

## Phase 1: Foundation (MVP)

### 1.1 Authentication & Authorization

**Database Schema**:
```sql
-- Add role column to users table
ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'customer';
-- Roles: 'customer', 'admin', 'super_admin'

CREATE INDEX idx_users_role ON users(role);
```

**Middleware**:
- `RequireAuth()` - Ensures user is logged in
- `RequireAdmin()` - Ensures user has admin role
- Redirect non-admins to 403 page

**Routes**:
```go
mux.HandleFunc("/admin", h.RequireAdmin(h.AdminDashboard))
mux.HandleFunc("/admin/products", h.RequireAdmin(h.AdminProducts))
mux.HandleFunc("/admin/orders", h.RequireAdmin(h.AdminOrders))
mux.HandleFunc("/admin/users", h.RequireAdmin(h.AdminUsers))
```

### 1.2 Dashboard (Home Page)

**Metrics to Display**:
- Total Orders (today, this week, this month)
- Total Revenue (today, this week, this month)
- Active Users (last 24h)
- Low Stock Alerts (< 10 items)
- Recent Orders (last 10)
- Top Selling Products (last 30 days)

**Implementation**:
```go
type DashboardData struct {
    TodayOrders    int
    TodayRevenue   float64
    WeekOrders     int
    WeekRevenue    float64
    MonthOrders    int
    MonthRevenue   float64
    ActiveUsers    int
    LowStockItems  []Product
    RecentOrders   []Order
    TopProducts    []ProductSales
}
```

### 1.3 Product Management

**Features**:
- ✅ List all products (paginated)
- ✅ Search/filter products
- ✅ Create new product
- ✅ Edit existing product
- ✅ Delete product
- ✅ Bulk operations (delete, update stock)
- ✅ Image upload via MinIO

**UI Components**:
- Data table with sorting
- Inline editing
- Modal for create/edit
- Drag-and-drop image upload
- Category selector
- Stock level indicators

### 1.4 Order Management

**Features**:
- ✅ List all orders (paginated)
- ✅ Filter by status, date range, customer
- ✅ View order details
- ✅ Update order status
- ✅ View customer information
- ✅ Print invoice/packing slip

**Order Statuses**:
- `pending` - Just placed
- `processing` - Being prepared
- `shipped` - On the way
- `delivered` - Completed
- `cancelled` - Cancelled by customer/admin

**Status Workflow**:
```
pending -> processing -> shipped -> delivered
   |                                     |
   +--------- cancelled ----------------+
```

### 1.5 User Management

**Features**:
- ✅ List all users
- ✅ Search by email/name
- ✅ View user details (orders, reviews)
- ✅ Change user role
- ✅ Disable/enable user account
- ⚠️ Do NOT allow password changes (security)

## Phase 2: Advanced Features

### 2.1 Category Management

**Features**:
- Create/edit/delete categories
- Reorder categories
- Set category images
- View products per category

### 2.2 Inventory Management

**Features**:
- Stock level tracking
- Low stock alerts
- Bulk stock updates
- Stock history (who changed what, when)
- Reorder point settings

### 2.3 Analytics & Reporting

**Reports**:
- Sales by category
- Sales by time period
- Customer lifetime value
- Product performance
- Revenue trends
- Export to CSV

### 2.4 Settings

**Configuration**:
- Site name, logo
- Email settings (SMTP)
- Payment gateway settings
- Shipping options
- Tax rates
- Currency settings

## Phase 3: VCF Integration Features

### 3.1 Cert-Manager Integration

**Purpose**: Secure admin access with HTTPS and mTLS

**Implementation**:
```yaml
# Certificate for admin subdomain
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: admin-tls
spec:
  secretName: admin-tls-secret
  issuerRef:
    name: letsencrypt-prod
  dnsNames:
    - admin.bookstore.example.com
```

**Ingress Configuration**:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: admin-ingress
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
    - hosts:
        - admin.bookstore.example.com
      secretName: admin-tls-secret
  rules:
    - host: admin.bookstore.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: app-service
                port:
                  number: 80
```

### 3.2 RBAC Integration

**Kubernetes RBAC**:
```yaml
# ServiceAccount for admin pods
apiVersion: v1
kind: ServiceAccount
metadata:
  name: admin-sa

---
# Role for admin operations
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: admin-role
rules:
  - apiGroups: [""]
    resources: ["pods", "services"]
    verbs: ["get", "list"]

---
# RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: admin-rolebinding
subjects:
  - kind: ServiceAccount
    name: admin-sa
roleRef:
  kind: Role
  name: admin-role
  apiGroup: rbac.authorization.k8s.io
```

### 3.3 Internal Ingress

**Purpose**: Admin console only accessible from internal network

**Implementation**:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: admin-internal-ingress
  annotations:
    nginx.ingress.kubernetes.io/whitelist-source-range: "10.0.0.0/8,172.16.0.0/12"
spec:
  ingressClassName: nginx-internal
  rules:
    - host: admin-internal.bookstore.local
      http:
        paths:
          - path: /admin
            pathType: Prefix
            backend:
              service:
                name: app-service
                port:
                  number: 80
```

## Technology Stack

### Backend
- **Language**: Go (same as main app)
- **Framework**: net/http (standard library)
- **Templates**: html/template
- **Authentication**: Session-based (existing system)

### Frontend
- **CSS Framework**: Pico CSS (already in use)
- **JavaScript**: Vanilla JS + HTMX for dynamic updates
- **Charts**: Chart.js for analytics

### Database
- **Schema**: Extend existing PostgreSQL schema
- **Migrations**: Add new migration files

## File Structure

```
internal/
  handlers/
    admin.go              # Admin-specific handlers
    admin_products.go     # Product management
    admin_orders.go       # Order management
    admin_users.go        # User management
  middleware/
    admin.go              # Admin authentication middleware
  models/
    admin.go              # Admin-specific models

templates/
  admin/
    base.html             # Admin layout (sidebar, nav)
    dashboard.html        # Dashboard/home
    products.html         # Product list
    product-edit.html     # Product edit form
    orders.html           # Order list
    order-detail.html     # Order details
    users.html            # User list
    user-detail.html      # User details

migrations/
  015_add_user_roles.sql
  016_add_order_status.sql
  017_add_stock_history.sql
```

## Implementation Timeline

### Week 1: Foundation
- Day 1-2: Authentication & authorization
- Day 3-4: Dashboard with basic metrics
- Day 5: Product list view

### Week 2: Core Features
- Day 1-2: Product CRUD operations
- Day 3-4: Order management
- Day 5: User management

### Week 3: Polish & VCF Integration
- Day 1-2: Analytics & reporting
- Day 3: Cert-Manager setup
- Day 4: RBAC configuration
- Day 5: Testing & documentation

## Security Considerations

1. **Authentication**: Always verify admin role before any operation
2. **CSRF Protection**: Add CSRF tokens to all forms
3. **Input Validation**: Sanitize all user inputs
4. **Audit Logging**: Log all admin actions (who did what, when)
5. **Rate Limiting**: Prevent brute force attacks
6. **Session Timeout**: Auto-logout after inactivity

## Testing Strategy

1. **Unit Tests**: Test all admin handlers
2. **Integration Tests**: Test admin workflows
3. **Security Tests**: Test authorization checks
4. **Smoke Tests**: Add admin-specific smoke tests

## Demo Script for VCF

1. **Show Admin Login**: Demonstrate role-based access
2. **Dashboard Metrics**: Show real-time business metrics
3. **Product Management**: Create/edit/delete products
4. **Order Processing**: Update order status
5. **Cert-Manager**: Show HTTPS certificate
6. **RBAC**: Show Kubernetes permissions
7. **Internal Access**: Show network restrictions

## Success Criteria

- ✅ Admin can manage all products
- ✅ Admin can view and update orders
- ✅ Admin can view user information
- ✅ Dashboard shows real-time metrics
- ✅ All actions are logged
- ✅ Secure HTTPS access
- ✅ Kubernetes RBAC integrated
- ✅ Internal-only access configured

