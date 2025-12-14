#!/bin/bash
# =============================================================================
# Initialize Databases Only (PostgreSQL already installed)
# =============================================================================
# Run with: sudo ./init-databases-only.sh
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🗄️  Creating wealist databases and users..."
sudo -u postgres psql -f "$SCRIPT_DIR/01-create-databases.sql"

echo ""
echo "✅ Done! Databases created:"
echo "   - user_db, board_db, chat_db, noti_db, storage_db, video_db"
