# Kind 클러스터 환경 가이드

## 환경 비교

| 항목 | kind-localhost | kind-dev |
|------|----------------|----------|
| **PostgreSQL** | Pod (내장) | 호스트 PC (분리) |
| **Redis** | Pod (내장) | 호스트 PC (분리) |
| **Frontend** | Pod (내장) | `npm run dev` (분리) |
| **Istio** | ✅ Ambient | ✅ Ambient |
| **Monitoring** | ✅ | ✅ |
| **용도** | 빠른 테스트, 데모 | 개발 (핫 리로드) |

---

## kind-localhost (통합 환경)

모든 컴포넌트가 클러스터 안에서 실행됩니다.

```bash
# 1. 클러스터 생성 + 이미지 로드
make kind-localhost-setup

# 2. Helm 배포
make helm-install-all ENV=localhost
```

**접속:** http://localhost

---

## kind-dev (분리 환경)

DB는 호스트 PC, Frontend는 로컬에서 실행합니다.

```bash
# 1. 클러스터 생성 + 이미지 로드 (DB 설치 확인 포함)
make kind-check-db-setup

# 2. Helm 배포
make helm-install-all ENV=dev

# 3. Frontend 실행 (별도 터미널)
cd services/frontend && npm run dev
```

**Backend:** http://localhost/svc/*
**Frontend:** http://localhost:5173

---

## 공통 명령어

```bash
# 클러스터 삭제
make kind-delete

# 클러스터 복구 (재부팅 후)
make kind-recover

# 상태 확인
kubectl get pods -n wealist-localhost  # localhost 환경
kubectl get pods -n wealist-dev        # dev 환경
```

---

## 모니터링 접속

- **Grafana:** http://localhost/monitoring/grafana
- **Prometheus:** http://localhost/monitoring/prometheus
