DROP TABLE IF EXISTS stripe_webhook_receipts;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS last_stripe_event_created;
