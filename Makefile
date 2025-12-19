# =============================================================================
# weAlist Project Makefile (Local Development)
# =============================================================================
# Self-documenting Makefile using ## comments
# Run 'make help' to see all available commands
#
# Supported Modes:
#   1. docker-compose: make dev-up / make dev-down
#   2. localhost (Kind + internal DB): make kind-local-all
#   3. localhost (Kind + external DB): make kind-local-all EXTERNAL_DB=true
#
# Structure:
#   makefiles/_variables.mk  - Common variables
#   makefiles/docker.mk      - Docker Compose commands (dev-*)
#   makefiles/kind.mk        - Kind cluster commands (kind-*)
#   makefiles/services.mk    - Per-service commands (*-build, *-load, *-redeploy)
#   makefiles/helm.mk        - Helm chart commands (helm-*)
# =============================================================================

.DEFAULT_GOAL := help

# Include all sub-makefiles
include makefiles/_variables.mk
include makefiles/docker.mk
include makefiles/kind.mk
include makefiles/services.mk
include makefiles/helm.mk

##@ General

.PHONY: help

help: ## Display this help
	@echo ""
	@echo "\033[1m🌿 weAlist Local Development\033[0m"
	@echo ""
	@echo "\033[1mQuick Start:\033[0m"
	@echo "  \033[36mmake dev-up\033[0m                    Start with docker-compose"
	@echo "  \033[36mmake kind-local-all\033[0m            Start Kind cluster (internal DB)"
	@echo "  \033[36mmake kind-local-all EXTERNAL_DB=true\033[0m  Start Kind (external DB)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } \
		/^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "\033[1mPer-Service Commands:\033[0m"
	@echo "  \033[36m<service>-build\033[0m           Build Docker image"
	@echo "  \033[36m<service>-load\033[0m            Build and push to registry"
	@echo "  \033[36m<service>-redeploy\033[0m        Rollout restart in k8s"
	@echo "  \033[36m<service>-all\033[0m             Build, push, and redeploy"
	@echo ""
	@echo "  Services: auth-service, board-service, chat-service, frontend,"
	@echo "            noti-service, storage-service, user-service, video-service"
	@echo ""
