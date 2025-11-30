# Project Diary: 12-Factor E-commerce Template

> **📋 See [`PLANNING.md`](PLANNING.md) for complete project vision, roadmap, and VCF 9.0 integration strategy**

## 🎯 Quick Status Summary
**Last Updated:** November 30, 2025  
**Project Status:** ✅ Phase 1 Complete + Phase 2 UI Polish In Progress  
**Recent Focus:** Product Images, Table/Tile Toggle, Compact Cart, Order History  
**Next Up:** Individual Product Detail Pages  
**Project Goal**: Demo platform to showcase VMware Cloud Foundation (VCF) 9.0 capabilities through real-world e-commerce application

### What's Working
- ✅ **Modern Architecture:** Refactored data layer into a `Repository Pattern` (Hexagonal Architecture ready).
- ✅ **Advanced Schema:** Added Categories, SKUs, Stock Levels, User Roles, and Order History linkage.
- ✅ **Search:** Full product search functionality (SQL-based, ready for Elasticsearch swap).
- ✅ **Responsive UI:** New header with mobile "Hamburger" menu and search bar.
- ✅ **Sticky Header:** Fixed position header that shrinks on scroll with smooth transitions
- ✅ **Product Images:** Cards with images, table/tile toggle, minimal controls
- ✅ **Order History:** Expandable cards with full order details and product images
- ✅ **Compact Cart:** Single-row items with 80x80px thumbnails
- ✅ **User Features:** "My Orders" page to view purchase history
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

## November 30, 2025

### UI Polish Session: Product Images, View Toggle, Compact Cart

**Goal**: Complete UI overhaul with product images, flexible viewing options, and compact layouts throughout the shopping experience.

#### Session Summary
Today we transformed the application from basic table-based views into a modern, image-rich e-commerce experience with multiple viewing options and compact, efficient layouts.

#### Major Features Completed

**1. Product Images & Card Grid**
- Modern card-based layout with large product images (250px)
- Responsive grid (auto-fill, 280px minimum)
- Hover effects (lift + shadow)
- 📦 fallback icons for missing/broken images
- Touch-friendly on mobile

**2. Table/Tile View Toggle**
- Toggle buttons in header (🔲 Tiles / ☰ Table)
- Tile view: Large cards with images
- Table view: Compact rows with 60px thumbnails
- Preference saved in localStorage (persists across sessions)
- Synchronized quantity controls across both views
- Mobile: Hides toggle, always shows tiles

**3. Compact Cart Layout**
- Single-row items (height = 80x80px thumbnail)
- All info visible at once (no wasted vertical space)
- Minimal quantity controls
- Smaller remove button
- Grid layout: Image | Info | Price | Qty | Subtotal | Remove

**4. Minimal UI Controls**
- Quantity buttons: 0.25rem padding, 0.9rem font, muted color
- Input width: 40px (compact)
- Remove button: 0.25rem padding
- Add to Cart: Smaller font and padding
- Hover effects for interaction feedback
- Inline styles for CSS specificity

**5. Enhanced Order History**
- Expandable order cards
- Product images in order items (📦 fallback)
- Alphabetically sorted items
- Full product details per order

#### Technical Implementation

**View Toggle System**:
```javascript
// Saves preference to localStorage
// Restores on page load
// Synchronizes quantity across views
```

**Image Fallback Pattern**:
```html
<img onerror="this.style.display='none'; this.nextElementSibling.style.display='flex';">
<div class="placeholder" style="display:none;">📦</div>
```

**Compact Grid Layout**:
```css
grid-template-columns: 80px 1fr auto auto auto auto;
gap: 1rem;
padding: 1rem;
```

#### User Testing & Iterations

**Iteration 1**: Card grid with large images
- Issue: Cart took too much vertical space
- Fix: Reduced to 80x80px, single-row layout

**Iteration 2**: Minimal controls attempt via CSS classes
- Issue: Pico CSS overrode button styles
- Fix: Used inline styles for higher specificity

**Iteration 3**: Quantity controls not shrinking
- Issue: CSS classes had lower specificity than Pico
- Fix: Applied inline styles directly to elements

