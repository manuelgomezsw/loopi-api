#!/usr/bin/env bash
# =============================================================================
# Loopi API — Cloud Monitoring Setup
# Crea uptime check, alert policies y dashboard en GCP Cloud Monitoring.
#
# Uso:
#   export PROJECT_ID="tu-gcp-project-id"
#   export APP_HOST="tu-app.appspot.com"       # hostname de App Engine sin https://
#   export NOTIFY_EMAIL="ops@tu-dominio.com"   # email para alertas (opcional)
#   ./scripts/monitoring/setup.sh
# =============================================================================
set -euo pipefail

# ── Validaciones ─────────────────────────────────────────────────────────────
: "${PROJECT_ID:?ERROR: Define PROJECT_ID antes de ejecutar este script}"
: "${APP_HOST:?ERROR: Define APP_HOST (ej: loopi-c048d.appspot.com)}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NOTIFY_EMAIL="${NOTIFY_EMAIL:-}"

echo "→ Proyecto : $PROJECT_ID"
echo "→ Host     : $APP_HOST"
echo ""

# ── 1. Notification channel (email) ──────────────────────────────────────────
CHANNEL_NAME=""
if [[ -n "$NOTIFY_EMAIL" ]]; then
  echo "▶ Creando notification channel para $NOTIFY_EMAIL ..."
  CHANNEL_NAME=$(gcloud alpha monitoring channels create \
    --display-name="Loopi API Ops" \
    --type="email" \
    --channel-labels="email_address=${NOTIFY_EMAIL}" \
    --project="$PROJECT_ID" \
    --format="value(name)")
  echo "  ✓ Channel: $CHANNEL_NAME"
fi

# ── 2. Uptime Check ──────────────────────────────────────────────────────────
echo "▶ Creando uptime check para /health ..."
gcloud monitoring uptime create "loopi-api /health" \
  --resource-type=uptime-url \
  --resource-labels="host=${APP_HOST},project_id=${PROJECT_ID}" \
  --path="/health" \
  --period=5 \
  --protocol=https \
  --project="$PROJECT_ID" \
  2>&1 || echo "  ⚠ Uptime check ya existe o requiere permisos adicionales."
echo "  ✓ Uptime check configurado (cada 5 min)"

# ── 3. Alert Policies ────────────────────────────────────────────────────────
import_policy() {
  local file="$1"
  local name="$2"
  # Sustituye el placeholder del proyecto en el JSON
  local tmp
  tmp=$(mktemp)
  sed "s/__PROJECT_ID__/${PROJECT_ID}/g" "$file" > "$tmp"

  # Agrega notification channel si existe
  if [[ -n "$CHANNEL_NAME" ]]; then
    # Inyecta el campo notificationChannels en el JSON temporal
    python3 -c "
import json, sys
data = json.load(open('$tmp'))
data['notificationChannels'] = ['${CHANNEL_NAME}']
json.dump(data, open('$tmp','w'), indent=2)
"
  fi

  gcloud monitoring policies create \
    --policy-from-file="$tmp" \
    --project="$PROJECT_ID" \
    2>&1 && echo "  ✓ $name" || echo "  ⚠ $name — verificar manualmente"
  rm -f "$tmp"
}

echo "▶ Importando alert policies ..."
import_policy "$SCRIPT_DIR/alerts/http_5xx.json"      "HTTP 5xx rate alert"
import_policy "$SCRIPT_DIR/alerts/latency_p95.json"   "Latency p95 alert"
import_policy "$SCRIPT_DIR/alerts/db_connections.json" "DB connections alert"

# ── 4. Dashboard ─────────────────────────────────────────────────────────────
echo "▶ Creando dashboard operacional ..."
tmp_dash=$(mktemp)
sed "s/__PROJECT_ID__/${PROJECT_ID}/g" "$SCRIPT_DIR/dashboard.json" > "$tmp_dash"
gcloud monitoring dashboards create \
  --config-from-file="$tmp_dash" \
  --project="$PROJECT_ID" \
  2>&1 && echo "  ✓ Dashboard creado" || echo "  ⚠ Dashboard — verificar manualmente"
rm -f "$tmp_dash"

echo ""
echo "✅ Setup completado. Revisa Cloud Monitoring en:"
echo "   https://console.cloud.google.com/monitoring?project=$PROJECT_ID"
