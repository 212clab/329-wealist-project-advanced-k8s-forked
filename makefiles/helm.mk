# =============================================================================
# Helm Chart Commands (Local Development Only)
# =============================================================================
# Supports: docker-compose, localhost (Kind with internal/external DB)
# =============================================================================

##@ Helm Charts

.PHONY: helm-deps-build helm-lint
.PHONY: helm-install-infra helm-install-services helm-install-monitoring
.PHONY: helm-install-all helm-upgrade-all helm-uninstall-all

# Detect OS for external DB host
# macOS (Darwin) uses host.docker.internal, Linux uses 172.18.0.1 (Kind bridge gateway)
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    EXTERNAL_DB_HOST := host.docker.internal
else
    EXTERNAL_DB_HOST := 172.18.0.1
endif

# All services including frontend (local development)
HELM_SERVICES = $(SERVICES)

helm-deps-build: ## Build all Helm dependencies
	@echo "Building all Helm dependencies..."
	@helm dependency update ./k8s/helm/charts/wealist-common 2>/dev/null || true
	@for chart in $(HELM_SERVICES); do \
		echo "Updating $$chart dependencies..."; \
		helm dependency update ./k8s/helm/charts/$$chart; \
	done
	@helm dependency update ./k8s/helm/charts/wealist-infrastructure
	@echo "All dependencies built!"

helm-lint: ## Lint all Helm charts
	@echo "Linting all Helm charts..."
	@helm lint ./k8s/helm/charts/wealist-common
	@helm lint ./k8s/helm/charts/wealist-infrastructure
	@for service in $(HELM_SERVICES); do \
		echo "Linting $$service..."; \
		helm lint ./k8s/helm/charts/$$service; \
	done
	@echo "All charts linted successfully!"

##@ Helm Installation

helm-install-infra: ## Install infrastructure chart (EXTERNAL_DB controls DB deployment)
	@echo "Installing infrastructure (ENV=$(ENV), NS=$(K8S_NAMESPACE), EXTERNAL_DB=$(EXTERNAL_DB))..."
ifeq ($(EXTERNAL_DB),true)
	@echo "Using EXTERNAL database (host: $(EXTERNAL_DB_HOST))"
	helm upgrade --install wealist-infrastructure ./k8s/helm/charts/wealist-infrastructure \
		-f $(HELM_BASE_VALUES) \
		-f $(HELM_ENV_VALUES) $(HELM_SECRETS_FLAG) \
		--set postgres.enabled=false \
		--set postgres.external.enabled=true \
		--set postgres.external.host=$(EXTERNAL_DB_HOST) \
		--set redis.enabled=false \
		--set redis.external.enabled=true \
		--set redis.external.host=$(EXTERNAL_DB_HOST) \
		-n $(K8S_NAMESPACE) --create-namespace
else
	@echo "Using INTERNAL database (PostgreSQL/Redis pods in cluster)"
	helm upgrade --install wealist-infrastructure ./k8s/helm/charts/wealist-infrastructure \
		-f $(HELM_BASE_VALUES) \
		-f $(HELM_ENV_VALUES) $(HELM_SECRETS_FLAG) \
		--set postgres.enabled=true \
		--set postgres.external.enabled=false \
		--set redis.enabled=true \
		--set redis.external.enabled=false \
		-n $(K8S_NAMESPACE) --create-namespace
endif
	@echo "Infrastructure installed!"

helm-install-services: ## Install all service charts
	@echo "Installing services (ENV=$(ENV), NS=$(K8S_NAMESPACE), EXTERNAL_DB=$(EXTERNAL_DB))..."
	@echo "Services to install: $(HELM_SERVICES)"
	@echo "Installing with DB_AUTO_MIGRATE=true (tables auto-created by services)"
	@for service in $(HELM_SERVICES); do \
		echo "Installing $$service..."; \
		helm upgrade --install $$service ./k8s/helm/charts/$$service \
			-f $(HELM_BASE_VALUES) \
			-f $(HELM_ENV_VALUES) $(HELM_SECRETS_FLAG) \
			--set shared.config.DB_AUTO_MIGRATE=true \
			-n $(K8S_NAMESPACE); \
	done
	@echo "All services installed!"
	@echo ""
	@echo "Next: make status"

helm-install-monitoring: ## Install monitoring stack (Prometheus, Loki, Grafana)
	@echo "Installing monitoring stack (ENV=$(ENV), NS=$(K8S_NAMESPACE), EXTERNAL_DB=$(EXTERNAL_DB))..."
ifeq ($(EXTERNAL_DB),true)
	@echo "Using EXTERNAL database exporters (host: $(EXTERNAL_DB_HOST))"
	helm upgrade --install wealist-monitoring ./k8s/helm/charts/wealist-monitoring \
		-f $(HELM_BASE_VALUES) \
		-f $(HELM_ENV_VALUES) $(HELM_SECRETS_FLAG) \
		--set global.namespace=$(K8S_NAMESPACE) \
		--set postgresExporter.config.host=$(EXTERNAL_DB_HOST) \
		--set redisExporter.config.host=$(EXTERNAL_DB_HOST) \
		-n $(K8S_NAMESPACE)
