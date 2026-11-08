-- Discord delivery for outbound webhooks (#83). A subscription's target type
-- decides the request body: the signed JSON event for "generic", a Discord
-- message with embeds for "discord".
ALTER TABLE webhook_subscriptions
    ADD COLUMN target_type VARCHAR(20) NOT NULL DEFAULT 'generic'
        CHECK (target_type IN ('generic', 'discord'));

-- The daily and countdown events are queued by schedulers and workers that
-- may run more than once for the same day or edition; one delivery per
-- subscription and event keeps them from posting twice.
CREATE UNIQUE INDEX idx_webhook_deliveries_once
    ON webhook_deliveries (subscription_id, event_type, event_id)
    WHERE event_type IN ('countdown.published', 'daily.top_clips');
