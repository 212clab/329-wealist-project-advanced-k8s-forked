# Istio Ambient Mesh 가이드

> 이 문서는 Claude Code가 Istio Ambient 관련 작업 시 참조하는 문서입니다.
> 인터넷 검색 없이 이 문서를 참조하세요.

## 버전 호환성 매트릭스

### Istio Ambient GA 버전 (2024.11~)

| Istio Version | Kubernetes | Gateway API | Status | 비고 |
|---------------|------------|-------------|--------|------|
| **1.24.x** | 1.28-1.31 | v1.2.0 | ✅ GA (Production Ready) | 프로젝트 사용 버전 |
| 1.25.x | 1.29-1.32 | v1.2.0 | ✅ GA | |
| 1.26.x | 1.29-1.32 | v1.2.0 | ✅ GA | |
| 1.27.x | 1.30-1.33 | v1.2.0 | ✅ GA | |
| 1.28.x | 1.30-1.33 | v1.2.0 | ✅ GA | 최신 (2025.12) |

### 필수 의존성

| Component | Version | 설치 방법 |
|-----------|---------|----------|
| Kubernetes Gateway API CRDs | v1.2.0 | `kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.0/standard-install.yaml` |
| Helm | 3.12+ | brew install helm |
| kubectl | 1.28+ | brew install kubectl |

### Helm Chart 버전 (istio-release repo)

```
istio/base       - Istio CRDs
istio/istiod     - Control plane
istio/cni        - CNI plugin (Ambient 필수)
istio/ztunnel    - L4 proxy (Ambient 필수)
istio/gateway    - Ingress gateway (선택, legacy용)
```

