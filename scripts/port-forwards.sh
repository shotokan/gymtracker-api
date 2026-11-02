#!/usr/bin/env bash
# Opens local access to the cluster services. Ctrl+C closes them all.
set -euo pipefail

trap 'kill $(jobs -p) 2>/dev/null || true' EXIT

forward() {
  # A forward drops when its target pod restarts, so reopen it.
  while true; do
    kubectl -n "$1" port-forward "svc/$2" "$3" >/dev/null 2>&1 || true
    sleep 1
  done
}

forward ingress traefik 8081:80 &
forward gymtracker dynamodb-local 8000:8000 &
forward monitoring monitoring-kube-prometheus-prometheus 9091:9090 &
forward monitoring monitoring-grafana 3000:80 &

cat <<MSG
API (por el Ingress)  http://127.0.0.1:8081   con Host: gymtracker.local
DynamoDB Local        http://127.0.0.1:8000
Prometheus            http://127.0.0.1:9091
Grafana               http://127.0.0.1:3000
MSG

wait
