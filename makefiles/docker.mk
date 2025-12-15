# =============================================================================
# Docker Compose Commands
# =============================================================================

##@ Development (Docker Compose)

.PHONY: dev-up dev-down dev-logs

dev-up: ## Start all services
	./docker/scripts/dev.sh up

dev-down: ## Stop all services
	./docker/scripts/dev.sh down

dev-logs: ## View logs
	./docker/scripts/dev.sh logs

##@ SonarQube (Code Quality)

.PHONY: sonar-up sonar-down sonar-logs sonar-status sonar-restart sonar-clean

sonar-up: ## Start SonarQube only (lightweight)
	@echo "Starting SonarQube standalone environment..."
	./docker/scripts/sonar.sh up

sonar-down: ## Stop SonarQube environment
	@echo "Stopping SonarQube standalone environment..."
	./docker/scripts/sonar.sh down

sonar-logs: ## View SonarQube logs
	./docker/scripts/sonar.sh logs

sonar-status: ## Check SonarQube status
	./docker/scripts/sonar.sh status

sonar-restart: ## Restart SonarQube environment
	@echo "Restarting SonarQube standalone environment..."
	./docker/scripts/sonar.sh restart

sonar-clean: ## Clean SonarQube data (destructive)
	@echo "Cleaning SonarQube standalone environment..."
	./docker/scripts/sonar.sh clean