#### Files Modified
- `templates/products.html`: Grid + table toggle, minimal controls
- `templates/cart.html`: Compact layout, minimal buttons
- `templates/orders.html`: Image fallbacks, alphabetical sort
- `internal/repository/postgres.go`: ORDER BY p.name for order items
- `diary.md`: This documentation

#### Test Results
- ✅ All 15 smoke tests passing
- ✅ View toggle works and persists
- ✅ Images load with fallbacks
- ✅ Cart is compact (single-row items)
- ✅ Controls are minimal but functional
- ✅ Mobile responsive
- ✅ User tested and approved

#### Commits Today
```
6986c49 feat: Add product images with table/tile toggle and minimal controls
5d3d407 fix: Order history improvements - images and alphabetical sorting
1230948 feat: Enhanced order history page with detailed order items
4199d83 fix: Resolve session cookie and cart bugs (all tests passing)
7a74ca1 feat: Add sticky header with scroll shrinking behavior
```

#### Next Steps
**Tomorrow's Focus: Product Detail Pages**
- Click on product card → detail page
- Larger images
- Full description
- Related products
- Reviews (future)
- Add to cart from detail page

---

### UI Enhancement: Product Images Throughout Site

**Goal**: Add product images to all shopping pages (products, cart) with modern card-based layouts.

#### Features Implemented

**Products Page (Homepage)**:
- **Card Grid Layout**: Responsive grid (280px min, auto-fill)
- **Product Images**: Large 250px images with fallback icons
- **Visual Cards**: Hover effects (lift + shadow)
- **Better Information Hierarchy**: Image → Title → Description → Price → Actions
- **Improved Actions**: Quantity controls + Add to Cart in same row

**Cart Page**:
- **List-Based Layout**: Horizontal cards showing all info at once
- **Product Thumbnails**: 100px × 130px images with fallback
- **Enhanced Summary**: Total displayed in prominent card
- **Better Empty State**: Icon + helpful message + call-to-action
- **Mobile Responsive**: Stacks to 2 columns on mobile

**Image Handling**:
- Fallback to 📦 icon for missing images
- `onerror` handler for broken image URLs
- Placeholder background color matches theme
- Graceful degradation

#### Design Improvements

**Before (Table Layout)**:
- Plain table rows
- No images
- Cramped on mobile
- Limited visual appeal

**After (Card Layout)**:
- Modern card design
- Product images prominent
- Touch-friendly spacing
- Professional e-commerce look

**CSS Grid Benefits**:
- Automatic responsive columns
- Equal height cards
- Better spacing control
- Modern browser support

#### Technical Details

**Products Grid**:
```css
grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
gap: 1.5rem;
```
- Auto-fills columns based on viewport width
- Minimum 280px per card
- Equal distribution of space

**Cart Items**:
```css
grid-template-columns: 100px 1fr auto auto auto;
```
- Fixed image column (100px)
- Flexible info column (1fr)
- Auto-sized action columns

#### Mobile Optimization
- Products: 240px minimum card width
- Cart: Stacks to 80px image + details
- Image heights reduced (250px → 200px, 130px → 100px)
- Touch-friendly buttons and controls

#### Files Modified
- `templates/products.html`: Complete redesign with card grid
- `templates/cart.html`: Card-based list layout
- Both: Added image handling and fallbacks

#### Test Results
- ✅ All 15 smoke tests passing
- ✅ Product grid displays correctly
- ✅ Images load or show fallback icon
- ✅ Cart shows product images
- ✅ Quantity controls still functional
- ✅ Add to cart works from new layout
- ✅ Mobile responsive breakpoints work

---

### UI Enhancement: Order History Page with Full Details

**Goal**: Transform the basic order history page into a rich, detailed view with expandable orders showing all items purchased.

#### Features Implemented

1. **Card-Based Layout**: Modern card design for each order
2. **Expandable Details**: Click to expand and see all items in an order
3. **Rich Product Display**: Shows product images, names, descriptions, quantities, and prices
4. **Status Badges**: Color-coded status indicators (completed, pending, processing)
5. **Responsive Design**: Mobile-optimized grid layout
6. **Enhanced Date Formatting**: Full date with time for better clarity
7. **Empty State**: Beautiful empty state with icon when no orders exist

