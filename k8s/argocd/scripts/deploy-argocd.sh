#!/bin/bash
set -e

echo "🚀 Starting ArgoCD deployment with Sealed Secrets..."

# 색상
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 스크립트 실행 위치 저장 (절대 경로로 변환)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ARGOCD_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"  # k8s/argocd (절대 경로)
PROJECT_ROOT="$(cd "$ARGOCD_DIR/../.." && pwd)"  # 프로젝트 루트 (절대 경로)

# GitHub 저장소 정보
REPO_URL="https://github.com/OrangesCloud/wealist-project-advanced-k8s.git"
SEALED_SECRETS_KEY="${1:-$SCRIPT_DIR/sealed-secrets-dev-20251227-220912.key}"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Wealist Platform Deployment"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📂 Paths (Absolute):"
echo "   Script:       $SCRIPT_DIR"
echo "   ArgoCD:       $ARGOCD_DIR"
echo "   Project Root: $PROJECT_ROOT"
echo ""

# ============================================
# 0. 디렉토리 구조 확인
# ============================================
echo -e "${YELLOW}🔍 Step 0: Verifying directory structure...${NC}"

# 필수 디렉토리 확인
REQUIRED_DIRS=(
    "$ARGOCD_DIR/apps"
    "$ARGOCD_DIR/sealed-secrets"
)

for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        echo "   ✅ Found: $dir"
    else
        echo "   ❌ Missing: $dir"
        echo ""
        echo -e "${YELLOW}Creating missing directory: $dir${NC}"
        mkdir -p "$dir"
    fi
done

echo ""

# ============================================
# 1. Sealed Secrets 키 확인
# ============================================
echo -e "${YELLOW}🔑 Step 1: Checking Sealed Secrets key...${NC}"

if [ -f "$SEALED_SECRETS_KEY" ]; then
    echo -e "${GREEN}✅ Found key backup: $SEALED_SECRETS_KEY${NC}"
    USE_EXISTING_KEY=true
else
    echo -e "${YELLOW}⚠️  Key file not found: $SEALED_SECRETS_KEY${NC}"
    echo ""
    echo "Options:"
    echo "  1) Provide key file path"
    echo "  2) Continue without key (new key will be generated)"
    echo ""
    read -p "Choose (1/2): " -n 1 -r
    echo ""
    
    if [[ $REPLY == "1" ]]; then
        read -p "Enter key file path: " SEALED_SECRETS_KEY
        if [ -f "$SEALED_SECRETS_KEY" ]; then
            USE_EXISTING_KEY=true
        else
            echo -e "${RED}❌ File not found: $SEALED_SECRETS_KEY${NC}"
            exit 1
        fi
    else
        echo -e "${YELLOW}⚠️  Proceeding without key backup${NC}"
        echo -e "${YELLOW}    New keys will be generated${NC}"
        echo -e "${YELLOW}    Existing SealedSecrets will NOT work!${NC}"
        USE_EXISTING_KEY=false
    fi
fi
echo ""

# ============================================
# 2. ArgoCD 설치
# ============================================
echo -e "${YELLOW}📦 Step 2: Installing ArgoCD...${NC}"
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
echo -e "${GREEN}✅ ArgoCD installed${NC}"
echo ""

# ============================================
# 3. Sealed Secrets 키 복원 (있으면)
# ============================================
if [ "$USE_EXISTING_KEY" = true ]; then
    echo -e "${YELLOW}🔑 Step 3: Restoring Sealed Secrets key...${NC}"
    
    # 기존 키 삭제 (있다면)
    kubectl delete secret -n kube-system -l sealedsecrets.bitnami.com/sealed-secrets-key 2>/dev/null || true
    
    # 키 복원
    kubectl apply -f "$SEALED_SECRETS_KEY"
    echo -e "${GREEN}✅ Key restored from backup${NC}"
else
    echo -e "${YELLOW}⏭️  Step 3: Skipping key restoration${NC}"
fi
echo ""

