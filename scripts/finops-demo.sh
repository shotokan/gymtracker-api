#!/usr/bin/env bash
# Runs the off-hours CronJobs on demand and checks that they stop and restart the API.
set -euo pipefail

run_job() {
  local cron="$1" job="$1-manual-$(date +%s)"
  kubectl -n gymtracker create job "$job" --from="cronjob/$cron"
  kubectl -n gymtracker wait --for=condition=complete "job/$job" --timeout=120s
  kubectl -n gymtracker logs "job/$job"
}

desired() { kubectl -n gymtracker get deployment gymtracker-api -o jsonpath='{.spec.replicas}'; }

kubectl -n gymtracker get cronjobs

echo "── Apagado"
run_job finops-scale-down
[ "$(desired)" = "0" ] || { echo "FAIL  expected 0 replicas" >&2; exit 1; }
kubectl -n gymtracker get deployment gymtracker-api

echo "── Encendido"
run_job finops-scale-up
[ "$(desired)" = "1" ] || { echo "FAIL  expected 1 replica" >&2; exit 1; }
kubectl -n gymtracker rollout status deployment/gymtracker-api --timeout=120s
kubectl -n gymtracker get deployment gymtracker-api
