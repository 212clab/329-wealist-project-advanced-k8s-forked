#!/bin/bash
# =============================================================================
# External Database Setup Script for EXTERNAL_DB=true mode
# =============================================================================
# Supports: macOS (Homebrew) and Linux (apt/systemctl)
#
# Usage: ./setup-external-db.sh [--auto-yes]
#   --auto-yes: Skip confirmation prompts (for CI/CD)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse arguments
AUTO_YES=false
if [[ "$1" == "--auto-yes" ]]; then
    AUTO_YES=true
fi

# Detect OS
OS_TYPE="unknown"
if [[ "$(uname)" == "Darwin" ]]; then
    OS_TYPE="macos"
elif [[ -f /etc/debian_version ]]; then
    OS_TYPE="debian"
elif [[ -f /etc/redhat-release ]]; then
    OS_TYPE="redhat"
fi

echo -e "${BLUE}=============================================${NC}"
echo -e "${BLUE}  외부 데이터베이스 설정 (EXTERNAL_DB=true)${NC}"
echo -e "${BLUE}=============================================${NC}"
echo -e "  감지된 OS (Detected OS): ${YELLOW}${OS_TYPE}${NC}"
echo ""

# Function to ask user confirmation
ask_confirm() {
    if [[ "$AUTO_YES" == "true" ]]; then
        return 0
    fi

    local prompt="$1"
    local response

    while true; do
        read -p "$prompt [Y/n]: " response
        case "$response" in
            [Yy]*|"") return 0 ;;
            [Nn]*) return 1 ;;
            *) echo "Y 또는 N으로 답해주세요. (Please answer Y or N)" ;;
        esac
    done
}

# Function to check if a command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# =============================================================================
# macOS Functions
# =============================================================================
macos_check_postgres() {
    if command_exists psql; then
        echo -e "  ${GREEN}✓${NC} PostgreSQL 설치됨 (installed)"
        if brew services list 2>/dev/null | grep -E "postgresql.*started" >/dev/null; then
            echo -e "  ${GREEN}✓${NC} PostgreSQL 실행 중 (running via brew services)"
            return 0
        elif pg_isready -h localhost &>/dev/null; then
            echo -e "  ${GREEN}✓${NC} PostgreSQL 실행 중 (running)"
            return 0
        else
            echo -e "  ${YELLOW}!${NC} PostgreSQL 설치됨, 실행 안됨 (installed but not running)"
            return 1
        fi
    else
        echo -e "  ${RED}✗${NC} PostgreSQL 미설치 (not installed)"
        return 2
    fi
}

macos_install_postgres() {
    echo -e "${BLUE}Homebrew로 PostgreSQL 설치 중... (Installing via Homebrew)${NC}"
    brew install postgresql@17 2>/dev/null || brew install postgresql
    echo -e "${GREEN}PostgreSQL 설치 완료! (installed!)${NC}"
}

macos_start_postgres() {
    echo -e "${BLUE}PostgreSQL 시작 중... (Starting)${NC}"
    brew services start postgresql@17 2>/dev/null || brew services start postgresql 2>/dev/null || true
    sleep 2
    echo -e "${GREEN}PostgreSQL 시작됨! (started!)${NC}"
}

macos_check_redis() {
    if command_exists redis-cli; then
        echo -e "  ${GREEN}✓${NC} Redis 설치됨 (installed)"
        if redis-cli ping &>/dev/null; then
            echo -e "  ${GREEN}✓${NC} Redis 실행 중 (running)"
            return 0
        else
            echo -e "  ${YELLOW}!${NC} Redis 설치됨, 실행 안됨 (installed but not running)"
            return 1
        fi
    else
        echo -e "  ${RED}✗${NC} Redis 미설치 (not installed)"
        return 2
    fi
}

macos_install_redis() {
    echo -e "${BLUE}Homebrew로 Redis 설치 중... (Installing via Homebrew)${NC}"
    brew install redis
    echo -e "${GREEN}Redis 설치 완료! (installed!)${NC}"
}

