# Project Diary: 12-Factor E-commerce Template

> **📋 See [`PLANNING.md`](PLANNING.md) for complete project vision, roadmap, and VCF 9.0 integration strategy**

## 🎯 Quick Status Summary
**Last Updated:** November 21, 2025  
**Project Status:** ✅ Phase 1 Complete + Cart Fixes  
**Recent Focus:** Cart Ordering & Duplicate Item Bugs  
**Project Goal**: Demo platform to showcase VMware Cloud Foundation (VCF) 9.0 capabilities through real-world e-commerce application

### What's Working
- ✅ **Modern Architecture:** Refactored data layer into a `Repository Pattern` (Hexagonal Architecture ready).
- ✅ **Advanced Schema:** Added Categories, SKUs, Stock Levels, User Roles, and Order History linkage.
- ✅ **Search:** Full product search functionality (SQL-based, ready for Elasticsearch swap).
- ✅ **Responsive UI:** New header with mobile "Hamburger" menu and search bar.
- ✅ **User Features:** "My Orders" page to view purchase history.
- ✅ **Cart Ordering:** Items display alphabetically and maintain stable order
- ✅ User authentication (register, login, logout)
- ✅ Product catalog display with quantity controls
- ✅ Shopping cart (add, remove, view) with proper quantity accumulation
- ✅ Hover-to-open cart preview with auto-close
- ✅ Quantity management with +/- buttons and manual input (cart & products)
- ✅ Cart count badge showing total quantities
- ✅ PII-free checkout process
- ✅ Cross-browser compatibility (Chrome, Firefox)
- ✅ Session support for anonymous users
- ✅ Modern UI with Pico.css and htmx

### Key Technical Achievements
- **Cart Deduplication:** Delete-then-insert pattern prevents duplicate cart_items
- **Query Optimization:** Simplified cart queries by removing unnecessary aggregation
- **Hover Logic:** Solved complex Pico CSS dropdown positioning issue with bounding box verification
- **Quantity System:** Proper accumulation when adding existing items
- **Cache Control:** Multi-layered approach (meta tags, headers, htmx attributes)
- **Firefox Compatibility:** Modern `oninput` event instead of deprecated `onkeypress`

### Next Steps (Phase 2)
- **UI Enhancements**:
    - Add **sticky header bar** that stays at the top when scrolling
    - Improve **product categories/filtering** UI
    - Enhance **order history** page for logged-in users with more details
- **Infrastructure**:
    - Deploy **Redis** for session management
    - Deploy **Elasticsearch** for advanced search
    - Deploy **MinIO** for image storage
- **Admin Features**:
    - Create **Admin Panel** for product management
    - Add product inventory management
- **Future Considerations**:
    - Payment integration
    - Automated integration tests (Selenium/Playwright)
    - Load testing for performance benchmarks

---

## November 21, 2025

### Bug Fixes: Cart Ordering & Item Management

#### Issues Identified
1. **Cart items displayed in random order** - Items appeared in database insertion order instead of alphabetically
2. **Adding existing items replaced quantity** - Adding a product already in cart would reset quantity instead of incrementing
3. **Homepage products unordered** - Product list had no ORDER BY clause

#### Root Causes
- Cart query had `ORDER BY p.name` but wasn't being honored due to browser caching
- `AddToCart()` function checked existing quantity AFTER deleting items (always returned 0)
- `ListProducts()` query had no ORDER BY clause

#### Solutions Implemented
1. **Fixed cart ordering**: Simplified query to remove unnecessary `GROUP BY` and `MIN()` aggregation since duplicates are now prevented
2. **Fixed AddToCart logic**: Moved existing quantity check BEFORE delete operation to properly accumulate quantities
3. **Added product ordering**: Added `ORDER BY name` to `ListProducts()` query for consistent alphabetical display
4. **Fixed template field reference**: Changed `.CartItemID` to `.ID` to match `models.CartItem` struct

