# =============================================================================
# Kubernetes (Kind) Commands
# =============================================================================

##@ Kubernetes (Kind)

.PHONY: kind-setup kind-setup-db kind-load-images kind-load-images-mono kind-delete kind-recover
.PHONY: _setup-db-macos _setup-db-debian _check-secrets
.PHONY: kind-local-all kind-check-db kind-verify-monitoring kind-status-all

# Detect OS for external DB host
# macOS (Darwin) uses host.docker.internal, Linux uses 172.18.0.1 (Kind bridge gateway)
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    EXTERNAL_DB_HOST := host.docker.internal
else
    EXTERNAL_DB_HOST := 172.18.0.1
endif

kind-setup: ## Create cluster + registry (lightweight, no local DB setup)
	@echo "=== Creating Kind cluster with local registry ==="
	./docker/scripts/dev/0.setup-cluster.sh
	@echo ""
	@echo "Cluster ready! Next: make kind-load-images"
	@echo ""
	@echo "NOTE: For full setup with external DB, use: make kind-local-all EXTERNAL_DB=true"

kind-setup-db: ## Setup local PostgreSQL/Redis for Kind
	@echo "=== Step 1: Setting up local PostgreSQL and Redis ==="
	@echo ""
	@# Detect OS
	@if [ "$$(uname)" = "Darwin" ]; then \
		echo "Detected: macOS"; \
		$(MAKE) _setup-db-macos; \
	elif [ -f /etc/debian_version ]; then \
		echo "Detected: Debian/Ubuntu"; \
		$(MAKE) _setup-db-debian; \
	else \
		echo "Unsupported OS. Please install PostgreSQL and Redis manually."; \
		echo "  PostgreSQL: listening on 0.0.0.0:5432"; \
		echo "  Redis: listening on 0.0.0.0:6379"; \
	fi
	@echo ""
	@echo "Local DB setup complete!"
	@echo ""

_setup-db-macos:
	@# PostgreSQL
	@if ! command -v psql >/dev/null 2>&1; then \
		echo "Installing PostgreSQL..."; \
		brew install postgresql@14; \
		brew services start postgresql@14; \
	else \
		echo "PostgreSQL already installed"; \
		brew services start postgresql@14 2>/dev/null || brew services start postgresql 2>/dev/null || true; \
	fi
	@# Redis
	@if ! command -v redis-cli >/dev/null 2>&1; then \
		echo "Installing Redis..."; \
		brew install redis; \
		brew services start redis; \
	else \
		echo "Redis already installed"; \
		brew services start redis 2>/dev/null || true; \
	fi
	@# Create wealist databases
	@echo "Creating wealist databases..."
	@psql -U postgres -c "SELECT 1" 2>/dev/null || createuser -s postgres 2>/dev/null || true
	@for db in wealist wealist_auth wealist_user wealist_board wealist_chat wealist_noti wealist_storage wealist_video; do \
		psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$$db'" | grep -q 1 || \
		psql -U postgres -c "CREATE DATABASE $$db" 2>/dev/null || true; \
	done
	@echo "PostgreSQL databases ready"

_setup-db-debian:
	@# PostgreSQL
	@if ! command -v psql >/dev/null 2>&1; then \
		echo "Installing PostgreSQL..."; \
		sudo apt-get update && sudo apt-get install -y postgresql postgresql-contrib; \
	else \
		echo "PostgreSQL already installed"; \
	fi
	@sudo systemctl start postgresql || true
	@# Configure PostgreSQL for external access
	@echo "Configuring PostgreSQL for Kind cluster access..."
	@PG_HBA=$$(sudo -u postgres psql -t -P format=unaligned -c "SHOW hba_file"); \
	if ! sudo grep -q "172.18.0.0/16" "$$PG_HBA" 2>/dev/null; then \
		echo "host    all    all    172.17.0.0/16    trust" | sudo tee -a "$$PG_HBA" >/dev/null; \
		echo "host    all    all    172.18.0.0/16    trust" | sudo tee -a "$$PG_HBA" >/dev/null; \
	fi
	@PG_CONF=$$(sudo -u postgres psql -t -P format=unaligned -c "SHOW config_file"); \
	sudo sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/" "$$PG_CONF" 2>/dev/null || true; \
	sudo sed -i "s/listen_addresses = 'localhost'/listen_addresses = '*'/" "$$PG_CONF" 2>/dev/null || true
	@sudo systemctl restart postgresql
	@# Create wealist databases
	@echo "Creating wealist databases..."
	@for db in wealist wealist_auth wealist_user wealist_board wealist_chat wealist_noti wealist_storage wealist_video; do \
		sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname = '$$db'" | grep -q 1 || \
		sudo -u postgres psql -c "CREATE DATABASE $$db" 2>/dev/null || true; \
	done
	@echo "PostgreSQL databases ready"
	@# Redis
	@if ! command -v redis-cli >/dev/null 2>&1; then \
		echo "Installing Redis..."; \
		sudo apt-get install -y redis-server; \
	else \
		echo "Redis already installed"; \
	fi
	@# Configure Redis for external access
	@echo "Configuring Redis for Kind cluster access..."
	@sudo sed -i 's/^bind 127.0.0.1/bind 0.0.0.0/' /etc/redis/redis.conf 2>/dev/null || true
	@sudo sed -i 's/^protected-mode yes/protected-mode no/' /etc/redis/redis.conf 2>/dev/null || true
	@sudo systemctl restart redis-server || sudo systemctl restart redis
	@echo "Redis ready"

