-- =============================================================================
-- wealist Database & User Creation Script
-- =============================================================================
-- Run as postgres superuser:
--   sudo -u postgres psql -f 01-create-databases.sql
-- Or on macOS:
--   psql -U postgres -f 01-create-databases.sql
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Create Users (Roles)
-- -----------------------------------------------------------------------------
DO $$
BEGIN
    -- user-service
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'user_service') THEN
        CREATE ROLE user_service WITH LOGIN PASSWORD 'user_service_password';
    END IF;

    -- board-service
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'board_service') THEN
        CREATE ROLE board_service WITH LOGIN PASSWORD 'board_service_password';
    END IF;

    -- chat-service
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'chat_service') THEN
        CREATE ROLE chat_service WITH LOGIN PASSWORD 'chat_service_password';
    END IF;

    -- noti-service
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'noti_service') THEN
        CREATE ROLE noti_service WITH LOGIN PASSWORD 'noti_service_password';
    END IF;

    -- storage-service
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'storage_service') THEN
        CREATE ROLE storage_service WITH LOGIN PASSWORD 'storage_service_password';
    END IF;

    -- video-service
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'video_service') THEN
        CREATE ROLE video_service WITH LOGIN PASSWORD 'video_service_password';
    END IF;
END
$$;

-- -----------------------------------------------------------------------------
-- Create Databases (matching Helm values DB_NAME)
-- -----------------------------------------------------------------------------
-- Note: CREATE DATABASE cannot be inside a transaction block, so we use \gexec

SELECT 'CREATE DATABASE wealist_user_service_db OWNER user_service'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'wealist_user_service_db')\gexec

SELECT 'CREATE DATABASE wealist_board_service_db OWNER board_service'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'wealist_board_service_db')\gexec

SELECT 'CREATE DATABASE wealist_chat_service_db OWNER chat_service'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'wealist_chat_service_db')\gexec

SELECT 'CREATE DATABASE wealist_noti_service_db OWNER noti_service'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'wealist_noti_service_db')\gexec

SELECT 'CREATE DATABASE wealist_storage_service_db OWNER storage_service'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'wealist_storage_service_db')\gexec

SELECT 'CREATE DATABASE wealist_video_service_db OWNER video_service'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'wealist_video_service_db')\gexec

-- -----------------------------------------------------------------------------
-- Grant Privileges
-- -----------------------------------------------------------------------------
GRANT ALL PRIVILEGES ON DATABASE wealist_user_service_db TO user_service;
GRANT ALL PRIVILEGES ON DATABASE wealist_board_service_db TO board_service;
GRANT ALL PRIVILEGES ON DATABASE wealist_chat_service_db TO chat_service;
GRANT ALL PRIVILEGES ON DATABASE wealist_noti_service_db TO noti_service;
GRANT ALL PRIVILEGES ON DATABASE wealist_storage_service_db TO storage_service;
GRANT ALL PRIVILEGES ON DATABASE wealist_video_service_db TO video_service;

-- -----------------------------------------------------------------------------
-- Enable UUID Extension (required for all databases)
-- -----------------------------------------------------------------------------
\c wealist_user_service_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
GRANT ALL ON SCHEMA public TO user_service;

\c wealist_board_service_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
GRANT ALL ON SCHEMA public TO board_service;

\c wealist_chat_service_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
GRANT ALL ON SCHEMA public TO chat_service;

\c wealist_noti_service_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
GRANT ALL ON SCHEMA public TO noti_service;

\c wealist_storage_service_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
GRANT ALL ON SCHEMA public TO storage_service;

\c wealist_video_service_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
GRANT ALL ON SCHEMA public TO video_service;

\echo '✅ All databases and users created successfully!'
