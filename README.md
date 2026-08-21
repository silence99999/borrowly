# Borrowly — Rent-Items Microservices Platform

A rental-marketplace application ("Borrowly") built as a set of independent Go
microservices behind an Nginx API gateway, with a React/TypeScript frontend,
per-service PostgreSQL databases, Prometheus/Grafana observability, and
Terraform + Ansible automation for cloud deployment to AWS EC2. CI/CD is
handled by GitHub Actions, publishing images to GHCR and deploying over SSH.

---

## Architecture

```
                                Browser
                                   │
                                   ▼
                     ┌─────────────────────────────┐
                     │   Gateway (Nginx) :8080      │
                     │   serves React SPA + proxies │
                     │   /api/* to services         │
                     └───┬──────┬──────┬──────┬─────┘
                         │      │      │      │
             ┌───────────┘      │      │      └───────────┐
             ▼                  ▼      ▼                   ▼
      auth-service :5001  item-service :5002  rental-service :5003
             │                  │      │                   │
             ▼                  ▼      │                   ▼
        db-auth :5433      db-items :5434                db-rentals :5435
                                        │
                    review-service :5004        chat-service :5005
                            │                          │
                            ▼                          ▼
                    db-reviews :5436             db-chat :5437

  Prometheus :9090  ── scrapes all services + node-exporter
  Grafana    :3000  ── dashboards on top of Prometheus
  pgAdmin    :5050  ── admin UI for all five databases
  node-exporter :9100 ── host-level metrics
```

Each microservice owns its own PostgreSQL database — there is no shared
database and no direct service-to-service DB access. Services talk to each
other only through their public HTTP APIs (e.g. `rental-service` calls
`auth-service` and `item-service` over HTTP).

### Services

| Service | Port | Own DB | Responsibility |
|---|---|---|---|
| `auth-service` | 5001 | db-auth :5433 | Signup/login, JWT issuance, email verification |
| `item-service` | 5002 | db-items :5434 | Item CRUD, pickup points, image upload |
| `rental-service` | 5003 | db-rentals :5435 | Rental orders, pricing, reminder worker (background goroutine), emails |
| `review-service` | 5004 | db-reviews :5436 | Item reviews |
| `chat-service` | 5005 | db-chat :5437 | User-to-user messaging |
| `gateway` | 8080 | — | Nginx reverse proxy + React SPA hosting |

### Support stack

| Component | Port | Purpose |
|---|---|---|
| Prometheus | 9090 | Metrics scraping + alert evaluation (`observability/prometheus`) |
| Grafana | 3000 | Dashboards, auto-provisioned (admin/admin) |
| node-exporter | 9100 | Host CPU/memory metrics for Prometheus |
| pgAdmin | 5050 | Web UI to inspect all 5 Postgres databases (admin@admin.com/admin) |

---

## Repository layout

```
final/
├── services/
│   ├── auth-service/       Go microservice — auth, JWT, email verification
│   ├── item-service/       Go microservice — items, pickup points, uploads
│   ├── rental-service/     Go microservice — rentals, pricing, reminder worker
│   ├── review-service/     Go microservice — item reviews
│   └── chat-service/       Go microservice — user-to-user chat
│       each service/ contains: main.go, api.go, handlers.go, middleware.go,
│       models.go, Dockerfile, migrations/ (SQL, applied via docker-entrypoint-initdb.d)
├── gateway/                 Nginx config + Dockerfile (routes /api/*, serves SPA)
├── frontend/                React 18 + TypeScript + Vite SPA
├── backend/                 Original Go monolith (kept for reference/history)
├── observability/
│   ├── prometheus/          prometheus.yml + alert_rules.yml
│   ├── grafana/provisioning/ datasources + auto-provisioned dashboard JSON
│   └── pgadmin/              pgAdmin server list (servers.json)
├── terraform/                AWS IaC: EC2 instance + security group (main.tf, variables.tf, outputs.tf)
├── ansible/                  Playbooks: docker / app / monitoring roles, inventory.ini
├── docs/                     Assignment reports (incident response, Terraform, automation, end-term report)
├── load-test.js              k6 load-test script (constant-arrival-rate scenario)
├── docker-compose.yaml       Full local/production stack definition
├── .env.example              Template for required environment variables
└── .github/workflows/        CI (test/vet/build) and CD (build+push to GHCR, deploy via SSH)
```

---