kind-load-images: ## Build/pull all images (infra + services)
	@echo "=== Step 2: Loading all images ==="
	@echo ""
	@echo "--- Loading infrastructure images ---"
	EXTERNAL_DB=$(EXTERNAL_DB) ./docker/scripts/dev/1.load_infra_images.sh
	@echo ""
	@echo "--- Building service images ---"
	./docker/scripts/dev/2.build_services_and_load.sh
	@echo ""
	@echo "All images loaded!"
	@echo ""
	@echo "Next: make helm-install-all ENV=local-kind"

kind-load-images-mono: ## Build Go services with monorepo pattern (faster rebuilds)
	@echo "=== Loading images using Monorepo Build (BuildKit cache) ==="
	@echo ""
	@echo "--- Loading infrastructure images ---"
	EXTERNAL_DB=$(EXTERNAL_DB) ./docker/scripts/dev/1.load_infra_images.sh
	@echo ""
	@echo "--- Building Go services (monorepo pattern) ---"
	./docker/scripts/dev-mono.sh build
	@echo ""
	@echo "--- Tagging and pushing to local registry ---"
	@for svc in user-service board-service chat-service noti-service storage-service video-service; do \
		echo "Pushing $$svc..."; \
		docker tag wealist/$$svc:latest $(LOCAL_REGISTRY)/$$svc:$(IMAGE_TAG); \
		docker push $(LOCAL_REGISTRY)/$$svc:$(IMAGE_TAG); \
	done
	@echo ""
	@echo "--- Building auth-service and frontend ---"
	@$(MAKE) auth-service-load frontend-load
	@echo ""
	@echo "All images loaded! (Monorepo pattern)"
	@echo ""
	@echo "Next: make helm-install-all ENV=local-kind"

kind-delete: ## Delete cluster
	kind delete cluster --name $(KIND_CLUSTER)
	@docker rm -f kind-registry 2>/dev/null || true

kind-recover: ## Recover cluster after reboot
	@echo "Recovering Kind cluster..."
	@docker restart $(KIND_CLUSTER)-control-plane $(KIND_CLUSTER)-worker $(KIND_CLUSTER)-worker2 kind-registry 2>/dev/null || true
	@sleep 30
	@kind export kubeconfig --name $(KIND_CLUSTER)
	@echo "Waiting for API server..."
	@until kubectl get nodes >/dev/null 2>&1; do sleep 5; done
	@echo "Cluster recovered!"
	@kubectl get nodes

##@ Local Domain (local.wealist.co.kr)

.PHONY: local-tls-secret

local-tls-secret: ## Create TLS secret for local.wealist.co.kr
	@echo "=== Creating TLS secret for local.wealist.co.kr ==="
	@if kubectl get secret local-wealist-tls -n $(K8S_NAMESPACE) >/dev/null 2>&1; then \
		echo "TLS secret already exists, skipping..."; \
	else \
		echo "Generating self-signed certificate..."; \
		openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
			-keyout /tmp/local-wealist-tls.key \
			-out /tmp/local-wealist-tls.crt \
			-subj "/CN=local.wealist.co.kr/O=wealist" \
			-addext "subjectAltName=DNS:local.wealist.co.kr"; \
		kubectl create secret tls local-wealist-tls \
			--cert=/tmp/local-wealist-tls.crt \
			--key=/tmp/local-wealist-tls.key \
			-n $(K8S_NAMESPACE); \
		rm -f /tmp/local-wealist-tls.key /tmp/local-wealist-tls.crt; \
		echo "TLS secret created"; \
	fi

