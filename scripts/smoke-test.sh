#!/usr/bin/env bash
# Runs the whole access flow against a deployed API, and fails on the first surprise.
set -euo pipefail

API_URL="${API_URL:-http://127.0.0.1:8081}"
API_HOST="${API_HOST:-gymtracker.local}"
DYNAMO_URL="${DYNAMO_URL:-http://127.0.0.1:8000}"
TABLE="${DYNAMO_TABLE:-gym-tracker}"
RUN_ID="$(date +%s)-$RANDOM"
ADMIN_EMAIL="admin-$RUN_ID@example.com"
USER_EMAIL="user-$RUN_ID@example.com"
PASSWORD="smoke-test-$RUN_ID"

RESPONSE="$(mktemp)"
trap 'rm -f "$RESPONSE"' EXIT
passed=0

# call <expected status> <description> <method> <path> [token] [json body]
call() {
  local expected="$1" description="$2" method="$3" path="$4" token="${5:-}" body="${6:-}"
  local args=(-sS -o "$RESPONSE" -w '%{http_code}' -X "$method" -H "Host: $API_HOST")
  [ -n "$token" ] && args+=(-H "Authorization: Bearer $token")
  [ -n "$body" ] && args+=(-H "Content-Type: application/json" -d "$body")

  local status
  status="$(curl "${args[@]}" "$API_URL$path")"
  if [ "$status" != "$expected" ]; then
    echo "FAIL  $description: expected $expected, got $status" >&2
    cat "$RESPONSE" >&2
    exit 1
  fi
  echo "ok    $status  $description"
  passed=$((passed + 1))
}

# Promotes an account by hand, as the guide explains for the first administrator.
promote_admin() {
  AWS_ACCESS_KEY_ID=local AWS_SECRET_ACCESS_KEY=local AWS_REGION=us-east-1 \
    aws dynamodb update-item --endpoint-url "$DYNAMO_URL" --table-name "$TABLE" \
    --key "{\"PK\":{\"S\":\"USER#$1\"},\"SK\":{\"S\":\"PROFILE\"}}" \
    --update-expression "SET #r = :role, #s = :status" \
    --condition-expression "attribute_exists(PK)" \
    --expression-attribute-names '{"#r":"role","#s":"status"}' \
    --expression-attribute-values '{":role":{"S":"admin"},":status":{"S":"active"}}'
}

credentials() { printf '{"email":"%s","password":"%s"}' "$1" "$PASSWORD"; }

echo "── Primer administrador"
call 201 "registrar admin" POST /users "" "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$PASSWORD\",\"name\":\"Admin\"}"
call 401 "una cuenta pendiente no entra" POST /auth/login "" "$(credentials "$ADMIN_EMAIL")"
promote_admin "$ADMIN_EMAIL"
call 200 "login de admin" POST /auth/login "" "$(credentials "$ADMIN_EMAIL")"
ADMIN_TOKEN="$(jq -r .token "$RESPONSE")"

echo "── Alta de un usuario"
call 201 "registrar usuario" POST /users "" "{\"email\":\"$USER_EMAIL\",\"password\":\"$PASSWORD\",\"name\":\"Ana\",\"role\":\"admin\"}"
USER_ID="$(jq -r .id "$RESPONSE")"
call 409 "correo repetido" POST /users "" "{\"email\":\"$USER_EMAIL\",\"password\":\"$PASSWORD\",\"name\":\"Ana\"}"
call 401 "usuario pendiente no entra" POST /auth/login "" "$(credentials "$USER_EMAIL")"
call 200 "admin lista pendientes" GET "/users?status=pending" "$ADMIN_TOKEN"
jq -e --arg id "$USER_ID" '.users | any(.id == $id)' "$RESPONSE" >/dev/null || { echo "FAIL  pending user not listed" >&2; exit 1; }
call 204 "admin aprueba" PATCH "/users/$USER_ID/activate" "$ADMIN_TOKEN"
call 200 "login de usuario" POST /auth/login "" "$(credentials "$USER_EMAIL")"
USER_TOKEN="$(jq -r .token "$RESPONSE")"

echo "── Permisos"
call 401 "sin token" GET "/users/$USER_ID"
call 403 "token inválido" GET "/users/$USER_ID" "not-a-token"
call 200 "perfil propio" GET "/users/$USER_ID" "$USER_TOKEN"
[ "$(jq -r .status "$RESPONSE")" = "active" ] || { echo "FAIL  status is not active" >&2; exit 1; }
call 403 "el admin no ve perfiles ajenos" GET "/users/$USER_ID" "$ADMIN_TOKEN"
call 403 "un usuario no lista cuentas" GET "/users?status=pending" "$USER_TOKEN"
call 403 "el rol enviado al registrarse se ignora" PATCH "/users/$USER_ID/activate" "$USER_TOKEN"
call 404 "ruta inexistente" GET /nope "$USER_TOKEN"

echo "── Plan y entrenamiento"
call 201 "crear plan" POST /plans "$USER_TOKEN" '{"name":"Fuerza"}'
PLAN_ID="$(jq -r .id "$RESPONSE")"
call 204 "activar plan" PATCH "/plans/$PLAN_ID/activate" "$USER_TOKEN"
call 200 "plan activo" GET /plans/active "$USER_TOKEN"
call 201 "agregar rutina" POST "/plans/$PLAN_ID/routines" "$USER_TOKEN" '{"name":"Empuje"}'
ROUTINE_ID="$(jq -r .id "$RESPONSE")"
call 201 "agregar ejercicio" POST "/plans/$PLAN_ID/routines/$ROUTINE_ID/exercises" "$USER_TOKEN" '{"name":"Press banca","sets":4,"reps":8,"weight":60}'
EXERCISE_ID="$(jq -r .id "$RESPONSE")"
call 200 "listar ejercicios" GET "/plans/$PLAN_ID/routines/$ROUTINE_ID/exercises" "$USER_TOKEN"
call 201 "registrar sesión" POST /sessions "$USER_TOKEN" "{\"plan_id\":\"$PLAN_ID\",\"routine_id\":\"$ROUTINE_ID\",\"routine_name\":\"Empuje\",\"exercises\":[{\"exercise_id\":\"$EXERCISE_ID\",\"name\":\"Press banca\",\"sets\":[{\"weight\":60,\"reps\":8,\"rest_seconds\":90}]}]}"
SESSION_ID="$(jq -r .id "$RESPONSE")"
[ "$(jq -r .total_volume "$RESPONSE")" = "480" ] || { echo "FAIL  total_volume is not 480" >&2; exit 1; }
call 200 "ver sesión" GET "/sessions/$SESSION_ID" "$USER_TOKEN"
call 200 "listar sesiones" GET /sessions "$USER_TOKEN"

echo "── Baja y último admin"
call 204 "darse de baja" DELETE "/users/$USER_ID" "$USER_TOKEN"
call 401 "una cuenta dada de baja no entra" POST /auth/login "" "$(credentials "$USER_EMAIL")"
call 200 "admin lista bajas" GET "/users?status=deactivated" "$ADMIN_TOKEN"

echo
echo "$passed comprobaciones correctas"
