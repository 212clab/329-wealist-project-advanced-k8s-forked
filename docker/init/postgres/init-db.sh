#!/bin/bash
set -e

# PostgreSQL 초기화 스크립트
# 세 개의 독립된 데이터베이스와 사용자를 생성합니다

echo "🚀 weAlist 데이터베이스 초기화 시작..."

# User Service Database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE ${USER_DB_NAME};
    CREATE USER ${USER_DB_USER} WITH PASSWORD '${USER_DB_PASSWORD}';
    GRANT ALL PRIVILEGES ON DATABASE ${USER_DB_NAME} TO ${USER_DB_USER};
    \c ${USER_DB_NAME}
    GRANT ALL ON SCHEMA public TO ${USER_DB_USER};
EOSQL

echo "✅ User 서비스 데이터베이스 생성 완료: ${USER_DB_NAME}"

# Board Service Database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE ${BOARD_DB_NAME};
    CREATE USER ${BOARD_DB_USER} WITH PASSWORD '${BOARD_DB_PASSWORD}';
    GRANT ALL PRIVILEGES ON DATABASE ${BOARD_DB_NAME} TO ${BOARD_DB_USER};
    \c ${BOARD_DB_NAME}
    GRANT ALL ON SCHEMA public TO ${BOARD_DB_USER};
EOSQL

echo "✅ Board 서비스 데이터베이스 생성 완료: ${BOARD_DB_NAME}"

# Chat Service Database (🔥 추가)
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE ${CHAT_DB_NAME};
    CREATE USER ${CHAT_DB_USER} WITH PASSWORD '${CHAT_DB_PASSWORD}';
    GRANT ALL PRIVILEGES ON DATABASE ${CHAT_DB_NAME} TO ${CHAT_DB_USER};
    \c ${CHAT_DB_NAME}
    GRANT ALL ON SCHEMA public TO ${CHAT_DB_USER};
EOSQL

echo "✅ Chat 서비스 데이터베이스 생성 완료: ${CHAT_DB_NAME}"

# Notification Service Database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE ${NOTI_DB_NAME};
    CREATE USER ${NOTI_DB_USER} WITH PASSWORD '${NOTI_DB_PASSWORD}';
    GRANT ALL PRIVILEGES ON DATABASE ${NOTI_DB_NAME} TO ${NOTI_DB_USER};
    \c ${NOTI_DB_NAME}
    GRANT ALL ON SCHEMA public TO ${NOTI_DB_USER};
EOSQL

echo "✅ Notification 서비스 데이터베이스 생성 완료: ${NOTI_DB_NAME}"

# Storage Service Database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE ${STORAGE_DB_NAME};
    CREATE USER ${STORAGE_DB_USER} WITH PASSWORD '${STORAGE_DB_PASSWORD}';
    GRANT ALL PRIVILEGES ON DATABASE ${STORAGE_DB_NAME} TO ${STORAGE_DB_USER};
    \c ${STORAGE_DB_NAME}
    GRANT ALL ON SCHEMA public TO ${STORAGE_DB_USER};
EOSQL

echo "✅ Storage 서비스 데이터베이스 생성 완료: ${STORAGE_DB_NAME}"

# Video Service Database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE ${VIDEO_DB_NAME:-wealist_video_db};
    CREATE USER ${VIDEO_DB_USER:-video_service} WITH PASSWORD '${VIDEO_DB_PASSWORD:-video_service_password}';
    GRANT ALL PRIVILEGES ON DATABASE ${VIDEO_DB_NAME:-wealist_video_db} TO ${VIDEO_DB_USER:-video_service};
    \c ${VIDEO_DB_NAME:-wealist_video_db}
    GRANT ALL ON SCHEMA public TO ${VIDEO_DB_USER:-video_service};
EOSQL

echo "✅ Video 서비스 데이터베이스 생성 완료: ${VIDEO_DB_NAME:-wealist_video_db}"

# SonarQube Database
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE ${SONARQUBE_DB_NAME};
    CREATE USER ${SONARQUBE_DB_USER} WITH PASSWORD '${SONARQUBE_DB_PASSWORD}';
    GRANT ALL PRIVILEGES ON DATABASE ${SONARQUBE_DB_NAME} TO ${SONARQUBE_DB_USER};
    \c ${SONARQUBE_DB_NAME}
    GRANT ALL ON SCHEMA public TO ${SONARQUBE_DB_USER};
EOSQL

echo "✅ SonarQube 데이터베이스 생성 완료: ${SONARQUBE_DB_NAME}"

echo "🎉 모든 데이터베이스 초기화 완료!"
echo "📋 생성된 데이터베이스:"
echo "   - ${USER_DB_NAME} (User Service)"
echo "   - ${BOARD_DB_NAME} (Board Service)"
echo "   - ${CHAT_DB_NAME} (Chat Service)"
echo "   - ${NOTI_DB_NAME} (Notification Service)"
echo "   - ${STORAGE_DB_NAME} (Storage Service)"
echo "   - ${VIDEO_DB_NAME:-wealist_video_db} (Video Service)"
echo "   - ${SONARQUBE_DB_NAME} (SonarQube - Code Quality)"
