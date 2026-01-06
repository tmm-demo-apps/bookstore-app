#!/bin/bash

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          Kubernetes Diagnostics                                            ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "📊 Pod Status:"
kubectl get pods -n bookstore
echo ""

echo "📊 Service Status:"
kubectl get svc -n bookstore
echo ""

echo "📊 App Logs (last 20 lines):"
kubectl logs -n bookstore deployment/app-deployment --tail=20
echo ""

echo "📊 Redis Connection Test:"
kubectl exec -n bookstore deployment/redis -- redis-cli ping
echo ""

echo "📊 Database Connection Test:"
kubectl exec -n bookstore postgres-0 -- pg_isready -U bookstore_user
echo ""

echo "📊 Elasticsearch Health:"
kubectl exec -n bookstore statefulset/elasticsearch -- curl -s http://localhost:9200/_cluster/health | grep -o '"status":"[^"]*"'
echo ""

echo "📊 MinIO Health:"
kubectl exec -n bookstore deployment/minio -- sh -c 'wget -q -O- http://localhost:9000/minio/health/live' && echo "✅ MinIO is live"
echo ""

echo "📊 Database Counts:"
kubectl exec -it -n bookstore postgres-0 -- psql -U bookstore_user -d bookstore -c "
SELECT 
    'Categories' as table_name, COUNT(*) as count FROM categories
UNION ALL
SELECT 'Products', COUNT(*) FROM products
UNION ALL
SELECT 'Users', COUNT(*) FROM users
UNION ALL
SELECT 'Cart Items', COUNT(*) FROM cart_items
UNION ALL
SELECT 'Orders', COUNT(*) FROM orders;
"

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ DIAGNOSTICS COMPLETE                                           ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"

