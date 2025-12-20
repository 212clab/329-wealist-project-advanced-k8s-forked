#!/bin/bash
# =============================================================================
# Istio Gateway hostPort 80 설정 (localhost 환경용)
# - localhost:80 으로 접근 가능하도록 Gateway Pod 패치
# - 0.setup-cluster.sh 실행 후 한 번 실행
# =============================================================================

set -e

echo "=== Istio Gateway hostPort 80 설정 ==="
echo ""

# Istio Gateway 존재 확인
if ! kubectl get deployment istio-ingressgateway -n istio-system &>/dev/null; then
    echo "ERROR: istio-ingressgateway deployment가 없습니다."
    echo "       0.setup-cluster.sh를 먼저 실행하세요."
    exit 1
fi

# Gateway Pod를 hostPort 80으로 패치
echo "⚙️ Gateway Deployment 패치 중..."
kubectl patch deployment istio-ingressgateway -n istio-system --type='json' -p='[
  {
    "op": "replace",
    "path": "/spec/template/spec/containers/0/ports",
    "value": [
      {"containerPort": 80, "hostPort": 80, "protocol": "TCP", "name": "http"},
      {"containerPort": 443, "hostPort": 443, "protocol": "TCP", "name": "https"}
    ]
  },
  {
    "op": "add",
    "path": "/spec/template/spec/nodeSelector",
    "value": {"ingress-ready": "true"}
  }
]'

# Gateway Pod 재시작 대기
echo "⏳ Gateway Pod 재시작 대기 중..."
kubectl rollout status deployment/istio-ingressgateway -n istio-system --timeout=120s

echo ""
echo "✅ Istio Gateway hostPort 80 설정 완료!"
echo ""
echo "📌 접근 방법:"
echo "   - Frontend: localhost/"
echo "   - API: localhost/svc/{service}/api/..."
echo ""