#### Technical Implementation

**Repository Enhancement** (`internal/repository/postgres.go`):
- Modified `GetOrdersByUserID` to eagerly load order items with product details
- Added JOIN query to fetch products for each order item:
  ```sql
  SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price,
         p.id, p.name, p.description, p.price, p.image_url
  FROM order_items oi
  JOIN products p ON oi.product_id = p.id
  WHERE oi.order_id = $1
  ```
- N+1 query pattern (acceptable for order history use case)

**Template Redesign** (`templates/orders.html`):
- Used HTML5 `<details>` element for native expandable functionality
- CSS Grid layout for responsive design
- Status badges with semantic colors
- Product cards with image thumbnails
- Mobile-first breakpoints at 768px

**Handler Improvements** (`internal/handlers/orders.go`):
- Added debug logging for troubleshooting
- Proper error handling with user-friendly messages
- Template execution error checking

#### Design Features

**Order Card Structure**:
- Order ID (clickable to expand)
- Date (Monday, January 2, 2006 at 3:04 PM)
- Status badge (color-coded)
- Total amount (prominent)

**Order Details** (expanded):
- Item count header
- Product grid with:
  - 60x80px image thumbnail
  - Product name and description
  - Quantity purchased
  - Price per item

**Status Color Scheme**:
- `completed`: Green (#d4edda / #155724)
- `pending`: Yellow (#fff3cd / #856404)
- `processing`: Blue (#d1ecf1 / #0c5460)

#### Bug Fixes

**Template Rendering Issue**:
- **Problem**: `orders.html` missing `{{template "base.html" .}}` directive
- **Result**: Page rendered empty (no base layout)
- **Fix**: Added template directive at top of file
- **Lesson**: All page templates must include base template

#### User Experience

**Before**: Simple table with order ID, date, status, total
**After**: 
- Rich, interactive cards
- Full order details on-demand
- Visual product display
- Professional e-commerce look

#### Mobile Optimization
- Simplified grid layout (2 columns vs 4)
- Stacked quantity and price info
- Touch-friendly tap targets
- Smaller image thumbnails (50px)

#### Files Modified
- `internal/repository/postgres.go` - Order items loading
- `templates/orders.html` - Complete redesign
- `internal/handlers/orders.go` - Debug logging

#### Test Results
- ✅ All 15 smoke tests passing
- ✅ Order history loads correctly
- ✅ Order details expand/collapse
- ✅ Product images and details display
- ✅ Status badges render correctly
- ✅ Empty state shows when no orders
- ✅ Mobile responsive layout works

---

### Critical Bug Fixes: Session Cookie & Cart Issues

**Problem**: 3 automated smoke tests were failing:
1. Cart count showing (0) after adding items
2. Cart not merged on signup
3. Cart quantities not merged correctly on login

#### Root Cause Analysis

**Issue #1: Session Cookie Configuration**
- Gorilla/sessions was setting `Secure; SameSite=None` by default
- These flags require HTTPS, breaking localhost testing
- Cookies weren't being sent back to server, so sessions didn't persist

**Issue #2: Session ID Logic Bug**
- In `AddToCart` handler, new session ID was created and saved
- But code then checked the OLD `sessionOk` variable (still false)
- This reset `sessionID` to empty string before database insert
- Result: Cart items inserted with NULL session_id and NULL user_id

#### Fixes Implemented

**1. Session Configuration** (`cmd/web/main.go`)
```go
store.Options = &sessions.Options{
    Path:     "/",
    MaxAge:   86400 * 30, // 30 days
    HttpOnly: true,
    Secure:   false,      // Allow HTTP for development
    SameSite: http.SameSiteLaxMode,
}
```

**2. AddToCart Session Logic** (`internal/handlers/cart.go`)
- Fixed session ID creation to update `sessionOk = true` after creating ID
- Prevents resetting sessionID to empty string
- Now properly tracks anonymous users with session IDs

**3. Dark Mode Header Fix** (`templates/base.html`)
- Removed `rgba()` color with white fallback that broke dark mode
- Header now respects Pico CSS `--background-color` in both scrolled/non-scrolled states
- Added `-webkit-backdrop-filter` for Safari support

#### Test Results
✅ **All 15 smoke tests passing!**
- Cart count works for anonymous users
- Cart persists across page loads
- Cart merges correctly on signup (anonymous → authenticated)
- Cart quantities add together on login (2 + 3 = 5)
- No duplicate cart items
- All database constraints verified

#### Technical Details

**Session Cookie Before Fix:**
```
Set-Cookie: cart-session=...; Secure; SameSite=None
```
→ Browser rejects on HTTP localhost

**Session Cookie After Fix:**
```
Set-Cookie: cart-session=...; HttpOnly; SameSite=Lax
```
→ Works on localhost, secure for production with HTTPS

**Database State Before Fix:**
```sql
SELECT * FROM cart_items;
 id | session_id | user_id | product_id | quantity
----|------------|---------|------------|----------
171 |    NULL    |  NULL   |     1      |    12     ← Orphaned!
```

**Database State After Fix:**
```sql
SELECT * FROM cart_items;
 id |     session_id      | user_id | product_id | quantity
----|---------------------|---------|------------|----------
175 | $uuid-valid-string  |  NULL   |     1      |    2      ← Valid!
```

#### Files Modified
- `cmd/web/main.go` - Session cookie configuration
- `internal/handlers/cart.go` - Fixed session ID logic
- `templates/base.html` - Dark mode header background

#### Production Notes
**Important**: Before deploying to production:
- Set `Secure: true` in session options
- Ensure HTTPS is enabled
- Consider using environment variable for secure flag: 
  ```go
  Secure: os.Getenv("ENV") == "production"
  ```

---

### Phase 2 Kickoff: Sticky Header Implementation

**Goal**: Implement a modern, sticky header that stays at the top when scrolling - a common UX pattern on e-commerce sites like Amazon.

#### Features Implemented
1. **Fixed Positioning**: Header remains visible at top of viewport on scroll
2. **Smooth Shrinking**: Header compacts when scrolled (reduced padding, smaller search bar)
3. **Visual Effects**: Enhanced shadow and backdrop blur when scrolled
4. **Smooth Transitions**: All size changes animated with CSS transitions (0.3s ease)
5. **Mobile Responsive**: Different padding for mobile devices (70px vs 80px)

#### Technical Implementation

**CSS Changes**:
- `position: fixed` with `z-index: 1000` for header wrapper
- Body padding-top to reserve space and prevent content jump
- `.scrolled` class applied dynamically for compact state
- Transitions on padding, font-size, and height for smooth animations
- Backdrop filter with blur effect for modern glass morphism look

**JavaScript**:
- Scroll event listener monitors `window.pageYOffset`
- Adds `.scrolled` class when scroll > 50px
- Removes class when scrolled back to top
- Minimal performance impact (simple class toggle)

**Scroll States**:
```
Not Scrolled:
- Full padding: 0.5rem
- Search bar height: 2.5rem
- Brand font: normal size
- Shadow: light (2px)

Scrolled (>50px):
- Compact padding: 0.25rem
- Search bar height: 2rem
- Brand font: 1.2rem
- Shadow: enhanced (4px + backdrop blur)
- Semi-transparent background (95% opacity)
```

#### Docker Build Fix
**Issue**: Alpine Linux base image had SSL certificate verification errors preventing package installation.

**Solution**: 
- Removed unnecessary `apk add ca-certificates git` command
- Git not needed for build process
- Simplified Dockerfile to just copy `go.mod` and `go.sum` and run `go mod download`

#### User Experience Benefits
- ✅ Navigation always accessible without scrolling up
- ✅ More screen space for content when scrolled
- ✅ Professional, modern e-commerce feel
- ✅ Visual feedback that page state has changed
- ✅ Works seamlessly with existing cart hover behavior

#### Files Modified
- `templates/base.html` - Added sticky header styles and scroll behavior JavaScript
- `Dockerfile` - Simplified build process, removed git dependency

#### Testing Results
- ✅ Header stays fixed at top on scroll
- ✅ Smooth transition to compact state
- ✅ All interactive elements (cart, search, menus) remain functional
- ✅ Mobile responsive behavior maintained
- ✅ No layout shift or content jump
- ✅ Compatible with existing htmx cart functionality

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
