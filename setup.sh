#!/usr/bin/env bash
# setup.sh - Run ONCE to configure Minikube, deploy everything, and configure Grafana.
set -euo pipefail

NS=devops-dsols

is_yes() {
  case "$1" in
    [Yy]|[Yy][Ee][Ss]) return 0 ;;
    *) return 1 ;;
  esac
}

# --- Configure Minikube resources and start fresh ---

read -r -p "Apply Minikube config for 4 CPUs and 4096 MB RAM? [y/N] " APPLY_MINIKUBE_CONFIG
if is_yes "${APPLY_MINIKUBE_CONFIG:-}"; then
  echo "Setting Minikube to use 4 CPUs and 4096 MB RAM..."
  minikube config set cpus 4
  minikube config set memory 4096

  read -r -p "Delete the current Minikube cluster now so the new config takes effect? [y/N] " RESET_MINIKUBE
  if is_yes "${RESET_MINIKUBE:-}"; then
    echo "Deleting existing Minikube cluster..."
    minikube delete
  else
    echo "Skipping Minikube delete. Existing cluster settings will remain until you recreate the cluster."
  fi
else
  echo "Skipping Minikube config changes."
fi

echo "Starting Minikube with Docker driver..."
minikube start --driver=docker


# --- Configure HELM charts ---

read -r -p "Add helm charts? [y/N] " APPLY_HELM
if is_yes "${APPLY_HELM:-}"; then
  echo "Setting HELM charts"
  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
  helm repo update
  helm install kube-state-metrics prometheus-community/kube-state-metrics -n devops-dsols
else
  echo "Skipping HELM charts."
fi


# --- Enable required addons ---

echo "Enabling addons..."
minikube addons enable metrics-server
minikube addons enable ingress


# --- Build image and deploy all Kubernetes resources ---

echo "Building Docker image and deploying to Kubernetes..."
chmod +x ./pipeline.sh
./pipeline.sh


# --- Wait for all pods to be ready ---

echo "Waiting for all pods to be ready (this may take a few minutes)..."
for app in postgres rabbitmq prometheus grafana user-controller user-service product-controller product-service; do
  echo "  Waiting for $app..."
  kubectl wait --for=condition=ready pod -l app=$app -n $NS --timeout=300s 2>/dev/null \
    || echo "  Warning: $app not ready in time, moving on."
done

# --- Done ---

echo ""
echo "Setup complete."
echo "Run ./start.sh to start the stack next time."
