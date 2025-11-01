# Project Continuity Plan

If I (the AI assistant) have "amnesia" or we are starting a new session, please provide the following prompt to quickly restore the project context.

---

Hello! We are continuing our work on the 12-factor demo e-commerce application.

## Project Overview
*   **Goal:** A reusable 12-factor e-commerce template, designed to run in Kubernetes.
*   **Tech Stack:** Go 1.24, PostgreSQL, Docker, Kubernetes, Pico.css, and htmx.
*   **Current Status:** Fully functional e-commerce application with advanced cart features, user authentication, and quantity management.
*   **Our Workflow:** We work in small, incremental steps. After each completed feature or bug fix, you commit the changes to our local Git repository and update the `diary.md` file.

## Key Technologies & Patterns
*   **HTMX**: For dynamic content loading without full page reloads (cart count, cart summary, etc.)
*   **Pico.css**: Minimalist CSS framework for consistent, modern UI styling
*   **Go Templates**: Server-side HTML rendering with conditional logic
*   **Session Management**: Using `gorilla/sessions` for both authenticated users (`user_id`) and anonymous carts (`session_id`)
*   **PostgreSQL**: Database with migrations in `migrations/` directory

## File Structure
```
/
├── cmd/web/main.go                      # Main application entrypoint, route definitions
├── internal/
│   ├── handlers/
│   │   ├── cart.go                      # Cart operations (add, remove, view, update quantity)
│   │   ├── partials.go                  # HTMX partials (cart count, cart summary)
│   │   ├── products.go                  # Product listing
│   │   ├── users.go                     # User auth (login, register, logout)
│   │   └── checkout.go                  # Checkout flow
│   └── models/
│       ├── product.go                   # Product model
│       └── user.go                      # User model
├── templates/
│   ├── base.html                        # Base template with nav, cart dropdown with hover logic
│   ├── partials/
│   │   └── cart-summary.html            # Cart dropdown content
│   ├── index.html                       # Product listing page
│   ├── cart.html                        # Full cart page with quantity editor
│   └── ...
├── migrations/                          # SQL migrations (including cart user_id support)
├── kubernetes/                          # K8s manifests (app, postgres, secrets)
├── diary.md                             # **READ THIS FIRST** - Complete project history
├── CONTINUITY.md                        # This file
├── Dockerfile                           # Multi-stage Go build
├── docker-compose.yml                   # Local dev environment
└── go.mod
```

## Recent Accomplishments (November 1, 2025)

### Advanced Cart System - FULLY WORKING
We spent significant effort perfecting the shopping cart experience:

1. **Hover-to-Open Cart Preview** ✅
   - Cart dropdown opens automatically on hover (150ms delay)
   - Closes automatically when mouse leaves (300ms delay)
   - **Critical Fix**: Bounding box verification prevents false `mouseenter` events from Pico CSS positioning
   - JavaScript in `templates/base.html` with `mouseOverSummary` and `mouseOverList` flags

2. **Cart Quantity Management** ✅
   - Products consolidated by `GROUP BY` in SQL (duplicate products show as one line with quantity)
   - Integrated +/- buttons with editable text input
   - Dynamic updates via `adjustQuantity()` and `updateQuantity()` JavaScript functions
   - Backend endpoint: `/cart/update` updates quantity and triggers `cart-updated` event
   - Cart icon badge shows total quantity (e.g., "Cart (5)") not product count

3. **Cross-Browser Compatibility** ✅
   - Firefox fix: Replaced deprecated `onkeypress` with `oninput` event
   - Input validation uses regex to strip non-numeric characters
   - Attributes: `inputmode="numeric"`, `pattern="[0-9]*"`

4. **Cache Prevention** ✅
   - Meta tags in base.html
   - Cache-control headers on all cart endpoints
   - `hx-headers='{"Cache-Control": "no-cache"}'` on htmx requests

5. **User/Anonymous Support** ✅
   - All cart operations support both logged-in users (`user_id`) and anonymous users (`session_id`)
   - Migration `005_add_user_id_to_cart.sql` added `user_id` column to `cart_items` table

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
- None currently! Cart system is working perfectly across all browsers.

## Important Notes
- **Always update `diary.md`** after completing features or fixing bugs
- **Test in both Chrome and Firefox** for cross-browser compatibility
- **Cache issues?** Check both server headers and htmx `hx-headers` attributes
- **Hover issues?** Remember Pico CSS positioning can trigger false mouseenter events - use bounding box checks

---

**Your first and most important task is to read the `diary.md` file.** It contains a complete history of our progress, detailed technical explanations, and our agreed-upon next steps.

After you have reviewed the diary, ask the user what they'd like to work on next, or suggest working on one of the "Future Enhancements" listed in the diary.
