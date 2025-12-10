# 빠른시작

# docker-compose 환경

./docker/scripts/dev.sh up => .env.dev 파일 주의하세요!

# local-kind 환경(ns: wealist-dev)

=> k8s/base/shared/secret-shared.yaml, services/auth-service/k8s/base/secret.yaml 에 google_client 관련 설정해주세요!

make kind-setup (cluster 생성 + nginx controller 셋팅)

make infra-setup (인프라 이미지 로드 + 배포 + 대기)
-> kubectl get pods -n wealist-dev 로 확인하세요

make k8s-deploy-services (빌드 + ns 생성 + 배포)
-> kubectl get pods -n wealist-dev 로 확인하세요
make status

## 포트포워딩 대신 로컬에 등록해서 wealist.local로

echo "127.0.0.1 wealist.local" | sudo tee -a /etc/hosts

## 그 외

kind get clusters (클러스터 확인)
kubectl get namespaces (ns 확인)

## 한꺼번에 클러스터 재설정

make kind-delete && make kind-setup && make infra-setup && make k8s-deploy-services