macos_start_redis() {
    echo -e "${BLUE}Redis 시작 중... (Starting)${NC}"
    brew services start redis
    sleep 1
    echo -e "${GREEN}Redis 시작됨! (started!)${NC}"
}

macos_check_redis_binding() {
    # Check if Redis is bound to 0.0.0.0
    local binding=$(lsof -i :6379 2>/dev/null | grep -E "redis.*LISTEN" | head -1)

    if echo "$binding" | grep -q "\*:6379"; then
        echo -e "  ${GREEN}✓${NC} Redis가 0.0.0.0에 바인딩됨 (bound to all interfaces)"
        return 0
    elif echo "$binding" | grep -q "localhost:6379"; then
        echo -e "  ${YELLOW}!${NC} Redis가 localhost에만 바인딩됨 (only bound to 127.0.0.1)"
        echo -e "      Kind pod에서 Redis에 접근 불가 (cannot access)"
        return 1
    else
        echo -e "  ${YELLOW}?${NC} Redis 바인딩 상태 확인 불가 (could not determine)"
        return 1
    fi
}

macos_configure_redis() {
    echo -e "${BLUE}Kind 클러스터 접근을 위한 Redis 설정 중... (Configuring for Kind)${NC}"
    echo ""
    echo -e "  ${YELLOW}필요한 변경사항 (Required changes):${NC}"
    echo -e "    - bind 0.0.0.0 (외부 연결 허용)"
    echo -e "    - protected-mode no (비밀번호 없이 연결 허용)"
    echo ""

    # Find Redis config file
    local REDIS_CONF=""
    local REDIS_PREFIX=$(brew --prefix redis 2>/dev/null || brew --prefix 2>/dev/null)

    if [[ -f "$REDIS_PREFIX/etc/redis.conf" ]]; then
        REDIS_CONF="$REDIS_PREFIX/etc/redis.conf"
    elif [[ -f "/opt/homebrew/etc/redis.conf" ]]; then
        REDIS_CONF="/opt/homebrew/etc/redis.conf"
    elif [[ -f "/usr/local/etc/redis.conf" ]]; then
        REDIS_CONF="/usr/local/etc/redis.conf"
    fi

    if [[ -z "$REDIS_CONF" || ! -f "$REDIS_CONF" ]]; then
        echo -e "  ${YELLOW}!${NC} redis.conf 파일을 찾을 수 없음 (could not find)"
        echo -e "  ${YELLOW}!${NC} 수동으로 Redis를 설정해주세요 (Please configure manually):"
        echo "      1. redis.conf 파일 찾기"
        echo "      2. 변경: bind 127.0.0.1 → bind 0.0.0.0"
        echo "      3. 변경: protected-mode yes → protected-mode no"
        echo "      4. 실행: brew services restart redis"
        return 1
    fi

    echo -e "  설정 파일 발견 (Found config): ${BLUE}$REDIS_CONF${NC}"

    # Backup config
    cp "$REDIS_CONF" "$REDIS_CONF.backup.$(date +%Y%m%d%H%M%S)" 2>/dev/null || true

    # Update bind address
    if grep -q "^bind 127.0.0.1" "$REDIS_CONF" 2>/dev/null; then
        sed -i '' 's/^bind 127.0.0.1.*/bind 0.0.0.0/' "$REDIS_CONF"
        echo -e "  ${GREEN}✓${NC} bind를 0.0.0.0으로 변경함 (updated)"
    elif grep -q "^bind " "$REDIS_CONF" 2>/dev/null; then
        sed -i '' 's/^bind .*/bind 0.0.0.0/' "$REDIS_CONF"
        echo -e "  ${GREEN}✓${NC} bind를 0.0.0.0으로 변경함 (updated)"
    else
        echo "bind 0.0.0.0" >> "$REDIS_CONF"
        echo -e "  ${GREEN}✓${NC} bind 0.0.0.0 추가됨 (added)"
    fi

    # Disable protected mode
    if grep -q "^protected-mode yes" "$REDIS_CONF" 2>/dev/null; then
        sed -i '' 's/^protected-mode yes/protected-mode no/' "$REDIS_CONF"
        echo -e "  ${GREEN}✓${NC} protected-mode 비활성화됨 (disabled)"
    elif ! grep -q "^protected-mode" "$REDIS_CONF" 2>/dev/null; then
        echo "protected-mode no" >> "$REDIS_CONF"
        echo -e "  ${GREEN}✓${NC} protected-mode no 추가됨 (added)"
    else
        echo -e "  ${GREEN}✓${NC} protected-mode 이미 설정됨 (already configured)"
    fi

    # Restart Redis
    echo -e "  ${BLUE}Redis 재시작 중... (Restarting)${NC}"
    brew services restart redis
    sleep 2
    echo -e "  ${GREEN}✓${NC} Redis 재시작됨 (restarted)"

    return 0
}

