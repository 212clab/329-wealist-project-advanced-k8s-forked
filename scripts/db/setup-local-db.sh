#!/bin/bash
# =============================================================================
# wealist Local Database Setup Script
# =============================================================================
# This script sets up PostgreSQL and Redis on local Ubuntu for persistent data
# Run with: sudo ./setup-local-db.sh
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN} wealist Local Database Setup${NC}"
echo -e "${GREEN}=========================================${NC}"

# -----------------------------------------------------------------------------
# Check if running as root
# -----------------------------------------------------------------------------
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root (sudo)${NC}"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# -----------------------------------------------------------------------------
# Step 1: Install PostgreSQL 17
# -----------------------------------------------------------------------------
echo -e "\n${YELLOW}[1/6] Installing PostgreSQL 17...${NC}"

if ! command -v psql &> /dev/null; then
    # Add PostgreSQL repository
    sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
    wget -qO- https://www.postgresql.org/media/keys/ACCC4CF8.asc | tee /etc/apt/trusted.gpg.d/pgdg.asc &>/dev/null
    apt-get update -y
    apt-get install postgresql-17 postgresql-client-17 -y
    echo -e "${GREEN}PostgreSQL 17 installed!${NC}"
else
    echo -e "${GREEN}PostgreSQL already installed: $(psql --version)${NC}"
fi

systemctl enable postgresql
systemctl start postgresql

# -----------------------------------------------------------------------------
# Step 2: Install Redis 7
# -----------------------------------------------------------------------------
echo -e "\n${YELLOW}[2/6] Installing Redis 7...${NC}"

if ! command -v redis-server &> /dev/null; then
    # Add Redis repository
    curl -fsSL https://packages.redis.io/gpg | gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg
    echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/redis.list
    apt-get update -y
    apt-get install redis -y
    echo -e "${GREEN}Redis 7 installed!${NC}"
else
    echo -e "${GREEN}Redis already installed: $(redis-server --version)${NC}"
fi

systemctl enable redis-server
systemctl start redis-server

# -----------------------------------------------------------------------------
# Step 3: Configure PostgreSQL for remote access
# -----------------------------------------------------------------------------
echo -e "\n${YELLOW}[3/6] Configuring PostgreSQL for remote access...${NC}"

PG_VERSION="17"
PG_CONF="/etc/postgresql/${PG_VERSION}/main/postgresql.conf"
PG_HBA="/etc/postgresql/${PG_VERSION}/main/pg_hba.conf"

# Allow listening on all interfaces
if grep -q "^listen_addresses" "$PG_CONF"; then
    sed -i "s/^listen_addresses.*/listen_addresses = '*'/" "$PG_CONF"
else
    echo "listen_addresses = '*'" >> "$PG_CONF"
fi

# Allow password authentication from local network and K8s pods
if ! grep -q "0.0.0.0/0" "$PG_HBA"; then
    echo "# Allow connections from any IP (for K8s pods)" >> "$PG_HBA"
    echo "host    all    all    0.0.0.0/0    scram-sha-256" >> "$PG_HBA"
fi

echo -e "${GREEN}PostgreSQL configured for remote access${NC}"

# -----------------------------------------------------------------------------
# Step 4: Configure Redis for remote access
# -----------------------------------------------------------------------------
echo -e "\n${YELLOW}[4/6] Configuring Redis for remote access...${NC}"

REDIS_CONF="/etc/redis/redis.conf"

# Allow binding to all interfaces
sed -i 's/^bind 127.0.0.1.*/bind 0.0.0.0/' "$REDIS_CONF"

# Disable protected mode (for local dev only!)
sed -i 's/^protected-mode yes/protected-mode no/' "$REDIS_CONF"

echo -e "${GREEN}Redis configured for remote access${NC}"

# -----------------------------------------------------------------------------
# Step 5: Create databases and users
# -----------------------------------------------------------------------------
echo -e "\n${YELLOW}[5/6] Creating databases and users...${NC}"

sudo -u postgres psql -f "$SCRIPT_DIR/01-create-databases.sql"

# -----------------------------------------------------------------------------
# Step 6: Restart services
# -----------------------------------------------------------------------------
echo -e "\n${YELLOW}[6/6] Restarting services...${NC}"

systemctl restart postgresql
systemctl restart redis-server

# Wait for services to be ready
sleep 3

# -----------------------------------------------------------------------------
# Verify
# -----------------------------------------------------------------------------
echo -e "\n${GREEN}=========================================${NC}"
echo -e "${GREEN} Setup Complete!${NC}"
echo -e "${GREEN}=========================================${NC}"

echo -e "\n${YELLOW}PostgreSQL Status:${NC}"
systemctl status postgresql --no-pager -l | head -5

echo -e "\n${YELLOW}Redis Status:${NC}"
systemctl status redis-server --no-pager -l | head -5

# Get local IP
LOCAL_IP=$(hostname -I | awk '{print $1}')

echo -e "\n${GREEN}=========================================${NC}"
echo -e "${GREEN} Connection Info${NC}"
echo -e "${GREEN}=========================================${NC}"
echo -e "PostgreSQL: ${LOCAL_IP}:5432"
echo -e "Redis:      ${LOCAL_IP}:6379"
echo -e ""
echo -e "${YELLOW}Databases created:${NC}"
echo -e "  - wealist_user_service_db     (user: user_service)"
echo -e "  - wealist_board_service_db    (user: board_service)"
echo -e "  - wealist_chat_service_db     (user: chat_service)"
echo -e "  - wealist_noti_service_db     (user: noti_service)"
echo -e "  - wealist_storage_service_db  (user: storage_service)"
echo -e "  - wealist_video_service_db    (user: video_service)"
echo -e ""
echo -e "${YELLOW}Test connections:${NC}"
echo -e "  psql -h ${LOCAL_IP} -U user_service -d wealist_user_service_db"
echo -e "  redis-cli -h ${LOCAL_IP} ping"
echo -e ""
echo -e "${YELLOW}Update Helm values (helm/environments/local-ubuntu.yaml):${NC}"
echo -e "  shared:"
echo -e "    config:"
echo -e "      POSTGRES_HOST: \"${LOCAL_IP}\""
echo -e "      DB_HOST: \"${LOCAL_IP}\""
echo -e "      REDIS_HOST: \"${LOCAL_IP}\""
