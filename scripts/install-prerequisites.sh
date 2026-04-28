#!/bin/bash
#
# Install Gateway API prerequisites for non-VKS clusters (kind, EKS, GKE, etc.)
#
# For VKS clusters, use AddonInstall resources on the Supervisor instead:
#   kubectl apply -f - <<EOF
#   apiVersion: addons.kubernetes.vmware.com/v1alpha1
#   kind: AddonInstall
#   metadata:
#     name: <cluster-name>-cert-manager
#     namespace: <cluster-namespace>
#   spec:
#     addonRef:
#       name: cert-manager
#     clusters:
#     - selector:
#         matchLabels:
#           cluster.x-k8s.io/cluster-name: <cluster-name>
#   ---
#   apiVersion: addons.kubernetes.vmware.com/v1alpha1
#   kind: AddonInstall
#   metadata:
#     name: <cluster-name>-istio
#     namespace: <cluster-namespace>
#   spec:
#     addonRef:
#       name: istio
#     clusters:
#     - selector:
#         matchLabels:
#           cluster.x-k8s.io/cluster-name: <cluster-name>
#   EOF
#
# Gateway API CRDs (step 1 below) are needed on ALL clusters, including VKS.
#
set -euo pipefail

# ─── Version pins (update these when upgrading) ───────────────────────────────
GATEWAY_API_VERSION="v1.4.0"
ISTIO_VERSION="1.29.1"
CERT_MANAGER_VERSION="v1.20.0"
# NOTE: Gateway API CRDs MUST be compatible with Istio.
#       Istio 1.28/1.29 requires v1.4.x (v1.5 crashes istiod).
#       Check: https://istio.io/latest/docs/setup/install/helm/
# ──────────────────────────────────────────────────────────────────────────────

echo "==> Installing Gateway API CRDs ${GATEWAY_API_VERSION}"
if kubectl get crd gateways.gateway.networking.k8s.io &>/dev/null; then
  echo "    Already installed, skipping"
else
  kubectl apply --server-side \
    -f "https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/standard-install.yaml"
fi

echo "==> Installing cert-manager ${CERT_MANAGER_VERSION}"
if kubectl get ns cert-manager &>/dev/null 2>&1; then
  echo "    Already installed, skipping"
else
  helm repo add jetstack https://charts.jetstack.io --force-update
  helm repo update jetstack
  helm upgrade --install cert-manager jetstack/cert-manager \
    -n cert-manager --create-namespace \
    --version "${CERT_MANAGER_VERSION}" \
    --set crds.enabled=true --wait
fi

echo "==> Installing Istio ${ISTIO_VERSION}"
if kubectl get ns istio-system &>/dev/null 2>&1; then
  echo "    Already installed, skipping"
else
  # Create namespace and label for Pod Security Admission (required for VKS/restricted clusters)
  kubectl create namespace istio-system --dry-run=client -o yaml | kubectl apply -f -
  kubectl label namespace istio-system pod-security.kubernetes.io/enforce=privileged --overwrite

  helm repo add istio https://istio-release.storage.googleapis.com/charts --force-update
  helm repo update istio
  helm upgrade --install istio-base istio/base \
    -n istio-system --version "${ISTIO_VERSION}"
  helm upgrade --install istiod istio/istiod \
    -n istio-system --version "${ISTIO_VERSION}" --wait
fi

echo ""
echo "==> All prerequisites installed successfully"
echo "    Gateway API CRDs: ${GATEWAY_API_VERSION}"
echo "    cert-manager:     ${CERT_MANAGER_VERSION}"
echo "    Istio:            ${ISTIO_VERSION}"
echo ""
echo "Next: helm install demo ./helm/demo-suite --set global.domain=<your-domain>"
