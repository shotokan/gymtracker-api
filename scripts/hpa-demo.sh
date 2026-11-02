#!/usr/bin/env bash
# Generates real load and records how the HPA scales up and back down.
# Fails if the Deployment does not scale.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
API_URL="${API_URL:-http://127.0.0.1:8081}"
API_HOST="${API_HOST:-gymtracker.local}"
DYNAMO_URL="${DYNAMO_URL:-http://127.0.0.1:8000}"
TABLE="${DYNAMO_TABLE:-gym-tracker}"
EMAIL=loadtest@example.com

sample() {
  printf '%s  ' "$(date -u +%H:%M:%S)"
  kubectl -n gymtracker get hpa gymtracker-api --no-headers \
    | awk '{print "cpu=" $3 " min=" $4 " max=" $5 " replicas=" $6}'
}

replicas() { kubectl -n gymtracker get deployment gymtracker-api -o jsonpath='{.status.readyReplicas}'; }

# The load user logs in for real, so it needs an active account.
curl -sS -o /dev/null -X POST "$API_URL/users" -H "Host: $API_HOST" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"load-test-password\",\"name\":\"Load\"}"
AWS_ACCESS_KEY_ID=local AWS_SECRET_ACCESS_KEY=local AWS_REGION=us-east-1 \
  aws dynamodb update-item --endpoint-url "$DYNAMO_URL" --table-name "$TABLE" \
  --key "{\"PK\":{\"S\":\"USER#$EMAIL\"},\"SK\":{\"S\":\"PROFILE\"}}" \
  --update-expression "SET #s = :status" --expression-attribute-names '{"#s":"status"}' \
  --expression-attribute-values '{":status":{"S":"active"}}'

echo "── Antes de la carga"
sample

kubectl -n gymtracker delete job gymtracker-loadtest --ignore-not-found
kubectl apply -f "$ROOT/deploy/k8s/loadtest/job.yaml"

echo "── Durante la carga"
peak=0
for _ in $(seq 1 16); do
  sleep 15
  sample
  current="$(replicas)"
  [ "${current:-0}" -gt "$peak" ] && peak="$current"
done
kubectl -n gymtracker wait --for=condition=complete job/gymtracker-loadtest --timeout=120s || true

echo "── Después de la carga"
for _ in $(seq 1 16); do
  sleep 15
  sample
  [ "$(replicas)" = "1" ] && break
done

kubectl -n gymtracker describe hpa gymtracker-api | sed -n '/Events:/,$p'

echo "Máximo de réplicas listas: $peak"
if [ "$peak" -le 1 ]; then
  echo "FAIL  the HPA did not scale up" >&2
  exit 1
fi
