# wealist-argo-helm

이 프로젝트는 Helm Chart와 ArgoCD를 사용하여 Wealist 서비스를 배포하는 저장소입니다.

## 📁 프로젝트 구조

- **charts/** - 각 서비스별 개별 Helm Chart
- **environments/** - 환경별 설정 파일 (localhost, dev, staging)
  - Chart 템플릿은 변경하지 않고, 모든 환경에서 공통으로 사용합니다
  - 환경별로 다른 값들만 이 디렉토리에서 관리합니다

> ⚠️ **중요**: 배포 전에 각 환경의 `secret` 파일에 환경변수 값을 반드시 설정해야 합니다.

---

## 🚀 빠른 시작 가이드

Kind 클러스터에서 Wealist 서비스를 배포하는 두 가지 방법이 있습니다.

| 환경          | 이미지 레지스트리                   | 사용 목적                               |
| ------------- | ----------------------------------- | --------------------------------------- |
| **localhost** | 로컬 레지스트리 (`localhost:5001`)  | 로컬에서 직접 빌드한 이미지로 테스트    |
| **dev**       | GitHub Container Registry (ghcr.io) | GitHub에 푸시된 이미지로 개발 환경 구성 |

---

## 🏠 방법 1: Localhost 환경 배포

로컬에서 빌드한 Docker 이미지를 사용하여 배포합니다.

### Step 1. Kind 클러스터 및 로컬 레지스트리 생성

```bash
make kind-localhost-setup
```

이 명령어는 다음을 수행합니다:

- Kind 클러스터 생성
- 로컬 Docker 레지스트리 생성 (`localhost:5001`)
- 클러스터와 레지스트리 연결

### Step 2. Docker 이미지 빌드 및 푸시

각 서비스 디렉토리에서 Docker 이미지를 빌드하고 로컬 레지스트리에 푸시합니다:

```bash
# 예시: 서비스 이미지 빌드 및 푸시
docker build -t localhost:5001/service-name:latest .
docker push localhost:5001/service-name:latest
```

이미지 업로드 확인:

```bash
curl -s http://localhost:5001/v2/_catalog | jq
```

### Step 3. Helm 차트 배포

```bash
make helm-install-all ENV=localhost
```

✅ **Localhost 환경 배포 완료!**

---

## 🌐 방법 2: Dev 환경 배포

GitHub Container Registry(ghcr.io)에 있는 이미지를 사용하여 배포합니다.

### Step 1. Kind 클러스터 생성

```bash
make kind-dev-setup
```

이 명령어는 다음을 수행합니다:

- Kind 클러스터 생성
- ghcr.io 접근을 위한 설정

### Step 2. Helm 차트 배포

```bash
make helm-install-all ENV=dev
```

✅ **Dev 환경 배포 완료!**

---

## 📋 환경별 비교

| 항목          | Localhost                             | Dev                             |
| ------------- | ------------------------------------- | ------------------------------- |
| 클러스터 생성 | `make kind-localhost-setup`           | `make kind-dev-setup`           |
| Helm 배포     | `make helm-install-all ENV=localhost` | `make helm-install-all ENV=dev` |
| 이미지 소스   | 로컬 레지스트리                       | ghcr.io                         |
| 이미지 빌드   | 직접 빌드 필요                        | GitHub Actions로 자동 빌드      |
| 적합한 상황   | 로컬 개발 및 테스트                   | CI/CD 통합 테스트               |

---

## 💡 주요 참고사항

- Chart 템플릿은 수정하지 말고, `environments/` 디렉토리의 값만 수정하세요
- 환경별 시크릿 설정을 잊지 마세요
- 클러스터 삭제: `kind delete cluster --name wealist`
