#!/bin/bash
# =============================================================================
# Generate Kustomization for Docker Hub Deployment
# =============================================================================
# 사용법:
#   DOCKER_HUB_ID=myid ./docker/scripts/docker-hub/generate-kustomization.sh
#
# 환경변수:
#   DOCKER_HUB_ID  - Docker Hub 사용자/조직 ID (필수)
#   IMAGE_TAG      - 이미지 태그 (기본값: latest)
# =============================================================================

set -e

# Docker Hub ID 확인
if [ -z "$DOCKER_HUB_ID" ]; then
    echo "Error: DOCKER_HUB_ID environment variable is required"
    echo ""
    echo "Usage:"
    echo "  DOCKER_HUB_ID=your-docker-id ./docker/scripts/docker-hub/generate-kustomization.sh"
    exit 1
fi

# 설정
TAG="${IMAGE_TAG:-latest}"
REPO_ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
TEMPLATE_DIR="$REPO_ROOT/k8s/overlays/develop-dockerhub/all-services"
TEMPLATE_FILE="$TEMPLATE_DIR/kustomization.yaml.template"
OUTPUT_FILE="$TEMPLATE_DIR/kustomization.yaml"

if [ ! -f "$TEMPLATE_FILE" ]; then
    echo "Error: Template file not found: $TEMPLATE_FILE"
    exit 1
fi

echo "Generating kustomization.yaml..."
echo "  Docker Hub ID: $DOCKER_HUB_ID"
echo "  Image Tag: $TAG"

# envsubst로 템플릿 치환
export DOCKER_HUB_ID
export IMAGE_TAG="$TAG"
envsubst < "$TEMPLATE_FILE" > "$OUTPUT_FILE"

echo "  Generated: $OUTPUT_FILE"
echo ""
echo "Preview:"
head -20 "$OUTPUT_FILE"
echo "..."
