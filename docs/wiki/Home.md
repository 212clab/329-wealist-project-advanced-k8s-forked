# weAlist - Cloud Native Project Management Platform

Production-ready 마이크로서비스 아키텍처로 구현된 협업 프로젝트 관리 플랫폼입니다.

![Architecture Overview](https://raw.githubusercontent.com/OrangesCloud/wealist-project-advanced-k8s/main/docs/images/wealist_aws_arch_v2.png)

---

## Tech Stack

**Backend**: Go 1.24 (Gin) · Spring Boot 3.4 (Java 21)
**Frontend**: React 19 · TypeScript 5 · Vite 5
**Infrastructure**: Kubernetes · Helm · ArgoCD · Terraform
**Service Mesh**: Istio 1.28 (Sidecar mTLS)
**Observability**: Prometheus · Grafana · Loki · Tempo · OpenTelemetry
**Database**: PostgreSQL 17 · Redis 7.2
**Cloud**: AWS EKS · RDS · ElastiCache · S3 · CloudFront

---

## Key Highlights

### Production-Grade Microservices
- **7개 서비스** 독립 배포 (5 Go + 1 Spring Boot + 1 React)
- Clean Architecture 패턴으로 테스트 가능한 코드 구조
- 서비스 간 통신: gRPC-like HTTP with JWT propagation

### GitOps & Progressive Delivery
- **ArgoCD** App-of-Apps 패턴으로 선언적 배포
- **Argo Rollouts** 카나리 배포 (10% → 30% → 50% → 100%)
- Terraform으로 AWS 인프라 전체 IaC 관리

### Full Observability (LGTM Stack)
- **Metrics**: Prometheus + Istio sidecar 메트릭 자동 수집
- **Traces**: OpenTelemetry SDK → Tempo, Span Metrics 자동 생성
- **Logs**: Alloy → Loki, trace_id 연동으로 상관분석

### Service Mesh Security
- Istio **mTLS** 전 서비스 암호화 통신
- **AuthorizationPolicy**로 서비스 간 접근 제어
- **RequestAuthentication**으로 JWT 검증 오프로드

---

## Architecture Documentation

| Document | Description |
|----------|-------------|
| **[Architecture Overview](Architecture)** | 전체 시스템 아키텍처, AWS, Terraform IaC |
| **[Kubernetes Architecture](Architecture-K8s)** | EKS 클러스터, Istio, ArgoCD, Helm 구성 |
| **[CI/CD Pipeline](Architecture-CICD)** | GitHub Actions, ArgoCD GitOps 플로우 |
| **[Monitoring Stack](Architecture-Monitoring)** | LGTM Stack, OTEL, Distributed Tracing |
| **[Security (VPC)](Architecture-VPC)** | 네트워크 보안, Private Subnet 구성 |

---

## Project Documentation

| Document | Description |
|----------|-------------|
| **[Requirements](Requirements)** | 기능 요구사항, 비기능 요구사항 |
| **[Business Flow](Business-Flow)** | 사용자 시나리오, 비즈니스 플로우 |
| **[ADR](ADR)** | 아키텍처 결정 기록 (Architecture Decision Records) |
| **[Cloud Proposal](Cloud-Proposal)** | 클라우드 인프라 제안서 |

---

## Backend Services

| Service | Tech | Port | Role |
|---------|------|------|------|
| **auth-service** | Spring Boot 3 | 8080 | JWT 발급/검증, OAuth2 |
| **user-service** | Go + Gin | 8081 | 사용자, 워크스페이스 관리 |
| **board-service** | Go + Gin | 8000 | 프로젝트, 보드, 댓글 |
| **chat-service** | Go + Gin | 8001 | 실시간 메시징 (WebSocket) |
| **noti-service** | Go + Gin | 8002 | 알림 (Server-Sent Events) |
| **storage-service** | Go + Gin | 8003 | 파일 저장소 (S3/MinIO) |
| **ops-service** | Go + Gin | 8004 | 운영 대시보드 (메트릭, 로그 조회) |

---

## External Documents

| Document | Link |
|----------|------|
| 클라우드 제안서 | [Google Docs](https://docs.google.com/document/d/1DiVO6p0NjmxzoEXwG3hZ7KoZLpqU-iiSdLWmOvTuH_s) |
| 요구사항 정의서 | [Google Docs](https://docs.google.com/document/d/1Cmc4fSrtqnJRTxgARCCyQGNgOiVJ-vvIkmSqktE_hx8) |
| 클라우드 설계/아키텍처 | [Google Docs](https://docs.google.com/document/d/1K2L1s3t15OCGDkmCfuXjLbpeDbREeuoT1OP1ldCSGY8) |

---

## Team & Progress

- **[Team & Contributions](Team)** - 팀 역할 및 진행 상황

---

## Repository

**GitHub**: [OrangesCloud/wealist-project-advanced-k8s](https://github.com/OrangesCloud/wealist-project-advanced-k8s)
