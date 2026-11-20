# Stats Central API

Stats Central API is a **REST API Backend** built with **Go + Gin Gonic** focused on **football match statistics** for analysis and sports betting purposes. The domain includes leagues, teams, matches, rounds, statistics, and analysis methods designed to feed AI models or automated bots.

## 🚀 Overview

The project provides a comprehensive API for:
- Tournament and season management
- Team registration and lookup
- Match listing with odds and pagination
- Statistical analysis endpoints for AI/bot consumption

## 🏗 Architecture

This project follows a **Feature-Based Clean Architecture** pattern. Each feature is implemented as an independent module within the `internal/` directory, ensuring high cohesion and low coupling.

### Key Architectural Pillars:
- **Dependency Injection**: All dependencies are managed and injected via a centralized DI container.
- **Interface Segregation**: Strict use of interfaces for service and repository layers.
- **Repository Pattern**: Abstraction of data access logic.
- **Structured Logging**: Using Go's native `slog` for consistent observability.

## 📂 Directory Structure

```text
├── cmd/
│   └── api/                # Application entry point
├── docs/                   # Documentation and samples
├── internal/
│   ├── analyse/            # Analysis feature module
│   ├── di/                 # Dependency Injection container
│   ├── match/              # Match feature module
│   ├── team/               # Team feature module
│   ├── tournament/         # Tournament feature module
│   └── shared/             # Shared utilities (config, logger, middleware)
├── install/                # Production deployment scripts (install.sh, update.sh)
├── migrations/             # Database schema migrations
├── Makefile                # Development automation commands
└── .env.sample             # Environment variables template
```

## 🛠 Prerequisites

Ensure you have the following installed:
- **Go**: 1.23+
- **MySQL**: 8.0+ (running locally or via Docker)
- **Make**: For running automation commands
- **golang-migrate**: For database migrations
- **reflex** (optional): For hot-reload during development
- **golangci-lint** (optional): For linting

Install development tools:
```bash
make install-tools
```

## ⚙️ Installation & Setup (Development)

### 1. Clone the repository
```bash
git clone <repository-url>
cd central-estatisticas/backend
```

### 2. Configure Environment Variables
Copy the sample environment file and adjust the values:
```bash
cp .env.sample .env
```

Default development values:
```env
SERVER_PORT=3030
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3301
DB_USER=betnator
DB_PASSWORD=
DB_NAME=betnator
LOG_LEVEL=error
```

### 3. Install Dependencies
```bash
make install-deps
```

### 4. Database Setup
Ensure MySQL is running and accessible with the credentials in `.env`, then run migrations:
```bash
make migrate
```

### 5. Run the Application

**Hot-reload mode (development):**
```bash
make run-dev
```
Server starts at `http://localhost:3030` with automatic reload on file changes.

**Compiled binary:**
```bash
make run
```

## 🗄 Database Migrations

Migrations are managed using `golang-migrate`.

- **Run migrations**: `make migrate`
- **Rollback last migration**: `make rollback`
- **Create new migration**: `make migrate-create name=<migration_name>`

## 🧪 Testing

- **Run all tests**: `make test`
- **Run tests with verbose output**: `make test-v`
- **Check test coverage**: `make test-cover`

## 📡 API Endpoints

### Health Check
- **URL**: `GET /health`
- **Description**: Verifies if the API and its dependencies are healthy.

### Tournaments
- `GET    /v1/tournaments` — List tournaments
- `POST   /v1/tournaments` — Create tournament
- `GET    /v1/tournaments/:tournamentId` — Tournament details
- `PUT    /v1/tournaments/:tournamentId` — Update tournament
- `PATCH  /v1/tournaments/:tournamentId/status` — Update tournament status

### Teams
- `GET    /v1/teams` — List teams
- `GET    /v1/teams/:teamId` — Team details

### Analysis
- `GET    /v1/teams/:teamId/analysis` — Team tournament analysis

### Matches
- `GET    /v1/matches` — List matches with filters and pagination
  - Query params: `tournament_id`, `season`, `round`, `status`, `home_team_id`, `away_team_id`, `page`, `page_size`

## 📜 Available Commands

| Command | Description |
|---------|-------------|
| `make run` | Build and run the API locally |
| `make run-dev` | Run with hot-reload (reflex) |
| `make build` | Build optimized binary (`bin/stats-central-api`) |
| `make test` | Run tests |
| `make test-v` | Run tests with verbose output |
| `make test-cover` | Generate coverage report |
| `make lint` | Run linter (golangci-lint) |
| `make install-deps` | Install Go dependencies |
| `make install-tools` | Install dev tools (reflex, golangci-lint) |
| `make migrate` | Run database migrations |
| `make rollback` | Rollback last migration |
| `make migrate-create name=<name>` | Create new migration |
| `make clean` | Remove binaries and coverage files |

---

## 🏭 Production Deployment (VPS)

The application is deployed to a Linux VPS using a **systemd service** to manage the Go binary directly (no Docker). MySQL must already be running and configured on the server.

### Architecture

```
Internet → <vps-ip>:3030 → Go binary (systemd)
                               ↓
                        MySQL (127.0.0.1:3301)
```

### Prerequisites on VPS
- Linux (Ubuntu 22.04+ / Debian 12+) with systemd
- MySQL 8.0+ running and accessible
- sudo access

### Installation Scripts

Two scripts in `install/` automate deployment:

| Script | Purpose |
|--------|---------|
| `install/install.sh` | First-time setup: creates `/opt/stats-central-api`, installs systemd unit, enables service |
| `install/update.sh`  | Deploy new versions: replaces binary and restarts service |

### First-time Install

**1. Build the binary locally:**
```bash
make build
```

**2. Copy binary, scripts, and `.env` to the VPS:**
```bash
scp bin/stats-central-api install/install.sh install/update.sh root@<vps>:/tmp/
scp .env root@<vps>:/tmp/.env.production
```

**3. On the VPS — prepare files:**
```bash
sudo install -d -o www-data -g www-data -m 0755 /opt/stats-central-api
sudo install -m 0755 -o www-data -g www-data /tmp/stats-central-api /opt/stats-central-api/stats-central-api
sudo install -m 0600 -o www-data -g www-data /tmp/.env.production /opt/stats-central-api/.env
# Adjust values in /opt/stats-central-api/.env for production (DB_HOST, DB_PASSWORD, LOG_LEVEL=warn)
```

**4. Run the installer:**
```bash
sudo bash /tmp/install.sh
```

This creates the `stats-central-api.service` systemd unit (with hardening), enables it, and starts the service.

### Updating to a New Version

**1. Build locally:**
```bash
make build
```

**2. Copy the new binary to the VPS:**
```bash
scp bin/stats-central-api root@<vps>:/opt/stats-central-api/stats-central-api
```

**3. Run the update script on the VPS:**
```bash
sudo bash /tmp/update.sh
```

The update script fixes permissions and restarts the service — no need to touch systemd or `.env`.

### Service Management

```bash
systemctl status stats-central-api                      # Check status
journalctl -u stats-central-api -f                      # Tail logs
systemctl restart stats-central-api                     # Manual restart
systemctl stop stats-central-api                        # Stop
sudo journalctl -u stats-central-api -n 50 --no-pager   # Service Logs
```