#### Technical Details
```sql
-- New simplified cart query (no duplicates = no need for GROUP BY)
SELECT ci.id, ci.product_id, p.name, p.description, p.price, ci.quantity
FROM cart_items ci
JOIN products p ON ci.product_id = p.id
WHERE ci.user_id = $1
ORDER BY p.name
```

**AddToCart Flow (Fixed)**:
1. Query existing quantity
2. Calculate new quantity (existing + new)
3. Delete old rows
4. Insert single consolidated row

#### Testing Results
✅ Cart items maintain alphabetical order when adding/removing  
✅ Adding duplicate items increments quantity correctly  
✅ Homepage products display alphabetically  
✅ No duplicate cart items created  
✅ Quantities properly accumulated across multiple adds

---

## November 20, 2025

### Phase 1: Retail Foundation & Data Prep
We executed a major architectural overhaul to prepare the application for enterprise VCF demos.

#### 1. Repository Pattern Refactoring
Moved all raw SQL queries from Handlers into a structured `internal/repository` package.
- **Why?**: Allows us to swap the backend database (e.g., Postgres -> MariaDB) or Search Engine (SQL -> Elasticsearch) by just changing the interface implementation, without touching the UI code.
- **Interfaces**: Defined `ProductRepository`, `OrderRepository`, `CartRepository`, `UserRepository`.
- **Implementation**: Created `PostgresRepository` that implements all interfaces.

#### 2. Schema Expansion
Upgraded the database schema to support realistic retail scenarios:
- **Categories**: Created `categories` table (Fiction, Non-Fiction, Tech, etc.).
- **Products**: Added `sku`, `stock_quantity`, `image_url`, `category_id`, `status`.
- **Users**: Added `full_name`, `role` (customer/admin), `created_at`.
- **Orders**: Added `user_id` (linked to Users table), `total_amount`, `status`, `shipping_info` (JSONB).

#### 3. New Features
- **Search**: Added search bar to header. Backend currently uses SQL `ILIKE`, but is designed to swap for Elasticsearch.
- **My Orders**: Authenticated users can now see their past orders at `/orders`.
- **Responsive Header**: Replaced the simple nav with a responsive Grid layout featuring a hamburger menu for mobile devices.

#### 4. Data Seeding
Created `migrations/008_seed_data.sql` to populate the store with ~20 realistic books across 4 categories with placeholder images.

---

## October 30, 2025

### Project Goal
Create a reusable 12-factor e-commerce template, designed for Kubernetes using the 12-factor app methodology.

### What We've Done So Far
- **Project Initialization:** Set up a Go project with a standard directory structure and initialized a local Git repository for version control.
- **Backend Development:** Implemented a basic Go web server that connects to a PostgreSQL database to fetch and display a list of books using server-side templates.
- **Containerization:** Wrote a multi-stage `Dockerfile` to create an optimized container image for the application and a `docker-compose.yml` file to run the app and database locally.
- **Deployment Setup:** Created Kubernetes manifests (`deployment.yaml`, `service.yaml`) for both the application and the PostgreSQL database.
- **CI/CD Pipeline:** Set up a basic CI/CD workflow using GitHub Actions to automate the building and pushing of the Docker image.
- **Core Shopping Cart:** Implemented the ability to add, view, and remove items from the shopping cart.
- **Checkout Process:** Implemented a PII-free checkout process where users can confirm an order without entering personal data.
- **UI Improvement**: Refactored the frontend using Pico.css and a base template structure to create a clean, modern, and consistent user interface.
- **Advanced Cart Features**: Implemented several UI/UX improvements for the shopping cart using htmx, including a dynamic cart count, a hover-enabled cart summary, and a total cost display.
- **Debugging:**
    - Resolved an initial error caused by Go not being installed on the system.
    - Fixed several build failures due to a Go version mismatch between the `go.mod` file and the Docker image, eventually standardizing on Go 1.23.
    - Added missing package dependencies (`github.com/google/uuid`) to resolve build errors.
