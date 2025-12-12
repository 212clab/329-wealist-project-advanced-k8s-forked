# Helm Charts - Quick Start Guide

## Prerequisites

- Helm 3.x installed
- kubectl configured
- Kubernetes cluster (Kind/Minikube for local, EKS/GKE for production)

## 🚀 Local Development Setup

### 1. Validate Charts

```bash
# Run comprehensive validation (84 tests)
./helm/scripts/validate-all-charts.sh
```

### 2. Install Infrastructure

```bash
helm install wealist-infrastructure ./helm/charts/wealist-infrastructure \
  -f ./helm/charts/wealist-infrastructure/values-develop-registry-local.yaml \
  -n wealist-dev --create-namespace
```

**What this installs**:
- PostgreSQL StatefulSet (6 databases)
- Redis StatefulSet
- MinIO StatefulSet
- Shared ConfigMap
- Ingress (routes to all services)

### 3. Install Services

**Option A: Install all at once**
```bash
for service in auth-service user-service board-service chat-service \
               noti-service storage-service video-service frontend; do
  echo "Installing $service..."
  helm install $service ./helm/charts/$service \
    -f ./helm/charts/$service/values-develop-registry-local.yaml \
    -n wealist-dev
done
```

**Option B: Install one by one**
```bash
# Auth service (JWT, OAuth2)
helm install auth-service ./helm/charts/auth-service \
  -f ./helm/charts/auth-service/values-develop-registry-local.yaml \
  -n wealist-dev

# User service (Users, workspaces)
helm install user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values-develop-registry-local.yaml \
  -n wealist-dev

# ... repeat for other services
```

### 4. Verify Deployment

```bash
# Check pod status
kubectl get pods -n wealist-dev

# Check services
kubectl get svc -n wealist-dev

# Check ingress
kubectl get ingress -n wealist-dev

# Follow logs
kubectl logs -f -l app.kubernetes.io/name=user-service -n wealist-dev
```

## 🔄 Update Workflow

### Upgrade a Single Service

```bash
helm upgrade user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values-develop-registry-local.yaml \
  -n wealist-dev
```

### Upgrade All Services

```bash
for service in auth-service user-service board-service chat-service \
               noti-service storage-service video-service frontend; do
  echo "Upgrading $service..."
  helm upgrade $service ./helm/charts/$service \
    -f ./helm/charts/$service/values-develop-registry-local.yaml \
    -n wealist-dev
done
```

### Rollback

```bash
# View history
helm history user-service -n wealist-dev

# Rollback to previous version
helm rollback user-service -n wealist-dev

# Rollback to specific revision
helm rollback user-service 2 -n wealist-dev
```

## 🧪 Testing & Debugging

### Dry Run (Template Rendering)

```bash
# See what will be deployed
helm template user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values-develop-registry-local.yaml \
  --debug

# Check specific resource
helm template user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values-develop-registry-local.yaml \
  -s templates/deployment.yaml
```

### Lint Chart

```bash
helm lint ./helm/charts/user-service
helm lint ./helm/charts/user-service \
  -f ./helm/charts/user-service/values-develop-registry-local.yaml
```

### Validate Required Values

```bash
# This will fail if required values are missing
helm template user-service ./helm/charts/user-service \
  --set image.repository=""  # Should fail validation
```

### Debug Deployed Release

```bash
# Get deployed values
helm get values user-service -n wealist-dev

# Get all resources
helm get all user-service -n wealist-dev

# Get manifest
helm get manifest user-service -n wealist-dev
```

## 🔧 Common Operations

### Override Values

```bash
# Command-line override
helm install user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values-develop-registry-local.yaml \
  --set replicaCount=3 \
  --set image.tag=v2.0.0 \
  -n wealist-dev

# Multiple values files
helm install user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values.yaml \
  -f ./helm/charts/user-service/values-develop-registry-local.yaml \
  -f custom-overrides.yaml \
  -n wealist-dev
```

### Uninstall

```bash
# Single service
helm uninstall user-service -n wealist-dev

# All services
for service in auth-service user-service board-service chat-service \
               noti-service storage-service video-service frontend; do
  helm uninstall $service -n wealist-dev 2>/dev/null || true
done

# Infrastructure (last)
helm uninstall wealist-infrastructure -n wealist-dev
```

### List Releases

```bash
helm list -n wealist-dev
helm list -A  # All namespaces
```

## 🏭 Production Deployment

### 1. Create Production Values

```yaml
# values-production.yaml
global:
  namespace: wealist
  environment: production
  domain: wealist.co.kr
  imageRegistry: your-registry.io

image:
  tag: v1.0.0  # Specific version, not 'latest'
  pullPolicy: IfNotPresent

replicaCount: 2  # Minimum for HA

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10

podDisruptionBudget:
  enabled: true
  minAvailable: 1

externalSecrets:
  enabled: true
  secretStore: aws-secrets-manager
```

