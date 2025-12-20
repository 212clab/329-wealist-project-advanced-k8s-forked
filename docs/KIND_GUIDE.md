# Kind 클러스터 환경 가이드

## 환경 비교

| 항목 | kind-localhost | kind-dev |
|------|----------------|----------|
| **PostgreSQL** | Pod (내장) | 호스트 PC (분리) |
| **Redis** | Pod (내장) | 호스트 PC (분리) |
| **Frontend** | Pod (내장) | CloudFront/S3 (분리) |
| **Istio** | ✅ Ambient (완화된 보안) | ✅ Ambient (강화된 보안) |
| **Monitoring** | ✅ | ✅ |
| **용도** | 빠른 테스트, 데모 | 개발 (프로덕션 유사) |

---

## Istio 기능 비교

| 기능 | localhost | dev |
|-----|-----------|-----|
| Gateway + HTTPRoute | ✅ | ✅ |
| mTLS | PERMISSIVE | STRICT |
| DestinationRules (Circuit Breaker) | ✅ | ✅ |
| AuthorizationPolicy | 기본 허용 | Zero Trust |
| ServiceAuthorization | ❌ | ✅ |
| JWT 인증 (Istio 검증) | ❌ | ✅ |
| Ambient Mode + Waypoint | ✅ | ✅ |
| Telemetry (Tracing) | ✅ 100% | ✅ 100% |

---

## kind-localhost (통합 환경)

모든 컴포넌트가 클러스터 안에서 실행됩니다.

```bash
# 1. 클러스터 생성 + 이미지 로드
make kind-localhost-setup

# 2. Helm 배포
make helm-install-all ENV=localhost
```

**접속:** http://localhost:8080

---

## kind-dev (분리 환경)

DB는 호스트 PC, Frontend는 CloudFront/S3에서 서빙합니다.

```bash
# 1. 클러스터 생성 + 이미지 로드 (외부 DB 사용)
make kind-dev-setup

# 2. Helm 배포
make helm-install-all ENV=dev
```

**Backend:** http://localhost:8080/svc/*

---

## 공통 명령어

```bash
# 클러스터 삭제
make kind-delete

# 클러스터 복구 (재부팅 후)
make kind-recover

# 상태 확인
make status
kubectl get pods -n wealist-localhost  # localhost 환경
kubectl get pods -n wealist-dev        # dev 환경

# Istio 상태 확인
kubectl get gateway -n istio-system
kubectl get httproute -n wealist-localhost
kubectl get peerauthentication -n wealist-localhost
```

---

## 모니터링 접속

- **Grafana:** http://localhost:8080/monitoring/grafana
- **Prometheus:** http://localhost:8080/monitoring/prometheus

---

## 트러블슈팅

### DB 연결 실패 (mTLS 문제)

```bash
# PeerAuthentication 확인
kubectl get peerauthentication -n wealist-localhost

# postgres/redis에 mTLS DISABLE이 있어야 함
NAME                    MODE          AGE
default                 PERMISSIVE    5m
postgres-disable-mtls   DISABLE       5m
redis-disable-mtls      DISABLE       5m
```

### TLS 에러 (Gateway → Service)

```bash
# namespace에 ambient 라벨 확인
kubectl get namespace wealist-localhost --show-labels

# istio.io/dataplane-mode=ambient 라벨이 있어야 함
```

### Pod CrashLoopBackOff

```bash
# 로그 확인
kubectl logs -n wealist-localhost -l app=board-service --tail=50

# 서비스 재시작
kubectl rollout restart deployment -n wealist-localhost
```