- **Kubernetes Security**: Refactored the Kubernetes manifests and application code to use Kubernetes Secrets for managing database credentials, removing sensitive data from version control.
- **User Management & Refactoring**: Implemented a complete user management system (registration, login, logout) with secure password hashing. As part of this, all Go handlers were refactored to use a centralized `Handlers` struct for cleaner code. This work also included several bug fixes and UX improvements, such as streamlining the signup flow, adding form validations, and fixing a major regression that prevented books from being displayed on the main page.
- **Generalization**: Refactored the entire application from a specific "bookstore" into a generic, reusable e-commerce template. This involved renaming models, handlers, database tables, and updating the UI to use generic "product" terminology.
- **Checkout Login Flow**: Implemented a forced-login flow at checkout. Unauthenticated users are now redirected to the login page and are returned to the checkout process after a successful login.
- **Project Rollback**: Reverted the project state to commit `47a98fd` to undo a series of buggy changes related to the shopping cart's dynamic features. We are now at a stable state where user management is functional, and the basic cart works.
- **Dynamic Cart Features**: Successfully implemented cart hover preview and fixed critical caching issues:
    - **Hover-to-Open Preview**: Cart dropdown now opens automatically when hovering over the cart button, with cart data loading immediately. Closes automatically when mouse moves away from the cart area. Implemented by:
        - Moved all hover logic to JavaScript (removed htmx mouseenter handler)
        - Separate timers for opening (150ms delay) and closing (300ms delay)
        - Boolean flags track mouse position over summary and list elements
        - **Critical fix**: Verify mouse position against element bounding box to detect false positive `mouseenter` events
        - Pico CSS's dropdown positioning was causing spurious events when mouse left dropdown list
        - Now checks if mouse coordinates are actually within summary bounds before opening
        - Prevents infinite open/close loop and allows proper hover-away closing
        - Click outside also closes the dropdown
        - Creates a smooth, professional dropdown menu UX
    - **Comprehensive Caching Fix**: 
        - Added cache-control meta tags to base HTML template
        - Added cache-control headers to all cart endpoints (`AddToCart`, `RemoveFromCart`, `ViewCart`, `CartCount`, `CartSummary`)
        - Added `hx-headers` attribute to htmx requests to prevent client-side caching
        - This multi-layered approach ensures fresh cart data without requiring users to clear browser cache
    - **User/Session Support**: Updated all cart handlers to properly support both authenticated users (via `user_id`) and anonymous users (via `session_id`).
    - **Auto-refresh**: Cart count and summary now automatically update when items are added or removed using htmx's `cart-updated` event.

## November 1, 2025

### Cart Quantity System & Advanced Features
Implemented a comprehensive quantity management system for the shopping cart with full UI/UX enhancements:

#### Quantity Consolidation
- **Database-level Aggregation**: Modified cart queries to use `GROUP BY` and `SUM(quantity)` to consolidate duplicate products into single line items
- **Subtotal Calculation**: Added `Subtotal` field to `CartItemView` struct, calculated as `Price * Quantity` in the Go handler
- **Display Updates**: Both cart page and dropdown summary now show quantity for each product

#### Interactive Quantity Editor
- **Integrated Quantity Controls**: Designed seamless +/- buttons with text input field:
    - Removed individual button borders for a cohesive look
    - Used `inline-flex` layout with shared border and border-radius
    - Transparent backgrounds that inherit from card background
    - Buttons and input visually integrated as a single component
- **Dynamic JavaScript Updates**:
    - `adjustQuantity()` function reads current input value and adjusts by ±1
    - `updateQuantity()` function validates (1-99 range) and sends update to server via fetch
    - Input validation prevents non-numeric characters
    - Page reloads after successful update to refresh all totals
- **Cross-browser Compatibility**: 
    - Replaced deprecated `onkeypress` with modern `oninput` event
    - Used regex (`/[^0-9]/g`) to strip non-numeric characters in real-time
    - Added `inputmode="numeric"` and `pattern="[0-9]*"` attributes
    - Fixed Firefox compatibility issue where quantity field wasn't editable

