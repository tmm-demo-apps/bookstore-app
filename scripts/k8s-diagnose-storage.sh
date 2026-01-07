#!/bin/bash
# Diagnose vSAN Storage Issues

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          Kubernetes vSAN Storage Diagnostics                              ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "1. Storage Classes:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
kubectl get storageclass -o wide
echo ""

echo "2. CSI Driver Pods:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
kubectl get pods -n kube-system | grep -i csi
echo ""

echo "3. CSI Driver Logs (recent errors):"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
kubectl logs -n kube-system -l app=vsphere-csi-controller --tail=20 2>/dev/null || echo "No vSphere CSI controller found"
echo ""

echo "4. Node Storage Capacity:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
kubectl get nodes -o custom-columns=NAME:.metadata.name,STORAGE:.status.allocatable.ephemeral-storage
echo ""

echo "5. PVC Events:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
kubectl describe pvc -n bookstore | grep -A 10 "Events:"
echo ""

echo "6. Test: Create a test PVC with immediate binding:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc-immediate
  namespace: bookstore
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: vsan-default-storage-policy
  resources:
    requests:
      storage: 1Gi
EOF

echo "Waiting 10 seconds..."
sleep 10

kubectl get pvc test-pvc-immediate -n bookstore
kubectl describe pvc test-pvc-immediate -n bookstore | grep -A 10 "Events:"

echo ""
echo "7. Test: Create a test PVC with late binding:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc-latebinding
  namespace: bookstore
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: vsan-default-storage-policy-latebinding
  resources:
    requests:
      storage: 1Gi
EOF

kubectl get pvc test-pvc-latebinding -n bookstore

echo ""
echo "Cleanup test PVCs? (y/n)"
read -p "> " CLEANUP
if [ "$CLEANUP" = "y" ]; then
    kubectl delete pvc test-pvc-immediate test-pvc-latebinding -n bookstore
    echo "✅ Test PVCs deleted"
fi

