# weAlist Helm Charts

Production-ready Helm charts for deploying the weAlist microservices platform on Kubernetes.

## 🚀 Quick Start

### Prerequisites

- Kubernetes cluster (Kind, Minikube, or production cluster)
- Helm 3.8+
- kubectl configured

### 1. Install Everything (Infrastructure + Services)

```bash
# From project root
make helm-install-all
```

This will:
1. Install infrastructure (PostgreSQL, Redis, MinIO, Ingress)
2. Wait 5 seconds for databases to initialize
3. Install all 8 service charts

### 2. Verify Deployment

```bash
# Check Helm releases
helm list -n wealist-dev

# Check pods
kubectl get pods -n wealist-dev

# Run validation (156 tests)
make helm-validate
```

### 3. Access Services

**Local Development** (localhost):
- Frontend: http://localhost
- API Gateway: http://localhost/api/users/health

**Local Development** (local.wealist.co.kr with TLS):
- Frontend: https://local.wealist.co.kr
- API Gateway: https://local.wealist.co.kr/api/users/health

---

## 📊 Chart Structure

```
helm/charts/
├── wealist-infrastructure/    # Core infrastructure (PostgreSQL, Redis, etc.)
├── wealist-common/            # Shared templates library (not deployed directly)
├── auth-service/              # JWT authentication service (Spring Boot)
├── user-service/              # User management (Go)
├── board-service/             # Project boards (Go)
├── chat-service/              # Real-time messaging (Go)
├── noti-service/              # Notifications (Go)
├── storage-service/           # File storage (Go)
├── video-service/             # Video calls (Go)
└── frontend/                  # React UI
```

### Chart Dependencies

All service charts depend on:
- **wealist-common** (library chart) - Provides shared templates
- **wealist-infrastructure** (runtime dependency) - Must be installed first

---

## 🎯 Common Commands

### Installation

```bash
# Install infrastructure only
make helm-install-infra

# Install all services (requires infrastructure)
make helm-install-services

# Install everything
make helm-install-all
```

### Upgrades

```bash
# Upgrade all charts
make helm-upgrade-all

# Upgrade specific service
helm upgrade user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values-develop-registry-local.yaml \
  -n wealist-dev
```

### Validation

```bash
# Lint all charts
make helm-lint

# Comprehensive validation (156 tests)
make helm-validate

# Dry-run before install
helm install user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values-develop-registry-local.yaml \
  -n wealist-dev --dry-run --debug
```

### Uninstallation

```bash
# Uninstall all (services + infrastructure)
make helm-uninstall-all

# Uninstall specific service
helm uninstall user-service -n wealist-dev
```

---

## ⚙️ Configuration

### Values Hierarchy

Each chart supports multiple values files for different environments:

```
values.yaml                         # Production defaults
values-develop.yaml                 # Development overrides
values-develop-registry-local.yaml  # Local Kind cluster with registry
values-prod.yaml                    # Production settings (future)
```

### Common Configuration Pattern

#### Infrastructure Chart (`wealist-infrastructure/values-develop-registry-local.yaml`)

```yaml
# PostgreSQL
postgres:
  image: postgres:17-bookworm
  storage: 2Gi

# Redis
redis:
  image: redis:7.2-bookworm

# Ingress
ingress:
  enabled: true
  host: local.wealist.co.kr
  tls:
    enabled: true
    secretName: local-wealist-tls
```

#### Service Chart (`user-service/values-develop-registry-local.yaml`)

```yaml
# Docker image
image:
  repository: localhost:5001/user-service
  tag: latest
  pullPolicy: Always

# Application configuration
config:
  PORT: "8081"
  ENV: "development"
  SERVER_BASE_PATH: "/api"
  S3_PUBLIC_ENDPOINT: "http://local.wealist.co.kr/storage"
  USER_SERVICE_URL: "http://user-service:8081"

# Health checks
healthCheck:
  liveness:
    path: /api/health/live
  readiness:
    path: /api/health/ready

# Resources
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

### Overriding Values at Install Time

```bash
# Override image tag
helm install user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values-develop-registry-local.yaml \
  --set image.tag=v1.2.3 \
  -n wealist-dev

# Override multiple values
helm upgrade user-service ./helm/charts/user-service \
  --set image.tag=latest \
  --set config.ENV=staging \
  --set resources.limits.memory=1Gi \
  -n wealist-dev
