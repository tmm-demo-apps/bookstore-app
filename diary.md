# Project Diary: 12-Factor E-commerce Template

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

### Technical Implementation Details
- **Go Template Limitations**: Addressed lack of arithmetic functions by pre-calculating values in handlers
- **HTMX Integration**: Leveraged `hx-trigger` with custom events for coordinated UI updates
- **Session Management**: Maintained support for both authenticated (`user_id`) and anonymous (`session_id`) carts throughout all quantity features
- **Cache Prevention**: Continued comprehensive cache-control strategy across all new endpoints

### Next Steps
- **Future Enhancements**:
    - Expanded product selection and categorization.
    - User management (settings, profile, etc.).
    - Order history page for users to view past orders.