macos_check_postgres_binding() {
    # Check if PostgreSQL is listening on all interfaces
    local binding=$(lsof -i :5432 2>/dev/null | grep -E "postgres.*LISTEN" | head -1)

    if echo "$binding" | grep -q "\*:5432\|postgresql:5432"; then
        echo -e "  ${GREEN}✓${NC} PostgreSQL이 모든 인터페이스에 바인딩됨 (bound to all interfaces)"
        return 0
    elif echo "$binding" | grep -q "localhost:5432"; then
        echo -e "  ${YELLOW}!${NC} PostgreSQL이 localhost에만 바인딩됨 (only bound to localhost)"
        return 1
    else
        # Could be listening on * but showing differently
        echo -e "  ${GREEN}✓${NC} PostgreSQL 바인딩 (기본 macOS 설정은 외부 허용)"
        return 0
    fi
}

macos_configure_postgres() {
    echo -e "${BLUE}Kind 클러스터 접근을 위한 PostgreSQL 설정 중... (Configuring for Kind)${NC}"
    echo ""
    echo -e "  ${YELLOW}필요사항 (Required):${NC}"
    echo -e "    - listen_addresses = '*' (외부 연결 허용)"
    echo -e "    - pg_hba.conf에 172.18.0.0/16 네트워크 허용"
    echo ""

    # Find PostgreSQL data directory
    local PG_DATA=$(psql postgres -t -c "SHOW data_directory" 2>/dev/null | tr -d ' ')

    if [[ -z "$PG_DATA" || ! -d "$PG_DATA" ]]; then
        # Try common Homebrew locations
        if [[ -d "/opt/homebrew/var/postgresql@17" ]]; then
            PG_DATA="/opt/homebrew/var/postgresql@17"
        elif [[ -d "/opt/homebrew/var/postgresql@14" ]]; then
            PG_DATA="/opt/homebrew/var/postgresql@14"
        elif [[ -d "/opt/homebrew/var/postgres" ]]; then
            PG_DATA="/opt/homebrew/var/postgres"
        elif [[ -d "/usr/local/var/postgres" ]]; then
            PG_DATA="/usr/local/var/postgres"
        fi
    fi

    if [[ -z "$PG_DATA" || ! -d "$PG_DATA" ]]; then
        echo -e "  ${YELLOW}!${NC} PostgreSQL 데이터 디렉토리를 찾을 수 없음 (could not find)"
        echo -e "  ${YELLOW}!${NC} 수동으로 PostgreSQL을 설정해주세요 (Please configure manually):"
        echo "      1. postgresql.conf 파일 찾기"
        echo "      2. 설정: listen_addresses = '*'"
        echo "      3. pg_hba.conf에 추가: host all all 172.18.0.0/16 trust"
        echo "      4. PostgreSQL 재시작"
        return 1
    fi

    echo -e "  데이터 디렉토리 발견 (Found data directory): ${BLUE}$PG_DATA${NC}"

    local PG_CONF="$PG_DATA/postgresql.conf"
    local PG_HBA="$PG_DATA/pg_hba.conf"

    # Backup configs
    cp "$PG_CONF" "$PG_CONF.backup.$(date +%Y%m%d%H%M%S)" 2>/dev/null || true
    cp "$PG_HBA" "$PG_HBA.backup.$(date +%Y%m%d%H%M%S)" 2>/dev/null || true

    # Update listen_addresses
    if grep -q "^listen_addresses" "$PG_CONF" 2>/dev/null; then
        sed -i '' "s/^listen_addresses.*/listen_addresses = '*'/" "$PG_CONF"
        echo -e "  ${GREEN}✓${NC} listen_addresses = '*' 변경됨 (updated)"
    elif grep -q "^#listen_addresses" "$PG_CONF" 2>/dev/null; then
        sed -i '' "s/^#listen_addresses.*/listen_addresses = '*'/" "$PG_CONF"
        echo -e "  ${GREEN}✓${NC} listen_addresses = '*' 활성화됨 (enabled)"
    else
        echo "listen_addresses = '*'" >> "$PG_CONF"
        echo -e "  ${GREEN}✓${NC} listen_addresses = '*' 추가됨 (added)"
    fi

    # Add Kind network to pg_hba.conf
    if ! grep -q "172.18.0.0/16" "$PG_HBA" 2>/dev/null; then
        echo "" >> "$PG_HBA"
        echo "# Kind cluster network access" >> "$PG_HBA"
        echo "host all all 172.18.0.0/16 trust" >> "$PG_HBA"
        echo "host all all 172.17.0.0/16 trust" >> "$PG_HBA"
        echo -e "  ${GREEN}✓${NC} pg_hba.conf에 Kind 네트워크 추가됨 (added Kind network)"
    else
        echo -e "  ${GREEN}✓${NC} Kind 네트워크가 이미 pg_hba.conf에 있음 (already exists)"
    fi

    # Restart PostgreSQL
    echo -e "  ${BLUE}PostgreSQL 재시작 중... (Restarting)${NC}"
    brew services restart postgresql@17 2>/dev/null || brew services restart postgresql 2>/dev/null || true
    sleep 3
    echo -e "  ${GREEN}✓${NC} PostgreSQL 재시작됨 (restarted)"

    return 0
}

