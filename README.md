# Devops-DSOLS
### Distributed System with Observability and Load Simulation

A distributed microservices-based retail backend deployed on a local Kubernetes (Minikube) cluster, demonstrating end-to-end DevOps practices: containerized builds, automated CI/CD via Jenkins, Kubernetes orchestration, asynchronous messaging via RabbitMQ, and full observability through Prometheus and Grafana.

**IIIT Bangalore — SPE Final Project**  
Devanshi Bavaria (MT2025041) · Parag Katoch (MT2025082)

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Deployment Workflow](#deployment-workflow)
- [CI/CD Pipeline](#cicd-pipeline)
- [API Reference](#api-reference)
- [Observability](#observability)
- [Load Testing](#load-testing)
- [Secret Management](#secret-management)
- [Configuration Management](#configuration-management)
- [Services & Ports](#services--ports)

---

## Overview

Devops-DSOLS simulates a realistic e-commerce workload across three business domains:

| Domain | Responsibility | Endpoints |
|--------|---------------|-----------|
| **Users** | Registration and retrieval of user accounts | `/api/user` |
| **Products** | Product listing, lookup, and inventory updates | `/api/product` |
| **Orders** | Order creation (multi-item) and listing | `/api/order` |

### Request Flow

```
k6 / External Client
        │
        ▼
  nginx Ingress (:80)          ← path-based routing
        │
   ┌────┴────────────────┐
   ▼         ▼           ▼
user-ctrl  product-ctrl  order-ctrl   ← publish events (fire-and-forget)
   │            │            │
   └────────────┴────────────┘
                │
            RabbitMQ
                │
   ┌────────────┴────────────┐
   ▼            ▼            ▼
user-svc   product-svc   order-svc    ← async consumers → PostgreSQL
                │
           Prometheus ← scrapes all components every 5s
                │
            Grafana  ← real-time dashboards
```

---

## Architecture

### Key Design Decisions

**Single binary, multiple roles** — All six application components (`user-controller`, `user-service`, `product-controller`, etc.) are compiled from a single Go binary and differentiated at runtime via CLI flags:

```bash
./app -type=controller -component=user   -config=config/local.yml
./app -type=service    -component=order  -config=config/local.yml
./app -type=init       -component=user   -config=config/local.yml
```

**Per-domain database isolation** — Each domain has its own dedicated PostgreSQL instance, with connection strings injected via ConfigMap/Secret.

**Async CQRS-inspired pattern** — HTTP controllers validate requests and immediately publish typed events to RabbitMQ, returning HTTP 200 to the client. Domain service workers consume these events and persist data, decoupling read/write latency from database I/O.

### Component Inventory

| Component | Type | Port | Role |
|-----------|------|------|------|
| `user-controller` | Go app | 9000 | HTTP handler for `/api/user`; publishes to RabbitMQ |
| `user-service` | Go app | 9000 | Async consumer; persists user data to PostgreSQL |
| `product-controller` | Go app | 9000 | HTTP handler for `/api/product` |
| `product-service` | Go app | 9000 | Async consumer; persists product/inventory data |
| `order-controller` | Go app | 9000 | HTTP handler for `/api/order` |
| `order-service` | Go app | 9000 | Async consumer; persists order records |
| `*-postgres` | PostgreSQL 15 | 5432 | Dedicated DB per domain (×3) |
| `rabbitmq` | RabbitMQ | 5672 / 15672 / 15692 | Async message broker + Prometheus plugin |
| `prometheus` | Prometheus | 9090 | Metrics scraping, TSDB, remote-write receiver |
| `grafana` | Grafana | 3000 | Metrics visualization |
| `kube-state-metrics` | Helm chart | 8080 | Kubernetes object state metrics |
| `nginx ingress` | K8s Ingress | 80 | API gateway; path-based routing |

---

## Technology Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.26.1 | Application runtime; all microservices |
| PostgreSQL | 15 | Persistent relational storage (per domain) |
| RabbitMQ | management image | Async AMQP message broker + Prometheus plugin |
| Kubernetes | Minikube | Container orchestration (local cluster) |
| Docker | Latest | Image build and Minikube image loading |
| Prometheus | prom/prometheus | Metrics collection and TSDB |
| Grafana | Latest | Dashboard visualization |
| kube-state-metrics | Helm chart | Kubernetes object state exposure |
| nginx | Minikube addon | Ingress controller / API gateway |
| k6 | Latest | Load testing with remote-write to Prometheus |
| Jenkins | Declarative pipeline | CI/CD automation |
| Vault | Dev mode | Secret management |
| Ansible | Latest | Idempotent cluster bootstrap |

---

## Prerequisites

- [Minikube](https://minikube.sigs.k8s.io/docs/start/) (configured with ≥4 CPUs, ≥4096 MB RAM)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Docker](https://docs.docker.com/get-docker/)
- [Helm](https://helm.sh/docs/intro/install/)
- [k6](https://k6.io/docs/get-started/installation/)
- [Ansible](https://docs.ansible.com/ansible/latest/installation_guide/) (for cluster bootstrap)
- [Jenkins](https://www.jenkins.io/doc/book/installing/) with `minikube`, `kubectl`, `docker`, and `k6` in `PATH`
- Docker Hub credentials configured in Jenkins as `dockerhub-credentials`

---

## Getting Started

### Initial Setup (first time)

```bash
chmod +x ./setup.sh ./start.sh ./pipeline.sh
./setup.sh
```

`setup.sh` will:
1. Configure Minikube with 4 CPUs / 4096 MB RAM (optional prompt)
2. Start Minikube with the Docker driver
3. Apply the `devops-dsols` namespace
4. Add the `prometheus-community` Helm repo and install `kube-state-metrics`
5. Enable `metrics-server` and `ingress` addons
6. Initialize Vault and sync secrets as native Kubernetes Secrets
7. Call `pipeline.sh` to build and deploy all components
8. Wait for all pods to reach `Ready` state (300s timeout)

### Subsequent Starts

```bash
sudo ./start.sh    # Starts Minikube, tunnel, port-forwards Prometheus and Grafana
```

### Incremental Updates

```bash
./pipeline.sh      # Rebuild image, reload to Minikube, re-apply YAMLs, rollout restart
```

---

## Deployment Workflow

### Kubernetes Resources

All resources are deployed in the `devops-dsols` namespace.

| Resource | Details |
|----------|---------|
| **Namespace** | `devops-dsols` |
| **ConfigMaps** | `app-config` (DB DSNs, RabbitMQ URL), `prometheus-config` (scrape targets) |
| **PVCs** | 1Gi each for 3× PostgreSQL, Prometheus, Grafana |
| **HPA** | All 6 app deployments: min=1, max=5, CPU target=70% |
| **Ingress** | nginx, path-based routing to controllers |

### Resource Limits

| Workload | CPU Request | CPU Limit | Mem Request | Mem Limit |
|----------|-------------|-----------|-------------|-----------|
| App controllers/services | 10m | 25m | 64Mi | 128Mi |
| PostgreSQL (each) | 200m | 500m | 256Mi | 512Mi |
| RabbitMQ | 300m | 900m | 512Mi | 1024Mi |
| Prometheus | 100m | 512m | 256Mi | 512Mi |
| Grafana | 100m | 500m | 256Mi | 1Gi |

### Grafana Dashboard Import

1. Open Grafana at `http://localhost:3000` (login: `admin` / `admin`)
2. Go to **Connections → Add new connection → Prometheus**, set URL to `http://prometheus:9090`, click **Save & Test**
3. Go to **Dashboards → New → Import**, paste `grafanaDashboard1.json`, select the Prometheus datasource, click **Import**

---

## CI/CD Pipeline

The Jenkins pipeline (`Jenkinsfile`) is triggered automatically on every `git push` via a GitHub webhook.

```
Checkout SCM
     │
Cluster Bootstrap (Ansible)   ← idempotent Minikube + Helm setup
     │
Verify Environment
     │
Build & Push Image             ← paragkatoch/devops-dsols-app:latest → Docker Hub
     │
Load Image to Minikube
     │
Deploy Configs & Infrastructure  ← Postgres, RabbitMQ, Prometheus, Grafana
     │
Deploy Applications             ← RabbitMQ stability gate → rolling restart
     │
Smoke / Load Test               ← k6 (5 VUs, 1 min) — fails if error rate >10% or P95 >1000ms
     │
Post Actions
```

### Quality Gates

The k6 smoke test acts as an automated quality gate on every deploy:

```js
thresholds: {
  http_req_failed: ['rate<0.10'],   // < 10% error rate
  http_req_duration: ['p(95)<1000'] // P95 latency < 1000ms
}
```

If either threshold is breached, the pipeline fails and the deployment is flagged.

### Manual Pipeline (no Jenkins)

```bash
./pipeline.sh
```

---

## API Reference

All endpoints are served via `http://127.0.0.1` (requires `sudo minikube tunnel`).

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/user` | List all users |
| `POST` | `/api/user` | Create a new user |
| `GET` | `/api/product` | List all products |
| `GET` | `/api/product/{id}` | Get a product by ID |
| `POST` | `/api/product/quantity` | Update product inventory |
| `GET` | `/api/order` | List all orders |
| `POST` | `/api/order` | Create a new order |

### Example: Create Order

```json
POST /api/order
{
  "user_id": "1",
  "total_amount": 1000,
  "currency": "INR",
  "items": [
    { "product_id": "1", "quantity": 5 },
    { "product_id": "2", "quantity": 3 }
  ]
}
```

---

## Observability

### Prometheus

- Scrapes all 6 Go apps, RabbitMQ plugin metrics, kube-state-metrics, and cAdvisor every **5 seconds**
- Accepts k6 remote-write metrics via `--web.enable-remote-write-receiver`
- TSDB retention: 2h / 200MB

### Grafana Dashboards

**Dashboard 1 — Distributed Microservices Observability** (`grafanaDashboard1.json`)

Panels: Total RPS, Error Rate %, Queue Backlog, Active VUs, K6 Request Rate, P95/P99 Latency, Requests by Endpoint, Queue Incoming/Outgoing/Lag, Pod CPU/Memory, Deployment Replicas, Orders Completed vs Failed, Inventory Levels.

**Dashboard 2 — K6 Load Testing** (`grafanaDashboard2.json`)

Panels: RPS, K6-RPS, Virtual Users, Latency P95, Queue Total/Outgoing/Incoming, Waiting P99, Replicas.

---

## Load Testing

```bash
k6 run test.js \
  --out experimental-prometheus-rw \
  --env K6_PROMETHEUS_RW_SERVER_URL=http://localhost:<prometheus-port>/api/v1/write
```

### Scenarios

| Scenario | VUs | Duration | Traffic Mix |
|----------|-----|----------|-------------|
| `normal_traffic` | 10 | 5 min | 25% GET user, 50% GET products, 25% GET product by ID |
| `distributed_traffic` | 10 | 5 min | 10% POST user, 40% POST quantity, 40% POST order, 10% GET orders |

The `distributed_traffic` scenario models realistic inventory events: 20% heavy deductions (stockout simulation), 30% heavy restocks (supplier deliveries), 50% normal fluctuation. Order creation intentionally produces ~15% failing requests to generate observable error spikes.

---

## Secret Management

Vault runs in dev mode (single pod, auto-unsealed) as a secure origin for Kubernetes Secrets. Application pods consume native Kubernetes Secrets directly — no sidecar injection required.

```
vault-init.sh (runs once on setup)
    │
    ├── Writes 4 secret paths to Vault KV-v2:
    │     secret/devops-dsols/postgres/order
    │     secret/devops-dsols/postgres/product
    │     secret/devops-dsols/postgres/user
    │     secret/devops-dsols/app
    │
    └── Creates native Kubernetes Secrets:
          order-postgres-creds, product-postgres-creds,
          user-postgres-creds, app-config
```

**Verify secrets:**

```bash
# Browse Vault UI (token: root)
kubectl port-forward -n devops-dsols svc/vault 8200:8200
# open http://localhost:8200

# Read a secret from Vault
kubectl exec -n devops-dsols vault-0 -- vault kv get secret/devops-dsols/postgres/user

# Inspect a Kubernetes Secret
kubectl get secret order-postgres-creds -n devops-dsols \
  -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d
```

> **Note:** Vault runs in dev mode (in-memory, no TLS). Secrets are lost if the `vault-0` pod restarts. Re-run `./k8s/vault-init.sh` to restore them.

---

## Configuration Management

Ansible handles idempotent cluster bootstrap, replacing ad-hoc shell scripts.

```bash
# Standard bootstrap
ansible-playbook -i ansible/inventory.ini ansible/setup.yml

# Reconfigure Minikube resources and reset cluster
ansible-playbook -i ansible/inventory.ini ansible/setup.yml \
  -e "configure_minikube=true reset_minikube=true"

# Skip Helm chart installation
ansible-playbook -i ansible/inventory.ini ansible/setup.yml \
  -e "install_helm_charts=false"
```

Key variables: `configure_minikube` (default: false), `reset_minikube` (false), `install_helm_charts` (true), `minikube_cpus` (4), `minikube_memory` (4096).

---

## Services & Ports

| Service | Port | Access |
|---------|------|--------|
| Ingress / API Gateway | 80 | `http://127.0.0.1` via `minikube tunnel` |
| Prometheus | 9090 | `minikube service prometheus -n devops-dsols` |
| Grafana | 3000 / 30000 | `minikube service grafana -n devops-dsols` |
| PostgreSQL (×3) | 5432 | Internal only (Kubernetes DNS) |
| RabbitMQ AMQP | 5672 | Internal only |
| RabbitMQ Management UI | 15672 | `kubectl port-forward` |
| RabbitMQ Metrics | 15692 | Scraped by Prometheus internally |
| App controllers/services | 9000 | Scraped by Prometheus; routed via Ingress |
| kube-state-metrics | 8080 | Scraped by Prometheus internally |
| Vault UI | 8200 | `kubectl port-forward svc/vault 8200:8200` |

---

## Repository Structure

```
Devops-DSOLS/
├── server/                  # Go application source
│   ├── main.go
│   ├── controllers/         # HTTP controllers (user, product, order)
│   ├── services/            # Async RabbitMQ consumers
│   ├── internal/
│   │   ├── http/handlers/
│   │   ├── storage/postgres/
│   │   └── prometheus/
│   ├── config/
│   ├── types/
│   └── util/
├── k8s/
│   ├── namespace.yaml
│   ├── configmaps.yaml
│   ├── infrastructure/      # Postgres, RabbitMQ, Prometheus, Grafana
│   ├── apps/                # user, product, order deployments + HPA
│   └── vault-init.sh
├── infra/
│   └── Dockerfile           # Two-stage Go build
├── ansible/
│   ├── inventory.ini
│   └── setup.yml
├── grafanaDashboard1.json
├── grafanaDashboard2.json
├── test.js                  # k6 load test script
├── Jenkinsfile
├── setup.sh
├── start.sh
└── pipeline.sh
```

---

## License

This project was developed as part of the Software Production Engineering (SPE) course at IIIT Bangalore, 2026.