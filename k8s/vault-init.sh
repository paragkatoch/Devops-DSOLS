#!/usr/bin/env bash
# vault-init.sh — Install Vault (dev mode) and sync secrets to K8s Secrets.
# Safe to re-run (idempotent via --dry-run=client | kubectl apply).
set -euo pipefail

NS=devops-dsols
VAULT_POD=vault-0

# ─────────────────────────────────────────────
# 1. Install Vault via Helm (dev mode, no sidecar injector)
# ─────────────────────────────────────────────
echo "► Adding HashiCorp Helm repo..."
helm repo add hashicorp https://helm.releases.hashicorp.com 2>/dev/null || true
helm repo update

echo "► Installing Vault in dev mode (single pod, no TLS, auto-unsealed)..."
helm upgrade --install vault hashicorp/vault \
  --namespace "$NS" \
  --set "server.dev.enabled=true" \
  --set "injector.enabled=false" \
  --wait --timeout 120s

# ─────────────────────────────────────────────
# 2. Wait for Vault pod to be Ready
# ─────────────────────────────────────────────
echo "► Waiting for Vault pod..."
kubectl wait pod/$VAULT_POD -n "$NS" --for=condition=Ready --timeout=120s

# ─────────────────────────────────────────────
# 3. Enable KV-v2 and write secrets into Vault
# ─────────────────────────────────────────────
echo "► Enabling KV-v2 secrets engine at 'secret/'..."
kubectl exec -n "$NS" $VAULT_POD -- vault secrets enable -path=secret kv-v2 2>/dev/null || true

echo "► Writing Postgres credentials..."
# ⚠️  Change the POSTGRES_PASSWORD values here to use your real passwords.
kubectl exec -n "$NS" $VAULT_POD -- vault kv put secret/devops-dsols/postgres/order \
  POSTGRES_DB=order_db POSTGRES_USER=postgres POSTGRES_PASSWORD=password

kubectl exec -n "$NS" $VAULT_POD -- vault kv put secret/devops-dsols/postgres/product \
  POSTGRES_DB=product_db POSTGRES_USER=postgres POSTGRES_PASSWORD=password

kubectl exec -n "$NS" $VAULT_POD -- vault kv put secret/devops-dsols/postgres/user \
  POSTGRES_DB=user_db POSTGRES_USER=postgres POSTGRES_PASSWORD=password

echo "► Writing app DSN config..."
kubectl exec -n "$NS" $VAULT_POD -- vault kv put secret/devops-dsols/app \
  DB_ORDER="postgres://postgres:password@order-postgres:5432/order_db?sslmode=disable" \
  DB_PRODUCT="postgres://postgres:password@product-postgres:5432/product_db?sslmode=disable" \
  DB_USER="postgres://postgres:password@user-postgres:5432/user_db?sslmode=disable" \
  RABBITMQ="amqp://guest:guest@rabbitmq:5672/"

# ─────────────────────────────────────────────
# 4. Sync Vault → native Kubernetes Secrets
#    (no pod annotations, no sidecar, just plain K8s Secrets)
# ─────────────────────────────────────────────
echo "► Syncing Vault secrets → Kubernetes Secrets..."

for DB in order product user; do
  PASS=$(kubectl exec -n "$NS" $VAULT_POD -- vault kv get -field=POSTGRES_PASSWORD "secret/devops-dsols/postgres/$DB")
  PG_USER=$(kubectl exec -n "$NS" $VAULT_POD -- vault kv get -field=POSTGRES_USER    "secret/devops-dsols/postgres/$DB")
  PG_DB=$(kubectl exec -n "$NS" $VAULT_POD -- vault kv get -field=POSTGRES_DB       "secret/devops-dsols/postgres/$DB")

  kubectl create secret generic "${DB}-postgres-creds" \
    -n "$NS" \
    --from-literal=POSTGRES_PASSWORD="$PASS" \
    --from-literal=POSTGRES_USER="$PG_USER" \
    --from-literal=POSTGRES_DB="$PG_DB" \
    --dry-run=client -o yaml | kubectl apply -f -

  echo "  ✓ ${DB}-postgres-creds"
done

# Render app local.yml from Vault data and store as a K8s Secret.
# The app pods mount this Secret the same way they used to mount the ConfigMap.
DB_ORDER_DSN=$(kubectl exec -n "$NS" $VAULT_POD -- vault kv get -field=DB_ORDER   secret/devops-dsols/app)
DB_PROD_DSN=$(kubectl exec -n "$NS" $VAULT_POD -- vault kv get -field=DB_PRODUCT  secret/devops-dsols/app)
DB_USER_DSN=$(kubectl exec -n "$NS" $VAULT_POD -- vault kv get -field=DB_USER     secret/devops-dsols/app)
RABBIT_DSN=$(kubectl exec -n "$NS" $VAULT_POD -- vault kv get -field=RABBITMQ    secret/devops-dsols/app)

kubectl create secret generic app-config \
  -n "$NS" \
  --from-literal=local.yml="env: \"local\"
databases:
  order: \"${DB_ORDER_DSN}\"
  product: \"${DB_PROD_DSN}\"
  user: \"${DB_USER_DSN}\"
queue_path: \"${RABBIT_DSN}\"
http_server:
  address: \":9000\"" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "  ✓ app-config secret"
echo ""
echo "✅  Vault init complete. All secrets are live in Kubernetes."
echo "   To inspect:  kubectl get secrets -n $NS"
echo "   Vault UI:    kubectl port-forward -n $NS svc/vault 8200:8200"