macos_create_databases() {
    echo -e "${YELLOW}[보너스] Wealist 데이터베이스 및 사용자 생성 중... (Creating databases & users)${NC}"

    # Define services with their roles and databases (Bash 3 compatible)
    SERVICES="user_service:wealist_user_service_db
board_service:wealist_board_service_db
chat_service:wealist_chat_service_db
noti_service:wealist_noti_service_db
storage_service:wealist_storage_service_db
video_service:wealist_video_service_db"

    echo "$SERVICES" | while IFS=: read -r role db; do
        password="${role}_password"

        # Create role if not exists
        if psql postgres -tc "SELECT 1 FROM pg_roles WHERE rolname = '$role'" 2>/dev/null | grep -q 1; then
            echo -e "  ${GREEN}✓${NC} 사용자 존재 (role exists): $role"
        else
            psql postgres -c "CREATE ROLE $role WITH LOGIN PASSWORD '$password'" &>/dev/null && \
                echo -e "  ${GREEN}✓${NC} 사용자 생성됨 (role created): $role" || \
                echo -e "  ${YELLOW}!${NC} 사용자 생성 실패 (could not create role): $role"
        fi

        # Create database if not exists
        if psql postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$db'" 2>/dev/null | grep -q 1; then
            echo -e "  ${GREEN}✓${NC} 데이터베이스 존재 (exists): $db"
        else
            psql postgres -c "CREATE DATABASE $db OWNER $role" &>/dev/null && \
                echo -e "  ${GREEN}✓${NC} 생성됨 (created): $db" || \
                echo -e "  ${YELLOW}!${NC} 생성 실패 (could not create): $db"
        fi

        # Grant privileges
        psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE $db TO $role" &>/dev/null

        # Enable uuid-ossp and grant schema access
        psql -d "$db" -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"" &>/dev/null
        psql -d "$db" -c "GRANT ALL ON SCHEMA public TO $role" &>/dev/null
    done
}

