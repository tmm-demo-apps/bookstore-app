# Project Continuity Plan

If I (the AI assistant) have "amnesia" or we are starting a new session, please provide the following prompt to quickly restore the project context.

---

Hello! We are continuing our work on the 12-factor demo e-commerce application.

## Project Overview
*   **Goal:** A demo platform to showcase **VMware Cloud Foundation (VCF) 9.0** capabilities through a real-world e-commerce application. See [`PLANNING.md`](PLANNING.md) for complete vision and three-phase roadmap.
*   **Tech Stack:** Go 1.24, PostgreSQL, Docker, Kubernetes, Pico.css, and htmx.
*   **Current Status:** 🚧 Phase 2 UI Polish - **INCOMPLETE WORK**: Compressing products page layout and standardizing button sizes.
*   **Active Issue:** Quantity controls height doesn't match "Add to Cart" button despite identical CSS padding values. Need to debug and fix.
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

## Current Work In Progress (December 5, 2025 - Evening)

### 🚧 Products Page Layout Compression (INCOMPLETE)

**What We're Doing**: Compressing the products listing page to reduce whitespace and standardize button sizes for a more polished, professional look.

**What's Working**:
- ✅ Vertical spacing between cards reduced to consistent `0.75rem`
- ✅ Symmetrical padding inside cards (`0.75rem` all sides)
- ✅ Created `.btn-small` class using Pico CSS variables for "Add to Cart" buttons
- ✅ Refactored quantity controls to use CSS classes instead of inline styles
- ✅ Both using `calc(var(--form-element-spacing-vertical) * 0.35)` for compact sizing

**What's NOT Working** ❌:
- **Quantity controls are taller than "Add to Cart" button** despite identical padding values
- We tried multiple approaches (align-items, height fit-content, line-height, etc.)
- Root cause unknown - possibly input element browser defaults, borders, or Pico CSS overrides
- **White space within cards still needs reduction** - internal spacing between elements too loose

**Files Modified**: 
- `templates/products.html` - All CSS and HTML changes

**Next Steps for New Session**:
1. **DEBUG height mismatch** - Use browser dev tools to find root cause
2. **Reduce card whitespace** - Tighten spacing between internal elements
3. **Visual polish** - Ensure everything looks good on mobile
4. **Test functionality** - Verify quantity controls still work after fixes

## Recent Accomplishments (December 5, 2025 - Afternoon)

### Product Detail Pages ✅
Implemented comprehensive individual product pages with professional e-commerce UX:

1. **Detail Page Features** ✅
   - Large product images (500px, responsive)
   - Breadcrumb navigation (Home > Products > Product Name)
   - Full product descriptions
   - Color-coded stock status badges
   - SKU, availability, and status metadata
   - Quantity controls with stock limits
   - Add to Cart from detail page
   - 404 handling for non-existent products

2. **Clickable Products** ✅
   - Product cards and names now link to detail pages
   - Works in both grid and table views
   - Maintains existing hover effects
   - Consistent styling across views

3. **Mobile Responsive** ✅
   - Single-column layout on mobile
   - Reduced image sizes
   - Full-width controls and buttons
   - Touch-optimized spacing

**Test Status**: All 15 smoke tests + new detail page tests passing ✅

### Earlier Work (November 30, 2025)

### Major UI Overhaul ✅
Transformed the application into a modern, image-rich e-commerce experience:

1. **Product Images & Layouts** ✅
   - Card grid with large product images (250px)
   - Table/Tile view toggle with localStorage persistence
   - Table view with compact 60px thumbnails
   - Responsive grid (auto-fill, 280px minimum)
   - 📦 fallback icons for missing images

2. **Compact Cart** ✅
   - Single-row layout (height = 80x80px thumbnail)
   - All info visible: Image | Name | Price | Qty | Subtotal | Remove
   - No wasted vertical space
   - Minimal, muted controls

3. **Minimal UI Controls** ✅
   - Quantity buttons: 0.25rem padding, muted colors
   - Input width: 40px (compact)
   - Inline styles for CSS specificity over Pico
   - Hover effects for feedback

4. **Enhanced Order History** ✅
   - Expandable cards with full order details
   - Product images in order items
   - Alphabetically sorted items
   - Color-coded status badges

5. **Bug Fixes** ✅
   - Session cookie configuration (HTTP localhost)
   - Cart session ID logic
   - Dark mode header background
   - Order items alphabetical sorting


### Earlier Work (November 20-21, 2025)

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

## Next Steps - Future Focus 🎯

### **IMMEDIATE PRIORITY: Fix Product Page Layout Issues** 🚨

**Must Complete Before Moving On:**

1. **Fix Quantity Control Height Mismatch**
   - Problem: Controls are taller than "Add to Cart" button despite same padding
   - Approach: Use browser dev tools to inspect computed styles
   - Possible solutions:
     - Check if input has default browser height
     - Verify border isn't adding extra pixels
     - Try explicit height matching button
     - Consider `-webkit-appearance: none` on input
     - Investigate Pico CSS form element defaults

2. **Reduce White Space Within Cards**
   - Problem: Too much space between internal elements
   - Current: `0.25rem` margins, `0.75rem` padding
   - Goal: Tighter, more compact look without feeling cramped
   - Adjust spacing between: image-name, name-description, description-price, price-actions

3. **Visual Polish & Testing**
   - Verify mobile responsiveness (< 768px breakpoint)
   - Test both grid and table views
   - Ensure quantity controls still function
   - Run smoke tests after completion
   - Get user approval before committing as complete

### **AFTER LAYOUT IS FIXED: Category Filtering UI**
Implement the category filtering system to help users browse products by category.

**What to Build:**
- Category sidebar on desktop
- Category dropdown on mobile
- Filter products by category (already have `category_id` in DB)
- Visual category cards with icons
- Update search to filter by category

**Technical Approach:**
1. Query categories from database
2. Add category filter to product list handler
3. Design category sidebar/filters
4. Update repository method to filter by category
5. Add "All Categories" option

### Other Phase 2 Priorities
- **Infrastructure**: Redis (sessions), Elasticsearch (search), MinIO (images)
- **Admin**: Product management panel
- **Microservices**: AI Support Chatbot (Python/FastAPI)

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
