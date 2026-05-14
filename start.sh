#!/usr/bin/env bash
# start.sh - Run every time to start the stack (after setup.sh has been run once).
set -euo pipefail

if [ "$EUID" -eq 0 ]; then
  echo "Error: Please do not run this script with sudo or as root."
  echo "Run it as your normal user: ./start.sh"
  exit 1
fi

NS=devops-dsols
PROM_PORT=9091
GRAF_PORT=3001

# Stop port-forwards and tunnel when script exits
cleanup() {
  echo "Shutting down port-forwards and tunnel..."
  kill $PROM_PF_PID $GRAF_PF_PID $TUNNEL_PID 2>/dev/null || true
}
trap cleanup EXIT


# --- Start Minikube if not already running ---

echo "Starting Minikube..."
minikube status &>/dev/null || minikube start --driver=docker


# --- Start Kubernetes dashboard in background ---

echo "Starting Kubernetes dashboard..."
nohup minikube dashboard &>/dev/null &


# --- Start minikube tunnel in background ---
# nohub sudo -E minikube tunnel

# --- Start API Gateway Port-Forward in background ---

echo "Starting API Gateway port-forward on port 80 (you will be prompted for your sudo password)..."
# Cache sudo credentials upfront so the background command doesn't hang on a password prompt
sudo -v
nohup sudo -E kubectl port-forward svc/ingress-nginx-controller -n ingress-nginx 80:80 >/dev/null 2>&1 &
TUNNEL_PID=$!
echo "API Gateway port-forward started."
sleep 2


# --- Port-forward Prometheus and Grafana ---

echo "Port-forwarding Prometheus to localhost:$PROM_PORT..."
kubectl port-forward svc/prometheus $PROM_PORT:9090 -n $NS &>/dev/null &
PROM_PF_PID=$!

echo "Port-forwarding Grafana to localhost:$GRAF_PORT..."
kubectl port-forward svc/grafana $GRAF_PORT:3000 -n $NS &>/dev/null &
GRAF_PF_PID=$!

sleep 2


# --- Summary ---

echo ""
echo "Everything is running."
echo "  Prometheus : http://127.0.0.1:$PROM_PORT"
echo "  Grafana    : http://127.0.0.1:$GRAF_PORT  (admin / admin)"
echo "  API        : http://127.0.0.1  (via minikube tunnel)"
echo "  K6 Script  : K6_PROMETHEUS_RW_SERVER_URL='http://127.0.0.1:${PROM_PORT}/api/v1/write' k6 run --out experimental-prometheus-rw test.js"


# --- Optionally run k6 load test ---

# read -rp "Run k6 load test? [y/N] " RUN_K6
# case "${RUN_K6:-}" in
#   [Yy]|[Yy][Ee][Ss])
#   echo "Running k6..."
#   K6_PROMETHEUS_RW_SERVER_URL="http://127.0.0.1:${PROM_PORT}/api/v1/write" \
#     k6 run --out experimental-prometheus-rw test.js
#   echo "k6 done. Check Grafana for results."
#   ;;
# esac




# --- Keep port-forwards and tunnel alive ---

echo ""
echo "Services are live. Press Ctrl+C to stop everything."
wait
