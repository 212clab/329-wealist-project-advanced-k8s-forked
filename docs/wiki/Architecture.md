# Architecture Overview

weAlist의 전체 시스템 아키텍처입니다.

---

## AWS Architecture

![AWS Architecture](https://raw.githubusercontent.com/OrangesCloud/wealist-project-advanced-k8s/main/docs/images/wealist_aws_arch_v2.png)

### AWS Components

| Category | Service | Description |
|----------|---------|-------------|
| **Compute** | Amazon EKS | Kubernetes 클러스터 (Spot Instances) |
| **Database** | Amazon RDS | PostgreSQL 17 (db.t4g.small) |
| **Cache** | Amazon ElastiCache | Redis 7.2 (cache.t4g.small) |
| **Storage** | Amazon S3 | 파일 스토리지, Loki/Tempo 백엔드 |
| **CDN** | CloudFront | 정적 파일 배포 (Frontend) |
| **Networking** | VPC, ALB, NAT | Private Subnet, Load Balancing |
| **Security** | Secrets Manager, IAM | 시크릿 관리, Pod Identity |

### Cost Optimization

| Resource | Strategy |
|----------|----------|
| EKS | 100% Spot Instances (월 ~$50) |
| RDS | Single-AZ for dev, Multi-AZ for prod |
| NAT | Single NAT Gateway |
| Scheduled Scaling | 야간/주말 노드 0대로 스케일 다운 |

### Terraform Infrastructure (IaC)

2-layer 아키텍처로 인프라를 관리합니다.

```
terraform/prod/
├── foundation/     # Layer 1: VPC, RDS, Redis, ECR, S3, KMS
└── compute/        # Layer 2: EKS, Istio, ArgoCD, Pod Identity
```

| Layer | 리소스 | 설명 |
|-------|--------|------|
| **Foundation** | VPC | 10.0.0.0/16, Public/Private Subnets |
| | RDS | PostgreSQL 17, Secrets Manager 연동 |
| | ElastiCache | Redis 7.2 |
| | S3, ECR | 파일 스토리지, 컨테이너 레지스트리 |
| **Compute** | EKS | Kubernetes 1.34, Managed Node Groups |
| | Helm Releases | Istio, AWS LB Controller, ESO |
| | ArgoCD | App of Apps 패턴 |
| | Pod Identity | S3, Secrets Manager 접근 권한 |

> 상세 설계: [클라우드 설계/아키텍처 (Google Docs)](https://docs.google.com/document/d/1K2L1s3t15OCGDkmCfuXjLbpeDbREeuoT1OP1ldCSGY8)

---

## Microservices Architecture

![Microservices](https://raw.githubusercontent.com/OrangesCloud/wealist-project-advanced-k8s/main/docs/images/wealist_microservices.png)

### 서비스 구성

| Service | Tech | Port | Description |
|---------|------|------|-------------|
| **auth-service** | Spring Boot 3 | 8080 | JWT 토큰 관리, OAuth2 인증 (Redis only) |
| **user-service** | Go + Gin | 8081 | 사용자, 워크스페이스 관리 |
| **board-service** | Go + Gin | 8000 | 프로젝트, 보드, 댓글 관리 |
| **chat-service** | Go + Gin | 8001 | 실시간 메시징 (WebSocket) |
| **noti-service** | Go + Gin | 8002 | 푸시 알림 (SSE) |
| **storage-service** | Go + Gin | 8003 | 파일 스토리지 (S3/MinIO) |
| **ops-service** | Go + Gin | 8004 | 운영 대시보드 (메트릭, 로그 조회) |

---

## Kubernetes Workloads

![K8s Workloads](https://raw.githubusercontent.com/OrangesCloud/wealist-project-advanced-k8s/main/docs/images/wealist_k8s_workloads.png)

> K8s 플랫폼 상세: [Architecture-K8s](Architecture-K8s)

---

## Infrastructure

| Component | Technology | Description |
|-----------|------------|-------------|
| **Database** | PostgreSQL 17 | 6개 DB (서비스별 분리) |
| **Cache** | Redis 7.2 | 캐시, 토큰 저장소 |
| **Object Storage** | S3 / MinIO | 파일 스토리지 |
| **API Gateway** | Istio Ingress Gateway | VirtualService 기반 라우팅 |
| **Service Mesh** | Istio 1.28 | mTLS, AuthorizationPolicy |
| **Monitoring** | Prometheus + Grafana + Loki + OTEL + Tempo | 메트릭/로그/트레이싱 |

---

## Service Communication

![Service Communication](https://raw.githubusercontent.com/OrangesCloud/wealist-project-advanced-k8s/main/docs/images/wealist_service_communication.png)

### Istio Gateway 라우팅 (Backend API)

| Path | Service | Port |
|------|---------|------|
| `/svc/auth/*` | auth-service | 8080 |
| `/svc/user/*` | user-service | 8081 |
| `/svc/board/*` | board-service | 8000 |
| `/svc/chat/*` | chat-service | 8001 |
| `/svc/noti/*` | noti-service | 8002 |
| `/svc/storage/*` | storage-service | 8003 |
| `/svc/ops/*` | ops-service | 8004 |

### Frontend (Static)

| Path | Service | Description |
|------|---------|-------------|
| `/*` | CloudFront + S3 | 정적 파일 (Istio 미경유) |

### Internal Communication

- **External**: JWT Bearer token in `Authorization` header (Istio RequestAuthentication 검증)
- **Internal**: mTLS로 암호화, AuthorizationPolicy로 접근 제어
- **Service Discovery**: Kubernetes DNS (`{service}.{namespace}.svc.cluster.local`)

---

## Related Pages

- [Kubernetes Architecture](Architecture-K8s)
- [CI/CD Pipeline](Architecture-CICD)
- [Security (VPC)](Architecture-VPC)
- [Monitoring Stack](Architecture-Monitoring)