# =============================================================================
# Linux (Debian/Ubuntu) Functions
# =============================================================================
linux_check_postgres() {
    if command_exists psql; then
        echo -e "  ${GREEN}✓${NC} PostgreSQL 설치됨 (installed)"
        if systemctl is-active --quiet postgresql 2>/dev/null; then
            echo -e "  ${GREEN}✓${NC} PostgreSQL 실행 중 (running)"
            return 0
        elif pg_isready -h localhost &>/dev/null; then
            echo -e "  ${GREEN}✓${NC} PostgreSQL 실행 중 (running)"
            return 0
        else
            echo -e "  ${YELLOW}!${NC} PostgreSQL 설치됨, 실행 안됨 (installed but not running)"
            return 1
        fi
    else
        echo -e "  ${RED}✗${NC} PostgreSQL 미설치 (not installed)"
        return 2
    fi
}

linux_install_postgres() {
    echo -e "${BLUE}PostgreSQL 설치 중... (Installing)${NC}"
    sudo apt-get update
    sudo apt-get install -y postgresql postgresql-contrib
    echo -e "${GREEN}PostgreSQL 설치 완료! (installed!)${NC}"
}

linux_start_postgres() {
    echo -e "${BLUE}PostgreSQL 시작 중... (Starting)${NC}"
    sudo systemctl start postgresql
    sudo systemctl enable postgresql
    sleep 2
    echo -e "${GREEN}PostgreSQL 시작됨! (started!)${NC}"
}

linux_configure_postgres() {
    echo -e "${BLUE}Kind 네트워크 접근을 위한 PostgreSQL 설정 중... (Configuring for Kind)${NC}"

    # Find PostgreSQL config directory
    PG_CONF_DIR=$(find /etc/postgresql -name "postgresql.conf" -exec dirname {} \; 2>/dev/null | head -1)

    if [[ -n "$PG_CONF_DIR" ]]; then
        # Update listen_addresses
        if ! grep -q "listen_addresses = '\*'" "$PG_CONF_DIR/postgresql.conf" 2>/dev/null; then
            sudo sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/" "$PG_CONF_DIR/postgresql.conf"
            sudo sed -i "s/listen_addresses = 'localhost'/listen_addresses = '*'/" "$PG_CONF_DIR/postgresql.conf"
            echo -e "  ${GREEN}✓${NC} listen_addresses = '*' 변경됨 (updated)"
        else
            echo -e "  ${GREEN}✓${NC} listen_addresses 이미 설정됨 (already configured)"
        fi

        # Add Kind network to pg_hba.conf
        if ! grep -q "172.18.0.0/16" "$PG_CONF_DIR/pg_hba.conf" 2>/dev/null; then
            echo "# Kind cluster network access" | sudo tee -a "$PG_CONF_DIR/pg_hba.conf" > /dev/null
            echo "host all all 172.18.0.0/16 trust" | sudo tee -a "$PG_CONF_DIR/pg_hba.conf" > /dev/null
            echo "host all all 172.17.0.0/16 trust" | sudo tee -a "$PG_CONF_DIR/pg_hba.conf" > /dev/null
            echo -e "  ${GREEN}✓${NC} pg_hba.conf에 Kind 네트워크 (172.18.0.0/16) 추가됨"
        else
            echo -e "  ${GREEN}✓${NC} Kind 네트워크가 이미 pg_hba.conf에 있음 (already exists)"
        fi

        # Restart PostgreSQL to apply changes
        sudo systemctl restart postgresql
        echo -e "  ${GREEN}✓${NC} PostgreSQL 재시작됨 (restarted)"
    else
        echo -e "  ${YELLOW}!${NC} PostgreSQL 설정 디렉토리를 찾을 수 없음 (could not find)"
    fi
}

linux_check_redis() {
    if command_exists redis-cli; then
        echo -e "  ${GREEN}✓${NC} Redis 설치됨 (installed)"
        if systemctl is-active --quiet redis-server 2>/dev/null || systemctl is-active --quiet redis 2>/dev/null; then
            echo -e "  ${GREEN}✓${NC} Redis 실행 중 (running)"
            return 0
        elif redis-cli ping &>/dev/null; then
            echo -e "  ${GREEN}✓${NC} Redis 실행 중 (running)"
            return 0
        else
            echo -e "  ${YELLOW}!${NC} Redis 설치됨, 실행 안됨 (installed but not running)"
            return 1
        fi
    else
        echo -e "  ${RED}✗${NC} Redis 미설치 (not installed)"
        return 2
    fi
}

