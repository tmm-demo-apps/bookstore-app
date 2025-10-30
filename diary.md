# Project Diary: 12-Factor Bookstore App

## October 30, 2025

### Project Goal
Create a demo shopping cart application for selling books, designed for Kubernetes using the 12-factor app methodology.

### What We've Done So Far
- **Project Initialization:** Set up a Go project with a standard directory structure and initialized a local Git repository for version control.
- **Backend Development:** Implemented a basic Go web server that connects to a PostgreSQL database to fetch and display a list of books using server-side templates.
- **Containerization:** Wrote a multi-stage `Dockerfile` to create an optimized container image for the application and a `docker-compose.yml` file to run the app and database locally.
- **Deployment Setup:** Created Kubernetes manifests (`deployment.yaml`, `service.yaml`) for both the application and the PostgreSQL database.
- **CI/CD Pipeline:** Set up a basic CI/CD workflow using GitHub Actions to automate the building and pushing of the Docker image.
- **Debugging:**
    - Resolved an initial error caused by Go not being installed on the system.
    - Fixed several build failures due to a Go version mismatch between the `go.mod` file and the Docker image. We standardized on Go version `1.21`.
    - Resolved further Go version mismatches by updating the `Dockerfile` to use `go 1.23` after a dependency update modified the `go.mod` file.
    - Added the `github.com/google/uuid` dependency to `go.mod` to fix a "no required module" build error.
- **Application Preview:** Successfully launched the application locally using Docker Compose and populated the database with sample data.

### Next Steps
The foundational structure is complete. The next phase will be to build out the core shopping cart features, including:
- Adding a book to a shopping cart.
- Viewing the items in the cart.
- Implementing a basic checkout process.
- We can also look at enhancing the Kubernetes manifests for better configuration management and security (e.g., using Secrets for database credentials).

### Current Focus: Finalizing Shopping Cart
Based on our discussion, we've decided to complete the core cart functionality before moving to checkout. The immediate next step is to implement the ability to remove items from the cart.

### Next Steps
- **Implement "Remove from Cart"**: Added a button and handler to remove items from the shopping cart. This completes the core cart functionality (add, view, remove).
- **Implement Checkout Process**: Implemented a PII-free checkout process. Users can now view an order summary, confirm their order without entering personal data, and receive a confirmation. Cart items are converted into a historical order in the database.

### Current Focus: UI Improvement
We are now focusing on improving the user interface. We will use the Pico.css framework to provide a clean, modern look and feel with minimal changes to the HTML structure.

### Next Steps
- **Future Enhancements**:
    - Expanded book selection and categorization
    - User management (settings, profile, etc.)
- We can also look at enhancing the Kubernetes manifests for better configuration management and security (e.g., using Secrets for database credentials).
