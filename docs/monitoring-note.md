EXTERNAL_DB=false (기본값):

PostgreSQL, Redis 파드가 클러스터 안에 배포됨
postgres.enabled=true, redis.enabled=true
DB_AUTO_MIGRATE=true로 서비스가 자동 마이그레이션
postgres, redis 이미지도 로드됨
EXTERNAL_DB=true:

PostgreSQL, Redis는 호스트 머신 사용
ExternalName Service로 host.docker.internal (macOS) 또는 172.18.0.1 (Linux) 연결
postgres, redis 이미지 로드 스킵
두 모드 다 livekit-svc, minio는 동일하게 배포

# helm(kind) 테스트 - DB 포드 포함

```
make kind-setup                      # 클러스터 생성
make kind-load-images-mono           # 이미지 빌드/로드 (프론트 포함)
make helm-install-all ENV=local-kind # 배포 (DB/Redis Pod + 서비스 + 모니터링)
```

# helm(kind) 테스트 - DB 포드 비포함

```
make kind-local-all EXTERNAL_DB=true
또는
make kind-setup
make kind-load-images-mono # 빠르게
make helm-install-all ENV=local-kind EXTERNAL_DB=true
```

```
localhost + externalDB구성:

PostgreSQL 설치 (sudo apt install postgresql)
Redis 설치 (sudo apt install redis-server)
PostgreSQL/Redis가 외부 접속 허용하도록 설정 (dev.yaml 주석 참고)
helm-install-all 하면:

Go 서비스들 → 호스트 PC의 PostgreSQL/Redis에 연결
postgres-exporter → 호스트 PC의 PostgreSQL 메트릭 수집
redis-exporter → 호스트 PC의 Redis 메트릭 수집
Prometheus → 이 Exporter들 + 서비스들 메트릭 수집
```

```
외부 DB 연결 확인 curl 명령어
1. Prometheus Targets (exporter 호스트 확인)
# postgres-exporter가 172.18.0.1 가리키는지
curl -s http://localhost/monitoring/prometheus/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job | contains("postgres")) | {job: .labels.job, target: .discoveredLabels.__address__, health: .health}'

# redis-exporter가 172.18.0.1 가리키는지
curl -s http://localhost/monitoring/prometheus/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job | contains("redis")) | {job: .labels.job, target: .discoveredLabels.__address__, health: .health}'

2. PostgreSQL Exporter 메트릭
# DB 연결 상태 (1 = 연결됨)
curl -s http://localhost/monitoring/prometheus/api/v1/query?query=pg_up | jq '.data.result[0].value[1]'

# DB 버전 확인
curl -s http://localhost/monitoring/prometheus/api/v1/query?query=pg_static | jq '.data.result[0].metric'

3. Redis Exporter 메트릭
# Redis 연결 상태 (1 = 연결됨)
curl -s http://localhost/monitoring/prometheus/api/v1/query?query=redis_up | jq '.data.result[0].value[1]'

# Redis 정보
curl -s http://localhost/monitoring/prometheus/api/v1/query?query=redis_connected_clients | jq '.data.result[0].value[1]'

4. 서비스 DB 연결 (ConfigMap 확인)
# 서비스들의 DB_HOST 설정 확인
kubectl get configmap -n wealist-kind-local -o yaml | grep -A2 "DB_HOST"

# postgres/redis Service가 외부 IP를 가리키는지
kubectl get endpoints postgres redis -n wealist-kind-local

5. 직접 Pod에서 DB 연결 테스트
# user-service에서 PostgreSQL 연결 테스트
kubectl exec -it deploy/user-service -n wealist-kind-local -- sh -c "nc -zv 172.18.0.1 5432"

# user-service에서 Redis 연결 테스트
kubectl exec -it deploy/user-service -n wealist-kind-local -- sh -c "nc -zv 172.18.0.1 6379"

한 줄 요약 체크
# 전체 상태 한번에 확인
echo "=== External DB Status ===" && \
curl -s http://localhost/monitoring/prometheus/api/v1/query?query=pg_up | jq -r '"PostgreSQL: " + .data.result[0].value[1]' && \
curl -s http://localhost/monitoring/prometheus/api/v1/query?query=redis_up | jq -r '"Redis: " + .data.result[0].value[1]'
```

# docker-compose

```
# 1) 환경 파일 확인 (이미 생성됨)
ls docker/env/.env.dev

# 2) 전체 서비스 + 모니터링 실행
make dev-up
# 또는 모노레포 빌드 (Go 서비스 빌드가 더 빠름)
make dev-mono-up

# 3) 로그 확인
make dev-logs

# 4) 접속 테스트
# Grafana: http://localhost:3001 (admin/admin)
# Prometheus: http://localhost:9090
# Loki: http://localhost:3100

# Prometheus 살아있나?
curl -s http://localhost:9090/-/ready
# Loki 살아있나?
curl -s http://localhost:3100/ready

# Promtail이 Loki로 보내는지 확인 (라벨 있으면 OK)
curl -s 'http://localhost:3100/loki/api/v1/labels' | jq .

# 5) 메트릭 수집 확인
# Prometheus UI → Status → Targets 에서 서비스들이 UP인지 확인
# 쿼리 테스트: {__name__=~".+_http_requests_total"}

# 6) 로그 수집 확인
# Grafana → Explore → Loki 선택
# 쿼리: {compose_service=~".+"}

# 7) 종료
make dev-down
```

# prometheus curl query

```
# 서비스 상태 확인 (up)
curl -s 'http://localhost:9090/api/v1/query?query=up' | jq .

# HTTP 요청 총계 (go-pkg 메트릭)
curl -s 'http://localhost:9090/api/v1/query?query={__name__=~".%2B_http_requests_total"}' | jq .

# 특정 서비스 메트릭 확인 (예: board-service)
curl -s 'http://localhost:9090/api/v1/query?query=board_service_http_requests_total' | jq .

# 전체 메트릭 이름 목록
curl -s 'http://localhost:9090/api/v1/label/__name__/values' | jq .

# Targets 상태 확인
curl -s 'http://localhost:9090/api/v1/targets' | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
```

# prometheus curl query

```
# 최근 로그 10개 (compose_service 라벨로)

curl -s 'http://localhost:3100/loki/api/v1/query_range' \
 --data-urlencode 'query={compose_service=~".+"}' \
 --data-urlencode 'limit=10' | jq .

# 특정 서비스 로그 (예: user-service)

curl -s 'http://localhost:3100/loki/api/v1/query_range' \
 --data-urlencode 'query={compose_service="user-service"}' \
 --data-urlencode 'limit=5' | jq .

# 사용 가능한 라벨 목록

curl -s 'http://localhost:3100/loki/api/v1/labels' | jq .

# compose_service 라벨 값들

curl -s 'http://localhost:3100/loki/api/v1/label/compose_service/values' | jq .

# Loki 준비 상태 확인

curl -s 'http://localhost:3100/ready'

```
