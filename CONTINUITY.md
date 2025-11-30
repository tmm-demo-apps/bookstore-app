# Project Continuity Plan

If I (the AI assistant) have "amnesia" or we are starting a new session, please provide the following prompt to quickly restore the project context.

---

Hello! We are continuing our work on the 12-factor demo e-commerce application.

## Project Overview
*   **Goal:** A demo platform to showcase **VMware Cloud Foundation (VCF) 9.0** capabilities through a real-world e-commerce application. See [`PLANNING.md`](PLANNING.md) for complete vision and three-phase roadmap.
*   **Tech Stack:** Go 1.24, PostgreSQL, Docker, Kubernetes, Pico.css, and htmx.
*   **Current Status:** Phase 1 Complete + Cart Fixes. Fully functional e-commerce application with Repository Pattern architecture, advanced cart features, user authentication, search, and "My Orders" page.
*   **Our Workflow:** We work in small, incremental steps. After each completed feature or bug fix, we **TEST FIRST** (see `TEST-CHECKLIST.md`), then commit the changes to our local Git repository and update the `diary.md` file.

### Testing Before Commits (MANDATORY)
**NEVER commit without testing!** Follow this workflow:
1. Make code changes
2. Rebuild: `docker compose down && docker compose up --build -d`
3. Run automated tests: `./test-smoke.sh`
4. Manual browser test (2-min spot check from `TEST-CHECKLIST.md`)
5. Only after tests pass: `git add -A && git commit -m "message"`

## Key Technologies & Patterns
*   **Repository Pattern**: All data access abstracted through interfaces in `internal/repository/`, allowing easy database swapping
*   **HTMX**: For dynamic content loading without full page reloads (cart count, cart summary, etc.)
*   **Pico.css**: Minimalist CSS framework for consistent, modern UI styling
*   **Go Templates**: Server-side HTML rendering with conditional logic
*   **Session Management**: Using `gorilla/sessions` for both authenticated users (`user_id`) and anonymous carts (`session_id`)
*   **PostgreSQL**: Database with migrations in `migrations/` directory
*   **12-Factor Methodology**: Externalized config, stateless processes, explicit dependencies

## File Structure
```
/
├── cmd/web/main.go                      # Main application entrypoint, route definitions
├── internal/
│   ├── handlers/
│   │   ├── base.go                      # Base handler struct, authentication helper
│   │   ├── cart.go                      # Cart operations (add, remove, view, update quantity)
│   │   ├── partials.go                  # HTMX partials (cart count, cart summary)
│   │   ├── products.go                  # Product listing & search
│   │   ├── auth.go                      # User auth (login, register, logout)
│   │   ├── checkout.go                  # Checkout flow
│   │   └── orders.go                    # Order history page
│   ├── models/
│   │   ├── product.go                   # Product & Category models
│   │   ├── user.go                      # User model with roles
│   │   ├── cart.go                      # CartItem model
│   │   └── order.go                     # Order & OrderItem models
│   └── repository/
│       ├── repository.go                # Repository interfaces (Product, Order, Cart, User)
│       └── postgres.go                  # PostgreSQL implementation of all repositories
├── templates/
│   ├── base.html                        # Base template with responsive header, search, cart dropdown
│   ├── products.html                    # Product listing page with search
│   ├── cart.html                        # Full cart page with quantity editor
│   ├── orders.html                      # Order history page
│   ├── checkout.html                    # Checkout page
│   └── ...
├── migrations/                          # SQL migrations (8 files including schema expansions)
├── kubernetes/                          # K8s manifests (app, postgres, secrets)
├── diary.md                             # **READ THIS FIRST** - Complete project history
├── PLANNING.md                          # **Project vision, roadmap, VCF integration strategy**
├── CONTINUITY.md                        # This file
├── Dockerfile                           # Multi-stage Go build
├── docker-compose.yml                   # Local dev environment
└── go.mod
```

## Recent Accomplishments (November 20-21, 2025)

### Phase 1: Repository Pattern & Schema Expansion ✅
Major architectural overhaul completed:

1. **Repository Pattern Refactoring** ✅
   - All SQL queries moved from handlers to `internal/repository/`
   - Defined interfaces: ProductRepository, OrderRepository, CartRepository, UserRepository
   - PostgreSQL implementation in `postgres.go`
   - Easy to swap databases (e.g., PostgreSQL → MariaDB) without touching handler code

