#!/bin/bash
# =============================================================================
# Migrate ArgoCD Applications from Kustomize to Helm
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APPS_DIR="$(cd "${SCRIPT_DIR}/../apps" && pwd)"

echo "🔄 Migrating ArgoCD Applications to Helm source..."
echo ""

# Services to update (existing files)
SERVICES=(
  "auth-service"
  "board-service"
  "chat-service"
  "noti-service"
  "frontend"
)

# Update existing service Applications
for service in "${SERVICES[@]}"; do
  app_file="${APPS_DIR}/${service}.yaml"

  if [ ! -f "$app_file" ]; then
    echo "⚠️  ${service}.yaml not found, skipping..."
    continue
  fi

  echo "📝 Updating ${service}..."

  # Backup original
  cp "$app_file" "${app_file}.backup"

  # Replace Kustomize path with Helm path and add helm config
  sed -i.tmp '
    s|path: services/'${service}'/k8s/overlays/local|path: helm/charts/'${service}'|
    /path: helm\/charts\/'${service}'/a\
\    helm:\
\      valueFiles:\
\        - values.yaml\
\        - values-develop-registry-local.yaml\
\      parameters:\
\        - name: image.tag\
\          value: "latest"
  ' "$app_file"

  # Remove temporary file
  rm -f "${app_file}.tmp"

  echo "✅ ${service} updated"
done

echo ""
echo "🎉 Migration complete!"
echo ""
echo "📋 Summary:"
echo "  - Updated: ${#SERVICES[@]} existing Applications"
echo "  - Backups: *.backup files created"
echo ""
echo "🔍 Next steps:"
echo "  1. Review changes: git diff argocd/apps/"
echo "  2. Create missing Applications (storage-service, video-service)"
echo "  3. Test ArgoCD sync"
