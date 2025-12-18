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

# 5) 메트릭 수집 확인
# Prometheus UI → Status → Targets 에서 서비스들이 UP인지 확인
# 쿼리 테스트: {__name__=~".+_http_requests_total"}

# 6) 로그 수집 확인
# Grafana → Explore → Loki 선택
# 쿼리: {compose_service=~".+"}

# 7) 종료
make dev-down

```