linux_install_redis() {
    echo -e "${BLUE}Redis 설치 중... (Installing)${NC}"
    sudo apt-get update
    sudo apt-get install -y redis-server
    echo -e "${GREEN}Redis 설치 완료! (installed!)${NC}"
}

linux_start_redis() {
    echo -e "${BLUE}Redis 시작 중... (Starting)${NC}"
    sudo systemctl start redis-server || sudo systemctl start redis
    sudo systemctl enable redis-server || sudo systemctl enable redis
    sleep 1
    echo -e "${GREEN}Redis 시작됨! (started!)${NC}"
}

linux_configure_redis() {
    echo -e "${BLUE}Kind 네트워크 접근을 위한 Redis 설정 중... (Configuring for Kind)${NC}"

    REDIS_CONF="/etc/redis/redis.conf"
    if [[ -f "$REDIS_CONF" ]]; then
        # Update bind address
        if grep -q "bind 127.0.0.1" "$REDIS_CONF" 2>/dev/null; then
            sudo sed -i 's/bind 127.0.0.1.*/bind 0.0.0.0/' "$REDIS_CONF"
            echo -e "  ${GREEN}✓${NC} bind를 0.0.0.0으로 변경함 (updated)"
        elif ! grep -q "bind 0.0.0.0" "$REDIS_CONF" 2>/dev/null; then
            echo "bind 0.0.0.0" | sudo tee -a "$REDIS_CONF" > /dev/null
            echo -e "  ${GREEN}✓${NC} bind 0.0.0.0 추가됨 (added)"
        else
            echo -e "  ${GREEN}✓${NC} bind 이미 설정됨 (already configured)"
        fi

        # Disable protected mode
        if grep -q "protected-mode yes" "$REDIS_CONF" 2>/dev/null; then
            sudo sed -i 's/protected-mode yes/protected-mode no/' "$REDIS_CONF"
            echo -e "  ${GREEN}✓${NC} protected-mode 비활성화됨 (disabled)"
        else
            echo -e "  ${GREEN}✓${NC} protected-mode 이미 비활성화됨 (already disabled)"
        fi

        # Restart Redis to apply changes
        sudo systemctl restart redis-server || sudo systemctl restart redis
        echo -e "  ${GREEN}✓${NC} Redis 재시작됨 (restarted)"
    else
        echo -e "  ${YELLOW}!${NC} Redis 설정 파일을 찾을 수 없음 (could not find): $REDIS_CONF"
    fi
}

linux_create_databases() {
    echo -e "${YELLOW}[보너스] Wealist 데이터베이스 및 사용자 생성 중... (Creating databases & users)${NC}"

    # Define services with their roles and databases (Bash 3 compatible)
    SERVICES="user_service:wealist_user_service_db
board_service:wealist_board_service_db
chat_service:wealist_chat_service_db
noti_service:wealist_noti_service_db
storage_service:wealist_storage_service_db
video_service:wealist_video_service_db"

    echo "$SERVICES" | while IFS=: read -r role db; do
        password="${role}_password"

        # Create role if not exists
        if sudo -u postgres psql -tc "SELECT 1 FROM pg_roles WHERE rolname = '$role'" 2>/dev/null | grep -q 1; then
            echo -e "  ${GREEN}✓${NC} 사용자 존재 (role exists): $role"
        else
            sudo -u postgres psql -c "CREATE ROLE $role WITH LOGIN PASSWORD '$password'" &>/dev/null && \
                echo -e "  ${GREEN}✓${NC} 사용자 생성됨 (role created): $role" || \
                echo -e "  ${YELLOW}!${NC} 사용자 생성 실패 (could not create role): $role"
        fi

        # Create database if not exists
        if sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname = '$db'" 2>/dev/null | grep -q 1; then
            echo -e "  ${GREEN}✓${NC} 데이터베이스 존재 (exists): $db"
        else
            sudo -u postgres psql -c "CREATE DATABASE $db OWNER $role" &>/dev/null && \
                echo -e "  ${GREEN}✓${NC} 생성됨 (created): $db" || \
                echo -e "  ${YELLOW}!${NC} 생성 실패 (could not create): $db"
        fi

        # Grant privileges
        sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $db TO $role" &>/dev/null

        # Enable uuid-ossp and grant schema access
        sudo -u postgres psql -d "$db" -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"" &>/dev/null
        sudo -u postgres psql -d "$db" -c "GRANT ALL ON SCHEMA public TO $role" &>/dev/null
    done
}

