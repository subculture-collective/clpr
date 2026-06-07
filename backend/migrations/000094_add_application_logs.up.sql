-- Application logs table for centralized log collection from frontend and mobile
CREATE TABLE IF NOT EXISTS application_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level VARCHAR(10) NOT NULL CHECK (level IN ('debug', 'info', 'warn', 'error')),
    message TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    service VARCHAR(50) NOT NULL, -- 'clpr-frontend', 'clpr-mobile', etc.
    platform VARCHAR(20), -- 'web', 'ios', 'android'
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(255),
    trace_id VARCHAR(255),
    url TEXT,
    user_agent TEXT,
    device_id VARCHAR(255),
    app_version VARCHAR(50),
    error TEXT,
    stack TEXT,
    context JSONB, -- Additional context fields
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying and cleanup
CREATE INDEX idx_application_logs_level ON application_logs(level);
CREATE INDEX idx_application_logs_timestamp ON application_logs(timestamp DESC);
CREATE INDEX idx_application_logs_user_id ON application_logs(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_application_logs_service ON application_logs(service);
CREATE INDEX idx_application_logs_created_at ON application_logs(created_at DESC);

-- Composite index for common queries (level + timestamp)
CREATE INDEX idx_application_logs_level_timestamp ON application_logs(level, timestamp DESC);

-- Add comment to table
COMMENT ON TABLE application_logs IS 'Centralized application logs from frontend and mobile clients';
COMMENT ON COLUMN application_logs.level IS 'Log level: debug, info, warn, or error';
COMMENT ON COLUMN application_logs.service IS 'Source service identifier (e.g., clpr-frontend, clpr-mobile)';
COMMENT ON COLUMN application_logs.platform IS 'Platform identifier: web, ios, or android';
COMMENT ON COLUMN application_logs.context IS 'Additional structured context data (JSONB)';
