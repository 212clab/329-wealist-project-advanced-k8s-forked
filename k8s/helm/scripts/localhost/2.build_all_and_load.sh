#!/bin/bash
# =============================================================================
# 모든 서비스 이미지 빌드 및 로드 (localhost 환경용)
# - Backend 서비스 + Frontend 포함
# - 레지스트리에 이미지가 있으면 스킵 (재사용)
# =============================================================================
#
# 사용법:
#   ./2.build_all_and_load.sh           # 레지스트리에 없는 이미지만 빌드
#   ./2.build_all_and_load.sh --force   # 모든 이미지 강제 재빌드
#   FORCE_BUILD=1 ./2.build_all_and_load.sh  # 환경변수로 강제 빌드

set -e

LOCAL_REG="localhost:5001"
TAG="${IMAGE_TAG:-latest}"
FORCE_BUILD="${FORCE_BUILD:-0}"

# --force 플래그 처리
if [[ "$1" == "--force" ]] || [[ "$1" == "-f" ]]; then
    FORCE_BUILD=1
fi

echo "=== 서비스 이미지 빌드 및 로드 (localhost 환경) ==="
echo ""
echo "레지스트리: ${LOCAL_REG}"
echo "태그: ${TAG}"
if [[ "$FORCE_BUILD" == "1" ]]; then
    echo "모드: 강제 재빌드 (--force)"
else
    echo "모드: 캐시 사용 (레지스트리에 있으면 스킵)"
fi
echo ""

# 레지스트리 확인
if ! curl -s "http://${LOCAL_REG}/v2/" > /dev/null 2>&1; then
    echo "ERROR: 레지스트리 없음. make kind-setup 먼저 실행"
    exit 1
fi

# 프로젝트 루트로 이동 (스크립트는 k8s/helm/scripts/localhost/ 에 위치)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
cd "$PROJECT_ROOT"
echo "Working directory: $PROJECT_ROOT"
echo ""

# 로컬 레지스트리에 이미지 있는지 확인
image_exists() {
    local name=$1 tag=$2
    curl -sf "http://${LOCAL_REG}/v2/${name}/manifests/${tag}" > /dev/null 2>&1
}

# 이미지 빌드 및 푸시 (캐시 체크 포함)
build_and_push() {
    local name=$1
    local context=$2
    local dockerfile=$3

    # 캐시 체크 (--force가 아니면)
    if [[ "$FORCE_BUILD" != "1" ]] && image_exists "$name" "$TAG"; then
        echo "✓ ${name}:${TAG} - 이미 있음 (스킵)"
        return 0
    fi

    echo "🔨 ${name}:${TAG} 빌드 중..."

    if [[ -n "$dockerfile" ]]; then
        docker build -t "${LOCAL_REG}/${name}:${TAG}" -f "$dockerfile" "$context"
    else
        docker build -t "${LOCAL_REG}/${name}:${TAG}" "$context"
    fi

    docker push "${LOCAL_REG}/${name}:${TAG}"
    echo "✅ ${name} 푸시 완료"
}

# =============================================================================
# Backend 서비스 빌드
# =============================================================================
echo "=========================================="
echo "  Backend 서비스 빌드"
echo "=========================================="

BACKEND_SERVICES=(
    "auth-service"
    "user-service"
    "board-service"
    "chat-service"
    "noti-service"
    "storage-service"
    "video-service"
)

for service in "${BACKEND_SERVICES[@]}"; do
    echo ""
    SERVICE_PATH="services/${service}"

    if [ ! -d "$SERVICE_PATH" ]; then
        echo "⚠️  ${SERVICE_PATH} 없음 - 스킵"
        continue
    fi

    # Dockerfile 확인 (루트 또는 docker/ 하위)
    if [ -f "${SERVICE_PATH}/Dockerfile" ]; then
        # 서비스 루트에 Dockerfile이 있으면 서비스 폴더를 컨텍스트로
        build_and_push "$service" "${SERVICE_PATH}" ""
    elif [ -f "${SERVICE_PATH}/docker/Dockerfile" ]; then
        # docker/ 하위에 Dockerfile이 있으면 프로젝트 루트를 컨텍스트로 (Go 모노레포)
        build_and_push "$service" "." "${SERVICE_PATH}/docker/Dockerfile"
    else
        echo "⚠️  ${SERVICE_PATH}/Dockerfile 없음 - 스킵"
    fi
done

# =============================================================================
# Frontend 빌드
# =============================================================================
echo ""
echo "=========================================="
echo "  Frontend 빌드"
echo "=========================================="

FRONTEND_PATH="services/frontend"
if [ -d "$FRONTEND_PATH" ] && [ -f "${FRONTEND_PATH}/Dockerfile" ]; then
    echo ""
    build_and_push "frontend" "${FRONTEND_PATH}" ""
else
    echo "⚠️  ${FRONTEND_PATH}/Dockerfile 없음 - 스킵"
fi

echo ""
echo "=========================================="
echo "  🎉 모든 서비스 이미지 처리 완료!"
echo "=========================================="
echo ""
echo "다음 단계:"
echo "  make helm-install-all ENV=localhost"
echo ""
echo "💡 팁: 이미지 강제 재빌드하려면:"
echo "  ./2.build_all_and_load.sh --force"