# =============================================================================
# Main Logic
# =============================================================================

# PostgreSQL Setup
echo -e "${YELLOW}[1/2] PostgreSQL 확인 중... (Checking PostgreSQL)${NC}"

case "$OS_TYPE" in
    macos)
        macos_check_postgres
        status=$?
        if [[ $status -eq 2 ]]; then
            echo ""
            echo -e "${YELLOW}EXTERNAL_DB=true 모드에서 PostgreSQL이 필요합니다.${NC}"
            if ask_confirm "PostgreSQL을 설치하시겠습니까? (Install PostgreSQL?)"; then
                macos_install_postgres
                macos_start_postgres
            else
                echo -e "${RED}PostgreSQL 설치 건너뜀. 종료합니다. (Skipped. Exiting.)${NC}"
                exit 1
            fi
        elif [[ $status -eq 1 ]]; then
            macos_start_postgres
        fi

        # Check PostgreSQL network configuration for Kind access
        echo ""
        if ! macos_check_postgres_binding; then
            echo ""
            echo -e "${YELLOW}Kind pod에서 접근하려면 PostgreSQL이 외부 연결을 허용해야 합니다.${NC}"
            echo -e "${YELLOW}postgresql.conf와 pg_hba.conf 수정이 필요합니다.${NC}"
            echo ""
            if ask_confirm "PostgreSQL을 외부 접근용으로 설정하시겠습니까? (Configure for external access?)"; then
                macos_configure_postgres
            else
                echo -e "${RED}PostgreSQL 설정 건너뜀. (Configuration skipped.)${NC}"
                echo -e "${RED}Kind pod에서 PostgreSQL에 연결하지 못할 수 있습니다!${NC}"
            fi
        fi
        ;;
    debian|redhat)
        linux_check_postgres
        status=$?
        if [[ $status -eq 2 ]]; then
            echo ""
            echo -e "${YELLOW}EXTERNAL_DB=true 모드에서 PostgreSQL이 필요합니다.${NC}"
            if ask_confirm "PostgreSQL을 설치하시겠습니까? (Install PostgreSQL?)"; then
                linux_install_postgres
                linux_start_postgres
                linux_configure_postgres
            else
                echo -e "${RED}PostgreSQL 설치 건너뜀. 종료합니다. (Skipped. Exiting.)${NC}"
                exit 1
            fi
        elif [[ $status -eq 1 ]]; then
            linux_start_postgres
            linux_configure_postgres
        else
            linux_configure_postgres
        fi
        ;;
    *)
        echo -e "${YELLOW}지원하지 않는 OS입니다. PostgreSQL을 수동으로 설치해주세요. (Unsupported OS)${NC}"
        echo "  - PostgreSQL이 0.0.0.0:5432에서 listen하도록 설정"
        echo "  - 172.18.0.0/16 (Kind 네트워크)에서 연결 허용"
        ;;
esac

# Verify PostgreSQL is ready
echo ""
if pg_isready -h localhost &>/dev/null; then
    echo -e "  ${GREEN}✓${NC} PostgreSQL 준비 완료, 연결 대기 중 (ready and accepting connections)"
else
    echo -e "  ${RED}✗${NC} PostgreSQL 준비 안됨 (not ready)"
    exit 1
fi

echo ""

# Redis Setup
echo -e "${YELLOW}[2/2] Redis 확인 중... (Checking Redis)${NC}"

