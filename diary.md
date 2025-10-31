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
    - **Hover Preview**: Cart summary now loads immediately when dropdown opens, no click required.
    - **Comprehensive Caching Fix**: 
        - Added cache-control meta tags to base HTML template
        - Added cache-control headers to all cart endpoints (`AddToCart`, `RemoveFromCart`, `ViewCart`, `CartCount`, `CartSummary`)
        - Added `hx-headers` attribute to htmx requests to prevent client-side caching
        - This multi-layered approach ensures fresh cart data without requiring users to clear browser cache
    - **User/Session Support**: Updated all cart handlers to properly support both authenticated users (via `user_id`) and anonymous users (via `session_id`).
    - **Auto-refresh**: Cart count and summary now automatically update when items are added or removed using htmx's `cart-updated` event.
    - **Loading Fix**: Fixed "Loading..." text by using `hx-on:toggle` event on the details element to trigger cart load immediately when dropdown opens, rather than waiting for mouseenter on the list.

### Next Steps
- **Future Enhancements**:
    - Expanded product selection and categorization.
    - User management (settings, profile, etc.).
    - Order history page for users to view past orders.
