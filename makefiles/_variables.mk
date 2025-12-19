# =============================================================================
# Common Variables (Local Development Only)
# =============================================================================
# Supports: docker-compose, localhost (Kind with internal/external DB)
# =============================================================================

# Kind cluster configuration
KIND_CLUSTER ?= wealist
LOCAL_REGISTRY ?= localhost:5001
IMAGE_TAG ?= latest

# External Database Configuration
# false (default): Deploy PostgreSQL/Redis as pods inside cluster
# true: Use host machine's PostgreSQL/Redis (requires local installation)
EXTERNAL_DB ?= false

# Environment is always local-kind for this simplified setup
ENV = local-kind
K8S_NAMESPACE = wealist-kind-local
DOMAIN = localhost
PROTOCOL = http

# Helm values file paths
HELM_BASE_VALUES = ./k8s/helm/environments/base.yaml
HELM_ENV_VALUES = ./k8s/helm/environments/local-kind.yaml
HELM_SECRETS_VALUES = ./k8s/helm/environments/local-kind-secrets.yaml

# Conditionally add secrets file if it exists
HELM_SECRETS_FLAG = $(shell test -f $(HELM_SECRETS_VALUES) && echo "-f $(HELM_SECRETS_VALUES)")

# Services list (all microservices)
BACKEND_SERVICES = auth-service user-service board-service chat-service noti-service storage-service video-service

# Frontend (deployed in local Kind environment)
FRONTEND_SERVICE = frontend

# All services for local development
SERVICES = $(BACKEND_SERVICES) $(FRONTEND_SERVICE)

# Services with project root build context (use shared package)
ROOT_CONTEXT_SERVICES = chat-service noti-service storage-service user-service video-service

# Services with local build context
LOCAL_CONTEXT_SERVICES = auth-service board-service frontend
