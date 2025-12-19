# Wealist 로컬 개발 가이드 (Local Development Guide)

이 레포지토리에서 지원하는 로컬 개발 환경 설정 방법을 설명합니다.

## 지원 모드 (Supported Modes)

| 모드 | 명령어 | 용도 |
|------|--------|------|
| **Docker Compose** | `make dev-up` | 빠른 개발/테스트 |
| **Kind + 내부 DB** | `make kind-local-all` | K8s 환경 테스트 |
| **Kind + 외부 DB** | `make kind-local-all EXTERNAL_DB=true` | 프로덕션 유사 환경 |

---

## 1. Docker Compose 모드

가장 간단한 로컬 개발 환경입니다. 모든 서비스가 Docker 컨테이너로 실행됩니다.

### 시작하기

```bash
# 전체 서비스 시작
make dev-up

# 서비스 중지
make dev-down

# 로그 확인
make dev-logs
```

### 특징

- ✅ 빠른 시작 (1-2분)
- ✅ 별도 설정 불필요
- ✅ Hot reload 지원
- ❌ Kubernetes 기능 미지원 (Ingress, ConfigMap 등)

### 접속 URL

| 서비스 | URL |
|--------|-----|
| Frontend | http://localhost:3000 |
| API Gateway | http://localhost:8080 |
| PostgreSQL | localhost:5432 |
| Redis | localhost:6379 |

---

## 2. Kind + 내부 DB 모드

Kubernetes 환경에서 테스트하고 싶을 때 사용합니다. PostgreSQL과 Redis가 클러스터 내 Pod로 실행됩니다.

### 시작하기

```bash
# 전체 설정 (클러스터 생성 → 이미지 빌드 → Helm 설치)
make kind-local-all
```

### 실행 흐름

```
1. 시크릿 파일 확인 (local-kind-secrets.yaml)
   ↓
2. 설정 확인 [Y/n] 프롬프트
   ↓
3. Kind 클러스터 생성
   ↓
4. Docker 이미지 빌드 및 로드
   ↓
5. Helm 차트 설치 (Infrastructure + Services + Monitoring)
   ↓
6. 상태 확인
```

### 특징

- ✅ 실제 Kubernetes 환경
- ✅ Ingress, ConfigMap, Secret 테스트 가능
- ✅ 모니터링 스택 포함 (Prometheus, Grafana, Loki)
- ✅ 별도 DB 설치 불필요
- ❌ 데이터 영속성 없음 (클러스터 삭제 시 데이터 손실)

### 접속 URL

| 서비스 | URL |
|--------|-----|
| Frontend | http://localhost |
| API | http://localhost/api/* |
| Grafana | http://localhost/monitoring/grafana |
| Prometheus | http://localhost/monitoring/prometheus |

### 유용한 명령어

```bash
# 클러스터 상태 확인
make kind-status-all

# 모니터링 연결 확인
make kind-verify-monitoring

# 특정 서비스 재배포
make auth-service-all

# 클러스터 삭제
make kind-delete
```

---

## 3. Kind + 외부 DB 모드 (EXTERNAL_DB=true)

호스트 머신의 PostgreSQL/Redis를 사용합니다. 데이터가 영속적으로 유지되어 프로덕션과 유사한 환경을 테스트할 수 있습니다.

### 사전 요구사항

- PostgreSQL 14+ 설치 및 실행
- Redis 7+ 설치 및 실행

### 시작하기

```bash
# 전체 설정 (DB 확인 → 클러스터 생성 → 이미지 빌드 → Helm 설치)
make kind-local-all EXTERNAL_DB=true
```

### 실행 흐름

```
1. 시크릿 파일 확인 (local-kind-secrets.yaml)
   ↓
2. 설정 확인 [Y/n] 프롬프트
   ↓
3. PostgreSQL/Redis 확인 및 설정 (자동)
   - 설치 여부 확인
   - 네트워크 설정 (0.0.0.0 바인딩)
   - 데이터베이스 및 사용자 생성
   ↓
4. Kind 클러스터 생성
   ↓
5. Docker 이미지 빌드 및 로드
   ↓
6. Helm 차트 설치
   ↓
7. 상태 확인
```

### 자동 생성되는 데이터베이스

| 서비스 | DB 이름 | 사용자 | 비밀번호 |
|--------|---------|--------|----------|
| user-service | wealist_user_service_db | user_service | user_service_password |
| board-service | wealist_board_service_db | board_service | board_service_password |
| chat-service | wealist_chat_service_db | chat_service | chat_service_password |
| noti-service | wealist_noti_service_db | noti_service | noti_service_password |
| storage-service | wealist_storage_service_db | storage_service | storage_service_password |
| video-service | wealist_video_service_db | video_service | video_service_password |

### 특징

