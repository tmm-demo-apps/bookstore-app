#!/bin/bash
set -e

echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          Kubernetes Database Fix Script                                    ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"
echo ""

# Get database credentials from Kubernetes secret
DB_PASSWORD=$(kubectl get secret app-secrets -n bookstore -o jsonpath='{.data.DB_PASSWORD}' | base64 -d)

echo "✅ Retrieved credentials from Kubernetes secrets"
echo ""

# Port forward to Postgres
echo "🔌 Setting up port forward to PostgreSQL..."
kubectl port-forward -n bookstore svc/postgres-service 5432:5432 &
PG_PID=$!

# Wait for port forward to be ready
sleep 3
echo "✅ Port forward established"
echo ""

# Cleanup function
cleanup() {
    echo ""
    echo "🧹 Cleaning up port forward..."
    kill $PG_PID 2>/dev/null || true
    echo "✅ Cleanup complete"
}
trap cleanup EXIT

# Apply missing categories migration
echo "📝 Adding missing categories..."
kubectl exec -it -n bookstore postgres-0 -- psql -U bookstore_user -d bookstore -c "
INSERT INTO categories (name, description) VALUES
('Philosophy', 'Philosophical works and treatises'),
('Science Fiction', 'Science fiction and speculative fiction'),
('Drama', 'Plays and dramatic works')
ON CONFLICT DO NOTHING;
"

echo "✅ Categories added"
echo ""

# Fix cart_items session_id constraint
echo "🔧 Fixing cart_items session_id constraint..."
kubectl exec -it -n bookstore postgres-0 -- psql -U bookstore_user -d bookstore -c "
-- Drop the old constraint
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS session_or_user;

-- Make session_id nullable
ALTER TABLE cart_items ALTER COLUMN session_id DROP NOT NULL;

-- Re-add the constraint
ALTER TABLE cart_items ADD CONSTRAINT session_or_user CHECK (
    (user_id IS NOT NULL AND session_id IS NULL) OR
    (user_id IS NULL AND session_id IS NOT NULL) OR
    (user_id IS NOT NULL AND session_id IS NOT NULL)
);
"

echo "✅ Cart constraint fixed"
echo ""

# Re-run book seeding to add missing books
echo "📚 Re-seeding books with all categories available..."
DB_HOST=localhost:5432 \
DB_USER=bookstore_user \
DB_PASSWORD="$DB_PASSWORD" \
DB_NAME=bookstore \
./scripts/bin/seed-gutenberg-books

echo ""
echo "✅ Books re-seeded"
echo ""

# Check counts
echo "📊 Database Statistics:"
kubectl exec -it -n bookstore postgres-0 -- psql -U bookstore_user -d bookstore -c "
SELECT 
    (SELECT COUNT(*) FROM categories) as total_categories,
    (SELECT COUNT(*) FROM products) as total_products;
"

echo ""
echo "📊 Books per Category:"
kubectl exec -it -n bookstore postgres-0 -- psql -U bookstore_user -d bookstore -c "
SELECT c.name, COUNT(p.id) as book_count
FROM categories c
LEFT JOIN products p ON p.category_id = c.id
GROUP BY c.name
ORDER BY book_count DESC;
"

echo ""
echo "╔════════════════════════════════════════════════════════════════════════════╗"
echo "║          ✅ DATABASE FIX COMPLETE                                          ║"
echo "╚════════════════════════════════════════════════════════════════════════════╝"