else
	@echo "Using INTERNAL database exporters (host: postgres/redis service)"
	helm upgrade --install wealist-monitoring ./k8s/helm/charts/wealist-monitoring \
		-f $(HELM_BASE_VALUES) \
		-f $(HELM_ENV_VALUES) $(HELM_SECRETS_FLAG) \
		--set global.namespace=$(K8S_NAMESPACE) \
		--set postgresExporter.config.host=postgres \
		--set redisExporter.config.host=redis \
		-n $(K8S_NAMESPACE)
endif
	@echo ""
	@echo "=============================================="
	@echo "  Monitoring Stack Installed Successfully!"
	@echo "=============================================="
	@echo ""
	@echo "  Access URLs (via Ingress):"
	@echo "    - Grafana:    $(PROTOCOL)://$(DOMAIN)/monitoring/grafana"
	@echo "    - Prometheus: $(PROTOCOL)://$(DOMAIN)/monitoring/prometheus"
	@echo "    - Loki:       $(PROTOCOL)://$(DOMAIN)/monitoring/loki"
	@echo ""
	@echo "  Grafana Login: admin / admin"
	@echo "=============================================="

_check-external-db: ## Check if external DB/Redis are accessible (internal target)
ifeq ($(EXTERNAL_DB),true)
	@echo "=== EXTERNAL_DB=true: Checking database connectivity ==="
	@echo ""
	@echo "Checking PostgreSQL..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		if ! psql -U postgres -h localhost -c "SELECT 1" >/dev/null 2>&1; then \
			echo ""; \
			echo "❌ PostgreSQL 연결 실패!"; \
			echo ""; \
			echo "🚫 설치 중단됨 (Installation cancelled)"; \
			echo ""; \
			echo "EXTERNAL_DB=true는 호스트 머신의 DB를 사용합니다."; \
			echo "먼저 다음 명령어로 DB를 설정해주세요:"; \
			echo ""; \
			echo "  make kind-local-all EXTERNAL_DB=true"; \
			echo ""; \
			exit 1; \
		fi; \
		echo "  ✓ PostgreSQL OK"; \
	else \
		if ! pg_isready -h 127.0.0.1 -p 5432 >/dev/null 2>&1; then \
			echo ""; \
			echo "❌ PostgreSQL 연결 실패!"; \
			echo ""; \
			echo "🚫 설치 중단됨 (Installation cancelled)"; \
			echo ""; \
			echo "EXTERNAL_DB=true는 호스트 머신의 DB를 사용합니다."; \
			echo "먼저 다음 명령어로 DB를 설정해주세요:"; \
			echo ""; \
			echo "  make kind-local-all EXTERNAL_DB=true"; \
			echo ""; \
			exit 1; \
		fi; \
		echo "  ✓ PostgreSQL OK"; \
	fi
	@echo "Checking Redis..."
	@if ! redis-cli -h localhost ping >/dev/null 2>&1; then \
		echo ""; \
		echo "❌ Redis 연결 실패!"; \
		echo ""; \
		echo "🚫 설치 중단됨 (Installation cancelled)"; \
		echo ""; \
		echo "EXTERNAL_DB=true는 호스트 머신의 Redis를 사용합니다."; \
		echo "먼저 다음 명령어로 설정해주세요:"; \
		echo ""; \
		echo "  make kind-local-all EXTERNAL_DB=true"; \
		echo ""; \
		exit 1; \
	fi
	@echo "  ✓ Redis OK"
	@echo ""
endif

helm-install-all: helm-deps-build _check-external-db helm-install-infra ## Install all charts (infra + services + monitoring)
	@sleep 5
	@$(MAKE) helm-install-services ENV=$(ENV) EXTERNAL_DB=$(EXTERNAL_DB)
	@sleep 3
	@$(MAKE) helm-install-monitoring ENV=$(ENV) EXTERNAL_DB=$(EXTERNAL_DB)

##@ Helm Upgrade/Uninstall

helm-upgrade-all: helm-deps-build ## Upgrade all charts
	@echo "Upgrading all charts (ENV=$(ENV), NS=$(K8S_NAMESPACE), EXTERNAL_DB=$(EXTERNAL_DB))..."
	@echo "Services to upgrade: $(HELM_SERVICES)"
ifeq ($(EXTERNAL_DB),true)
	@helm upgrade wealist-infrastructure ./k8s/helm/charts/wealist-infrastructure \
		-f $(HELM_BASE_VALUES) \
		-f $(HELM_ENV_VALUES) $(HELM_SECRETS_FLAG) \
		--set postgres.enabled=false \
		--set postgres.external.enabled=true \
		--set postgres.external.host=$(EXTERNAL_DB_HOST) \
		--set redis.enabled=false \
		--set redis.external.enabled=true \
		--set redis.external.host=$(EXTERNAL_DB_HOST) \
		-n $(K8S_NAMESPACE)
