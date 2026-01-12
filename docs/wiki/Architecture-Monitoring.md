# Monitoring Stack

weAlist의 모니터링, 로깅, 분산 트레이싱 아키텍처입니다.

---

## Architecture Overview

![Monitoring Architecture](https://raw.githubusercontent.com/OrangesCloud/wealist-project-advanced-k8s/main/docs/images/wealist_k8s_monitoring.png)

LGTM Stack (Loki, Grafana, Tempo, Mimir/Prometheus) + OpenTelemetry 기반의 통합 관측성(Observability) 플랫폼입니다.

---

## Stack Components

| Component | Role | Port | Storage (Prod) |
|-----------|------|------|----------------|
| **Prometheus** | 메트릭 수집/저장 | 9090 | AMP (AWS Managed Prometheus) |
| **Tempo** | 분산 트레이싱 | 4317/4318 | S3 |
| **Loki** | 로그 수집/저장 | 3100 | S3 |
| **Grafana** | 시각화 대시보드 | 3000 | - |
| **OTEL Collector** | 트레이스 수집/처리 | 4317/4318 | - |
| **Alloy** | 로그 에이전트 (K8s) | - | - |
| **Kiali** | Istio 서비스 메시 대시보드 | 20001 | - |

---

## Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Applications (6 Services)                     │
│  auth(Spring) / user / board / chat / noti / storage (Go)       │
└─────────────────────────────────────────────────────────────────┘
         │                    │                         │
         │ /metrics           │ OTLP traces             │ stdout
         │ /actuator          │                         │
         ▼                    ▼                         ▼
┌─────────────┐     ┌─────────────────┐          ┌─────────┐
│ Prometheus  │     │  OTEL Collector │          │  Alloy  │
│  (scrape)   │     │    (gateway)    │          │(K8s log)│
└─────────────┘     └─────────────────┘          └─────────┘
         │                    │                         │
         │                    ├─── spanmetrics ──▶ Prometheus
         │                    │                         │
         ▼                    ▼                         ▼
┌─────────────┐     ┌─────────────────┐          ┌─────────┐
│    AMP      │     │      Tempo      │          │   Loki  │
│   (prod)    │     │    (traces)     │          │  (logs) │
└─────────────┘     └─────────────────┘          └─────────┘
         └─────────────────┬───────────────────────────┘
                           │
                           ▼
                   ┌─────────────────┐
                   │     Grafana     │
                   │  (visualization)│
                   └─────────────────┘
```

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

### Architecture

```
Application (OTEL SDK)
    │
    │ HTTP/protobuf (port 4318)
    │ gRPC/protobuf (port 4317)
    ▼
OTEL Collector (Gateway)
    │
    ├── Processors
    │   └── Tail Sampling (prod)
    │
    ├── Connectors
    │   ├── spanmetrics → Prometheus
    │   └── servicegraph → Prometheus
    │
    └── Exporters
        └── otlp → Tempo
```

### Tail Sampling (Production)

Production 환경에서는 트레이스 볼륨 최적화를 위해 Tail Sampling 적용:

| Policy | Sampling Rate | 조건 |
|--------|--------------|------|
| Error traces | 100% | status_code == ERROR |
| Slow traces | 100% | latency > 1s |
| Normal traces | 10% | 기타 |

### Span Metrics Connector

트레이스에서 RED 메트릭 자동 추출:

```
# 생성되는 메트릭
traces_spanmetrics_calls_total{service_name, operation, status_code}
traces_spanmetrics_latency_bucket{service_name, operation}
```

### Service Graph Connector

서비스 간 호출 관계를 메트릭으로 추출:

```
# 서비스 토폴로지 메트릭
traces_service_graph_request_total{client, server}
traces_service_graph_request_failed_total{client, server}
```

### Exemplars

Prometheus 메트릭과 Tempo 트레이스 연동:

- Grafana에서 메트릭 쿼리 시 exemplar 아이콘 클릭
- 해당 시점의 트레이스로 직접 이동 가능

---

## Log Aggregation

### Log Collection Pipeline (K8s)

```
Application (stdout/stderr)
    │
    ▼
Alloy (DaemonSet)
    │
    ├── discovery.kubernetes
    │   └── Pod 자동 감지
    │
    ├── loki.source.kubernetes
    │   └── Container logs 수집
    │
    └── loki.write
        └── Loki로 전송
```

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
```

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
