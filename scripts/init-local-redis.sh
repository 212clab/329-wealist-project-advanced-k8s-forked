#!/bin/bash
# =============================================================================
# Configure Local Redis for Wealist Services
# =============================================================================
# This script configures Redis to accept connections from Docker network
# (Kind cluster) on the local Ubuntu installation.
#
# Prerequisites:
#   1. Redis installed: sudo apt install redis-server
#   2. Redis service running: sudo systemctl start redis
#   3. Run this script as root or with sudo
#
# Usage:
#   sudo ./scripts/init-local-redis.sh
#
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

REDIS_CONF="/etc/redis/redis.conf"

# Check if running as root or can use sudo
check_permissions() {
    if [ "$EUID" -ne 0 ]; then
        log_error "Please run this script with sudo or as root"
        exit 1
    fi
}

# Check if Redis is installed
check_redis_installed() {
    if ! command -v redis-server &> /dev/null; then
        log_error "Redis is not installed. Install with: sudo apt install redis-server"
        exit 1
    fi
    log_info "Redis is installed"
}

# Backup original config
backup_config() {
    if [ -f "$REDIS_CONF" ] && [ ! -f "${REDIS_CONF}.backup" ]; then
        log_info "Backing up original redis.conf..."
        cp "$REDIS_CONF" "${REDIS_CONF}.backup"
    fi
}

# Configure Redis to accept connections from all interfaces
configure_bind() {
    log_info "Configuring Redis bind address..."

    # Check current bind configuration
    CURRENT_BIND=$(grep "^bind " "$REDIS_CONF" 2>/dev/null || echo "")

    if [[ "$CURRENT_BIND" == *"0.0.0.0"* ]]; then
        log_info "Redis already configured to bind to all interfaces"
    else
        log_info "Setting bind to 0.0.0.0 (all interfaces)"
        # Comment out existing bind and add new
        sed -i 's/^bind /#bind /' "$REDIS_CONF"
        echo "bind 0.0.0.0" >> "$REDIS_CONF"
    fi
}

# Disable protected mode (since we're binding to all interfaces)
configure_protected_mode() {
    log_info "Configuring protected-mode..."

    CURRENT_PROTECTED=$(grep "^protected-mode" "$REDIS_CONF" 2>/dev/null || echo "")

    if [[ "$CURRENT_PROTECTED" == *"no"* ]]; then
        log_info "protected-mode already disabled"
    else
        log_info "Disabling protected-mode for Docker network access"
        sed -i 's/^protected-mode yes/protected-mode no/' "$REDIS_CONF"
        # If not found, add it
        if ! grep -q "^protected-mode" "$REDIS_CONF"; then
            echo "protected-mode no" >> "$REDIS_CONF"
        fi
    fi
}

# Restart Redis to apply changes
restart_redis() {
    log_info "Restarting Redis..."
    systemctl restart redis-server || systemctl restart redis

    # Wait for Redis to be ready
    sleep 2

    if systemctl is-active --quiet redis-server || systemctl is-active --quiet redis; then
        log_info "Redis restarted successfully"
    else
        log_error "Failed to restart Redis"
        exit 1
    fi
}

# Verify connection
verify_connection() {
    log_info "Verifying Redis is accessible..."

    # Check if Redis is listening
    if netstat -tlnp 2>/dev/null | grep -q ":6379.*LISTEN" || ss -tlnp | grep -q ":6379.*LISTEN"; then
        log_info "Redis is listening on port 6379"
    else
        log_warn "Redis may not be listening on all interfaces"
    fi

    # Test ping
    if redis-cli ping | grep -q "PONG"; then
        log_info "Redis responding to PING"
    else
        log_warn "Redis may not be responding"
    fi
}

# Print summary
print_summary() {
    echo ""
    echo "============================================="
    echo "Redis Configuration Complete!"
    echo "============================================="
    echo ""
    echo "Connection Info (for Kind cluster):"
    echo "  Host: 172.17.0.1 (Docker bridge gateway)"
    echo "  Port: 6379"
    echo "  Password: (none)"
    echo ""
    echo "Configuration Changes:"
    echo "  - bind 0.0.0.0"
    echo "  - protected-mode no"
    echo ""
    echo "Security Note:"
    echo "  This configuration is for local development only!"
    echo "  Do not use in production without proper security."
    echo ""
}

# Main execution
main() {
    log_info "Starting Redis configuration for Wealist..."
    echo ""

    check_permissions
    check_redis_installed
    backup_config
    configure_bind
    configure_protected_mode
    restart_redis
    verify_connection
    print_summary
}

main "$@"