```

---

## 🏗️ Chart Development

### Creating a New Service Chart

1. **Copy existing chart structure**:
   ```bash
   cp -r helm/charts/user-service helm/charts/new-service
   ```

2. **Update Chart.yaml**:
   ```yaml
   apiVersion: v2
   name: new-service
   version: 1.0.0
   appVersion: "1.0.0"
   dependencies:
     - name: wealist-common
       version: "1.0.0"
       repository: "file://../wealist-common"
   ```

3. **Update values files**:
   - `values.yaml` - Production defaults
   - `values-develop-registry-local.yaml` - Local development

4. **Templates use common library**:
   ```yaml
   # templates/deployment.yaml
   {{- include "wealist-common.deployment" . }}

   # templates/service.yaml
   {{- include "wealist-common.service" . }}
   ```

5. **Validate**:
   ```bash
   helm lint ./helm/charts/new-service
   helm install new-service ./helm/charts/new-service --dry-run --debug
   ```

### Updating wealist-common Library

When updating shared templates in `wealist-common/`:

1. **Update template files** in `wealist-common/templates/`

2. **Bump version** in `wealist-common/Chart.yaml`:
   ```yaml
   version: 1.1.0  # Increment for breaking changes
   ```

3. **Update dependencies** in all service charts:
   ```yaml
   # service/Chart.yaml
   dependencies:
     - name: wealist-common
       version: "1.1.0"  # Match new version
   ```

4. **Update dependency**:
   ```bash
   cd helm/charts/user-service
   helm dependency update
   ```

---

## 🔍 Troubleshooting

### Common Issues

#### 1. Pod CrashLoopBackOff

**Symptom**: Service pod constantly restarts

**Check**:
```bash
# View pod logs
kubectl logs -f <pod-name> -n wealist-dev

# Describe pod for events
kubectl describe pod <pod-name> -n wealist-dev
```

**Common Causes**:
- Database not ready (check PostgreSQL pod)
- Incorrect environment variables (check ConfigMap)
- Health check path mismatch (verify `/health/live`)

**Fix**:
```bash
# Check infrastructure pods first
kubectl get pods -n wealist-dev | grep postgres
kubectl get pods -n wealist-dev | grep redis

# Restart service pod
kubectl delete pod <pod-name> -n wealist-dev
```

#### 2. Helm Install Fails with "Release Already Exists"

**Symptom**: `Error: cannot re-use a name that is still in use`

**Fix**:
```bash
# Check existing releases
helm list -n wealist-dev

# Uninstall old release
helm uninstall <release-name> -n wealist-dev

# Or use upgrade instead
helm upgrade --install <release-name> ./helm/charts/<chart>
```

#### 3. Values Not Applied

**Symptom**: Service uses default values instead of overrides

**Check**:
```bash
# View rendered values
helm get values <release-name> -n wealist-dev

# View all values (including defaults)
helm get values <release-name> -n wealist-dev --all
```

**Fix**:
```bash
# Ensure correct values file path
helm upgrade <release-name> ./helm/charts/<chart> \
  -f ./helm/charts/<chart>/values-develop-registry-local.yaml \
  -n wealist-dev
```

#### 4. Template Rendering Errors

**Symptom**: `Error: parse error` during install

**Debug**:
```bash
# Render templates without installing
helm template <release-name> ./helm/charts/<chart> \
  -f ./helm/charts/<chart>/values-develop-registry-local.yaml \
  --debug

# Check specific template
helm template <release-name> ./helm/charts/<chart> \
  -s templates/deployment.yaml
```

**Common Causes**:
- Missing required value
- Incorrect indentation in YAML
- Invalid template syntax

#### 5. Ingress Not Working

**Symptom**: 404 or 503 on ingress routes

**Check**:
```bash
# Check ingress resource
kubectl get ingress -n wealist-dev
kubectl describe ingress wealist-ingress -n wealist-dev

# Check ingress controller
kubectl get pods -n ingress-nginx

# Test service directly
kubectl port-forward svc/user-service 8081:8081 -n wealist-dev
curl http://localhost:8081/api/health/live
```

**Fix**:
- Verify service is running
- Check ingress path matches service actual path
- Ensure ingress controller is installed (for Kind: `make kind-setup`)

---

## 📚 Validation Tests

The project includes comprehensive validation:

### Helm Chart Tests (84 tests)
```bash
./helm/scripts/validate-all-charts.sh
```

**Categories**:
- Chart linting (9 tests)
- Template rendering (27 tests)
- Values validation (27 tests)
- Dependency checks (21 tests)

### ArgoCD Application Tests (72 tests)
```bash
./argocd/scripts/validate-applications.sh
```

**Categories**:
- Application syntax (9 tests)
- Helm source configuration (18 tests)
- Sync policy (18 tests)
- Destination (18 tests)
- Health assessment (9 tests)

### Run All Tests
```bash
make helm-validate  # Runs both scripts (156 total tests)
```

---

## 🎨 Best Practices

### 1. Always Validate Before Deploy

```bash
# Lint charts
make helm-lint

# Dry-run
helm install <release> ./helm/charts/<chart> \
  -f values-develop-registry-local.yaml \
  -n wealist-dev --dry-run --debug

# Run full validation
make helm-validate
```

### 2. Use Environment-Specific Values

Never modify `values.yaml` directly for environment-specific settings.

**Good**:
```bash
helm install user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values.yaml \
  -f ./helm/charts/user-service/values-develop.yaml \
  -n wealist-dev
```

**Bad**:
```bash
# Don't edit values.yaml for dev settings
helm install user-service ./helm/charts/user-service \
  -n wealist-dev