##@ Local Database

.PHONY: init-local-db

init-local-db: ## Init local PostgreSQL/Redis (Ubuntu, ENV=local-ubuntu)
	@echo "Initializing local PostgreSQL and Redis for Wealist..."
	@echo ""
	@echo "This will configure your local PostgreSQL and Redis to accept"
	@echo "connections from the Kind cluster (Docker network)."
	@echo ""
	@echo "Prerequisites:"
	@echo "  - PostgreSQL installed: sudo apt install postgresql postgresql-contrib"
	@echo "  - Redis installed: sudo apt install redis-server"
	@echo ""
	@echo "Running scripts with sudo..."
	@sudo ./scripts/init-local-postgres.sh
	@sudo ./scripts/init-local-redis.sh
	@echo ""
	@echo "Local database initialization complete!"
	@echo ""
	@echo "Next: make helm-install-all ENV=local-ubuntu"

##@ One-Stop Local Development

SECRETS_FILE = ./k8s/helm/environments/local-kind-secrets.yaml
SECRETS_EXAMPLE = ./k8s/helm/environments/secrets.example.yaml

_check-secrets:
	@echo ""
	@echo "=============================================="
	@echo "  시크릿 파일 확인 (Checking Secrets)"
	@echo "=============================================="
	@echo ""
	@# Create secrets file if not exists
	@if [ ! -f "$(SECRETS_FILE)" ]; then \
		echo "  ⚠️  시크릿 파일이 없습니다. (Secrets file not found)"; \
		echo "     $(SECRETS_FILE)"; \
		echo ""; \
		echo "  📝 예시 파일에서 복사합니다... (Copying from example)"; \
		cp "$(SECRETS_EXAMPLE)" "$(SECRETS_FILE)"; \
		echo "  ✅ 생성됨 (Created): $(SECRETS_FILE)"; \
		echo ""; \
	else \
		echo "  ✅ 시크릿 파일 존재 (Secrets file exists)"; \
	fi
	@# Check for placeholder values
	@if grep -q "your-google-client-id\|your-client-id\|your-google-client-secret\|your-client-secret" "$(SECRETS_FILE)" 2>/dev/null; then \
		echo ""; \
		echo "  ⚠️  기본 시크릿 값이 감지되었습니다! (Default secrets detected!)"; \
		echo ""; \
		echo "  현재 값 (Current values):"; \
		grep -E "GOOGLE_CLIENT_ID|GOOGLE_CLIENT_SECRET" "$(SECRETS_FILE)" | sed 's/^/    /'; \
		echo ""; \
		echo "  📌 Google OAuth2 설정 방법:"; \
		echo "     1. https://console.cloud.google.com/apis/credentials 접속"; \
		echo "     2. OAuth 2.0 클라이언트 ID 생성"; \
		echo "     3. $(SECRETS_FILE) 파일 수정"; \
		echo ""; \
		echo "  ⚠️  기본값으로 진행하면 Google 로그인이 작동하지 않습니다!"; \
		echo ""; \
		read -p "  기본값으로 계속 진행하시겠습니까? (Continue with defaults?) [y/N]: " confirm; \
		if [ "$$confirm" != "y" ] && [ "$$confirm" != "Y" ]; then \
			echo ""; \
			echo "  취소되었습니다. 시크릿 파일을 수정 후 다시 실행해주세요."; \
			echo "  (Cancelled. Please edit secrets file and try again.)"; \
			echo ""; \
			echo "  파일 위치: $(SECRETS_FILE)"; \
			echo ""; \
			exit 1; \
		fi; \
		echo ""; \
		echo "  ⚡ 기본값으로 진행합니다... (Proceeding with defaults...)"; \
	else \
		echo "  ✅ 시크릿 값 설정됨 (Secrets configured)"; \
	fi
	@echo ""

kind-local-all: _check-secrets ## [ONE-STOP] Setup everything for local-kind with external DB and monitoring
	@echo ""
	@echo "=============================================="
	@echo "  Wealist Local Kind - 설정 확인 (Configuration)"
	@echo "=============================================="
	@echo ""
	@echo "  현재 설정 (Current Settings):"
	@echo "    - Cluster:     $(KIND_CLUSTER)"
	@echo "    - Namespace:   $(K8S_NAMESPACE)"
	@echo "    - Registry:    $(LOCAL_REGISTRY)"
	@echo "    - EXTERNAL_DB: $(EXTERNAL_DB)"
