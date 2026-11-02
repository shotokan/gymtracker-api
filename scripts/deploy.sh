#!/usr/bin/env bash
# Deploys the application to the current kubectl context.
#   scripts/deploy.sh <image> [overlay]
set -euo pipefail

IMAGE="${1:?usage: deploy.sh <image> [overlay]}"
OVERLAY="${2:-local}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

kubectl apply -f "$ROOT/deploy/k8s/base/namespace.yaml"

# Keep the existing secret so a redeploy does not invalidate the sessions in use.
if ! kubectl -n gymtracker get secret gymtracker-secrets >/dev/null 2>&1; then
  kubectl -n gymtracker create secret generic gymtracker-secrets \
    --from-literal=jwt-secret="$(openssl rand -base64 48)"
fi

kubectl kustomize "$ROOT/deploy/k8s/overlays/$OVERLAY" \
  | sed -E "s#image: (gymtracker-api:local|.*\.amazonaws\.com/gymtracker-api:latest)#image: ${IMAGE}#" \
  | kubectl apply -f -
kubectl apply -k "$ROOT/deploy/k8s/monitoring"

if kubectl -n gymtracker get deployment dynamodb-local >/dev/null 2>&1; then
  kubectl -n gymtracker rollout status deployment/dynamodb-local --timeout=180s
fi
kubectl -n gymtracker rollout status deployment/gymtracker-api --timeout=240s