### 2. Deploy to Production

```bash
# Infrastructure
helm install wealist-infrastructure ./helm/charts/wealist-infrastructure \
  -f ./helm/charts/wealist-infrastructure/values.yaml \
  -f ./helm/charts/wealist-infrastructure/values-production.yaml \
  -n wealist --create-namespace

# Services
helm install user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values.yaml \
  -f ./helm/charts/user-service/values-production.yaml \
  -n wealist
```

### 3. Enable Security Features

```yaml
# Enable in production values
networkPolicy:
  enabled: true

securityContext:
  readOnlyRootFilesystem: true  # Where possible

podSecurityContext:
  runAsNonRoot: true
  seccompProfile:
    type: RuntimeDefault
```

## 📊 Monitoring

### Check Metrics

```bash
# If Prometheus annotations are enabled
kubectl get --raw /api/v1/namespaces/wealist-dev/pods/user-service-xxx/proxy/metrics
```

### View Autoscaling Status

```bash
kubectl get hpa -n wealist-dev
kubectl describe hpa user-service -n wealist-dev
```

### Check Pod Disruption Budgets

```bash
kubectl get pdb -n wealist-dev
kubectl describe pdb user-service -n wealist-dev
```

## 🆕 Add New Service

### Using the Generator Script

```bash
# Generate new service chart
./helm/scripts/generate-service-chart.sh new-service 8005 true

# Customize values
vim helm/charts/new-service/values.yaml
vim helm/charts/new-service/values-develop-registry-local.yaml

# Update dependencies
cd helm/charts/new-service && helm dependency update && cd -

# Validate
helm lint ./helm/charts/new-service

# Deploy
helm install new-service ./helm/charts/new-service \
  -f ./helm/charts/new-service/values-develop-registry-local.yaml \
  -n wealist-dev
```

## 🐛 Troubleshooting

### Pod Not Starting

```bash
# Check events
kubectl describe pod user-service-xxx -n wealist-dev

# Check logs
kubectl logs user-service-xxx -n wealist-dev

# Check previous container (if crashed)
kubectl logs user-service-xxx -n wealist-dev --previous
```

### Template Rendering Error

```bash
# Debug template rendering
helm template user-service ./helm/charts/user-service \
  -f ./helm/charts/user-service/values-develop-registry-local.yaml \
  --debug
```

### Values Not Applied

```bash
# Check what values are actually used
helm get values user-service -n wealist-dev --all
```

### Dependency Issues

```bash
# Re-download dependencies
cd helm/charts/user-service
rm -rf charts/
helm dependency update
cd -

# Verify
ls -la helm/charts/user-service/charts/
```

## 📚 Reference

### Chart Structure

```
user-service/
├── Chart.yaml                              # Chart metadata
├── values.yaml                             # Production baseline (200-300 lines)
├── values-develop-registry-local.yaml      # Dev overrides (50-100 lines)
├── charts/                                 # Dependencies
│   └── wealist-common-1.0.0.tgz
└── templates/
    ├── deployment.yaml                     # Uses wealist-common
    ├── service.yaml
    ├── configmap.yaml
    ├── hpa.yaml
    ├── serviceaccount.yaml
    ├── poddisruptionbudget.yaml
    └── networkpolicy.yaml
```

### Key Values Paths

```yaml
# Image
image.repository: "user-service"
image.tag: "1.0.0"

# Service
service.port: 8081
service.targetPort: 8081

# Config (environment variables)
config.ENV: "production"
config.DB_HOST: "postgres"

# Security
podSecurityContext.runAsNonRoot: true
securityContext.allowPrivilegeEscalation: false

# Scaling
replicaCount: 2
autoscaling.enabled: true
podDisruptionBudget.enabled: true

# Health checks
healthCheck.liveness.path: "/health/live"
healthCheck.readiness.path: "/health/ready"
```

## 💡 Tips

1. **Always validate before deploy**: `helm lint` and `helm template --debug`
2. **Use specific image tags in production**: Never use `latest`
3. **Test rollback procedures**: Ensure you can rollback quickly
4. **Monitor resource usage**: Tune limits based on actual usage
5. **Use External Secrets in production**: Don't commit secrets to Git
6. **Enable Network Policies**: Add defense in depth
7. **Keep values DRY**: Only override what changes between environments

## 🔗 Additional Resources

- [Helm Documentation](https://helm.sh/docs/)
- [PRODUCTION_READY_SUMMARY.md](./PRODUCTION_READY_SUMMARY.md) - Detailed feature overview
- [../CLAUDE.md](../CLAUDE.md) - Project-specific patterns and troubleshooting
- [Validation Script](./scripts/validate-all-charts.sh) - Automated testing

---

**Need help?** Check the validation results:
```bash
./helm/scripts/validate-all-charts.sh
```
