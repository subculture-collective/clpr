-- Drop the tables of features removed in November 2026 (#72). Nothing in the
-- application reads or writes them any more.
--
--   ads and campaigns (#65):            000021, 000022
--   account types (#69):                000040
--   creator verification (#69):         000051, 000085
--   site chat (#63):                    000052, 000070, 000098
--   recommendation engine (#68):        000055 (user_clip_interactions), 000119
--   forum (#64):                        000069, 000080, 000081, 000115
--   watch parties (#62):                000075 to 000079
--   CDN clip mirroring (#67):           000083
--
-- Kept on purpose:
--   user_preferences (000055)           creator-first onboarding preferences
--   subscriptions, Stripe (000012, 000043, 000118)
--                                       retained historical billing records
--   MFA (000045, 000047)                MFA is live; see #66
--   webhook retry/DLQ (000017, 000082)  outbound webhooks keep a retry queue; see #83
--
-- There is no CASCADE: if anything still depends on one of these tables,
-- the migration fails instead of dropping it silently. The down migration
-- recreates the tables empty; restore a backup to get the rows back.

DROP MATERIALIZED VIEW IF EXISTS forum_vote_counts, watch_party_analytics;

DROP TABLE IF EXISTS
    ads, ad_impressions, ad_frequency_caps, ad_frequency_limits,
    ad_experiments, ad_targeting_rules, ad_campaign_analytics, ad_experiment_analytics,
    account_type_conversions,
    creator_verification_applications, creator_verification_decisions, verification_audit_logs,
    chat_channels, chat_messages, chat_bans, chat_moderation_log, channel_members, channel_moderators,
    user_clip_interactions, recommendation_feedback,
    forum_threads, forum_replies, moderation_actions, user_bans, content_flags, forum_votes, user_reputation,
    watch_parties, watch_party_participants, user_presence, watch_party_messages,
    watch_party_reactions, watch_party_events,
    clip_mirrors, mirror_metrics, cdn_configurations, cdn_metrics;