ifeq ($(EXTERNAL_DB),true)
	@echo ""
	@echo "  ⚠️  EXTERNAL_DB=true 모드:"
	@echo "    - 호스트 머신의 PostgreSQL/Redis 사용"
	@echo "    - DB Host: $(EXTERNAL_DB_HOST)"
	@echo "    - PostgreSQL이 :5432에서 실행 중이어야 함"
	@echo "    - Redis가 :6379에서 실행 중이어야 함"
else
	@echo ""
	@echo "  📦 EXTERNAL_DB=false 모드 (기본값):"
	@echo "    - 클러스터 내 PostgreSQL/Redis Pod 사용"
	@echo "    - 별도 DB 설치 불필요"
endif
	@echo ""
	@echo "=============================================="
	@read -p "  계속 진행하시겠습니까? (Continue?) [Y/n]: " confirm; \
	if [ "$$confirm" = "n" ] || [ "$$confirm" = "N" ]; then \
		echo ""; \
		echo "  취소되었습니다. (Cancelled)"; \
		echo ""; \
		echo "  도움말 (Help):"; \
		echo "    - DB 분리 모드: make kind-local-all EXTERNAL_DB=true"; \
		echo "    - DB 포함 모드: make kind-local-all"; \
		echo ""; \
		exit 1; \
	fi
	@echo ""
	@echo "=============================================="
	@echo "  Wealist Local Kind - One-Stop Setup"
	@echo "  EXTERNAL_DB=$(EXTERNAL_DB)"
	@echo "=============================================="
	@echo ""
	@# Step 1: Check and setup external databases
ifeq ($(EXTERNAL_DB),true)
	@echo "=== Step 1/5: Checking External Databases ==="
	@$(MAKE) kind-check-db
else
	@echo "=== Step 1/5: Skipping DB setup (EXTERNAL_DB=false) ==="
endif
	@echo ""
	@# Step 2: Create Kind cluster
	@echo "=== Step 2/5: Creating Kind Cluster ==="
	@if kind get clusters 2>/dev/null | grep -q "^$(KIND_CLUSTER)$$"; then \
		echo "Kind cluster '$(KIND_CLUSTER)' already exists, skipping creation..."; \
		kind export kubeconfig --name $(KIND_CLUSTER); \
	else \
		./docker/scripts/dev/0.setup-cluster.sh; \
	fi
	@echo ""
	@# Step 3: Build and load images
	@echo "=== Step 3/5: Building and Loading Images ==="
	@$(MAKE) kind-load-images-mono
	@echo ""
	@# Step 4: Install Helm charts (infra + services + monitoring)
	@echo "=== Step 4/5: Installing Helm Charts ==="
	@$(MAKE) helm-install-all ENV=local-kind
	@echo ""
	@# Step 5: Verify deployment
	@echo "=== Step 5/5: Verifying Deployment ==="
	@$(MAKE) kind-status-all
	@echo ""
	@echo "=============================================="
	@echo "  Setup Complete! "
	@echo "=============================================="
	@echo ""
	@echo "  Access URLs:"
	@echo "    - Frontend:   http://localhost"
	@echo "    - Grafana:    http://localhost/monitoring/grafana"
	@echo "    - Prometheus: http://localhost/monitoring/prometheus"
	@echo "    - Loki:       http://localhost/monitoring/loki"
	@echo ""
	@echo "  Grafana Login: admin / admin"
	@echo ""
	@echo "  Quick Commands:"
	@echo "    make kind-status-all        # Check all services status"
	@echo "    make kind-verify-monitoring # Verify monitoring connections"
	@echo "=============================================="

