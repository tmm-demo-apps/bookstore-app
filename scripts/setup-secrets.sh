#!/bin/bash
# Setup secrets for all demo apps
# Run this on cli-vm with kubectl access to VKS-04

set -e

echo "=== Demo Apps Secret Setup ==="
echo ""

# Prompt for Harbor password (hidden input)
echo -n "Enter Harbor admin password: "
read -s HARBOR_PASSWORD
echo ""

if [ -z "$HARBOR_PASSWORD" ]; then
  echo "Error: Harbor password cannot be empty"
  exit 1
fi

# Generate random passwords
echo "Generating secure random values..."
READER_PG_PASS=$(openssl rand -hex 16)
READER_SESSION=$(openssl rand -hex 32)

# MinIO credentials - fetch from bookstore namespace (shared MinIO)
echo "Fetching MinIO credentials from bookstore namespace..."
MINIO_ACCESS=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.MINIO_ACCESS_KEY}' 2>/dev/null | base64 -d)
MINIO_SECRET=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.MINIO_SECRET_KEY}' 2>/dev/null | base64 -d)

if [ -z "$MINIO_ACCESS" ] || [ -z "$MINIO_SECRET" ]; then
  echo "Warning: Could not fetch MinIO credentials from bookstore namespace."
  echo "Using default credentials (may not work with existing MinIO)."
  MINIO_ACCESS="minioadmin"
  MINIO_SECRET="minioadmin"
else
  echo "  Found MinIO credentials from bookstore."
fi

echo ""
echo "Creating Kubernetes secrets..."

# Reader namespace secrets
echo "  [1/4] Creating reader-secrets in reader namespace..."
READER_DB_URL="postgres://reader:${READER_PG_PASS}@reader-postgres:5432/reader?sslmode=disable"
kubectl create secret generic reader-secrets \
  --namespace=reader \
  --from-literal=database-url="$READER_DB_URL" \
  --from-literal=postgres-password="$READER_PG_PASS" \
  --from-literal=minio-access-key="$MINIO_ACCESS" \
  --from-literal=minio-secret-key="$MINIO_SECRET" \
  --from-literal=session-secret="$READER_SESSION" \
  --dry-run=client -o yaml | kubectl apply -f -

# Chatbot namespace secrets (optional - only needed if using VCF Private AI)
echo "  [2/4] Creating chatbot-secrets in chatbot namespace..."
kubectl create secret generic chatbot-secrets \
  --namespace=chatbot \
  --from-literal=llm-api-key="" \
  --dry-run=client -o yaml | kubectl apply -f -

# Harbor registry secret for reader namespace
echo "  [3/4] Creating harbor-registry-secret in reader namespace..."
kubectl create secret docker-registry harbor-registry-secret \
  --namespace=reader \
  --docker-server=harbor.corp.vmbeans.com \
  --docker-username=admin \
  --docker-password="$HARBOR_PASSWORD" \
  --dry-run=client -o yaml | kubectl apply -f -

# Harbor registry secret for chatbot namespace
echo "  [4/4] Creating harbor-registry-secret in chatbot namespace..."
kubectl create secret docker-registry harbor-registry-secret \
  --namespace=chatbot \
  --docker-server=harbor.corp.vmbeans.com \
  --docker-username=admin \
  --docker-password="$HARBOR_PASSWORD" \
  --dry-run=client -o yaml | kubectl apply -f -

echo ""
echo "=== Kubernetes Secrets Created ==="
echo ""
echo "To verify:"
echo "  kubectl get secrets -n reader"
echo "  kubectl get secrets -n chatbot"
echo ""

# Generate VCF Secret Store YAML files for future migration
echo "=== Generating VCF Secret Store YAML Files ==="
echo ""

mkdir -p vcf-secrets

cat > vcf-secrets/reader-secrets.yaml << EOF
# VCF Secret Store - Reader App Secrets
# Apply with: vcf secret create --file vcf-secrets/reader-secrets.yaml
kind: KeyValueSecret
apiVersion: secretstore.vmware.com/v1alpha1
metadata:
  name: reader-secrets
  namespace: reader
spec:
  name: reader-secrets
  data:
  - key: database-url
    value: "$READER_DB_URL"
  - key: postgres-password
    value: "$READER_PG_PASS"
  - key: minio-access-key
    value: "$MINIO_ACCESS"
  - key: minio-secret-key
    value: "$MINIO_SECRET"
  - key: session-secret
    value: "$READER_SESSION"
EOF

cat > vcf-secrets/chatbot-secrets.yaml << EOF
# VCF Secret Store - Chatbot App Secrets
# Apply with: vcf secret create --file vcf-secrets/chatbot-secrets.yaml
kind: KeyValueSecret
apiVersion: secretstore.vmware.com/v1alpha1
metadata:
  name: chatbot-secrets
  namespace: chatbot
spec:
  name: chatbot-secrets
  data:
  - key: llm-api-key
    value: ""
EOF

cat > vcf-secrets/harbor-registry.yaml << EOF
# VCF Secret Store - Harbor Registry Credentials
# Apply with: vcf secret create --file vcf-secrets/harbor-registry.yaml
# Note: Create separately for each namespace that needs it
kind: KeyValueSecret
apiVersion: secretstore.vmware.com/v1alpha1
metadata:
  name: harbor-registry
spec:
  name: harbor-registry
  data:
  - key: docker-server
    value: "harbor.corp.vmbeans.com"
  - key: docker-username
    value: "admin"
  - key: docker-password
    value: "$HARBOR_PASSWORD"
EOF

echo "VCF Secret Store YAML files created in ./vcf-secrets/"
echo "  - vcf-secrets/reader-secrets.yaml"
echo "  - vcf-secrets/chatbot-secrets.yaml"
echo "  - vcf-secrets/harbor-registry.yaml"
echo ""
echo "WARNING: These files contain sensitive values. Do not commit to git!"
echo ""
echo "=== VCF Secret Store Migration (Future) ==="
echo ""
echo "To migrate to VCF Secret Store:"
echo "  1. vcf context use <your-context>"
echo "  2. vcf secret create --file vcf-secrets/reader-secrets.yaml"
echo "  3. vcf secret create --file vcf-secrets/chatbot-secrets.yaml"
echo "  4. Update deployments to use Vault Agent Injector annotations"
echo ""
echo "=== Setup Complete ==="
echo ""