# ============================================
# 4. Sealed Secrets Controller 설치
# ============================================
echo -e "${YELLOW}🔐 Step 4: Installing Sealed Secrets Controller...${NC}"
helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets 2>/dev/null || true
helm repo update

helm upgrade --install sealed-secrets sealed-secrets/sealed-secrets \
  -n kube-system \
  --set fullnameOverride=sealed-secrets \
  --wait --timeout=300s
echo -e "${GREEN}✅ Controller installed${NC}"
echo ""

# ============================================
# 5. Controller 재시작 (키 로드)
# ============================================
if [ "$USE_EXISTING_KEY" = true ]; then
    echo -e "${YELLOW}🔄 Step 5: Restarting controller to load key...${NC}"
    kubectl delete pod -n kube-system -l app.kubernetes.io/name=sealed-secrets 2>/dev/null || true
    sleep 5
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=sealed-secrets -n kube-system --timeout=300s
    echo -e "${GREEN}✅ Controller ready with restored key${NC}"
else
    echo -e "${YELLOW}⏭️  Step 5: Controller ready with new key${NC}"
fi
echo ""

# ============================================
# 6. ArgoCD 준비 대기
# ============================================
echo -e "${YELLOW}⏳ Step 6: Waiting for ArgoCD server...${NC}"
kubectl wait --for=condition=available --timeout=600s deployment/argocd-server -n argocd
echo -e "${GREEN}✅ ArgoCD ready${NC}"
echo ""

# ============================================
# 7. 네임스페이스 생성
# ============================================
echo -e "${YELLOW}📁 Step 7: Creating application namespace...${NC}"
kubectl create namespace wealist-dev --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Namespace created${NC}"
echo ""

# ============================================
# 8. CRD 확인
# ============================================
echo -e "${YELLOW}🔍 Step 8: Verifying Sealed Secrets CRD...${NC}"
if kubectl get crd sealedsecrets.bitnami.com &> /dev/null; then
    echo -e "${GREEN}✅ CRD verified${NC}"
else
    echo -e "${RED}❌ CRD not found${NC}"
    exit 1
fi
echo ""

# ============================================
# 9. GitHub 저장소 인증 정보 수집
# ============================================
echo -e "${YELLOW}🔗 Step 9: Collecting GitHub repository credentials...${NC}"
echo ""
read -p "Enter your GitHub username: " GITHUB_USERNAME
echo -n "Enter your GitHub Personal Access Token (with repo permissions): "
read -s GITHUB_TOKEN
echo ""
echo ""

# 입력값 검증
if [ -z "$GITHUB_USERNAME" ] || [ -z "$GITHUB_TOKEN" ]; then
    echo -e "${RED}❌ GitHub credentials are required for ArgoCD repository access${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Credentials collected${NC}"
echo ""

# ============================================
# 10. SealedSecret 적용 (개별 파일들)
# ============================================
echo -e "${YELLOW}🔐 Step 10: Applying SealedSecrets...${NC}"

# dev 환경의 sealed secrets 디렉토리
SEALED_SECRETS_DIR="$ARGOCD_DIR/sealed-secrets/dev"

echo "   Looking for sealed secrets in: $SEALED_SECRETS_DIR"

