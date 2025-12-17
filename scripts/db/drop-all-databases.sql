-- =============================================================================
-- Drop All wealist Databases (CAUTION: Destroys all data!)
-- =============================================================================
-- Run as postgres superuser:
--   sudo -u postgres psql -f drop-all-databases.sql
-- =============================================================================

-- Terminate all connections first
SELECT pg_terminate_backend(pid) FROM pg_stat_activity
WHERE datname IN ('user_db', 'board_db', 'chat_db', 'noti_db', 'storage_db', 'video_db')
AND pid <> pg_backend_pid();

-- Drop databases
DROP DATABASE IF EXISTS user_db;
DROP DATABASE IF EXISTS board_db;
DROP DATABASE IF EXISTS chat_db;
DROP DATABASE IF EXISTS noti_db;
DROP DATABASE IF EXISTS storage_db;
DROP DATABASE IF EXISTS video_db;

-- Drop users
DROP ROLE IF EXISTS user_service;
DROP ROLE IF EXISTS board_service;
DROP ROLE IF EXISTS chat_service;
DROP ROLE IF EXISTS noti_service;
DROP ROLE IF EXISTS storage_service;
DROP ROLE IF EXISTS video_service;

\echo '🗑️  All databases and users dropped!'