#### Cart Count Fix
- **Total Quantity Display**: Changed cart icon badge from counting rows (`COUNT(id)`) to summing quantities (`SUM(quantity)`)
- **Real-time Updates**: Cart icon now correctly reflects total item count after quantity adjustments
- **Example**: Cart with 3× Product A and 2× Product B shows (5), not (2)

#### Backend Support
- **New Endpoint**: Added `/cart/update` route and `UpdateCartQuantity` handler
- **SQL Updates**: `UPDATE cart_items SET quantity = $1 WHERE id = $2`
- **Event Triggering**: Sets `HX-Trigger: cart-updated` header to refresh cart count and summary
- **Validation**: Server-side quantity limits (1-99) prevent invalid values

#### UI Refinements
- **Dropdown Order**: Reordered cart summary display to show `ProductName | ×Qty | $Price`
- **Seamless Design**: Removed boxy containers and bright blue buttons from dropdown for a more integrated look
- **Consistent Styling**: Used CSS variables (`var(--muted-color)`, `var(--muted-border-color)`) for theme consistency

#### Bug Fixes
- **Login Cart Error**: Fixed "pq: column ci.user_id does not exist" by applying migration `005_add_user_id_to_cart.sql`
- **Template Function Error**: Resolved "function mul not defined" by moving subtotal calculation from template to Go handler
- **Stale Quantity Bug**: Fixed +/- buttons using static template values by implementing dynamic JavaScript that reads current input value
- **Logitech Cart Count**: Fixed an issue where cart count would not update when removing items from the cart summary dropdown.

### Technical Implementation Details
- **Go Template Limitations**: Addressed lack of arithmetic functions by pre-calculating values in handlers
- **HTMX Integration**: Leveraged `hx-trigger` with custom events for coordinated UI updates
- **Session Management**: Maintained support for both authenticated (`user_id`) and anonymous (`session_id`) carts throughout all quantity features
- **Cache Prevention**: Continued comprehensive cache-control strategy across all new endpoints

## November 4, 2025

### Product Page Quantity Controls
Enhanced the product listing page with the same quantity management controls used in the cart:

#### Frontend Implementation
- **Quantity Controls on Products**: Added integrated +/- buttons with text input to each product row
- **Visual Consistency**: Used identical styling to cart page (inline-flex layout, shared border, transparent backgrounds)
- **JavaScript Functions**:
  - `adjustProductQuantity()` - Increments/decrements quantity with 1-99 limits
  - `setQuantity()` - Validates and syncs quantity value to hidden form input on submit
  - Input validation with regex to strip non-numeric characters
  - Same cross-browser compatibility (inputmode="numeric", pattern="[0-9]*")

#### Backend Enhancement
- **Updated `AddToCart` Handler**: Now accepts optional `quantity` parameter from form
- **Smart Defaults**: Quantity defaults to 1 if not provided (backward compatible)
- **Validation**: Server-side enforcement of 1-99 quantity limits
- **Error Handling**: Added proper error checking for invalid product IDs

#### User Experience
- **Multi-item Adding**: Users can now add multiple quantities of a product in one action
- **Intuitive Interface**: Familiar +/- button pattern from cart page
- **Real-time Feedback**: Cart count badge updates immediately via htmx `cart-updated` event
- **No Page Reload**: Seamless addition to cart without leaving product page

#### Technical Details
- Modified `INSERT` statements in `AddToCart` to use variable quantity instead of hardcoded `1`
- Added hidden form field `quantity` that syncs with visible input on form submit
- Form submission via htmx with `hx-post="/cart/add"` maintains existing cart update flow

### Critical Bug Fix: Duplicate Cart Items
**Problem Discovered**: When adjusting quantities in the cart, values were jumping erratically (e.g., 28→35→40). Investigation revealed the root cause: multiple duplicate rows existed in the database for the same product+user combination.

#### Root Cause Analysis
1. **Original Bug**: `AddToCart` handler always executed `INSERT`, creating new rows instead of checking for existing items
2. **Display Masking**: `ViewCart` used `GROUP BY` with `SUM(quantity)` to consolidate display, hiding the underlying duplicates
3. **Update Failure**: `UpdateCartQuantity` only updated one row (using `MIN(ci.id)`), leaving other duplicates untouched
4. **Result**: Cart displayed consolidated view, but updates affected only partial data

