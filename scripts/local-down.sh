#!/usr/bin/env bash
# FinOps: frees the memory and disk of the local environment.
#   scripts/local-down.sh           stops the cluster and keeps its data
#   scripts/local-down.sh --delete  deletes the cluster and the Terraform state
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

if [ "${1:-}" = "--delete" ]; then
  minikube delete
  rm -f "$ROOT"/terraform/environments/local/terraform.tfstate*
else
  minikube stop
fi