## Prerequisites

- Docker Desktop ≥ 24 with Compose v2
- Go 1.24+ (only needed for local service development / running tests)
- Node.js 20 (only needed for local frontend development)
- Terraform ≥ 1.5 and an AWS account (only needed for cloud provisioning)
- Ansible (only needed for provisioning an existing host)
- [k6](https://k6.io/) (only needed to run `load-test.js`)

---

## Quick start (local, Docker Compose)

```bash
git clone <repo-url>
cd final
cp .env.example .env   # fill in real secrets (SMTP credentials, JWT secret, DB creds)

docker compose up --build
```

Compose brings up, in dependency order: five per-service Postgres databases →
pgAdmin → the five Go microservices → the gateway → Prometheus/Grafana/node-exporter.
Each service's SQL migrations are mounted into its database's
`docker-entrypoint-initdb.d` and applied automatically on first boot.

| URL | Description |
|---|---|
| http://localhost:8080 | Web app (React SPA + API via gateway) |
| http://localhost:9090 | Prometheus |
| http://localhost:3000 | Grafana (admin/admin) |
| http://localhost:5050 | pgAdmin (admin@admin.com/admin) |

To tear down and reset all data:

```bash
docker compose down -v
```

---

## Environment variables

See `.env.example` for the full list. Each service needs its own DB
credentials/DSN, plus shared configuration:

```
AUTH_DB_USER / AUTH_DB_PASSWORD / AUTH_DB_NAME / AUTH_DB_DSN
ITEM_DB_USER / ITEM_DB_PASSWORD / ITEM_DB_NAME / ITEM_DB_DSN
RENTAL_DB_USER / RENTAL_DB_PASSWORD / RENTAL_DB_NAME / RENTAL_DB_DSN
REVIEW_DB_USER / REVIEW_DB_PASSWORD / REVIEW_DB_NAME / REVIEW_DB_DSN
CHAT_DB_USER / CHAT_DB_PASSWORD / CHAT_DB_NAME / CHAT_DB_DSN

ITEM_SERVICE_URL / AUTH_SERVICE_URL   # inter-service base URLs
SECRET                                # JWT signing secret

SMTP_HOST / SMTP_PORT / SMTP_USERNAME / SMTP_PASSWORD / SMTP_SENDER
```

`IMAGE_REGISTRY` and `IMAGE_TAG` (used by `docker-compose.yaml`) default to
`local`/`latest` for local builds and are overridden in CD to point at GHCR.

---

## Running a single service locally (outside Docker)

```bash
cd services/auth-service
export DB_DSN="postgres://auth_user:auth_pass@localhost:5433/auth_db?sslmode=disable"
export SECRET="your-jwt-secret"
export SMTP_HOST=sandbox.smtp.mailtrap.io
export SMTP_PORT=2525
export SMTP_USERNAME=...
export SMTP_PASSWORD=...
export SMTP_SENDER="rent_item <no-reply@rent_item.com>"
go run .
```

Start the matching `db-*` container from `docker-compose.yaml` first (e.g.
`docker compose up -d db-auth`) if you're not running the whole stack.

---

## Frontend development

```bash
cd frontend
npm install
npm run dev        # Vite dev server
npm run build       # tsc -b && vite build — production bundle used by the gateway image
```

---

## API routes (via gateway at :8080)

All routes are exposed under the `/api` prefix; the gateway strips it before
proxying to the owning service.

### Auth (`auth-service`)
```
POST /api/auth/signup
POST /api/auth/login
POST /api/auth/verify-email
POST /api/auth/logout
GET  /api/auth/me
```

### Items (`item-service`)
```
GET    /api/items
GET    /api/items/{id}
POST   /api/items                    (auth)
PATCH  /api/items/{id}               (auth, owner/admin)
DELETE /api/items/{id}               (auth, owner/admin)
POST   /api/items/{id}/images        (auth)
GET    /api/pickup-points
POST   /api/pickup-points            (auth, admin)
```

### Rentals (`rental-service`)
```
POST  /api/rentals                   (auth)
GET   /api/rentals/my                (auth)
PATCH /api/rental/{id}/cancel        (auth)
PATCH /api/rental/{id}/pay           (auth)
```

### Reviews (`review-service`)
```
GET    /api/items/{id}/reviews
POST   /api/items/{id}/reviews              (auth)
PATCH  /api/items/{id}/reviews/{reviewID}   (auth)
DELETE /api/items/{id}/reviews/{reviewID}   (auth)
```

### Chat (`chat-service`)
```
GET  /api/chats             (auth) — list conversations
GET  /api/chats/{userID}    (auth) — messages with a user
POST /api/chats/{userID}    (auth) — send a message
```

---

## Observability

- **Prometheus** (`observability/prometheus/prometheus.yml`) scrapes `/metrics`
  on every microservice plus `node-exporter` and itself.
- **Alert rules** (`observability/prometheus/alert_rules.yml`) cover:
  - `RentalServiceDown` / `AuthServiceDown` / `ItemServiceDown` — service unreachable
  - `RentalCreationFailureHigh` — >10% of rental creations failing over 5 min
  - `RentalHighRequestLatency` — P95 latency over 1 s for 5 min
  - `HighCPUUsage` / `HighMemoryUsage` — host resource pressure
- **Grafana** dashboards are auto-provisioned from
  `observability/grafana/provisioning/` (datasource + dashboard JSON) — no
  manual setup needed after `docker compose up`.

### Incident simulation

```bash
# 1. Kill rental-service's DB connection to simulate an outage
docker compose stop rental-service
DB_DSN=postgres://wrong:wrong@wrong-host:5432/wrong docker compose up -d rental-service

# 2. Watch Prometheus fire RentalServiceDown / Grafana panels go red

# 3. Restore
docker compose stop rental-service
docker compose up -d rental-service
```

See `docs/assignment4-incident-report.md` for the full writeup and postmortem.

---

## Load testing

`load-test.js` is a [k6](https://k6.io/) script that drives a
constant-arrival-rate scenario (10,000 req/s target, 1 minute, up to 10,000
VUs) against a deployed gateway:

```bash
k6 run load-test.js
```

Edit the target URL inside the script before pointing it at your own
deployment — do not run this against infrastructure you don't own or without
authorization.

---

## Infrastructure as code (Terraform)

`terraform/` provisions a single AWS EC2 instance for the whole stack:

- Security group opening 22 (SSH, split between an admin CIDR and the
  GitHub Actions IP ranges used for CD), 80/8080 (gateway), 3000 (Grafana),
  9090 (Prometheus)
- `t3.small` Ubuntu 22.04 instance with a 20 GB gp3 root volume

```bash
cd terraform
terraform init
terraform plan
terraform apply
```

Key outputs: `instance_public_ip`, `app_url`, `grafana_url`,
`prometheus_url`, `ssh_command`. See `docs/assignment5-terraform.md` for the
full writeup.

---

## Configuration management (Ansible)

`ansible/site.yml` targets the `app_servers` group (`ansible/inventory.ini`)
and applies three roles in order:

1. `docker` — installs Docker + Compose plugin
2. `app` — clones/updates the repo and starts `docker-compose.yaml`
3. `monitoring` — ensures Prometheus/Grafana/node-exporter are healthy

```bash
ansible-playbook -i ansible/inventory.ini ansible/site.yml
```

The playbook finishes with a post-task that runs `docker compose ps` on the
remote host and prints container status for verification.

---

## CI/CD (GitHub Actions)

- **CI** (`.github/workflows/ci.yml`) — on every push/PR:
  - `go vet` + `go test` for each of the 5 services (matrix build)
  - Frontend type-check (`tsc --noEmit`) and production build
  - Docker image build (no push) for all 5 services + gateway, using GHA
    layer caching
- **CD** (`.github/workflows/cd.yml`) — on push to `main`:
  - Builds and pushes all 6 images to `ghcr.io/<owner>/rent-items-<service>`
    (tagged by commit SHA and `latest`)
  - SSHes into the production EC2 host, pulls the new images, and runs
    `docker compose up -d --remove-orphans`, then prunes dangling images

Required repository secrets: `GHCR_PUSH_TOKEN`, `EC2_HOST`, `EC2_USER`,
`EC2_SSH_PRIVATE_KEY`, `GHCR_TOKEN`.

---

## Documentation

Detailed assignment reports live in `docs/`:

- `assignment4-incident-report.md` — incident simulation & postmortem
- `assignment5-terraform.md` — Terraform provisioning writeup
- `assignment6-report.md` / `assignment6-automation-capacity.md` — Ansible automation & capacity planning
- `end-term-report.md` — full end-of-term project report

---

## License

Student project — no license specified.