- ✅ 데이터 영속성 (클러스터 재생성해도 데이터 유지)
- ✅ 프로덕션 유사 환경
- ✅ DB 디버깅 용이 (호스트에서 직접 접근)
- ❌ PostgreSQL/Redis 설치 필요

### macOS vs Linux DB 호스트

| OS | Kind Pod에서 호스트 접근 경로 |
|----|------------------------------|
| macOS | `host.docker.internal` |
| Linux | `172.18.0.1` (Kind 브릿지 게이트웨이) |

> 자동으로 OS를 감지하여 적절한 호스트를 설정합니다.

---

## 시크릿 설정 (Secrets Configuration)

`make kind-local-all` 실행 시 자동으로 시크릿 파일을 확인합니다.

### 시크릿 파일 위치

```
k8s/helm/environments/local-kind-secrets.yaml
```

### 필수 설정 항목

```yaml
shared:
  secrets:
    # Google OAuth2 (Google 로그인용)
    GOOGLE_CLIENT_ID: "your-client-id.apps.googleusercontent.com"
    GOOGLE_CLIENT_SECRET: "your-client-secret"

    # JWT Secret
    JWT_SECRET: "your-jwt-secret-at-least-64-characters-long"
```

### Google OAuth2 설정 방법

1. [Google Cloud Console](https://console.cloud.google.com/apis/credentials) 접속
2. OAuth 2.0 클라이언트 ID 생성
3. 승인된 리디렉션 URI 추가:
   - `http://localhost/login/oauth2/code/google`
4. 발급받은 Client ID/Secret을 시크릿 파일에 입력

---

## 모니터링 (Monitoring)

Kind 모드에서는 모니터링 스택이 자동으로 설치됩니다.

### 포함된 도구

| 도구 | 용도 | 접속 URL |
|------|------|----------|
| **Grafana** | 대시보드 | http://localhost/monitoring/grafana |
| **Prometheus** | 메트릭 수집 | http://localhost/monitoring/prometheus |
| **Loki** | 로그 수집 | http://localhost/monitoring/loki |
| **Promtail** | 로그 전송 | (내부 사용) |

### Grafana 로그인

- **Username:** admin
- **Password:** admin

### 포트 포워딩 (직접 접근)

```bash
# 개별 포트 포워딩
make port-forward-grafana     # localhost:3001
make port-forward-prometheus  # localhost:9090
make port-forward-loki        # localhost:3100

# 전체 포트 포워딩 (백그라운드)
make port-forward-monitoring
```

---

## 문제 해결 (Troubleshooting)

### Pod가 CrashLoopBackOff 상태일 때

```bash
# 로그 확인
kubectl logs -l app=<service-name> -n wealist-kind-local --tail=50

# 예시
kubectl logs -l app=chat-service -n wealist-kind-local --tail=50
```

### DB 연결 실패 (EXTERNAL_DB=true)

```bash
# DB 연결 테스트
make kind-status-all

# PostgreSQL 상태 확인 (macOS)
brew services list | grep postgresql

# Redis 상태 확인 (macOS)
brew services list | grep redis
```

### 클러스터 재시작 후 복구

```bash
# 컨테이너 재시작 및 kubeconfig 복구
make kind-recover
```

### 전체 초기화

```bash
# 클러스터 삭제
make kind-delete

# 다시 시작
make kind-local-all
```

---

## 명령어 요약 (Command Summary)

### Docker Compose

| 명령어 | 설명 |
|--------|------|
| `make dev-up` | 서비스 시작 |
| `make dev-down` | 서비스 중지 |
| `make dev-logs` | 로그 확인 |
| `make dev-rebuild` | 이미지 재빌드 후 시작 |

### Kind

| 명령어 | 설명 |
|--------|------|
| `make kind-local-all` | 전체 설정 (내부 DB) |
| `make kind-local-all EXTERNAL_DB=true` | 전체 설정 (외부 DB) |
| `make kind-status-all` | 상태 확인 |
| `make kind-verify-monitoring` | 모니터링 연결 확인 |
| `make kind-delete` | 클러스터 삭제 |
| `make kind-recover` | 클러스터 복구 |

### 서비스별

| 명령어 | 설명 |
|--------|------|
| `make <service>-build` | Docker 이미지 빌드 |
| `make <service>-load` | 이미지 빌드 및 레지스트리 푸시 |
| `make <service>-redeploy` | K8s 재배포 |
| `make <service>-all` | 빌드 + 푸시 + 재배포 |

**서비스 목록:** auth-service, user-service, board-service, chat-service, noti-service, storage-service, video-service, frontend

---

## 추가 정보

- Helm 차트 위치: `k8s/helm/charts/`
- 환경 설정: `k8s/helm/environments/`
- 스크립트: `docker/scripts/dev/`
