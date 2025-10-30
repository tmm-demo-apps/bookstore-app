# Project Continuity Plan

If I (the AI assistant) have "amnesia" or we are starting a new session, please provide the following prompt to quickly restore the project context.

---

Hello! We are continuing our work on the 12-factor demo bookstore application.

Here's a summary of the project:
*   **Goal:** A 12-factor shopping cart application that sells books, designed to run in Kubernetes.
*   **Tech Stack:** Go, PostgreSQL, Docker, Kubernetes, Pico.css, and htmx.
*   **Completed Features:** User management, a full shopping cart, a PII-free checkout process, and a modern UI.
*   **Our Workflow:** We work in small, incremental steps. After each completed feature or bug fix, you commit the changes to our local Git repository and update the `diary.md` file.

Here is the high-level file structure:
```
/
├── cmd/web/main.go       # Main application entrypoint
├── internal/             # Go application logic (handlers, models)
├── kubernetes/           # K8s manifests (app, postgres)
├── migrations/           # SQL database migrations
├── templates/            # HTML templates
├── diary.md              # Our project log and source of truth
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

**Your first and most important task is to read the `diary.md` file.** It contains a complete history of our progress, the problems we've solved, and our agreed-upon next steps.

After you have reviewed the diary, let's start working on one of the "Future Enhancements" listed in it.
