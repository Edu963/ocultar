#!/usr/bin/env bash
# Build and push the OCULTAR community image to ECR Public.
# Usage: ./scripts/ecr-push.sh [TAG]   (default TAG: latest)
#
# Prerequisites:
#   aws-cli v2 configured with credentials that have ecr-public:* access
#   docker buildx (for --platform linux/amd64 cross-builds)

set -euo pipefail

REPO="public.ecr.aws/ocultar/refinery"
TAG="${1:-latest}"
REGION="us-east-1"  # ECR Public is always us-east-1

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(dirname "$SCRIPT_DIR")"

echo "==> Building OCULTAR community image (linux/amd64)..."
docker build \
    --platform linux/amd64 \
    -f "$ROOT/docker/Dockerfile.marketplace" \
    -t "${REPO}:${TAG}" \
    "$ROOT"

echo "==> Authenticating with ECR Public..."
aws ecr-public get-login-password --region "$REGION" | \
    docker login --username AWS --password-stdin public.ecr.aws

echo "==> Pushing ${REPO}:${TAG}..."
docker push "${REPO}:${TAG}"

if [[ "$TAG" != "latest" ]]; then
    docker tag "${REPO}:${TAG}" "${REPO}:latest"
    docker push "${REPO}:latest"
    echo "==> Also tagged and pushed as latest"
fi

echo ""
echo "Done: ${REPO}:${TAG}"
echo "ECR Public URI: ${REPO}:${TAG}"
