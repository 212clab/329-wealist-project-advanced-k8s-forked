# 빠른시작

# docker-compose 환경

./docker/scripts/dev.sh up => .env.dev 파일 주의하세요!

# local-kind 환경(ns: wealist-dev)

=> k8s/base/shared/secret-shared.yaml, services/auth-service/k8s/base/secret.yaml 에 google_client 관련 설정해주세요!

# 1. 클러스터 생성 (동일)

make kind-setup

# 2. 이미지 빌드/로드 (동일)

make kind-load-images

# 3. 배포 (도메인 선택)

make kind-apply # localhost 접속용
make local-kind-apply # local.wealist.co.kr 접속용

## 그 외

kind get clusters (클러스터 확인)
kubectl get namespaces (ns 확인)

## 한꺼번에 클러스터 재설정

make kind-delete && make kind-setup && make infra-setup && make k8s-deploy-services
