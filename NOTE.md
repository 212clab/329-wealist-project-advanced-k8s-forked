# 빠른시작

# docker-compose 환경

./docker/scripts/dev.sh up => .env.dev 파일 주의하세요!

# local-kind 환경(ns: wealist-dev)

=> k8s/base/shared/secret-shared.yaml, services/auth-service/k8s/base/secret.yaml 에 google_client 관련 설정해주세요!

## ####### 1. local에서 클러스터 생성 (동일)

```
# 1-1. 클러스터 생성
make kind-setup

# 1-2. 이미지 빌드 및 로드
make kind-load-images

# 1-3. Secrets 설정 (필수!)
cp helm/environments/secrets.example.yaml helm/environments/local-kind-secrets.yaml

# 파일 편집하여 Google OAuth 자격증명 입력 (GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET)
# 1-4. Helm으로 전체 배포

make helm-install-all ENV=local-kind

# 1-5. 상태 확인
make status

# 접속: http://localhost
```

## 그 외

kind get clusters (클러스터 확인)
kubectl get namespaces (ns 확인)

## 한꺼번에 클러스터 재설정

make kind-delete && make kind-setup && make infra-setup && make k8s-deploy-services
