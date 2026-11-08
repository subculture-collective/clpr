DROP INDEX IF EXISTS idx_webhook_deliveries_once;
ALTER TABLE webhook_subscriptions DROP COLUMN IF EXISTS target_type;