#### Comprehensive Fix Implemented

**Backend Changes (cart.go)**:
1. **AddToCart** - Complete rewrite with deduplication:
   - Checks for existing cart items: `SELECT SUM(quantity) WHERE user_id/session_id AND product_id`
   - If exists: `DELETE` all duplicates, then `INSERT` single consolidated row with updated quantity
   - If new: `INSERT` fresh row
   - Prevents duplicates at the source

2. **UpdateCartQuantity** - Consolidation logic:
   - Queries cart item to get `product_id`, `user_id`, `session_id`
   - Uses `sql.NullInt64` and `sql.NullString` for proper NULL handling
   - `DELETE` all duplicate rows for product+user/session combination
   - `INSERT` single row with new quantity
   - Ensures atomic consolidation on every update

3. **RemoveFromCart** - Complete removal:
   - Queries cart item metadata (product_id, user_id, session_id)
   - `DELETE` all duplicates for that product+user/session
   - Prevents partial removal bugs

**Database Changes (Migration 006)**:
- Created `006_consolidate_duplicate_cart_items.sql` migration
- Consolidated 73 duplicate rows from existing cart data
- Added unique constraints to prevent future duplicates:
  - `idx_cart_items_user_product` - UNIQUE (user_id, product_id) WHERE user_id IS NOT NULL
  - `idx_cart_items_session_product` - UNIQUE (session_id, product_id) WHERE session_id IS NOT NULL
- Database now enforces one row per product per user/session at schema level

#### Testing Results
- ✅ Existing duplicates cleaned up (73 rows consolidated)
- ✅ Quantity adjustments now work correctly
- ✅ AddToCart properly updates existing items instead of creating duplicates
- ✅ RemoveFromCart removes all instances of a product
- ✅ Database constraints prevent future duplicate creation

#### Technical Notes
- Used `sql.NullInt64` and `sql.NullString` for nullable foreign keys
- Delete-then-insert pattern ensures atomic consolidation
- Unique partial indexes (with WHERE clauses) allow NULL values while preventing duplicates
- All handlers now guarantee single-row-per-product invariant

### Checkout Flow Bug Fix
**Problem Discovered**: "Proceed to Checkout" button wasn't working for authenticated users. Checkout page showed empty cart despite items being present.

#### Root Cause
1. **CheckoutPage Handler**: Only queried cart items using `session_id`, but authenticated users' cart items have `user_id` set
2. **Missing Quantity Support**: Checkout didn't display or calculate quantities properly
3. **ProcessOrder Handler**: Also only used `session_id`, couldn't find items for authenticated users

#### Fix Implemented

**CheckoutPage Updates**:
- Added logic to check both `user_id` (authenticated) and `session_id` (anonymous)
- Updated SQL queries to use `GROUP BY` and `SUM(quantity)` for consolidated display
- Added quantity and subtotal calculations
- Redirects to cart if empty (better UX)
- Now displays: Product, Description, Price, Quantity, Subtotal, Total

**ProcessOrder Updates**:
- Added user authentication check
- Support for both `user_id` and `session_id` based cart clearing
- Updated order item insertion to use `SUM(quantity)` with `GROUP BY`
- Ensures proper cart cleanup for authenticated users

**Template Updates** (checkout.html):
- Added Quantity and Subtotal columns
- Updated total calculation to show accurate sum
- Changed terminology from "Title/Author" to "Product/Description"

#### Testing Results
- ✅ Checkout page loads with items (authenticated users)
- ✅ Quantities displayed correctly
- ✅ Subtotals calculated: price × quantity
- ✅ Total calculated: sum of all subtotals
- ✅ Order processing works
- ✅ Cart cleared after order completion

---

## Testing Strategy Implementation

### Problem Statement
We realized with each change we were breaking something else. Need systematic testing to prevent regressions.

