#!/usr/bin/env bash
#
# Deploy aivaldebatten to Cloud Run.
#
# Builds the container with Cloud Build (multi-stage Dockerfile: Svelte bundle ->
# embedded Go binary) and deploys it as a single always-on instance so the
# in-memory debate engine runs continuously and all viewers share one debate.
#
# Auth: the active gcloud account may not have access to the target project, so
# this script authenticates gcloud with Application Default Credentials (ADC).
# Make sure ADC is set up once:
#
#     gcloud auth application-default login
#
# Usage:
#     ./deploy.sh                # deploy with defaults below
#     PROJECT=other-proj ./deploy.sh
#     REGION=europe-north1 ./deploy.sh
#
set -euo pipefail

# --- Config (override via environment) ----------------------------------------
PROJECT="${PROJECT:-wedding-eg}"
REGION="${REGION:-europe-north2}"          # Stockholm
SERVICE="${SERVICE:-aivaldebatten}"

# Run from the repo root (directory containing this script) regardless of CWD.
cd "$(dirname "$0")"

# --- Authenticate gcloud via ADC ----------------------------------------------
# Drive gcloud with the ADC access token instead of the active CLI account, and
# bill API calls to the target project (avoids the shared SDK quota project).
echo ">> Authenticating with Application Default Credentials..."
if ! TOKEN="$(gcloud auth application-default print-access-token 2>/dev/null)"; then
  echo "ERROR: no Application Default Credentials found." >&2
  echo "       Run: gcloud auth application-default login" >&2
  exit 1
fi
export CLOUDSDK_AUTH_ACCESS_TOKEN="$TOKEN"
export CLOUDSDK_CORE_PROJECT="$PROJECT"
export CLOUDSDK_BILLING_QUOTA_PROJECT="$PROJECT"

# --- Ensure required APIs are enabled -----------------------------------------
echo ">> Ensuring required APIs are enabled on $PROJECT..."
gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  --quiet

# --- Build & deploy -----------------------------------------------------------
echo ">> Building and deploying $SERVICE to Cloud Run ($REGION)..."
gcloud run deploy "$SERVICE" \
  --source . \
  --region "$REGION" \
  --port 8080 \
  --allow-unauthenticated \
  --min-instances 1 \
  --max-instances 1 \
  --no-cpu-throttling \
  --session-affinity \
  --concurrency 250 \
  --timeout 3600 \
  --cpu 1 \
  --memory 512Mi \
  --quiet

# --- Report -------------------------------------------------------------------
URL="$(gcloud run services describe "$SERVICE" --region "$REGION" \
  --format='value(status.url)')"
echo
echo ">> Deployed: $URL"
