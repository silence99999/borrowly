# Rent-Items — Containerised Microservices Platform

A production-grade rental marketplace demonstrating microservices architecture, Docker-based
orchestration, Prometheus/Grafana observability, and Terraform infrastructure provisioning.

---

## Architecture

```
Browser
  │
  ▼
┌─────────────────────────────┐   port 8080
│  API Gateway (Nginx)        │◄──────────────
│  + React SPA static files   │
└──┬──────┬────┬──────┬───┬───┘
   │      │    │      │   │
   ▼      ▼    ▼      ▼   ▼
 Auth   Item Rental Review Chat
 :5001 :5002  :5003  :5004 :5005
   │      │    │      │   │
   └──────┴────┴──────┴───┘
              │
              ▼
         PostgreSQL :5432

Prometheus :9090  ←  scrapes all services
Grafana    :3000  ←  visualises Prometheus data
```

### Microservices

| Service | Port | Responsibility |
|---|---|---|
| `auth-service` | 5001 | Registration, login, JWT, email verification |
| `item-service` | 5002 | Item CRUD, pickup points, image upload |
| `rental-service` | 5003 | Rental orders, pricing, reminder worker |
| `review-service` | 5004 | Item reviews |
| `chat-service` | 5005 | User-to-user messaging (new feature) |
| `gateway` | 8080 | Nginx reverse proxy + SPA serving |

---

## Prerequisites

- Docker Desktop ≥ 24 with Compose v2
- Node.js 20 (only needed for local frontend dev)
- Go 1.22+ (only needed for local service dev)
- Terraform ≥ 1.5 (for cloud provisioning)

---

## Quick Start

### 1. Clone & configure

```bash
git clone <repo-url>
cd final
cp .env.example .env   # edit credentials if needed
```

### 2. Run all database migrations

```bash
# Install goose if not already installed
go install github.com/pressly/goose/v3/cmd/goose@latest

# Apply all migrations (runs against local DB via goose env vars in .env)
goose up
```

### 3. Build and start all services

```bash
docker compose up --build
```

Services start in this order: `db` → microservices → `gateway` + observability.

### 4. Access the application

| URL | Description |
|---|---|
| http://localhost:8080 | Web UI |
| http://localhost:9090 | Prometheus |
| http://localhost:3000 | Grafana (admin/admin) |

---

## Development — Running a Single Service Locally

```bash
cd services/auth-service
export DB_DSN="postgres://rent_item:pass@localhost:5434/rent_item?sslmode=disable"
export SECRET="your-jwt-secret"
export SMTP_HOST=sandbox.smtp.mailtrap.io
export SMTP_PORT=2525
export SMTP_USERNAME=...
export SMTP_PASSWORD=...
export SMTP_SENDER="rent_item <no-reply@rent_item.com>"
go run .
```

---

## Terraform — Cloud Provisioning

See `terraform/` directory and `docs/assignment5-terraform.md` for full details.

```bash
cd terraform
terraform init
terraform plan
terraform apply
```

Outputs: public IP, SSH command, Grafana URL, Prometheus URL.

---

## Incident Simulation

See `docs/assignment4-incident-report.md` for the full incident report and postmortem.

**Quick reproduction:**

```bash
# 1. Introduce the fault — bad DB DSN for rental-service only
docker compose stop rental-service
DB_DSN=postgres://wrong:wrong@wrong-host:5432/wrong docker compose up -d rental-service

# 2. Observe in Grafana / Prometheus
# 3. Restore
docker compose stop rental-service
docker compose up -d rental-service
```

---

## API Routes (via Gateway at :8080)

All API calls use the `/api` prefix.

### Authentication
```
POST /api/auth/signup
POST /api/auth/login
POST /api/auth/verify-email
POST /api/auth/logout
GET  /api/auth/me
```

### Items
```
GET    /api/items
GET    /api/items/{id}
POST   /api/items           (auth)
PATCH  /api/items/{id}      (auth, owner/admin)
DELETE /api/items/{id}      (auth, owner/admin)
POST   /api/items/{id}/images (auth)
GET    /api/pickup-points
POST   /api/pickup-points   (auth, admin)
```

### Rentals
```
POST  /api/rentals          (auth)
GET   /api/rentals/my       (auth)
PATCH /api/rental/{id}/cancel (auth)
PATCH /api/rental/{id}/pay    (auth)
```

### Reviews
```
GET    /api/items/{id}/reviews
POST   /api/items/{id}/reviews              (auth)
PATCH  /api/items/{id}/reviews/{reviewID}   (auth)
DELETE /api/items/{id}/reviews/{reviewID}   (auth)
```

### Chat
```
GET  /api/chats             (auth) — list conversations
GET  /api/chats/{userID}    (auth) — messages with a user
POST /api/chats/{userID}    (auth) — send a message
```

---

## Directory Structure

```
final/
├── services/
│   ├── auth-service/       Auth microservice (Go)
│   ├── item-service/       Item/Product microservice (Go)
│   ├── rental-service/     Rental/Order microservice (Go)
│   ├── review-service/     Review microservice (Go)
│   └── chat-service/       Chat microservice (Go, new feature)
├── gateway/                Nginx API gateway + Dockerfile
├── terraform/              AWS IaC (main.tf, variables.tf, outputs.tf)
├── backend/                Original Go monolith (reference)
│   └── migrations/         SQL migrations (run with goose)
├── frontend/               React + TypeScript SPA
├── observability/
│   ├── prometheus/         prometheus.yml + alert_rules.yml
│   └── grafana/provisioning/ datasources + auto-provisioned dashboard
├── docs/
│   ├── assignment4-incident-report.md
│   └── assignment5-terraform.md
└── docker-compose.yaml
```
