# Devops-DSOLS

A Go microservices system deployed on Kubernetes (Minikube) with full observability via Prometheus & Grafana and load testing via k6.

---

## Table of Contents

- [Devops-DSOLS](#devops-dsols)
  - [Table of Contents](#table-of-contents)
  - [Architecture Overview](#architecture-overview)
  - [Prerequisites](#prerequisites)
  - [Quick Start](#quick-start)
  - [Step-by-Step Run Guide](#step-by-step-run-guide)
    - [Setup scripts](#setup-scripts)
    - [Minikube and k8 setup](#minikube-and-k8-setup)
    - [Start Dashboard, tunnel, port-forwarding for prometheus and grafana](#start-dashboard-tunnel-port-forwarding-for-prometheus-and-grafana)
    - [Manually update when needed](#manually-update-when-needed)
    - [Configure Grafana](#configure-grafana)
    - [Run k6 load](#run-k6-load)
  - [Services \& Ports](#services--ports)
  - [Kubernetes Resources](#kubernetes-resources)
    - [Namespace](#namespace)
    - [ConfigMaps](#configmaps)
    - [Persistent Volume Claims](#persistent-volume-claims)
    - [Resource Limits Summary](#resource-limits-summary)
  - [CI/CD Pipeline (Jenkins)](#cicd-pipeline-jenkins)
  - [Troubleshooting](#troubleshooting)

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

| Component            | Type                  | Role                                                 |
| -------------------- | --------------------- | ---------------------------------------------------- |
| `user-controller`    | Go app                | HTTP handler for `/api/user` routes                  |
| `user-service`       | Go app                | Business logic consumer for user events              |
| `product-controller` | Go app                | HTTP handler for `/api/product` routes               |
| `product-service`    | Go app                | Business logic consumer for product events           |
| `postgres`           | PostgreSQL 15         | Primary datastore (`retail` DB)                      |
| `rabbitmq`           | RabbitMQ + management | Async message queue between controllers and services |
| `prometheus`         | Prometheus            | Metrics scraping & remote-write receiver             |
| `grafana`            | Grafana               | Metrics visualization                                |
| `nginx ingress`      | Kubernetes Ingress    | API gateway routing to services                      |

---

<!--
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
├── pipeline.sh               # One-shot build + deploy script
├── test.js                 # k6 load test script
├── grafanaDashboard.json   # Pre-built Grafana dashboard
├── docker-compose.yml      # Local dev alternative (no Kubernetes)
└── Jenkinsfile             # Jenkins CI/CD pipeline definition
```

--- -->

## Prerequisites

Make sure the following are installed before starting:

| Tool       | Purpose                  | Install                          |
| ---------- | ------------------------ | -------------------------------- |
| `minikube` | Local Kubernetes cluster | `brew install minikube`          |
| `kubectl`  | Kubernetes CLI           | `brew install kubectl`           |
| `docker`   | Build app image          | [docker.com](https://docker.com) |
| `k6`       | Load testing             | `brew install k6`                |

---

## Quick Start

Two scripts handle the entire workflow:

| Script          | When to run                                      |
| --------------- | ------------------------------------------------ |
| `./setup.sh`    | Configures Minikube, builds & deploys everything |
| `./start.sh`    | Starts Minikube, tunnel, port-forwards           |
| `./pipeline.sh` | Run to update the deployment                     |

## Step-by-Step Run Guide

### Setup scripts

```bash
chmod +x ./setup.sh
chmod +x ./start.sh
chmod +x ./pipeline.sh
```

### Minikube and k8 setup

```bash
./setup.sh
```

### Start Dashboard, tunnel, port-forwarding for prometheus and grafana

```bash
sudo ./start.sh
```

### Manually update when needed

```bash
./pipeline.sh
```

### Configure Grafana

1. Open the Grafana URL, login with `admin` / `admin`, click **Skip**.
2. Go to **Connections → Add new connection → Prometheus**.
3. Set URL to `http://prometheus:9090`, click **Save & Test**.
4. Go to **Dashboards → New → Import**, paste `grafanaDashboard.json`, select the Prometheus datasource, click **Import**.

### Run k6 load

Copy k6 script from terminal and run it

Check the Grafana dashboard for live results.

---

## Services & Ports

| Service                  | Internal Port             | Access Method                                 |
| ------------------------ | ------------------------- | --------------------------------------------- |
| Ingress / API Gateway    | `80`                      | `http://127.0.0.1` (via `minikube tunnel`)    |
| Prometheus               | `9090`                    | `minikube service prometheus -n devops-dsols` |
| Grafana                  | `3000` / NodePort `30000` | `minikube service grafana -n devops-dsols`    |
| Postgres                 | `5432`                    | Internal only                                 |
| RabbitMQ AMQP            | `5672`                    | Internal only                                 |
| RabbitMQ UI              | `15672`                   | Internal only                                 |
| RabbitMQ Metrics         | `15692`                   | Scraped by Prometheus                         |
| App services/controllers | `9000`                    | Scraped by Prometheus                         |

---

## Kubernetes Resources

### Namespace

All resources live in the `devops-dsols` namespace:

```bash
kubectl get all -n devops-dsols
```

### ConfigMaps

| ConfigMap           | Purpose                                               |
| ------------------- | ----------------------------------------------------- |
| `app-config`        | App config: DB URL, RabbitMQ URL, HTTP listen address |
| `prometheus-config` | Prometheus scrape targets                             |

### Persistent Volume Claims

| PVC              | Size | Used By                                   |
| ---------------- | ---- | ----------------------------------------- |
| `postgres-pvc`   | 1Gi  | Postgres data                             |
| `prometheus-pvc` | 1Gi  | Prometheus TSDB (2h retention, 200MB cap) |
| `grafana-pvc`    | 1Gi  | Grafana dashboards & settings             |

### Resource Limits Summary

| Workload                          | CPU Request | CPU Limit | Memory Request | Memory Limit |
| --------------------------------- | ----------- | --------- | -------------- | ------------ |
| `user/product-controller/service` | 10m         | 50m       | 128Mi          | 256Mi        |
| `postgres`                        | 200m        | 500m      | 256Mi          | 512Mi        |
| `rabbitmq`                        | 250m        | 500m      | 512Mi          | 1024Mi       |
| `prometheus`                      | 100m        | 512m      | 256Mi          | 512Mi        |
| `grafana`                         | 100m        | 500m      | 256Mi          | 1Gi          |

---

## CI/CD Pipeline (Jenkins)

The `Jenkinsfile` defines a declarative pipeline that mirrors `pipeline.sh` for automated CI/CD:

| Stage                           | Action                                                          |
| ------------------------------- | --------------------------------------------------------------- |
| Verify Environment              | `minikube status \|\| minikube start` + enable ingress addon    |
| Build Docker Image              | `docker build -t devops-dsols-app:latest -f infra/Dockerfile .` |
| Load Image to Minikube          | `minikube image load devops-dsols-app:latest`                   |
| Deploy Configs & Infrastructure | Apply namespace, ConfigMaps, and infrastructure YAMLs           |
| Deploy Applications             | Apply app YAMLs + `kubectl rollout restart`                     |

**Requirements for Jenkins:**

- Jenkins agent must have `minikube`, `kubectl`, and `docker` in PATH.
- PATH is configured in the `environment` block: `/opt/homebrew/bin:/usr/local/bin`.

---

<!--
## Docker Compose (Local Dev Alternative)

If you want to run without Kubernetes, use Docker Compose for a quick local setup:

```bash
docker compose up --build
```

| Service             | Port    |
| ------------------- | ------- |
| API Gateway (nginx) | `9000`  |
| Prometheus          | `9090`  |
| Grafana             | `3000`  |
| RabbitMQ UI         | `15672` |
| Postgres            | `5432`  |

> **Note:** The Kubernetes setup (`pipeline.sh`) is the primary workflow. Docker Compose is provided for local development convenience only.

--- -->

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