### Solution: Comprehensive Testing Framework

**Created TESTING.md** - Complete test plan including:
1. **Smoke Tests**: Quick tests after every change (13 core tests)
2. **Regression Tests**: Verify previously fixed bugs don't resurface
3. **Integration Tests**: Cross-component and cross-browser testing
4. **Database Integrity Tests**: SQL queries to verify constraints
5. **Manual Test Script**: 10-minute end-to-end flow for commits

**Created test-smoke.sh** - Automated smoke test script:
- Uses curl to test all endpoints
- Creates test users automatically
- Verifies database integrity
- Checks for duplicate cart items
- Validates unique constraints exist
- Color-coded pass/fail output
- Exit code 0 = all pass, 1 = failures

**Test Categories**:
```
✅ Product Listing (1 test)
✅ Anonymous Cart Operations (4 tests)
✅ Authenticated Cart Operations (3 tests)
✅ Checkout Flow (1 test)
✅ Database Connectivity (1 test)
✅ Duplicate Prevention (1 test)
✅ Constraint Verification (1 test)
---
Total: 13 automated tests
```

#### Usage
```bash
# Run smoke tests
./test-smoke.sh

# Run before every commit
./test-smoke.sh && git commit -m "message"
```

#### Test Results (November 4, 2025)
All 13 smoke tests **PASSED** ✅
- Server responding
- Products page loading
- Cart operations working (anonymous + authenticated)
- No duplicate cart items in database
- Unique constraints in place
- Checkout flow functional

### Testing Workflow Going Forward
1. **Before coding**: Review TESTING.md for affected areas
2. **During coding**: Think about test cases
3. **After coding**: Run `./test-smoke.sh`
4. **Before commit**: Run full smoke tests + manual spot checks
5. **After commit**: Note any new test cases needed

### Cart Merging on Login/Signup
**Problem Discovered**: When users add items to cart while not logged in, then login or signup, their anonymous cart items were lost. This is poor UX - users expect their cart to persist.

#### Root Cause
The Login handler had basic cart transfer logic (`UPDATE cart_items SET user_id = ... WHERE session_id = ...`), but this:
1. **Failed silently**: If the user already had the same products in their cart, it created duplicates (violating unique constraints)
2. **Replaced instead of merged**: Didn't add quantities together
3. **Only in Login**: Signup handler had no cart merging at all

#### Solution: Intelligent Cart Merging

**Created `mergeAnonymousCart()` helper function**:
- Queries anonymous cart items (grouped by product with summed quantities)
- For each product:
  - Checks if user already has that product
  - Deletes all existing rows for that product+user
  - Calculates merged quantity (existing + anonymous, capped at 99)
  - Inserts single consolidated row
- Deletes anonymous cart items
- All wrapped in transaction for atomicity

**Technical Implementation**:
- Query anonymous cart OUTSIDE transaction (avoid PostgreSQL protocol error)
- Store results in slice of structs
- Process each product inside transaction
- Delete-then-insert pattern ensures no duplicates
- Works with existing unique constraints

**Applied to both**:
- **Login handler**: Merges cart when existing user logs in
- **Signup handler**: Merges cart when new user registers

#### User Experience
**Before**: 
- Anonymous: Add Product A (qty 3) → Login → Cart empty 😞

**After**:
- Anonymous: Add Product A (qty 3) → Login → Cart shows Product A (qty 3) ✅
- Better: Already have Product A (qty 2) → Add Product A (qty 3) as anonymous → Login → Cart shows Product A (qty 5) ✅

#### Testing
Added 2 new automated tests (now 15 total):
1. **Test 14**: Cart merge on signup (anonymous → new user)
2. **Test 15**: Cart merge with existing items (quantities add together)

**All 15 smoke tests passing** ✅

#### Edge Cases Handled
- No anonymous cart: No-op, continues normally
- Quantity overflow: Caps at 99
- Merge failure: Logs error, continues with login (graceful degradation)
- Database constraints: Delete-then-insert avoids unique constraint violations