kind-check-db: ## Check and install PostgreSQL/Redis if needed (interactive)
	@echo "Checking external database requirements..."
	@./docker/scripts/dev/setup-external-db.sh || exit 1
	@echo ""
	@echo "=== Ensuring all database roles and databases exist ==="
	@if [ "$$(uname)" = "Darwin" ]; then \
		echo "Running SQL script on macOS..."; \
		if ! psql -U postgres -f scripts/db/01-create-databases.sql; then \
			echo ""; \
			echo "❌ SQL 스크립트 실행 실패! (SQL script failed!)"; \
			echo ""; \
			echo "수동으로 실행해주세요 (Run manually):"; \
			echo "  psql -U postgres -f scripts/db/01-create-databases.sql"; \
			echo ""; \
			exit 1; \
		fi; \
	else \
		echo "Running SQL script on Linux..."; \
		if ! sudo -u postgres psql -f scripts/db/01-create-databases.sql; then \
			echo ""; \
			echo "❌ SQL 스크립트 실행 실패! (SQL script failed!)"; \
			echo ""; \
			echo "수동으로 실행해주세요 (Run manually):"; \
			echo "  sudo -u postgres psql -f scripts/db/01-create-databases.sql"; \
			echo ""; \
			exit 1; \
		fi; \
	fi
	@echo ""
	@echo "✅ All database roles and databases are ready!"

_install-postgres:
	@if [ "$$(uname)" = "Darwin" ]; then \
		brew install postgresql@14 && brew services start postgresql@14; \
	elif [ -f /etc/debian_version ]; then \
		sudo apt-get update && sudo apt-get install -y postgresql postgresql-contrib; \
		sudo systemctl enable postgresql && sudo systemctl start postgresql; \
	else \
		echo "Please install PostgreSQL manually"; exit 1; \
	fi

_start-postgres:
	@if [ "$$(uname)" = "Darwin" ]; then \
		brew services start postgresql@14 2>/dev/null || brew services start postgresql; \
	else \
		sudo systemctl start postgresql; \
	fi
	@sleep 2

_install-redis:
	@if [ "$$(uname)" = "Darwin" ]; then \
		brew install redis && brew services start redis; \
	elif [ -f /etc/debian_version ]; then \
		sudo apt-get update && sudo apt-get install -y redis-server; \
		sudo systemctl enable redis-server && sudo systemctl start redis-server; \
	else \
		echo "Please install Redis manually"; exit 1; \
	fi

_start-redis:
	@if [ "$$(uname)" = "Darwin" ]; then \
		brew services start redis; \
	else \
		sudo systemctl start redis-server 2>/dev/null || sudo systemctl start redis; \
	fi
	@sleep 1

_init-wealist-dbs: _configure-db-access
	@echo "Creating Wealist databases..."
	@DATABASES="wealist wealist_user_service_db wealist_board_service_db wealist_chat_service_db wealist_noti_service_db wealist_storage_service_db wealist_video_service_db"; \
	if [ "$$(uname)" = "Darwin" ]; then \
		psql -U postgres -c "SELECT 1" 2>/dev/null || createuser -s postgres 2>/dev/null || true; \
		for db in $$DATABASES; do \
			psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$$db'" 2>/dev/null | grep -q 1 || \
			psql -U postgres -c "CREATE DATABASE $$db" 2>/dev/null && echo "  Created: $$db" || echo "  Exists: $$db"; \
		done; \
	else \
		for db in $$DATABASES; do \
			sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname = '$$db'" 2>/dev/null | grep -q 1 || \
			sudo -u postgres psql -c "CREATE DATABASE $$db" 2>/dev/null && echo "  Created: $$db" || echo "  Exists: $$db"; \
		done; \
	fi

_configure-db-access:
	@echo "Configuring database access for Kind cluster..."
	@if [ -f /etc/debian_version ] && [ "$$(uname)" != "Darwin" ]; then \
		PG_HBA=$$(sudo -u postgres psql -t -P format=unaligned -c "SHOW hba_file" 2>/dev/null); \
		if [ -n "$$PG_HBA" ] && ! sudo grep -q "172.18.0.0/16" "$$PG_HBA" 2>/dev/null; then \
			echo "  Adding Kind network to pg_hba.conf..."; \
			echo "host    all    all    172.17.0.0/16    trust" | sudo tee -a "$$PG_HBA" >/dev/null; \
			echo "host    all    all    172.18.0.0/16    trust" | sudo tee -a "$$PG_HBA" >/dev/null; \
		fi; \
		PG_CONF=$$(sudo -u postgres psql -t -P format=unaligned -c "SHOW config_file" 2>/dev/null); \
		if [ -n "$$PG_CONF" ]; then \
			sudo sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/" "$$PG_CONF" 2>/dev/null || true; \
			sudo sed -i "s/listen_addresses = 'localhost'/listen_addresses = '*'/" "$$PG_CONF" 2>/dev/null || true; \
		fi; \
		sudo systemctl restart postgresql 2>/dev/null || true; \
		sudo sed -i 's/^bind 127.0.0.1/bind 0.0.0.0/' /etc/redis/redis.conf 2>/dev/null || true; \
		sudo sed -i 's/^protected-mode yes/protected-mode no/' /etc/redis/redis.conf 2>/dev/null || true; \
		sudo systemctl restart redis-server 2>/dev/null || sudo systemctl restart redis 2>/dev/null || true; \
	fi