else
	@helm upgrade wealist-infrastructure ./k8s/helm/charts/wealist-infrastructure \
		-f $(HELM_BASE_VALUES) \
		-f $(HELM_ENV_VALUES) $(HELM_SECRETS_FLAG) \
		--set postgres.enabled=true \
		--set postgres.external.enabled=false \
		--set redis.enabled=true \
		--set redis.external.enabled=false \
		-n $(K8S_NAMESPACE)
endif
	@for service in $(HELM_SERVICES); do \
		echo "Upgrading $$service..."; \
		helm upgrade $$service ./k8s/helm/charts/$$service \
			-f $(HELM_BASE_VALUES) \
			-f $(HELM_ENV_VALUES) $(HELM_SECRETS_FLAG) \
			--set shared.config.DB_AUTO_MIGRATE=true \
			-n $(K8S_NAMESPACE); \
	done
ifeq ($(EXTERNAL_DB),true)
	@helm upgrade wealist-monitoring ./k8s/helm/charts/wealist-monitoring \
		-f $(HELM_BASE_VALUES) \
		-f $(HELM_ENV_VALUES) $(HELM_SECRETS_FLAG) \
		--set global.namespace=$(K8S_NAMESPACE) \
		--set postgresExporter.config.host=$(EXTERNAL_DB_HOST) \
		--set redisExporter.config.host=$(EXTERNAL_DB_HOST) \
		-n $(K8S_NAMESPACE) 2>/dev/null || echo "Monitoring not installed, skipping upgrade"
else
	@helm upgrade wealist-monitoring ./k8s/helm/charts/wealist-monitoring \
		-f $(HELM_BASE_VALUES) \
		-f $(HELM_ENV_VALUES) $(HELM_SECRETS_FLAG) \
		--set global.namespace=$(K8S_NAMESPACE) \
		--set postgresExporter.config.host=postgres \
		--set redisExporter.config.host=redis \
		-n $(K8S_NAMESPACE) 2>/dev/null || echo "Monitoring not installed, skipping upgrade"
endif
	@echo "All charts upgraded!"

helm-uninstall-all: ## Uninstall all charts
	@echo "Uninstalling all charts (ENV=$(ENV), NS=$(K8S_NAMESPACE))..."
	@# Uninstall monitoring first
	@helm uninstall wealist-monitoring -n $(K8S_NAMESPACE) 2>/dev/null || true
	@# Uninstall all services (including frontend if it was installed)
	@for service in $(SERVICES); do \
		echo "Uninstalling $$service..."; \
		helm uninstall $$service -n $(K8S_NAMESPACE) 2>/dev/null || true; \
	done
	@helm uninstall wealist-infrastructure -n $(K8S_NAMESPACE) 2>/dev/null || true
	@echo "All charts uninstalled!"

##@ Port Forwarding (Monitoring)

.PHONY: port-forward-grafana port-forward-prometheus port-forward-loki port-forward-monitoring

port-forward-grafana: ## Port forward Grafana (localhost:3001 -> 3000)
	@echo "Forwarding Grafana: http://localhost:3001"
	@echo "Press Ctrl+C to stop"
	kubectl port-forward svc/grafana -n $(K8S_NAMESPACE) 3001:3000

port-forward-prometheus: ## Port forward Prometheus (localhost:9090 -> 9090)
	@echo "Forwarding Prometheus: http://localhost:9090"
	@echo "Press Ctrl+C to stop"
	kubectl port-forward svc/prometheus -n $(K8S_NAMESPACE) 9090:9090

port-forward-loki: ## Port forward Loki (localhost:3100 -> 3100)
	@echo "Forwarding Loki: http://localhost:3100"
	@echo "Press Ctrl+C to stop"
	kubectl port-forward svc/loki -n $(K8S_NAMESPACE) 3100:3100

port-forward-monitoring: ## Port forward all monitoring services (background)
	@echo "Starting port forwarding for all monitoring services..."
	@echo ""
	@kubectl port-forward svc/grafana -n $(K8S_NAMESPACE) 3001:3000 &
	@kubectl port-forward svc/prometheus -n $(K8S_NAMESPACE) 9090:9090 &
	@kubectl port-forward svc/loki -n $(K8S_NAMESPACE) 3100:3100 &
	@echo ""
	@echo "=============================================="
	@echo "  Monitoring Services Port Forwarding Active"
	@echo "=============================================="
	@echo "  Grafana:    http://localhost:3001"
	@echo "  Prometheus: http://localhost:9090"
	@echo "  Loki:       http://localhost:3100"
	@echo "=============================================="
	@echo ""
	@echo "To stop: pkill -f 'kubectl port-forward'"
