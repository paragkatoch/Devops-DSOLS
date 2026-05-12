# Devops-DSOLS

A Go microservices system deployed on Kubernetes (Minikube) with full observability via Prometheus & Grafana and load testing via k6.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Step-by-Step Run Guide](#step-by-step-run-guide)
  - [1. Start Minikube](#1-start-minikube)
  - [2. Enable Metrics Addon](#2-enable-metrics-addon)
  - [3. Start Kubernetes Dashboard](#3-start-kubernetes-dashboard)
  - [4. Deploy Everything (script.sh)](#4-deploy-everything-scriptsh)
  - [5. Open Required Terminals](#5-open-required-terminals)
  - [6. Configure Grafana](#6-configure-grafana)
  - [7. Import Grafana Dashboard](#7-import-grafana-dashboard)
  - [8. Verify Deployments](#8-verify-deployments)
  - [9. Run k6 Load Test](#9-run-k6-load-test)
- [Services & Ports](#services--ports)
- [Kubernetes Resources](#kubernetes-resources)
- [CI/CD Pipeline (Jenkins)](#cicd-pipeline-jenkins)
- [Docker Compose (Local Dev Alternative)](#docker-compose-local-dev-alternative)

---

## Architecture Overview

```
                          ┌─────────────────────────────────────────────────┐
                          │              Minikube Cluster                   │
                          │           Namespace: devops-dsols               │
                          │                                                 │
  k6 Load Test ──────────▶│  Ingress (nginx)                               │
  (127.0.0.1)             │    /api/user    ──▶ user-controller :9000      │
                          │    /api/product ──▶ product-controller :9000   │
                          │                                                 │
                          │  user-controller ──[RabbitMQ]──▶ user-service  │
                          │  product-controller ──[RabbitMQ]──▶ product-service│
                          │                          │                      │
                          │                       postgres                  │
                          │                                                 │
                          │  prometheus :9090  ◀── scrapes all services    │
                          │  grafana    :3000  ◀── reads prometheus        │
                          └─────────────────────────────────────────────────┘
```

**Component Roles:**

| Component | Type | Role |
|---|---|---|
| `user-controller` | Go app | HTTP handler for `/api/user` routes |
| `user-service` | Go app | Business logic consumer for user events |
| `product-controller` | Go app | HTTP handler for `/api/product` routes |
| `product-service` | Go app | Business logic consumer for product events |
| `postgres` | PostgreSQL 15 | Primary datastore (`retail` DB) |
| `rabbitmq` | RabbitMQ + management | Async message queue between controllers and services |
| `prometheus` | Prometheus | Metrics scraping & remote-write receiver |
| `grafana` | Grafana | Metrics visualization |
| `nginx ingress` | Kubernetes Ingress | API gateway routing to services |

---

## Project Structure

```
Devops-DSOLS/
├── infra/
│   ├── Dockerfile          # Multi-stage Go builder → debian:slim runner
│   ├── nginx.conf          # Nginx config (Docker Compose only)
│   └── prometheus.yml      # Prometheus scrape config (Docker Compose only)
├── k8s/
│   ├── namespace.yaml      # Creates 'devops-dsols' namespace
│   ├── configmaps.yaml     # app-config + prometheus-config ConfigMaps
│   ├── infrastructure/
│   │   ├── postgres.yaml   # Postgres Deployment + PVC + Service
│   │   ├── rabbitmq.yaml   # RabbitMQ Deployment + Service (AMQP/UI/metrics)
│   │   ├── prometheus.yaml # Prometheus Deployment + PVC + Service
│   │   └── grafana.yaml    # Grafana Deployment + PVC + NodePort Service
│   └── apps/
│       ├── user.yaml            # user-controller + user-service Deployments & Services
│       ├── product.yaml         # product-controller + product-service Deployments & Services
│       ├── ingress.yaml         # Nginx Ingress routing /api/user and /api/product
│       └── postgres-setup-job.yaml  # One-time DB init Job
├── server/                 # Go source code
│   ├── main.go
│   ├── cmd/
│   ├── config/
│   ├── controllers/
│   ├── services/
│   ├── internal/
│   ├── types/
│   └── util/
├── script.sh               # One-shot build + deploy script
├── test.js                 # k6 load test script
├── grafanaDashboard.json   # Pre-built Grafana dashboard
├── docker-compose.yml      # Local dev alternative (no Kubernetes)
└── Jenkinsfile             # Jenkins CI/CD pipeline definition
```

---

## Prerequisites

Make sure the following are installed before starting:

| Tool | Purpose | Install |
|---|---|---|
| `minikube` | Local Kubernetes cluster | `brew install minikube` |
| `kubectl` | Kubernetes CLI | `brew install kubectl` |
| `docker` | Build app image | [docker.com](https://docker.com) |
| `k6` | Load testing | `brew install k6` |

---

## Step-by-Step Run Guide

### 1. Start Minikube

First, configure Minikube to use enough CPU and memory resources, then start fresh:

```bash
# Set resource limits (applied on next start)
# This step is one time only , if already project is setuped skip it
minikube config set cpus 4
minikube config set memory 4096

# Delete any existing cluster to apply the new config cleanly
# This step is one time only , if already project is setuped skip it
minikube delete

# Start Minikube with the Docker driver
minikube start --driver=docker
```

> **Why delete first?** `minikube config set` only takes effect when a new cluster is created. Deleting the old cluster ensures the CPU/memory settings are applied correctly.

Verify it's running:

```bash
minikube status
```

---

### 2. Enable Metrics Addon

Enable the Kubernetes metrics server so the dashboard shows CPU/memory:

```bash
minikube addons enable metrics-server
```

---

### 3. Start Kubernetes Dashboard

Open the Minikube dashboard in your browser to monitor pods and deployments:

```bash
minikube dashboard
```

> This will open a browser tab automatically. Keep this terminal running.

---

### 4. Deploy Everything (script.sh)

Make the deployment script executable, then run it:

```bash
chmod +x ./script.sh
./script.sh
```

**What `script.sh` does, step by step:**

1. Verifies Minikube is running (starts it if not)
2. Enables the `ingress` addon
3. Builds the Go app Docker image: `devops-dsols-app:latest`
4. Loads the image directly into Minikube's internal Docker daemon
5. Applies the `devops-dsols` namespace
6. Applies ConfigMaps (`app-config` + `prometheus-config`)
7. Applies all infrastructure: Postgres, RabbitMQ, Prometheus, Grafana
8. Applies all app deployments: user-controller, user-service, product-controller, product-service, Ingress, postgres-setup Job
9. Rolls out a restart to ensure latest image is active
10. Prints all pod statuses

**Verify pods are running:**

```bash
kubectl get pods -n devops-dsols
```

Wait until all pods show `Running` or `Completed` (the postgres-setup Job will be `Completed`).

---

### 5. Open Required Terminals

Open **3 separate terminal tabs/windows** and run one command in each:

#### Terminal 1 — Minikube Tunnel (required for Ingress/LoadBalancer)

```bash
sudo minikube tunnel
```

> This exposes Kubernetes services to `127.0.0.1`. Keep this running. It may prompt for your sudo password.

#### Terminal 2 — Expose Prometheus

```bash
minikube service prometheus -n devops-dsols
```

> Minikube will print and open a local URL like `http://127.0.0.1:<PORT>`.  
> **Note the port** — you will need it to run k6 with remote-write.

#### Terminal 3 — Expose Grafana

```bash
minikube service grafana -n devops-dsols
```

> Minikube will print and open the Grafana URL (e.g. `http://127.0.0.1:<PORT>`).  
> This is the URL you'll use to access the Grafana UI.

---

### 6. Configure Grafana

1. Open the Grafana URL from **Terminal 3** in your browser.
2. **Login** with:
   - Username: `admin`
   - Password: `admin`
3. When prompted to change the password, click **Skip**.

#### Add Prometheus as a Data Source

4. In the left sidebar, go to **Connections → Add new connection**.
5. Search for and select **Prometheus**.
6. Click **Add new data source**.
7. Set the **Connection URL** to:
   ```
   http://prometheus:9090
   ```
8. Scroll down and click **Save & Test**.
   - You should see a green ✅ `Successfully queried the Prometheus API.`

---

### 7. Import Grafana Dashboard

1. In the left sidebar, go to **Dashboards**.
2. Click **New → Import**.
3. Open `grafanaDashboard.json` from this repo and copy its entire contents.
4. Paste the JSON into the **Import via dashboard JSON model** text box.
5. Click **Load**.
6. Under **Prometheus**, select the Prometheus data source you just added.
7. Click **Import**.

The dashboard will load and start showing metrics as soon as traffic hits the services.

---

### 8. Verify Deployments

In the Kubernetes dashboard (from Step 3), verify there are **8 deployments** in the `devops-dsols` namespace:

| # | Deployment |
|---|---|
| 1 | `user-controller` |
| 2 | `user-service` |
| 3 | `product-controller` |
| 4 | `product-service` |
| 5 | `postgres` |
| 6 | `rabbitmq` |
| 7 | `prometheus` |
| 8 | `grafana` |

> The `postgres-setup` is a **Job**, not a Deployment — it will show as `Completed`.

---

### 9. Run k6 Load Test

Use the port shown in **Terminal 2** (the Prometheus service URL) for the remote-write endpoint.

```bash
K6_PROMETHEUS_RW_SERVER_URL=http://127.0.0.1:<PROMETHEUS_PORT>/api/v1/write \
k6 run --out experimental-prometheus-rw test.js
```

**Example** (replace `63964` with your actual port from Terminal 2):

```bash
K6_PROMETHEUS_RW_SERVER_URL=http://127.0.0.1:63964/api/v1/write \
k6 run --out experimental-prometheus-rw test.js
```

**What `test.js` does:**

The k6 script sends load in 3 stages against `http://127.0.0.1` (via the Minikube tunnel + Ingress):

| Stage | Duration | Target VUs |
|---|---|---|
| Stage 1 | 1 minute | 40 |
| Stage 2 | 1 minute | 45 |
| Stage 3 | 1 minute | 50 |

It randomly sends:
- **POST** `http://127.0.0.1/api/product/quantity` — updates product stock
- **GET** `http://127.0.0.1/api/product` — fetches all products

After the test runs, **check the Grafana dashboard** for live metrics (request rate, latency, VU count, etc.).

---

## Services & Ports

| Service | Internal Port | Access Method |
|---|---|---|
| Ingress / API Gateway | `80` | `http://127.0.0.1` (via `minikube tunnel`) |
| Prometheus | `9090` | `minikube service prometheus -n devops-dsols` |
| Grafana | `3000` / NodePort `30000` | `minikube service grafana -n devops-dsols` |
| Postgres | `5432` | Internal only |
| RabbitMQ AMQP | `5672` | Internal only |
| RabbitMQ UI | `15672` | Internal only |
| RabbitMQ Metrics | `15692` | Scraped by Prometheus |
| App services/controllers | `9000` | Scraped by Prometheus |

---

## Kubernetes Resources

### Namespace

All resources live in the `devops-dsols` namespace:

```bash
kubectl get all -n devops-dsols
```

### ConfigMaps

| ConfigMap | Purpose |
|---|---|
| `app-config` | App config: DB URL, RabbitMQ URL, HTTP listen address |
| `prometheus-config` | Prometheus scrape targets |

### Persistent Volume Claims

| PVC | Size | Used By |
|---|---|---|
| `postgres-pvc` | 1Gi | Postgres data |
| `prometheus-pvc` | 1Gi | Prometheus TSDB (2h retention, 200MB cap) |
| `grafana-pvc` | 1Gi | Grafana dashboards & settings |

### Resource Limits Summary

| Workload | CPU Request | CPU Limit | Memory Request | Memory Limit |
|---|---|---|---|---|
| `user/product-controller/service` | 10m | 50m | 128Mi | 256Mi |
| `postgres` | 200m | 500m | 256Mi | 512Mi |
| `rabbitmq` | 250m | 500m | 512Mi | 1024Mi |
| `prometheus` | 100m | 512m | 256Mi | 512Mi |
| `grafana` | 100m | 500m | 256Mi | 1Gi |

---

## CI/CD Pipeline (Jenkins)

The `Jenkinsfile` defines a declarative pipeline that mirrors `script.sh` for automated CI/CD:

| Stage | Action |
|---|---|
| Verify Environment | `minikube status \|\| minikube start` + enable ingress addon |
| Build Docker Image | `docker build -t devops-dsols-app:latest -f infra/Dockerfile .` |
| Load Image to Minikube | `minikube image load devops-dsols-app:latest` |
| Deploy Configs & Infrastructure | Apply namespace, ConfigMaps, and infrastructure YAMLs |
| Deploy Applications | Apply app YAMLs + `kubectl rollout restart` |

**Requirements for Jenkins:**
- Jenkins agent must have `minikube`, `kubectl`, and `docker` in PATH.
- PATH is configured in the `environment` block: `/opt/homebrew/bin:/usr/local/bin`.

---

## Docker Compose (Local Dev Alternative)

If you want to run without Kubernetes, use Docker Compose for a quick local setup:

```bash
docker compose up --build
```

| Service | Port |
|---|---|
| API Gateway (nginx) | `9000` |
| Prometheus | `9090` |
| Grafana | `3000` |
| RabbitMQ UI | `15672` |
| Postgres | `5432` |

> **Note:** The Kubernetes setup (`script.sh`) is the primary workflow. Docker Compose is provided for local development convenience only.

---

## Troubleshooting

**Pods stuck in `Pending` or `Init`:**
```bash
kubectl describe pod <pod-name> -n devops-dsols
kubectl logs <pod-name> -n devops-dsols
```

**Ingress not routing (404 on `127.0.0.1`):**
- Ensure `sudo minikube tunnel` is running in Terminal 1.
- Check ingress addon: `minikube addons enable ingress`

**Prometheus data source test fails in Grafana:**
- Make sure you used `http://prometheus:9090` (the Kubernetes service DNS name), not `localhost`.

**k6 remote write fails:**
- Double-check the port from Terminal 2 output.
- Ensure `--web.enable-remote-write-receiver` is set on Prometheus (it is, via `prometheus.yaml` args).

**Image not found in Minikube:**
- Re-run: `minikube image load devops-dsols-app:latest`
- Or rebuild: `docker build -t devops-dsols-app:latest -f infra/Dockerfile .`