**주의**: `istio/gateway` chart는 Helm 스키마 이슈가 있음 (GitHub #57354).
Ambient 모드에서는 Kubernetes Gateway API + Waypoint를 사용하므로 필요 없음.

---

## Ambient vs Sidecar 비교

### 아키텍처 차이

```
Sidecar Mode (Legacy)              Ambient Mode (Recommended)
┌─────────────────────┐            ┌─────────────────────┐
│ Pod                 │            │ Pod                 │
│ ┌─────┐ ┌────────┐  │            │ ┌─────────────────┐ │
│ │ App │ │ Envoy  │  │            │ │      App        │ │
│ │     │ │Sidecar │  │            │ │   (No Sidecar)  │ │
│ └─────┘ └────────┘  │            │ └─────────────────┘ │
└─────────────────────┘            └──────────┬──────────┘
                                              │
                                   ┌──────────▼──────────┐
                                   │      ztunnel        │
                                   │   (Node DaemonSet)  │
                                   └─────────────────────┘
```

### 기능 비교

| 기능 | Sidecar | Ambient (ztunnel) | Ambient (+ Waypoint) |
|------|---------|-------------------|----------------------|
| mTLS | ✅ | ✅ | ✅ |
| L4 AuthorizationPolicy | ✅ | ✅ | ✅ |
| L7 AuthorizationPolicy | ✅ | ❌ | ✅ |
| VirtualService Routing | ✅ | ❌ | ✅ |
| Retry/Timeout | ✅ | ❌ | ✅ |
| JWT Validation | ✅ | ❌ | ✅ |
| 리소스 사용 | Pod당 ~128MB | Node당 공유 | 네임스페이스당 1개 |
| Pod 재시작 필요 | ✅ | ❌ | ❌ |

---

## 설치 순서

### 1. Gateway API CRDs 설치 (필수)

```bash
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.0/standard-install.yaml
```

### 2. Istio Ambient 설치

```bash
# Helm repo 추가
helm repo add istio https://istio-release.storage.googleapis.com/charts
helm repo update

# 순서대로 설치 (의존성 있음)
helm upgrade --install istio-base istio/base -n istio-system --create-namespace --version 1.24.0
helm upgrade --install istio-cni istio/cni -n istio-system --version 1.24.0 --set profile=ambient
helm upgrade --install istiod istio/istiod -n istio-system --version 1.24.0 --set profile=ambient
helm upgrade --install ztunnel istio/ztunnel -n istio-system --version 1.24.0
```

### 3. 네임스페이스 라벨링

```bash
# Ambient 모드 활성화
kubectl label namespace <namespace> istio.io/dataplane-mode=ambient

# 기존 sidecar injection 라벨 제거 (있으면)
kubectl label namespace <namespace> istio-injection-
```

### 4. Waypoint Proxy 배포 (L7 기능 필요시)

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: <namespace>-waypoint
  namespace: <namespace>
  labels:
    istio.io/waypoint-for: service
spec:
  gatewayClassName: istio-waypoint
  listeners:
  - name: mesh
    port: 15008
    protocol: HBONE
    allowedRoutes:
      namespaces:
        from: Same
```

---

## 핵심 컴포넌트

### ztunnel (Zero Trust Tunnel)

- **역할**: L4 mTLS, 기본 인증/인가
- **배포**: DaemonSet (각 노드당 1개)
- **네임스페이스**: istio-system
- **포트**: 15008 (HBONE), 15020 (health)

```bash
# 상태 확인
kubectl get pods -n istio-system -l app=ztunnel

# 로그 확인 (트래픽 로그)
kubectl logs -l app=ztunnel -n istio-system --tail=50
```

### Waypoint Proxy

- **역할**: L7 기능 (VirtualService, JWT, 재시도)
- **배포**: Deployment (네임스페이스당 1개)
- **GatewayClass**: istio-waypoint

```bash
# 상태 확인
kubectl get gateway -n <namespace> -l istio.io/waypoint-for

# Waypoint pod 확인
kubectl get pods -n <namespace> -l gateway.networking.k8s.io/gateway-name
```

### istiod

- **역할**: Control plane, 설정 배포
- **배포**: Deployment
- **네임스페이스**: istio-system

---

## 트러블슈팅

### 문제 1: Gateway API CRD 없음

```
error: no matches for kind "Gateway" in version "gateway.networking.k8s.io/v1"
```

**해결**:
```bash
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.0/standard-install.yaml
```

### 문제 2: istio/gateway chart 스키마 오류

```
additional properties '_internal_defaults_do_not_set' not allowed
```

**원인**: Helm 스키마 검증 이슈 (GitHub #57354)
**해결**: Ambient 모드에서는 legacy gateway 불필요. Kubernetes Gateway API 사용.

### 문제 3: 서비스 간 통신 실패 (401/403)

```
error="allow policies exist, but none allowed"
```

**원인**: AuthorizationPolicy가 트래픽 차단
**해결**: 해당 서비스에 대한 AuthorizationPolicy ALLOW 규칙 추가

### 문제 4: mTLS 미적용 확인

```bash
# ztunnel 로그에서 SPIFFE ID 확인
kubectl logs -l app=ztunnel -n istio-system | grep "identity"

# 정상 예시:
# src.identity="spiffe://cluster.local/ns/<ns>/sa/<sa>"
# dst.identity="spiffe://cluster.local/ns/<ns>/sa/<sa>"
```

### 문제 5: Waypoint가 Programmed 상태가 아님

```bash
kubectl get gateway -n <namespace>
# PROGRAMMED = False
```

**확인**:
```bash
kubectl describe gateway <name> -n <namespace>
kubectl get pods -n <namespace> -l gateway.networking.k8s.io/gateway-name=<name>
```

---

## Makefile 명령어 (wealist 프로젝트)

```bash
make istio-install-ambient    # Istio Ambient 설치 (권장)
make istio-install            # Istio Sidecar 설치 (legacy)
make istio-install-gateway    # Ingress Gateway 설치 (선택)
make istio-install-config     # istio-config chart 설치
make istio-label-ns-ambient   # 네임스페이스 Ambient 라벨링
make istio-label-ns           # 네임스페이스 Sidecar 라벨링
make istio-status             # Istio 상태 확인
make istio-uninstall          # Istio 제거
```

---

## 관련 파일 (wealist 프로젝트)

```
makefiles/helm.mk                              # Istio 설치 명령어
helm/charts/istio-config/                      # Istio 설정 차트
├── templates/
│   ├── waypoint.yaml                          # Waypoint Proxy
│   ├── gateway.yaml                           # Gateway (Ingress)
│   ├── virtualservice.yaml                    # 라우팅 규칙
│   ├── destination-rules.yaml                 # 로드밸런싱, 서킷브레이커
│   ├── peer-authentication.yaml               # mTLS 설정
│   ├── authorization-policy.yaml              # 접근 제어
│   ├── request-authentication.yaml            # JWT 검증
│   └── telemetry.yaml                         # 메트릭, 로깅
└── values.yaml                                # 설정값

helm/environments/
├── base.yaml                                  # 공통 설정
└── local-kind.yaml                            # Kind 환경 설정

docs/ISTIO_OBSERVABILITY.md                    # 메트릭, 로깅 가이드
```

---

## 참고 자료

- [Istio Ambient Getting Started](https://istio.io/latest/docs/ambient/getting-started/)
- [Istio Ambient Install with Helm](https://istio.io/latest/docs/ambient/install/helm/)
- [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/)
- [Istio GitHub Issues](https://github.com/istio/istio/issues)

---

## 변경 이력

| 날짜 | 버전 | 내용 |
|------|------|------|
| 2025-12-18 | 1.0 | 최초 작성. Istio 1.24.0 기준 |