2. **Enhanced Schema** ✅
   - Added `categories` table (Fiction, Non-Fiction, Tech, Science)
   - Products: Added `sku`, `stock_quantity`, `image_url`, `category_id`, `status`
   - Users: Added `full_name`, `role` (customer/admin), `created_at`
   - Orders: Added `user_id`, `total_amount`, `status`, `shipping_info` (JSONB)

3. **New Features** ✅
   - Search bar in header (SQL ILIKE, ready for Elasticsearch)
   - "My Orders" page (`/orders`) for authenticated users
   - Responsive header with hamburger menu for mobile
   - Data seeding: ~20 realistic books with categories and images

### Cart Bug Fixes (November 21, 2025) ✅
Fixed critical cart ordering and quantity issues:

1. **Alphabetical Ordering** ✅
   - Cart items now display in alphabetical order by product name
   - Homepage products also alphabetically ordered
   - Order remains stable when adding/removing items
   - Simplified SQL query (removed unnecessary GROUP BY)

2. **Quantity Accumulation** ✅
   - Adding existing products now increments quantity instead of replacing
   - Fixed: Query existing quantity BEFORE deleting rows
   - Delete-then-insert pattern prevents duplicate cart_items

3. **Template Fix** ✅
   - Changed `.CartItemID` to `.ID` to match `models.CartItem` struct

### Advanced Cart System (Previous Work)
Fully working cart with sophisticated features:

- **Hover-to-Open Preview**: Automatic dropdown with bounding box verification
- **Quantity Management**: +/- buttons, editable input, 1-99 range validation
- **Cross-Browser Compatible**: Works in Chrome, Firefox, Safari
- **Cache Prevention**: Multi-layered approach (meta tags, headers, htmx)
- **Cart Merging**: Anonymous cart merges with user account on login/signup
- **No Duplicates**: Unique constraints + delete-then-insert pattern

## How to Run Locally
```bash
docker compose up --build -d
# App available at http://localhost:8080
# Database: postgres://user:password@localhost:5432/bookstore
```

## Development Commands
```bash
# Rebuild and restart
docker compose down && docker compose up --build -d

# View logs
docker compose logs -f app

# Run migrations manually
docker compose exec db psql -U user -d bookstore -f /path/to/migration.sql

# Git workflow
git add -A
git commit -m "descriptive message"
```

## Known Issues / Edge Cases
- None currently! All cart bugs fixed as of November 21, 2025.

## Next Steps (Phase 2)
See `diary.md` for complete list. Key priorities:
- **UI**: Sticky header bar, enhanced order history page
- **Infrastructure**: Redis (sessions), Elasticsearch (search), MinIO (images)
- **Admin**: Product management panel

## Important Notes
- **Always update `diary.md`** after completing features or fixing bugs (latest entries at TOP)
- **Test in both Chrome and Firefox** for cross-browser compatibility
- **Follow CONTINUITY.md workflow** for git commands and commits
- **Repository Pattern**: All database queries go through `internal/repository/`, NOT directly in handlers
- **Cart items**: Must prevent duplicates - use delete-then-insert pattern
- **Alphabetical ordering**: All product/cart lists should ORDER BY name

## Common Debugging Tips
- **Cart issues?** Check both `user_id` and `session_id` logic in queries
- **Cache issues?** Multi-layer approach: meta tags + headers + htmx hx-headers
- **Ordering issues?** Verify `ORDER BY` clause in SQL and check browser cache
- **Hover issues?** Pico CSS positioning can trigger false mouseenter - use bounding box checks

---

**🎯 RESTORE PROCEDURE (After Amnesia)**

**Priority reading order**:
1. **`PLANNING.md`** - Project vision, VCF 9.0 goals, three-phase roadmap, and feature brainstorming
2. **`diary.md`** - Complete project history in reverse chronological order (latest at top), detailed technical explanations, bug fixes
3. **`CONTINUITY.md`** - This file (quick reference)

After reviewing these documents, ask the user what they'd like to work on next, or suggest working on one of the Phase 2 items from `PLANNING.md`.
