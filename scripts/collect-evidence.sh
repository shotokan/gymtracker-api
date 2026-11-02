#!/usr/bin/env bash
# Saves the state of the cluster, Prometheus and Grafana into a directory.
#   scripts/collect-evidence.sh <dir>
set -uo pipefail

DIR="${1:?usage: collect-evidence.sh <dir>}"
PROMETHEUS="${PROMETHEUS_URL:-http://127.0.0.1:9091}"
GRAFANA="${GRAFANA_URL:-http://127.0.0.1:3000}"
CHROME="${CHROME_BIN:-google-chrome}"
mkdir -p "$DIR"

{
  kubectl version
  kubectl get nodes -o wide
  kubectl -n gymtracker get deploy,rs,pods,svc,ingress,hpa,pdb,cronjob,job,networkpolicy,resourcequota -o wide
  kubectl -n gymtracker describe resourcequota
  kubectl -n ingress get pods,svc
  kubectl -n monitoring get pods,svc
  kubectl top pods -n gymtracker
} > "$DIR/cluster-state.txt" 2>&1

kubectl -n gymtracker logs deployment/gymtracker-api --tail=30 > "$DIR/api-logs.txt" 2>&1

query() {
  curl -sS --get "$PROMETHEUS/api/v1/query" --data-urlencode "query=$1"
}

{
  echo "── Objetivos de Prometheus para la API"
  curl -sS "$PROMETHEUS/api/v1/targets?state=active" \
    | jq '[.data.activeTargets[] | select(.labels.namespace == "gymtracker") | {job: .labels.job, pod: .labels.pod, health, lastError}]'
  echo "── Peticiones totales por ruta y código"
  query 'sum by (route, status) (gymtracker_http_requests_total)' \
    | jq -r '.data.result | sort_by(.metric.route)[] | "\(.value[1])\t\(.metric.status)\t\(.metric.route)"'
  echo "── Latencia p95, en segundos, en los últimos 15 minutos"
  query 'histogram_quantile(0.95, sum by (le) (rate(gymtracker_http_request_duration_seconds_bucket[15m])))' \
    | jq -r '.data.result[0].value[1]'
  echo "── Máximo de réplicas del HPA en los últimos 30 minutos"
  query 'max_over_time(kube_horizontalpodautoscaler_status_current_replicas{horizontalpodautoscaler="gymtracker-api"}[30m])' \
    | jq -r '.data.result[0].value[1]'
  echo "── Reglas de alerta cargadas"
  curl -sS "$PROMETHEUS/api/v1/rules?type=alert" \
    | jq -r '.data.groups[] | select(.name == "gymtracker-api") | .rules[] | "\(.state)\t\(.name)"'
} > "$DIR/prometheus-queries.txt" 2>&1

shot() {
  "$CHROME" --headless=new --no-sandbox --disable-gpu --hide-scrollbars \
    --window-size="$3" --virtual-time-budget=30000 --screenshot="$DIR/$1" "$2" >/dev/null 2>&1 \
    || echo "could not capture $1" >> "$DIR/errors.txt"
}

shot grafana-dashboard.png "$GRAFANA/d/gymtracker/gymtracker-api?orgId=1&from=now-30m&to=now&kiosk" 1600,1500
shot prometheus-targets.png "$PROMETHEUS/targets?search=gymtracker" 1600,900
shot prometheus-replicas.png "$PROMETHEUS/graph?g0.expr=kube_horizontalpodautoscaler_status_current_replicas&g0.tab=0&g0.range_input=30m" 1600,900

exit 0