kind-verify-monitoring: ## Verify monitoring stack connections
	@echo "=============================================="
	@echo "  Verifying Monitoring Stack Connections"
	@echo "=============================================="
	@echo ""
	@echo "--- Checking Pods ---"
	@kubectl get pods -n $(K8S_NAMESPACE) -l "app in (prometheus,grafana,loki,promtail)" 2>/dev/null || echo "No monitoring pods found"
	@echo ""
	@echo "--- Prometheus Targets ---"
	@kubectl exec -n $(K8S_NAMESPACE) deployment/prometheus -- wget -qO- http://localhost:9090/api/v1/targets 2>/dev/null | \
		grep -o '"health":"[^"]*"' | sort | uniq -c || echo "Could not fetch Prometheus targets"
	@echo ""
	@echo "--- Loki Ready Check ---"
	@kubectl exec -n $(K8S_NAMESPACE) deployment/loki -- wget -qO- http://localhost:3100/ready 2>/dev/null && echo "Loki: Ready" || echo "Loki: Not ready"
	@echo ""
	@echo "--- Grafana Datasources ---"
	@kubectl exec -n $(K8S_NAMESPACE) deployment/grafana -- wget -qO- http://admin:admin@localhost:3000/api/datasources 2>/dev/null | \
		grep -o '"name":"[^"]*"' || echo "Could not fetch Grafana datasources"
	@echo ""
	@echo "--- Service Endpoints ---"
	@echo "Prometheus:"
	@kubectl get endpoints prometheus -n $(K8S_NAMESPACE) 2>/dev/null || echo "  Not found"
	@echo "Grafana:"
	@kubectl get endpoints grafana -n $(K8S_NAMESPACE) 2>/dev/null || echo "  Not found"
	@echo "Loki:"
	@kubectl get endpoints loki -n $(K8S_NAMESPACE) 2>/dev/null || echo "  Not found"
	@echo ""
	@echo "=============================================="

kind-status-all: ## Show status of all Kind cluster components
	@echo "=============================================="
	@echo "  Kind Cluster Status ($(KIND_CLUSTER))"
	@echo "=============================================="
	@echo ""
	@echo "--- Cluster Info ---"
	@kubectl cluster-info 2>/dev/null | head -2 || echo "Cluster not accessible"
	@echo ""
	@echo "--- Nodes ---"
	@kubectl get nodes -o wide 2>/dev/null || echo "No nodes found"
	@echo ""
	@echo "--- Pods ($(K8S_NAMESPACE)) ---"
	@kubectl get pods -n $(K8S_NAMESPACE) -o wide 2>/dev/null || echo "No pods found"
	@echo ""
	@echo "--- Services ---"
	@kubectl get svc -n $(K8S_NAMESPACE) 2>/dev/null || echo "No services found"
	@echo ""
	@echo "--- Ingress ---"
	@kubectl get ingress -n $(K8S_NAMESPACE) 2>/dev/null || echo "No ingress found"
	@echo ""
	@echo "--- External DB Connection Test ---"
ifeq ($(EXTERNAL_DB),true)
	@echo "PostgreSQL ($(EXTERNAL_DB_HOST):5432):"
	@kubectl run pg-test --rm -i --restart=Never --image=postgres:17-alpine -n $(K8S_NAMESPACE) \
		-- pg_isready -h $(EXTERNAL_DB_HOST) -p 5432 -U postgres 2>/dev/null && echo "  [OK] Connected" || echo "  [FAIL] Not reachable"
	@echo "Redis ($(EXTERNAL_DB_HOST):6379):"
	@kubectl run redis-test --rm -i --restart=Never --image=redis:7.2-alpine -n $(K8S_NAMESPACE) \
		-- redis-cli -h $(EXTERNAL_DB_HOST) -p 6379 ping 2>/dev/null && echo "  [OK] Connected" || echo "  [FAIL] Not reachable"
else
	@echo "Using internal databases (EXTERNAL_DB=false)"
endif
	@echo ""
	@echo "=============================================="
