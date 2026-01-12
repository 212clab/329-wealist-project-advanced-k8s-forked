# Monitoring Stack

weAlist의 모니터링, 로깅, 분산 트레이싱 아키텍처입니다.

---

## Architecture Overview

![Monitoring Architecture](https://raw.githubusercontent.com/OrangesCloud/wealist-project-advanced-k8s/main/docs/images/wealist_k8s_monitoring.png)

LGTM Stack + OpenTelemetry 기반 통합 Observability 플랫폼입니다.

### Data Pipelines

| Pipeline | Source | Collector | Storage | 용도 |
|----------|--------|-----------|---------|------|
| **Metrics** | `/metrics`, `/actuator` | Prometheus (scrape) | AMP (prod) | RED 메트릭, SLI |
| **Traces** | OTEL SDK (OTLP) | OTEL Collector | Tempo (S3) | 분산 추적, 성능 분석 |
| **Logs** | stdout/stderr | Alloy (DaemonSet) | Loki (S3) | 에러 분석, 디버깅 |

### Key Components

| Component | Port | 역할 |
|-----------|------|------|
| **Prometheus** | 9090 | 메트릭 수집, Span Metrics 수신 |
| **Tempo** | 4317/4318 | 분산 트레이스 저장 |
| **Loki** | 3100 | 로그 저장/쿼리 |
| **Grafana** | 3000 | 시각화 (모든 데이터소스 통합) |
| **OTEL Collector** | 4317/4318 | 트레이스 처리, Span Metrics/Service Graph 생성 |
| **Kiali** | 20001 | Istio 서비스 메시 대시보드 |

---

## Metrics Collection

### Application Metrics

| Service | Endpoint | Tech Stack |
|---------|----------|------------|
| auth-service | `/actuator/prometheus` | Spring Boot Actuator |
| user-service | `/metrics` | Prometheus Go client |
| board-service | `/metrics` | Prometheus Go client |
| chat-service | `/metrics` | Prometheus Go client |
| noti-service | `/metrics` | Prometheus Go client |
| storage-service | `/metrics` | Prometheus Go client |

### Go 서비스 메트릭 (공통)

```
# HTTP Request metrics
http_requests_total{method,path,status}
http_request_duration_seconds{method,path}

# Business metrics
board_service_boards_total
chat_service_messages_total

# Runtime metrics
go_goroutines
go_memstats_alloc_bytes
```

### Istio Sidecar 메트릭

```
# Request metrics
istio_requests_total{source_app,destination_app,response_code}
istio_request_duration_milliseconds_bucket

# TCP metrics
istio_tcp_connections_opened_total
istio_tcp_connections_closed_total
```

### Infrastructure Exporters

| Exporter | Target | Metrics |
|----------|--------|---------|
| postgres-exporter | PostgreSQL | Connections, query stats |
| redis-exporter | Redis | Memory, commands, clients |
| kube-state-metrics | K8s API | Pod/Deployment state |
| node-exporter | Nodes | CPU, Memory, Disk |

---

## Distributed Tracing (OpenTelemetry)

![OTEL Tracing Flow](https://raw.githubusercontent.com/OrangesCloud/wealist-project-advanced-k8s/main/docs/images/wealist_k8s_monitoring_otel.png)

### 동작 방식

1. **Application SDK**: 각 서비스에서 OTEL SDK가 HTTP 요청, DB 쿼리, Redis 호출을 자동 추적
2. **OTLP 전송**: HTTP/protobuf (port 4318) 또는 gRPC (port 4317)로 Collector에 전송
3. **Collector 처리**: 배치 처리, Tail Sampling, Span Metrics 생성
4. **저장 및 시각화**: Tempo에 트레이스 저장, Grafana에서 조회

### 서비스별 Instrumentation

| 서비스 | HTTP | DB | Redis | SDK |
|--------|------|-----|-------|-----|
| auth (Spring) | Micrometer | - | - | Spring OTEL |
| Go 서비스 (5개) | otelgin | gorm-otel | redisotel | go.opentelemetry.io |

Go 서비스는 HTTP, GORM, Redis 모두 자동 트레이싱됩니다. 별도 코드 없이 미들웨어 등록만으로 전체 요청 흐름을 추적할 수 있습니다.

### Tail Sampling (Production)

Production 환경에서는 트레이스 볼륨 최적화를 위해 Tail Sampling 적용:

| Policy | Sampling Rate | 조건 |
|--------|--------------|------|
| Error traces | 100% | status_code == ERROR |
| Slow traces | 100% | latency > 1s |
| Normal traces | 10% | 기타 |

### OTEL Connectors

OTEL Collector의 Connector는 트레이스 데이터를 메트릭으로 변환합니다. 별도 설정 없이 트레이스만 전송하면 RED 메트릭이 자동 생성됩니다.

**Span Metrics Connector**

트레이스의 각 span에서 RED(Rate, Errors, Duration) 메트릭을 추출합니다:

| 메트릭 | 설명 | 라벨 |
|--------|------|------|
| `traces_spanmetrics_calls_total` | 서비스 호출 수 | service_name, operation, status_code |
| `traces_spanmetrics_latency_bucket` | 지연시간 히스토그램 | service_name, operation |

**Service Graph Connector**

서비스 간 호출 관계를 추출하여 의존성 토폴로지를 파악합니다:

| 메트릭 | 설명 |
|--------|------|
| `traces_service_graph_request_total{client, server}` | 서비스 간 요청 수 |
| `traces_service_graph_request_failed_total{client, server}` | 서비스 간 실패 수 |

### Exemplars (Metrics ↔ Traces 연동)

Exemplar는 메트릭 데이터 포인트에 trace_id를 첨부하여 **메트릭에서 트레이스로 직접 이동**할 수 있게 합니다.

**활용 예시**:
1. Grafana에서 latency 스파이크 발견
2. 해당 시점의 exemplar 아이콘 클릭
3. 원인이 된 트레이스로 즉시 이동
4. 어떤 서비스/쿼리에서 지연이 발생했는지 확인

이를 통해 "무엇이 문제인가" (메트릭)에서 "왜 문제인가" (트레이스)로 빠르게 전환할 수 있습니다.

---

## Log Aggregation

![Log Collection Pipeline](https://raw.githubusercontent.com/OrangesCloud/wealist-project-advanced-k8s/main/docs/images/wealist_k8s_monitoring_logs.png)

### 동작 방식

1. **Application 로그 출력**: 모든 서비스가 stdout/stderr로 JSON 형식 로그 출력
2. **Container Runtime**: 로그를 `/var/log/containers/*.log`에 저장
3. **Alloy 수집**: DaemonSet으로 각 노드에서 로그 파일 감시 및 수집
4. **Loki 저장**: 라벨 기반으로 인덱싱하여 저장 (S3 backend in prod)
5. **Grafana 조회**: LogQL로 검색, 필터링, 집계

### Alloy Pipeline

Alloy는 Grafana Agent의 후속 제품으로, 선언적 파이프라인 구성을 제공합니다:

- `discovery.kubernetes`: K8s API로 Pod 자동 감지
- `loki.source.kubernetes`: Container 로그 수집 + 메타데이터 추출
- `loki.write`: Loki로 배치 전송

### Log Labels

| Label | Source | Example |
|-------|--------|---------|
| `namespace` | K8s metadata | wealist-prod |
| `pod` | K8s metadata | user-service-xxx |
| `container` | K8s metadata | user-service |
| `app` | Pod label | user-service |
| `level` | Log parsing | info, warn, error |

### LogQL 예시

```logql
# 특정 서비스 에러 로그
{app="board-service"} |= "error"

# 최근 5분 에러 카운트
count_over_time({namespace="wealist-prod", level="error"}[5m])

# JSON 파싱 후 필터
{app="auth-service"} | json | status >= 400

# Trace ID로 관련 로그 검색
{namespace="wealist-prod"} |= "abc123def456"
```

### Traces-Logs Correlation

Go 서비스는 로그에 자동으로 trace_id를 포함합니다:

```json
{"level":"info","trace_id":"abc123","span_id":"def456","msg":"Creating board","service":"board-service"}
```

**활용 흐름**:
1. Grafana Tempo에서 느린 트레이스 발견
2. trace_id 복사
3. Loki에서 `{app="board-service"} |= "abc123"` 검색
4. 해당 요청의 상세 로그 확인

이를 통해 트레이스의 각 span에서 어떤 작업이 수행되었는지 로그로 상세히 파악할 수 있습니다.

---

## Grafana Dashboards (V4)

### Role-based Dashboard 구조

| Dashboard | 대상 | 주요 내용 |
|-----------|------|----------|
| **Landing** | All | 서비스 선택, 빠른 링크 |
| **Service Overview** | Dev | 요청률, 에러율, 지연시간 |
| **Service Detail** | Dev | 엔드포인트별 상세, 에러 분석 |
| **Infrastructure** | Platform | DB, Redis, Node 리소스 |
| **Istio Mesh** | Platform | 서비스 메시, mTLS 상태 |
| **SRE Golden Signals** | SRE | SLI/SLO, 가용성 |
| **Alert Overview** | SRE | 알림 현황, 온콜 |

### 주요 대시보드

#### Service Overview Dashboard
- Request Rate (by service)
- Error Rate (4xx, 5xx)
- Latency Percentiles (p50, p95, p99)
- Trace 연동 (클릭 시 Tempo 이동)

#### Infrastructure Dashboard
- PostgreSQL: Connections, Query duration
- Redis: Memory, Commands/sec
- Node: CPU, Memory, Disk usage

#### Istio Service Mesh Dashboard
- Request volume by service pair
- Success rate by destination
- mTLS coverage

---

## Access URLs

### Localhost (Kind Cluster)

| Service | URL | 용도 |
|---------|-----|------|
| Grafana | `/api/monitoring/grafana` | 대시보드 |
| Prometheus | `/api/monitoring/prometheus` | 메트릭 쿼리 |
| Kiali | `/api/monitoring/kiali` | Istio 메시 |

> localhost 환경: `http://localhost:8080` 접두사 사용

### Production (AWS)

| Service | URL |
|---------|-----|
| Grafana | `https://grafana.wealist.co.kr` |
| Prometheus | Internal (AMP) |

---

## Alert Rules

### Critical Alerts

| Alert | Condition | Severity |
|-------|-----------|----------|
| ServiceDown | up == 0 for 5m | critical |
| HighErrorRate | error_rate > 5% for 5m | critical |
| HighLatency | p99 > 2s for 10m | critical |
| PodCrashLooping | restart > 5 in 15m | critical |

### Warning Alerts

| Alert | Condition | Severity |
|-------|-----------|----------|
| HighMemory | memory > 85% for 10m | warning |
| HighCPU | cpu > 80% for 10m | warning |
| DiskPressure | disk > 85% | warning |
| PendingPods | pending > 0 for 10m | warning |

---

## Environment Comparison

| Feature | Docker Compose | Localhost (Kind) | Production (EKS) |
|---------|---------------|------------------|------------------|
| Prometheus | ✅ Local | ✅ Local | ✅ AMP |
| Loki | ✅ Local | ✅ Local | ✅ S3 backend |
| Tempo | ✅ Local | ✅ Local | ✅ S3 backend |
| Grafana | ✅ Port 3001 | ✅ Port 3000 | ✅ ALB Ingress |
| OTEL Collector | ✅ | ✅ | ✅ |
| Alloy | ❌ | ✅ | ✅ |
| Tail Sampling | ❌ 100% | ❌ 100% | ✅ 10% |
| Kiali | ❌ | ✅ | ✅ |
| Span Metrics | ✅ | ✅ | ✅ |

---

## Troubleshooting

### 메트릭 수집 안됨

```bash
# Prometheus targets 확인
kubectl port-forward svc/prometheus 9090:9090 -n wealist-localhost
# http://localhost:9090/targets

# 서비스 메트릭 직접 확인
kubectl exec deploy/user-service -n wealist-localhost -- \
  wget -qO- http://localhost:8081/metrics
```

### 트레이스 안 보임

```bash
# OTEL Collector 로그
kubectl logs deploy/otel-collector -n wealist-localhost

# 서비스 OTEL 환경변수 확인
kubectl exec deploy/board-service -n wealist-localhost -- env | grep OTEL
```

### 로그 수집 안됨

```bash
# Alloy 상태 확인
kubectl logs ds/alloy -n wealist-localhost

# Loki 연결 테스트
kubectl exec ds/alloy -n wealist-localhost -- \
  wget -qO- http://loki:3100/ready
```

---

## Related Pages

- [Architecture Overview](Architecture)
- [Architecture-K8s](Architecture-K8s)
- [Istio Observability](.claude/docs/ISTIO_OBSERVABILITY.md)
- [ADR-005: Prometheus + Loki](ADR#adr-005-prometheus--loki-모니터링)
