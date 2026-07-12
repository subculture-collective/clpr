CREATE TABLE IF NOT EXISTS stripe_webhook_receipts (
    event_id VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('processing', 'completed')),
    locked_until TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_stripe_webhook_receipts_status_lease
    ON stripe_webhook_receipts (status, locked_until);

ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS last_stripe_event_created BIGINT NOT NULL DEFAULT 0;
