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
- **Core Shopping Cart:** Implemented the ability to add, view, and remove items from the shopping cart.
- **Checkout Process:** Implemented a PII-free checkout process where users can confirm an order without entering personal data.
- **UI Improvement**: Refactored the frontend using Pico.css and a base template structure to create a clean, modern, and consistent user interface.
- **Advanced Cart Features**: Implemented several UI/UX improvements for the shopping cart using htmx, including a dynamic cart count, a hover-enabled cart summary, and a total cost display.
- **Debugging:**
    - Resolved an initial error caused by Go not being installed on the system.
    - Fixed several build failures due to a Go version mismatch between the `go.mod` file and the Docker image, eventually standardizing on Go 1.23.
    - Added missing package dependencies (`github.com/google/uuid`) to resolve build errors.
- **Kubernetes Security**: Refactored the Kubernetes manifests and application code to use Kubernetes Secrets for managing database credentials, removing sensitive data from version control.
- **User Management & Refactoring**: Implemented a complete user management system (registration, login, logout) with secure password hashing. As part of this, all Go handlers were refactored to use a centralized `Handlers` struct for cleaner code and better dependency management.

### Current Focus: User Management Enhancements
We are now working on a series of improvements to the user management and cart systems based on user feedback. The current tasks are to fix the login functionality and improve error handling.

### Next Steps
- Implement auto-login after signup.
- Add password strength and email validation to the signup process.
- Implement a "show/hide password" feature.
- Allow anonymous users to use the shopping cart.
- **Future Enhancements**:
    - Expanded book selection and categorization.
