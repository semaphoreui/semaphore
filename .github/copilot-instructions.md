# Semaphore UI Development Instructions

**Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.**

Semaphore UI is a modern web interface for managing popular DevOps tools like Ansible, Terraform, PowerShell, and Bash scripts. It's built with Go (backend) and Vue.js (frontend), using Task runner for build automation.

## Working Effectively

### Bootstrap, build, and test the repository:
- Install Go 1.21+ (currently requires go version 1.21 or higher)
- Install Node.js 16+ 
- Install Task runner: `go install github.com/go-task/task/v3/cmd/task@latest`
- Install dependencies: `task deps` -- takes 3 minutes first time (faster with cache). NEVER CANCEL. Set timeout to 5+ minutes.
- Build the application: `task build` -- takes 1.5 minutes. NEVER CANCEL. Set timeout to 3+ minutes.
- Run tests: `task test` -- takes 33 seconds. NEVER CANCEL. Set timeout to 2+ minutes.

### Run the application:
- ALWAYS run the bootstrapping steps first
- Setup database and admin user: `./bin/semaphore setup` (interactive, use BoltDB option 2 for development)
- Start server: `./bin/semaphore server --config ./config.json`
- Web UI: http://localhost:3000 (login: admin / changeme)
- API: http://localhost:3000/api/ (test with: `curl http://localhost:3000/api/ping`)

## Validation

- **CRITICAL**: Always manually validate any new code by building and running the application.
- ALWAYS run through at least one complete end-to-end scenario after making changes:
  1. Build the application: `task build`
  2. Start the server: `./bin/semaphore server --config ./config.json`
  3. Test API endpoint: `curl http://localhost:3000/api/ping` (should return "pong")
  4. Access web UI at http://localhost:3000 and verify it loads
  5. For auth changes: Test login with admin/changeme
- For significant changes, run full setup process to ensure setup still works
- Always build and exercise your changes before considering the task complete

### Complete Validation Scenario (for major changes):
```bash
# 1. Clean build
task build

# 2. Setup database (if config.json doesn't exist)
./bin/semaphore setup  # Choose option 2 (BoltDB), use admin/changeme

# 3. Start server
./bin/semaphore server --config ./config.json

# 4. Test in another terminal
curl http://localhost:3000/api/ping  # Should return "pong"
curl -I http://localhost:3000/       # Should return HTTP 200

# 5. Test web interface manually in browser at http://localhost:3000
# 6. Test login with admin/changeme if auth-related changes
```

### Linting and Code Quality

- Frontend linting: `cd web && npm run lint` (has known warnings about console statements and asset sizes - ignore existing issues)
- Backend linting: `golangci-lint run --timeout=3m` (has known type errors due to module import issues - ignore existing issues) 
- **DO NOT** try to fix existing linting issues unless specifically asked to
- Always run linting on new code you add to ensure it follows project standards
- Install golangci-lint if needed: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.2`

## Common Tasks

### Repository Structure
```
.
├── README.md           - Project documentation  
├── CONTRIBUTING.md     - Development guidelines
├── Taskfile.yml       - Task runner configuration
├── go.mod             - Go module dependencies
├── web/               - Vue.js frontend application
│   ├── package.json   - Frontend dependencies
│   ├── src/           - Vue.js source code
│   └── public/        - Static assets
├── cli/               - Go CLI application entry point  
├── api/               - Go API server endpoints
├── db/                - Database models and interfaces
├── services/          - Business logic services
├── util/              - Utility functions and configuration
├── bin/               - Built binaries (after build)
└── config.json       - Runtime configuration (after setup)
```

### Key Commands Reference
```bash
# Install task runner
go install github.com/go-task/task/v3/cmd/task@latest

# Install all dependencies (backend + frontend + tools)
task deps

# Build application (frontend + backend)
task build

# Run tests
task test  

# Run linting
task lint

# Setup application (interactive)
./bin/semaphore setup

# Start server
./bin/semaphore server --config ./config.json

# View available task commands
task --list
```

### Database Options for Development
During setup, choose option 2 (BoltDB) for simplest development setup:
- No external database required
- Database file stored as `database.boltdb` in project root
- Perfect for development and testing

### Frontend Development
- Vue.js 2.x application in `web/` directory
- Built with Vue CLI and Vuetify components
- Build output goes to `api/public/` for serving by Go backend
- Development server not typically used - Go server serves built assets

### Backend Development  
- Go application with CLI and API server
- Uses Gorilla Mux for routing
- Supports multiple databases: MySQL, PostgreSQL, SQLite, BoltDB
- Configuration via JSON file or environment variables

## Troubleshooting

### Build Issues
- If `task` command not found: Install with `go install github.com/go-task/task/v3/cmd/task@latest`
- If Go version errors: Ensure Go 1.21+ is installed
- If npm install fails: Ensure Node.js 16+ is installed
- If build takes too long: This is normal - frontend build can take 60+ seconds

### Runtime Issues  
- If server won't start: Check config.json exists and database is accessible
- If web UI shows errors: Check that frontend build completed successfully in `api/public/`
- If API returns errors: Check server logs for specific error messages

### Database Issues
- For development, always use BoltDB (option 2) during setup
- Database file will be created automatically
- If database errors occur, delete `database.boltdb` and run setup again

## Important Notes

- **NEVER CANCEL** long-running builds or dependency installations
- Set appropriate timeouts: deps (5+ min), build (3+ min), tests (2+ min)  
- The application serves the frontend from the Go backend - no separate frontend server needed
- Configuration is stored in `config.json` after running setup
- Default admin credentials after setup: admin / changeme
- Linting has known issues - focus on not introducing new ones
- Always test changes by running the full application, not just unit tests