```

### 3. Version Your Charts

Update `version` in `Chart.yaml` for changes:

- **Major** (1.0.0 → 2.0.0): Breaking changes
- **Minor** (1.0.0 → 1.1.0): New features, backwards compatible
- **Patch** (1.0.0 → 1.0.1): Bug fixes

### 4. Document Custom Values

Add comments in `values.yaml`:

```yaml
# Image configuration
image:
  # Container image repository
  repository: localhost:5001/user-service

  # Image tag (override for specific versions)
  # Default: latest
  tag: latest

  # Image pull policy
  # Options: Always, IfNotPresent, Never
  pullPolicy: Always
```

### 5. Use Helm Hooks for Migrations

For database migrations or initialization:

```yaml
# templates/migration-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-weight": "-5"
    "helm.sh/hook-delete-policy": before-hook-creation
spec:
  template:
    spec:
      containers:
        - name: migrate
          image: {{ .Values.image.repository }}:{{ .Values.image.tag }}
          command: ["./migrate", "up"]
```

---

## 🔗 Related Documentation

### Project Documentation
- [`CLAUDE.md`](../CLAUDE.md) - Developer guide and architecture overview
- [`MIGRATION_COMPLETE.md`](../MIGRATION_COMPLETE.md) - Kustomize → Helm migration details
- [`docs/CONFIGURATION.md`](../docs/CONFIGURATION.md) - Ports, naming conventions, URLs

### Helm Chart Documentation
- [`helm/PRODUCTION_READY_SUMMARY.md`](./PRODUCTION_READY_SUMMARY.md) - Chart validation and features
- [`helm/charts/wealist-infrastructure/README.md`](./charts/wealist-infrastructure/README.md) - Infrastructure chart
- [`argocd/ARGOCD_HELM_INTEGRATION.md`](../argocd/ARGOCD_HELM_INTEGRATION.md) - GitOps with ArgoCD

### External Resources
- [Helm Documentation](https://helm.sh/docs/)
- [Helm Best Practices](https://helm.sh/docs/chart_best_practices/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)

---

## 🆘 Getting Help

### Common Questions

**Q: How do I update a single service?**
```bash
# Build new image
make user-service-load

# Upgrade Helm release
helm upgrade user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values-develop-registry-local.yaml \
  -n wealist-dev
```

**Q: How do I rollback a deployment?**
```bash
# List revisions
helm history user-service -n wealist-dev

# Rollback to previous
helm rollback user-service -n wealist-dev

# Rollback to specific revision
helm rollback user-service 3 -n wealist-dev
```

**Q: How do I check what values are being used?**
```bash
# Current values
helm get values user-service -n wealist-dev

# All values (including defaults)
helm get values user-service -n wealist-dev --all
```

**Q: How do I see the rendered manifests?**
```bash
# All manifests
helm get manifest user-service -n wealist-dev

# Specific template
helm template user-service ./helm/charts/user-service \
  -s templates/deployment.yaml
```

### Support

- **Internal Documentation**: Check `CLAUDE.md` first
- **Validation Tests**: Run `make helm-validate` to diagnose issues
- **Helm Debugging**: Use `--dry-run --debug` for detailed output
- **Kubernetes Events**: Check with `kubectl describe` for error messages

---

## 📊 Chart Versions

| Chart | Version | App Version | Status |
|-------|---------|-------------|--------|
| wealist-infrastructure | 1.0.0 | - | ✅ Production Ready |
| wealist-common | 1.0.0 | - | ✅ Production Ready |
| auth-service | 1.0.0 | 1.0.0 | ✅ Production Ready |
| user-service | 1.0.0 | 1.0.0 | ✅ Production Ready |
| board-service | 1.0.0 | 1.0.0 | ✅ Production Ready |
| chat-service | 1.0.0 | 1.0.0 | ✅ Production Ready |
| noti-service | 1.0.0 | 1.0.0 | ✅ Production Ready |
| storage-service | 1.0.0 | 1.0.0 | ✅ Production Ready |
| video-service | 1.0.0 | 1.0.0 | ✅ Production Ready |
| frontend | 1.0.0 | 1.0.0 | ✅ Production Ready |

**Last Updated**: 2025-12-12
**Migration Status**: Complete (Kustomize → Helm)
**Validation**: 156/156 tests passing

---

## 🎉 Quick Reference Card

```bash
# Deploy everything
make helm-install-all

# Validate (156 tests)
make helm-validate

# Upgrade all
make helm-upgrade-all

# Check status
helm list -n wealist-dev
kubectl get pods -n wealist-dev

# View logs
kubectl logs -f <pod-name> -n wealist-dev

# Uninstall all
make helm-uninstall-all
```

**Default Namespace**: `wealist-dev`
**Default Values**: `values-develop-registry-local.yaml`
**Validation Script**: `./helm/scripts/validate-all-charts.sh`

---

**Happy Deploying! 🚀**
