#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          Re-index Elasticsearch                                            ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "📊 Current Elasticsearch index count:"
kubectl exec -n bookstore statefulset/elasticsearch -- curl -s http://localhost:9200/products/_count | grep -o '"count":[0-9]*'
echo ""

echo "🗑️  Deleting old index..."
kubectl exec -n bookstore statefulset/elasticsearch -- curl -X DELETE http://localhost:9200/products
echo ""

echo "⏳ Waiting 2 seconds..."
sleep 2

echo "🔄 Restarting app pods to trigger re-indexing..."
kubectl rollout restart deployment/app-deployment -n bookstore

echo ""
echo "⏳ Waiting for pods to restart..."
kubectl rollout status deployment/app-deployment -n bookstore

echo ""
echo "⏳ Waiting 10 seconds for indexing to complete..."
sleep 10

echo ""
echo "📊 New Elasticsearch index count:"
kubectl exec -n bookstore statefulset/elasticsearch -- curl -s http://localhost:9200/products/_count | grep -o '"count":[0-9]*'

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ RE-INDEXING COMPLETE                                           ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"

