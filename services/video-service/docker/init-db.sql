-- =============================================================================
-- Video Service Database Initialization
-- =============================================================================
-- 이 스크립트는 docker-compose 시작 시 자동으로 실행됩니다.
-- PostgreSQL 초기화 시 video_service 유저와 권한을 설정합니다.
-- =============================================================================

-- Create video_service user if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'video_service') THEN
        CREATE USER video_service WITH PASSWORD 'video_service_password';
    END IF;
END
$$;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE wealist_video_db TO video_service;
GRANT ALL ON SCHEMA public TO video_service;
