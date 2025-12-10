.PHONY: help dev-up dev-down dev-logs kind-setup kind-delete infra-setup k8s-deploy-services status clean

# Kind cluster name
KIND_CLUSTER ?= wealist
LOCAL_REGISTRY ?= localhost:5001

help:
	@echo "Wealist Project"
	@echo ""
	@echo "  Development (Docker Compose):"
	@echo "    make dev-up       - Start all services"
	@echo "    make dev-down     - Stop all services"
	@echo "    make dev-logs     - View logs"
	@echo ""
	@echo "  Kubernetes (Local):"
	@echo "    make kind-setup         - 1. Create cluster + registry"
	@echo "    make infra-setup        - 2. Load infra images + deploy"
	@echo "    make k8s-deploy-services - 3. Build + deploy services"
	@echo "    make kind-delete        - Delete cluster"
	@echo ""
	@echo "  Utility:"
	@echo "    make status       - Show pods status"
	@echo "    make clean        - Clean up"

# =============================================================================
# Development (Docker Compose)
# =============================================================================

dev-up:
	./docker/scripts/dev.sh up

dev-down:
	./docker/scripts/dev.sh down

dev-logs:
	./docker/scripts/dev.sh logs

# =============================================================================
# Kubernetes (Local - Kind)
# =============================================================================

kind-setup:
	@echo "Setting up Kind cluster with local registry..."
	./docker/scripts/dev/0.setup-cluster.sh

kind-delete:
	kind delete cluster --name $(KIND_CLUSTER)
	@docker rm -f kind-registry 2>/dev/null || true

infra-setup:
	@echo "Loading infrastructure images..."
	./docker/scripts/dev/1.load_infra_images.sh
	@echo ""
	@echo "Deploying infrastructure..."
	kubectl apply -k infrastructure/overlays/develop
	@echo ""
	@echo "Waiting for pods..."
	kubectl wait --namespace wealist-dev --for=condition=ready pod --selector=app=postgres --timeout=120s || true
	kubectl wait --namespace wealist-dev --for=condition=ready pod --selector=app=redis --timeout=120s || true
	@echo ""
	@echo "Done! Next: make k8s-deploy-services"

k8s-deploy-services:
	@echo "Building services..."
	./docker/scripts/dev/2.build_services_and_load.sh
	@echo ""
	@echo "Deploying services..."
	kubectl apply -k k8s/overlays/develop-registry/all-services
	@echo ""
	@echo "✅ Done! Check: make status"

# =============================================================================
# Utility
# =============================================================================

status:
	@echo "=== Kubernetes Pods ==="
	@kubectl get pods -n wealist-dev 2>/dev/null || echo "Namespace not found"

clean:
	./docker/scripts/clean.sh