case "$OS_TYPE" in
    macos)
        macos_check_redis
        status=$?
        if [[ $status -eq 2 ]]; then
            echo ""
            echo -e "${YELLOW}EXTERNAL_DB=true 모드에서 Redis가 필요합니다.${NC}"
            if ask_confirm "Redis를 설치하시겠습니까? (Install Redis?)"; then
                macos_install_redis
                macos_start_redis
            else
                echo -e "${RED}Redis 설치 건너뜀. 종료합니다. (Skipped. Exiting.)${NC}"
                exit 1
            fi
        elif [[ $status -eq 1 ]]; then
            macos_start_redis
        fi

        # Check Redis binding (must be 0.0.0.0 for Kind access)
        echo ""
        if ! macos_check_redis_binding; then
            echo ""
            echo -e "${YELLOW}Kind pod에서 접근하려면 Redis가 0.0.0.0에 바인딩되어야 합니다.${NC}"
            echo -e "${YELLOW}redis.conf 수정 및 Redis 재시작이 필요합니다.${NC}"
            echo ""
            if ask_confirm "Redis를 외부 접근용으로 설정하시겠습니까? (Configure for external access?)"; then
                macos_configure_redis
            else
                echo -e "${RED}Redis 설정 건너뜀. (Configuration skipped.)${NC}"
                echo -e "${RED}Kind pod에서 Redis에 연결하지 못합니다!${NC}"
            fi
        fi
        ;;
    debian|redhat)
        linux_check_redis
        status=$?
        if [[ $status -eq 2 ]]; then
            echo ""
            echo -e "${YELLOW}EXTERNAL_DB=true 모드에서 Redis가 필요합니다.${NC}"
            if ask_confirm "Redis를 설치하시겠습니까? (Install Redis?)"; then
                linux_install_redis
                linux_start_redis
                linux_configure_redis
            else
                echo -e "${RED}Redis 설치 건너뜀. 종료합니다. (Skipped. Exiting.)${NC}"
                exit 1
            fi
        elif [[ $status -eq 1 ]]; then
            linux_start_redis
            linux_configure_redis
        else
            linux_configure_redis
        fi
        ;;
    *)
        echo -e "${YELLOW}지원하지 않는 OS입니다. Redis를 수동으로 설치해주세요. (Unsupported OS)${NC}"
        echo "  - Redis가 0.0.0.0:6379에서 listen하도록 설정"
        echo "  - protected-mode 비활성화"
        ;;
esac

# Verify Redis is ready
echo ""
if redis-cli ping &>/dev/null; then
    echo -e "  ${GREEN}✓${NC} Redis 준비 완료 (PONG 응답 받음)"
else
    echo -e "  ${RED}✗${NC} Redis 응답 없음 (not responding)"
    exit 1
fi

echo ""

# =============================================================================
# Create Wealist Databases
# =============================================================================
case "$OS_TYPE" in
    macos)
        macos_create_databases
        ;;
    debian|redhat)
        linux_create_databases
        ;;
esac

echo ""

# =============================================================================
# Summary
# =============================================================================
echo -e "${GREEN}=============================================${NC}"
echo -e "${GREEN}  외부 데이터베이스 설정 완료! (Setup Complete!)${NC}"
echo -e "${GREEN}=============================================${NC}"
echo ""
echo -e "  PostgreSQL: ${GREEN}준비됨 (Ready)${NC} (localhost:5432)"
echo -e "  Redis:      ${GREEN}준비됨 (Ready)${NC} (localhost:6379)"
echo ""

case "$OS_TYPE" in
    macos)
        echo -e "  Kind pod 접근 경로 (access via): ${BLUE}host.docker.internal${NC}"
        echo ""
        echo -e "  ${YELLOW}참고 (Note):${NC} macOS에서는 local-kind.yaml 업데이트 필요:"
        echo -e "    postgresExporter.config.host: host.docker.internal"
        echo -e "    redisExporter.config.host: host.docker.internal"
        ;;
    *)
        echo -e "  Kind pod 접근 경로 (access via): ${BLUE}172.18.0.1${NC}"
        ;;
esac
echo ""
