#!/usr/bin/env bash
# Creates the whole local environment on minikube: cluster, platform and application.
# Needs about 6 GB of memory and 15 GB of disk. It costs nothing.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IMAGE="gymtracker-api:local"

for tool in docker minikube kubectl terraform jq aws; do
  command -v "$tool" >/dev/null || { echo "missing tool: $tool" >&2; exit 1; }
done

minikube status >/dev/null 2>&1 || minikube start --cpus=4 --memory=6g --disk-size=15g

terraform -chdir="$ROOT/terraform/environments/local" init -input=false
terraform -chdir="$ROOT/terraform/environments/local" apply -input=false -auto-approve

minikube image build -t "$IMAGE" "$ROOT"
"$ROOT/scripts/deploy.sh" "$IMAGE"

cat <<MSG

Listo. En otra terminal abre los accesos:
  scripts/port-forwards.sh

Y prueba la API:
  scripts/smoke-test.sh
MSG