if [ -d "$SEALED_SECRETS_DIR" ]; then
    echo "   ✅ Found sealed secrets directory"
    
    SEALED_COUNT=0
    for sealed_file in "$SEALED_SECRETS_DIR"/*.yaml; do
        if [ -f "$sealed_file" ]; then
            filename=$(basename "$sealed_file")
            echo "   📄 Applying: $filename"
            kubectl apply -f "$sealed_file"
            SEALED_COUNT=$((SEALED_COUNT + 1))
        fi
    done
    
    if [ $SEALED_COUNT -gt 0 ]; then
        echo ""
        echo -e "${GREEN}   ✅ Applied $SEALED_COUNT SealedSecret(s)${NC}"
        
        # 복호화 확인
        echo "   ⏳ Waiting for decryption (15 seconds)..."
        sleep 15
        
        echo ""
        echo "   Checking decrypted secrets in wealist-dev:"
        kubectl get secrets -n wealist-dev 2>/dev/null || echo "   No secrets found yet"
    else
        echo -e "${YELLOW}   ⚠️  No sealed secret files found in $SEALED_SECRETS_DIR${NC}"
    fi
else
    echo -e "${YELLOW}   ⚠️  Sealed secrets directory not found: $SEALED_SECRETS_DIR${NC}"
    echo ""
    echo "   To create sealed secrets, run:"
    echo "   ./k8s/argocd/scripts/reseal-secrets.sh"
fi
echo ""

# ============================================
# 11. GitHub 저장소 인증
# ============================================
echo -e "${YELLOW}🔗 Step 11: Setting up GitHub repository access...${NC}"

kubectl create secret generic wealist-repo -n argocd \
  --from-literal=type=git \
  --from-literal=url=$REPO_URL \
  --from-literal=username=$GITHUB_USERNAME \
  --from-literal=password=$GITHUB_TOKEN \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl label secret wealist-repo -n argocd \
  argocd.argoproj.io/secret-type=repository --overwrite

echo -e "${GREEN}✅ Repository configured${NC}"
echo ""

# ============================================
# 12. ArgoCD 추가 대기
# ============================================
echo -e "${YELLOW}⏳ Step 12: Final preparations...${NC}"
sleep 10
echo -e "${GREEN}✅ Ready${NC}"
echo ""

# ============================================
# 13. AppProject 생성
# ============================================
echo -e "${YELLOW}🎯 Step 13: Creating AppProject...${NC}"
PROJECT_FILE="$ARGOCD_DIR/apps/project.yaml"

echo "   Looking for: $PROJECT_FILE"

if [ -f "$PROJECT_FILE" ]; then
    echo "   ✅ Found project file"
    kubectl apply -f "$PROJECT_FILE"
    echo -e "${GREEN}✅ AppProject created from file${NC}"
else
    echo -e "${RED}   ❌ Project file not found!${NC}"
    echo ""
    echo "   Expected location: $PROJECT_FILE"
    echo "   Please ensure the project.yaml file exists."
    exit 1
fi
echo ""

# ============================================
# 14. ArgoCD Applications 배포
# ============================================
echo -e "${YELLOW}🌟 Step 14: Deploying ArgoCD Applications...${NC}"

APPS_DIR="$ARGOCD_DIR/apps"
echo "   Scanning directory: $APPS_DIR"
echo ""

if [ -d "$APPS_DIR" ]; then
    # 디렉토리 내용 확인
    echo "   Found application files:"
    ls -1 "$APPS_DIR"/*.yaml 2>/dev/null | while read file; do
        echo "     - $(basename "$file")"
    done
    echo ""
    
    APPLICATION_COUNT=0
    SKIPPED_COUNT=0
    
    for app_file in "$APPS_DIR"/*.yaml; do
        # 파일이 실제로 존재하는지 확인
        if [ -f "$app_file" ]; then
            filename=$(basename "$app_file")
            
            # project.yaml 제외
            if [[ "$filename" == "project.yaml" ]]; then
                echo "   ⏭️  Skipping: $filename (already applied)"
                SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
                continue
            fi
            
            echo "   📄 Applying: $filename"
            if kubectl apply -f "$app_file"; then
                APPLICATION_COUNT=$((APPLICATION_COUNT + 1))
            else
                echo -e "${RED}      ❌ Failed to apply $filename${NC}"
            fi
        fi
    done
    
    echo ""
    
    if [ $APPLICATION_COUNT -gt 0 ]; then
        echo -e "${GREEN}✅ Successfully created $APPLICATION_COUNT Application(s)${NC}"
        if [ $SKIPPED_COUNT -gt 0 ]; then
            echo "   (Skipped $SKIPPED_COUNT file(s))"
        fi
        
        echo ""
        echo "   ⏳ Waiting for applications to sync (10 seconds)..."
        sleep 10
        
        echo ""
        echo "   Current application status:"
        kubectl get applications -n argocd
    else
        echo -e "${YELLOW}⚠️  No application files were applied${NC}"
    fi
else
    echo -e "${RED}   ❌ Applications directory not found: $APPS_DIR${NC}"
    exit 1
fi
echo ""

# ============================================
# 15. 새 키 백업
# ============================================
if [ "$USE_EXISTING_KEY" = false ]; then
    echo -e "${YELLOW}💾 Step 15: Backing up new keys...${NC}"
    NEW_KEY_FILE="$SCRIPT_DIR/sealed-secrets-new-$(date +%Y%m%d-%H%M%S).key"
    kubectl get secret -n kube-system -l sealedsecrets.bitnami.com/sealed-secrets-key -o yaml > "$NEW_KEY_FILE"
    echo -e "${GREEN}✅ New key backed up: $NEW_KEY_FILE${NC}"
    echo -e "${RED}⚠️  IMPORTANT: Store this file securely!${NC}"
else
    echo -e "${YELLOW}⏭️  Step 15: Using existing key${NC}"
fi
echo ""

# ============================================
# 16. 최종 정보
# ============================================
ARGOCD_PASSWORD=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d 2>/dev/null || echo "Password not found")

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ Deployment Complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "🌐 ArgoCD Access:"
echo "   URL:      https://localhost:8079"
echo "   Username: admin"
echo "   Password: $ARGOCD_PASSWORD"
echo ""
echo "🔐 Sealed Secrets:"
echo "   Controller: sealed-secrets (kube-system)"
if [ "$USE_EXISTING_KEY" = true ]; then
    echo "   Key:        Restored from backup ✅"
else
    echo "   Key:        Newly generated ⚠️"
    echo "   Backup:     $NEW_KEY_FILE"
fi
echo ""
echo "📂 Directory Paths:"
echo "   Apps:           $ARGOCD_DIR/apps/"
echo "   Sealed Secrets: $ARGOCD_DIR/sealed-secrets/dev/"
echo "   Scripts:        $SCRIPT_DIR/"
echo ""
echo "🔍 Verification Commands:"
echo "   kubectl get applications -n argocd"
echo "   kubectl get appprojects -n argocd"
echo "   kubectl get pods -n wealist-dev"
echo "   kubectl get sealedsecrets -n wealist-dev"
echo "   kubectl get secrets -n wealist-dev"
echo ""
echo "📊 Current Status:"
echo ""
echo "AppProjects:"
kubectl get appprojects -n argocd 2>/dev/null || echo "   No projects found"
echo ""
echo "Applications:"
kubectl get applications -n argocd 2>/dev/null || echo "   No applications found"
echo ""
echo "Pods in wealist-dev:"
kubectl get pods -n wealist-dev 2>/dev/null || echo "   No pods running yet"
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "📝 Next Steps:"
echo ""
echo "   1. Monitor ArgoCD applications:"
echo "      watch kubectl get applications -n argocd"
echo ""
echo "   2. Check application sync status:"
echo "      argocd app list"
echo "      argocd app get wealist-apps-dev"
echo "      argocd app get wealist-sealed-secrets"
echo ""
echo "   3. View application details in UI:"
echo "      kubectl port-forward svc/argocd-server -n argocd 8079:443"
echo "      Open: https://localhost:8079"
echo ""
echo "   4. Monitor pod deployments:"
echo "      watch kubectl get pods -n wealist-dev"
echo ""
echo "   5. If applications are OutOfSync:"
echo "      argocd app sync wealist-apps-dev"
echo "      argocd app sync wealist-sealed-secrets"
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Port-forward 옵션
read -p "Start port-forward to ArgoCD now? [Y/n]: " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    echo ""
    echo "🌐 Starting port-forward to ArgoCD..."
    echo "   Access ArgoCD at: https://localhost:8079"
    echo "   Username: admin"
    echo "   Password: $ARGOCD_PASSWORD"
    echo ""
    echo "   Press Ctrl+C to stop port-forward"
    echo ""
    kubectl port-forward svc/argocd-server -n argocd 8079:443
fi