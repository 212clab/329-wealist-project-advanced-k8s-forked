# Deprecated: Kustomize Manifests

**⚠️ These files are deprecated and kept for reference only.**

## Migration Complete

All Kubernetes deployments have been migrated from Kustomize to Helm.

- **Old**: Kustomize overlays (110 YAML files, 18 ConfigMap patches)
- **New**: Helm charts (9 production-ready charts)

## What Happened

- `k8s/` → Moved to `deprecated/kustomize/k8s/`
- `infrastructure/` → Moved to `deprecated/kustomize/infrastructure/`
- All functionality now in `helm/charts/`

## Migration Details

See:
- `helm/PRODUCTION_READY_SUMMARY.md`
- `argocd/ARGOCD_HELM_INTEGRATION.md`
- Root `MIGRATION_COMPLETE.md`

## Removal Timeline

These files may be removed in a future release. Ensure all teams have migrated to Helm before deletion.

**Date Archived**: $(date +%Y-%m-%d)
**Helm Charts Version**: 1.0.0
