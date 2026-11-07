-- Recreate the retired-feature tables dropped by 000148, empty, with the
-- schema they had at migration 000147 (generated with pg_dump --schema-only).
-- Their rows are not restored; use a database backup for that.

CREATE TABLE public.account_type_conversions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    old_type character varying(50) NOT NULL,
    new_type character varying(50) NOT NULL,
    reason text,
    converted_by uuid,
    converted_at timestamp without time zone DEFAULT now() NOT NULL,
    metadata jsonb,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);

COMMENT ON TABLE public.account_type_conversions IS 'Audit trail of account type changes for compliance and analytics';

CREATE TABLE public.ad_campaign_analytics (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    ad_id uuid NOT NULL,
    date date NOT NULL,
    slot_id character varying(100),
    impressions integer DEFAULT 0,
    viewable_impressions integer DEFAULT 0,
    clicks integer DEFAULT 0,
    spend_cents bigint DEFAULT 0,
    unique_users integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

COMMENT ON TABLE public.ad_campaign_analytics IS 'Aggregated campaign analytics by date and slot';

CREATE TABLE public.ad_experiment_analytics (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    experiment_id uuid NOT NULL,
    variant character varying(50) NOT NULL,
    date date NOT NULL,
    impressions integer DEFAULT 0,
    viewable_impressions integer DEFAULT 0,
    clicks integer DEFAULT 0,
    conversions integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now()
);

COMMENT ON TABLE public.ad_experiment_analytics IS 'Aggregated experiment analytics by variant and date';

CREATE TABLE public.ad_experiments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    status character varying(50) DEFAULT 'draft'::character varying,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    traffic_percent integer DEFAULT 100,
    winning_variant character varying(50),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT ad_experiments_traffic_percent_check CHECK (((traffic_percent >= 0) AND (traffic_percent <= 100)))
);

COMMENT ON TABLE public.ad_experiments IS 'A/B experiments for comparing ad variants';

CREATE TABLE public.ad_frequency_caps (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    ad_id uuid NOT NULL,
    user_id uuid,
    session_id character varying(255),
    impression_count integer DEFAULT 0,
    window_start timestamp with time zone DEFAULT now(),
    window_type character varying(20) NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

COMMENT ON TABLE public.ad_frequency_caps IS 'Per-user/session impression counts for frequency capping';

CREATE TABLE public.ad_frequency_limits (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    ad_id uuid NOT NULL,
    window_type character varying(20) NOT NULL,
    max_impressions integer NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);

COMMENT ON TABLE public.ad_frequency_limits IS 'Configurable frequency limits per ad';

CREATE TABLE public.ad_impressions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    ad_id uuid NOT NULL,
    user_id uuid,
    session_id character varying(255),
    platform character varying(50) NOT NULL,
    ip_address inet,
    user_agent text,
    page_url text,
    viewability_time_ms integer DEFAULT 0,
    is_viewable boolean DEFAULT false,
    is_clicked boolean DEFAULT false,
    clicked_at timestamp with time zone,
    cost_cents integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    slot_id character varying(100),
    country character varying(2),
    device_type character varying(50),
    experiment_id uuid,
    experiment_variant character varying(50)
);

COMMENT ON TABLE public.ad_impressions IS 'Tracks ad impressions with viewability and click data';

CREATE TABLE public.ad_targeting_rules (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    ad_id uuid NOT NULL,
    rule_type character varying(50) NOT NULL,
    operator character varying(20) DEFAULT 'include'::character varying NOT NULL,
    "values" text[] NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);

COMMENT ON TABLE public.ad_targeting_rules IS 'Structured targeting rules for ads (country, device, interests)';

CREATE TABLE public.ads (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    advertiser_name character varying(255) NOT NULL,
    ad_type character varying(50) NOT NULL,
    content_url text NOT NULL,
    click_url text,
    alt_text character varying(500),
    width integer,
    height integer,
    priority integer DEFAULT 0,
    weight integer DEFAULT 100,
    daily_budget_cents bigint,
    total_budget_cents bigint,
    spent_today_cents bigint DEFAULT 0,
    spent_total_cents bigint DEFAULT 0,
    cpm_cents integer DEFAULT 100,
    is_active boolean DEFAULT true,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    targeting_criteria jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    slot_id character varying(100),
    experiment_id uuid,
    experiment_variant character varying(50)
);

COMMENT ON TABLE public.ads IS 'Stores ad campaigns with targeting, budgets, and scheduling';

COMMENT ON COLUMN public.ads.slot_id IS 'Identifier for the ad placement slot (e.g., header, sidebar, feed)';

COMMENT ON COLUMN public.ads.experiment_id IS 'Reference to A/B experiment if ad is part of an experiment';

COMMENT ON COLUMN public.ads.experiment_variant IS 'Variant name within the experiment (control, variant_a, etc.)';

CREATE TABLE public.cdn_configurations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    provider character varying(50) NOT NULL,
    region character varying(50),
    is_active boolean DEFAULT true NOT NULL,
    priority integer DEFAULT 0 NOT NULL,
    config jsonb NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.cdn_metrics (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    provider character varying(50) NOT NULL,
    region character varying(50),
    metric_type character varying(50) NOT NULL,
    metric_value double precision NOT NULL,
    recorded_at timestamp without time zone DEFAULT now() NOT NULL,
    metadata jsonb
);

CREATE TABLE public.channel_members (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    channel_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role character varying(20) DEFAULT 'member'::character varying NOT NULL,
    joined_at timestamp without time zone DEFAULT now() NOT NULL,
    invited_by uuid,
    CONSTRAINT channel_members_role_check CHECK (((role)::text = ANY ((ARRAY['owner'::character varying, 'admin'::character varying, 'moderator'::character varying, 'member'::character varying])::text[])))
);

COMMENT ON TABLE public.channel_members IS 'Tracks membership and roles in chat channels';

COMMENT ON COLUMN public.channel_members.role IS 'User role in the channel: owner (creator), admin (elevated), moderator, or member';

COMMENT ON COLUMN public.channel_members.invited_by IS 'User who invited this member (NULL for owner/self-join)';

CREATE TABLE public.channel_moderators (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    channel_id uuid NOT NULL,
    moderator_id uuid NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

COMMENT ON TABLE public.channel_moderators IS 'Junction table linking chat channels to their moderators';

COMMENT ON COLUMN public.channel_moderators.channel_id IS 'Reference to the chat channel';

COMMENT ON COLUMN public.channel_moderators.moderator_id IS 'Reference to the user who is a moderator';

COMMENT ON COLUMN public.channel_moderators.is_active IS 'Whether the moderator assignment is currently active';

COMMENT ON COLUMN public.channel_moderators.created_at IS 'Timestamp when the moderator was assigned';

CREATE TABLE public.chat_bans (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    channel_id uuid NOT NULL,
    user_id uuid NOT NULL,
    banned_by uuid NOT NULL,
    reason text,
    expires_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.chat_channels (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    creator_id uuid NOT NULL,
    channel_type character varying(50) DEFAULT 'public'::character varying NOT NULL,
    is_active boolean DEFAULT true,
    max_participants integer,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.chat_messages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    channel_id uuid NOT NULL,
    user_id uuid NOT NULL,
    content text NOT NULL,
    is_deleted boolean DEFAULT false,
    deleted_at timestamp without time zone,
    deleted_by uuid,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.chat_moderation_log (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    channel_id uuid NOT NULL,
    moderator_id uuid NOT NULL,
    target_user_id uuid,
    action character varying(50) NOT NULL,
    reason text,
    metadata jsonb,
    created_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.clip_mirrors (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    clip_id uuid NOT NULL,
    region character varying(50) NOT NULL,
    mirror_url text NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    storage_provider character varying(50) NOT NULL,
    size_bytes bigint,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    last_accessed_at timestamp without time zone,
    access_count integer DEFAULT 0,
    expires_at timestamp without time zone,
    failure_reason text
);

CREATE TABLE public.content_flags (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    target_type character varying(50) NOT NULL,
    target_id uuid NOT NULL,
    reason character varying(100) NOT NULL,
    details text,
    status character varying(20) DEFAULT 'pending'::character varying,
    reviewed_by uuid,
    reviewed_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.creator_verification_applications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    twitch_channel_url character varying(500) NOT NULL,
    follower_count integer,
    subscriber_count integer,
    avg_viewers integer,
    content_description text,
    social_media_links jsonb,
    status character varying(20) DEFAULT 'pending'::character varying,
    priority integer DEFAULT 50,
    reviewed_by uuid,
    reviewed_at timestamp without time zone,
    reviewer_notes text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT verification_app_valid_priority CHECK (((priority >= 0) AND (priority <= 100))),
    CONSTRAINT verification_app_valid_status CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'approved'::character varying, 'rejected'::character varying])::text[])))
);

CREATE TABLE public.creator_verification_decisions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    application_id uuid NOT NULL,
    reviewer_id uuid NOT NULL,
    decision character varying(20) NOT NULL,
    notes text,
    metadata jsonb,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT verification_decision_valid_decision CHECK (((decision)::text = ANY ((ARRAY['approved'::character varying, 'rejected'::character varying])::text[])))
);

CREATE TABLE public.forum_replies (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    thread_id uuid NOT NULL,
    user_id uuid NOT NULL,
    parent_reply_id uuid,
    content text NOT NULL,
    is_deleted boolean DEFAULT false,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    depth integer DEFAULT 0,
    path public.ltree,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english'::regconfig, content)) STORED,
    flagged_as_spam boolean DEFAULT false,
    hidden boolean DEFAULT false,
    CONSTRAINT forum_replies_depth_check CHECK ((depth <= 9))
);

CREATE TABLE public.forum_threads (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    title character varying(255) NOT NULL,
    content text NOT NULL,
    locked boolean DEFAULT false,
    pinned boolean DEFAULT false,
    flag_count integer DEFAULT 0,
    reply_count integer DEFAULT 0,
    view_count integer DEFAULT 0,
    is_deleted boolean DEFAULT false,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    game_id uuid,
    tags character varying[] DEFAULT '{}'::character varying[],
    locked_at timestamp without time zone,
    locked_by uuid,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english'::regconfig, (((title)::text || ' '::text) || content))) STORED
);

CREATE TABLE public.forum_votes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    reply_id uuid NOT NULL,
    vote_value smallint NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT forum_votes_vote_value_check CHECK ((vote_value = ANY (ARRAY['-1'::integer, 0, 1])))
);

CREATE MATERIALIZED VIEW public.forum_vote_counts AS
 SELECT reply_id,
    (sum(
        CASE
            WHEN (vote_value = 1) THEN 1
            ELSE 0
        END))::integer AS upvote_count,
    (sum(
        CASE
            WHEN (vote_value = '-1'::integer) THEN 1
            ELSE 0
        END))::integer AS downvote_count,
    (sum(vote_value))::integer AS net_votes
   FROM public.forum_votes
  WHERE (vote_value <> 0)
  GROUP BY reply_id
  WITH NO DATA;

CREATE TABLE public.mirror_metrics (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    clip_id uuid NOT NULL,
    region character varying(50) NOT NULL,
    metric_type character varying(50) NOT NULL,
    metric_value double precision NOT NULL,
    recorded_at timestamp without time zone DEFAULT now() NOT NULL,
    metadata jsonb
);

CREATE TABLE public.moderation_actions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    moderator_id uuid NOT NULL,
    action_type character varying(50) NOT NULL,
    target_type character varying(50) NOT NULL,
    target_id uuid NOT NULL,
    reason text,
    metadata jsonb,
    created_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.recommendation_feedback (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    clip_id uuid NOT NULL,
    feedback_type character varying(10) NOT NULL,
    algorithm character varying(20),
    score double precision,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT recommendation_feedback_algorithm_check CHECK (((algorithm)::text = ANY ((ARRAY['content'::character varying, 'collaborative'::character varying, 'hybrid'::character varying, 'trending'::character varying])::text[]))),
    CONSTRAINT recommendation_feedback_feedback_type_check CHECK (((feedback_type)::text = ANY ((ARRAY['positive'::character varying, 'negative'::character varying])::text[]))),
    CONSTRAINT recommendation_feedback_score_check CHECK (((score >= (0)::double precision) AND (score <= (1)::double precision)))
);

CREATE TABLE public.user_bans (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    banned_by uuid NOT NULL,
    reason text NOT NULL,
    expires_at timestamp without time zone,
    active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

CREATE TABLE public.user_clip_interactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    clip_id uuid NOT NULL,
    interaction_type character varying(20) NOT NULL,
    dwell_time integer,
    "timestamp" timestamp without time zone DEFAULT now()
);

CREATE TABLE public.user_presence (
    user_id uuid NOT NULL,
    channel_id uuid NOT NULL,
    status character varying(20) DEFAULT 'online'::character varying NOT NULL,
    last_seen timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.user_reputation (
    user_id uuid NOT NULL,
    total_votes integer DEFAULT 0,
    threads_created integer DEFAULT 0,
    replies_created integer DEFAULT 0,
    reputation_score integer DEFAULT 0,
    reputation_badge character varying(50) DEFAULT 'new'::character varying,
    last_updated timestamp without time zone DEFAULT now()
);

CREATE TABLE public.verification_audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    audit_type character varying(50) NOT NULL,
    status character varying(20) NOT NULL,
    findings jsonb,
    notes text,
    audited_by uuid,
    action_taken character varying(50),
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT audit_log_valid_action CHECK ((((action_taken)::text = ANY ((ARRAY['none'::character varying, 'warning_sent'::character varying, 'verification_revoked'::character varying, 'further_review_required'::character varying])::text[])) OR (action_taken IS NULL))),
    CONSTRAINT audit_log_valid_status CHECK (((status)::text = ANY ((ARRAY['passed'::character varying, 'flagged'::character varying, 'revoked'::character varying])::text[]))),
    CONSTRAINT audit_log_valid_type CHECK (((audit_type)::text = ANY ((ARRAY['periodic_check'::character varying, 'manual_review'::character varying, 'abuse_detection'::character varying])::text[])))
);

CREATE TABLE public.watch_parties (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    host_user_id uuid NOT NULL,
    title character varying(200) NOT NULL,
    playlist_id uuid,
    current_clip_id uuid,
    current_position_seconds integer DEFAULT 0,
    is_playing boolean DEFAULT false,
    visibility character varying(20) DEFAULT 'private'::character varying,
    invite_code character varying(10) NOT NULL,
    max_participants integer DEFAULT 100,
    created_at timestamp without time zone DEFAULT now(),
    started_at timestamp without time zone,
    ended_at timestamp without time zone,
    password character varying(255),
    CONSTRAINT ended_at_after_started_or_created CHECK (((ended_at IS NULL) OR (ended_at >= COALESCE(started_at, created_at)))),
    CONSTRAINT started_at_after_created_at CHECK (((started_at IS NULL) OR (started_at >= created_at))),
    CONSTRAINT watch_parties_current_position_seconds_check CHECK ((current_position_seconds >= 0)),
    CONSTRAINT watch_parties_max_participants_check CHECK (((max_participants > 0) AND (max_participants <= 1000))),
    CONSTRAINT watch_parties_visibility_check CHECK (((visibility)::text = ANY ((ARRAY['private'::character varying, 'public'::character varying, 'friends'::character varying, 'invite'::character varying])::text[])))
);

COMMENT ON TABLE public.watch_parties IS 'Watch parties for synchronized video viewing with friends';

COMMENT ON COLUMN public.watch_parties.current_position_seconds IS 'Current playback position in seconds';

COMMENT ON COLUMN public.watch_parties.invite_code IS 'Unique code for joining the party (6-10 characters)';

COMMENT ON COLUMN public.watch_parties.password IS 'Optional bcrypt-hashed password for invite-only parties';

CREATE TABLE public.watch_party_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    "time" timestamp without time zone DEFAULT now() NOT NULL,
    party_id uuid NOT NULL,
    user_id uuid,
    event_type character varying(50) NOT NULL,
    metadata jsonb,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT watch_party_events_event_type_check CHECK (((event_type)::text = ANY ((ARRAY['join'::character varying, 'leave'::character varying, 'chat'::character varying, 'reaction'::character varying, 'sync'::character varying])::text[])))
);

COMMENT ON TABLE public.watch_party_events IS 'Event log for watch party analytics tracking';

COMMENT ON COLUMN public.watch_party_events.event_type IS 'Type of event: join, leave, chat, reaction, sync';

COMMENT ON COLUMN public.watch_party_events.metadata IS 'Additional event data in JSON format';

CREATE MATERIALIZED VIEW public.watch_party_analytics AS
 WITH user_durations AS (
         SELECT watch_party_events.party_id,
            watch_party_events.user_id,
            min(watch_party_events."time") FILTER (WHERE ((watch_party_events.event_type)::text = 'join'::text)) AS first_join,
            max(watch_party_events."time") FILTER (WHERE ((watch_party_events.event_type)::text = 'leave'::text)) AS last_leave
           FROM public.watch_party_events
          WHERE ((watch_party_events.event_type)::text = ANY ((ARRAY['join'::character varying, 'leave'::character varying])::text[]))
          GROUP BY watch_party_events.party_id, watch_party_events.user_id
        ), party_stats AS (
         SELECT wp.id AS party_id,
            wp.host_user_id,
            count(DISTINCT wpe.user_id) FILTER (WHERE (((wpe.event_type)::text = 'join'::text) AND (wpe.user_id IS NOT NULL))) AS unique_viewers,
            count(*) FILTER (WHERE ((wpe.event_type)::text = 'chat'::text)) AS chat_messages,
            count(*) FILTER (WHERE ((wpe.event_type)::text = 'reaction'::text)) AS reactions
           FROM (public.watch_parties wp
             LEFT JOIN public.watch_party_events wpe ON ((wpe.party_id = wp.id)))
          GROUP BY wp.id, wp.host_user_id
        ), avg_durations AS (
         SELECT user_durations.party_id,
            COALESCE(round(avg(EXTRACT(epoch FROM (user_durations.last_leave - user_durations.first_join)))), (0)::numeric) AS avg_watch_duration_seconds
           FROM user_durations
          WHERE ((user_durations.first_join IS NOT NULL) AND (user_durations.last_leave IS NOT NULL))
          GROUP BY user_durations.party_id
        ), peak_concurrent AS (
         SELECT concurrent.party_id,
            max(concurrent.concurrent_count) AS peak_concurrent_viewers
           FROM ( SELECT watch_party_events.party_id,
                    watch_party_events."time",
                    sum(
                        CASE
                            WHEN ((watch_party_events.event_type)::text = 'join'::text) THEN 1
                            WHEN ((watch_party_events.event_type)::text = 'leave'::text) THEN '-1'::integer
                            ELSE 0
                        END) OVER (PARTITION BY watch_party_events.party_id ORDER BY watch_party_events."time") AS concurrent_count
                   FROM public.watch_party_events
                  WHERE ((watch_party_events.event_type)::text = ANY ((ARRAY['join'::character varying, 'leave'::character varying])::text[]))) concurrent
          GROUP BY concurrent.party_id
        )
 SELECT ps.party_id,
    ps.host_user_id,
    ps.unique_viewers,
    ps.chat_messages,
    ps.reactions,
    COALESCE(ad.avg_watch_duration_seconds, (0)::numeric) AS avg_watch_duration_seconds,
    COALESCE(pc.peak_concurrent_viewers, (0)::bigint) AS peak_concurrent_viewers
   FROM ((party_stats ps
     LEFT JOIN avg_durations ad ON ((ps.party_id = ad.party_id)))
     LEFT JOIN peak_concurrent pc ON ((ps.party_id = pc.party_id)))
  WITH NO DATA;

COMMENT ON MATERIALIZED VIEW public.watch_party_analytics IS 'Aggregated analytics for watch parties';

CREATE TABLE public.watch_party_messages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    watch_party_id uuid NOT NULL,
    user_id uuid NOT NULL,
    message text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT watch_party_messages_message_check CHECK (((length(message) > 0) AND (length(message) <= 1000)))
);

COMMENT ON TABLE public.watch_party_messages IS 'Chat messages in watch parties';

COMMENT ON COLUMN public.watch_party_messages.message IS 'Chat message text (max 1000 characters)';

CREATE TABLE public.watch_party_participants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    party_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role character varying(20) DEFAULT 'viewer'::character varying,
    joined_at timestamp without time zone DEFAULT now(),
    left_at timestamp without time zone,
    last_sync_at timestamp without time zone,
    sync_offset_ms integer DEFAULT 0,
    CONSTRAINT valid_left_at CHECK (((left_at IS NULL) OR (left_at >= joined_at))),
    CONSTRAINT watch_party_participants_role_check CHECK (((role)::text = ANY ((ARRAY['host'::character varying, 'co-host'::character varying, 'viewer'::character varying])::text[])))
);

COMMENT ON TABLE public.watch_party_participants IS 'Junction table tracking participants in watch parties';

COMMENT ON COLUMN public.watch_party_participants.role IS 'Participant role: host (creator), co-host (elevated), or viewer';

COMMENT ON COLUMN public.watch_party_participants.sync_offset_ms IS 'Participant sync offset in milliseconds for monitoring';

CREATE TABLE public.watch_party_reactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    watch_party_id uuid NOT NULL,
    user_id uuid NOT NULL,
    emoji character varying(10) NOT NULL,
    video_timestamp numeric(10,2),
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT watch_party_reactions_emoji_check CHECK ((length((emoji)::text) > 0))
);

COMMENT ON TABLE public.watch_party_reactions IS 'Emoji reactions in watch parties with optional video timestamps';

COMMENT ON COLUMN public.watch_party_reactions.emoji IS 'Emoji character or unicode (max 10 characters)';

COMMENT ON COLUMN public.watch_party_reactions.video_timestamp IS 'Video playback position when reaction was sent (seconds with 2 decimal places)';

ALTER TABLE ONLY public.account_type_conversions
    ADD CONSTRAINT account_type_conversions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ad_campaign_analytics
    ADD CONSTRAINT ad_campaign_analytics_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ad_experiment_analytics
    ADD CONSTRAINT ad_experiment_analytics_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ad_experiments
    ADD CONSTRAINT ad_experiments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ad_frequency_caps
    ADD CONSTRAINT ad_frequency_caps_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ad_frequency_limits
    ADD CONSTRAINT ad_frequency_limits_ad_id_window_type_key UNIQUE (ad_id, window_type);

ALTER TABLE ONLY public.ad_frequency_limits
    ADD CONSTRAINT ad_frequency_limits_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ad_impressions
    ADD CONSTRAINT ad_impressions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ad_targeting_rules
    ADD CONSTRAINT ad_targeting_rules_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ads
    ADD CONSTRAINT ads_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.cdn_configurations
    ADD CONSTRAINT cdn_configurations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.cdn_configurations
    ADD CONSTRAINT cdn_configurations_provider_region_key UNIQUE (provider, region);

ALTER TABLE ONLY public.cdn_metrics
    ADD CONSTRAINT cdn_metrics_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.channel_members
    ADD CONSTRAINT channel_members_channel_id_user_id_key UNIQUE (channel_id, user_id);

ALTER TABLE ONLY public.channel_members
    ADD CONSTRAINT channel_members_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.channel_moderators
    ADD CONSTRAINT channel_moderators_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.chat_bans
    ADD CONSTRAINT chat_bans_channel_id_user_id_key UNIQUE (channel_id, user_id);

ALTER TABLE ONLY public.chat_bans
    ADD CONSTRAINT chat_bans_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.chat_channels
    ADD CONSTRAINT chat_channels_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.chat_messages
    ADD CONSTRAINT chat_messages_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.chat_moderation_log
    ADD CONSTRAINT chat_moderation_log_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.clip_mirrors
    ADD CONSTRAINT clip_mirrors_clip_id_region_key UNIQUE (clip_id, region);

ALTER TABLE ONLY public.clip_mirrors
    ADD CONSTRAINT clip_mirrors_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.content_flags
    ADD CONSTRAINT content_flags_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.creator_verification_applications
    ADD CONSTRAINT creator_verification_applications_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.creator_verification_decisions
    ADD CONSTRAINT creator_verification_decisions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.forum_replies
    ADD CONSTRAINT forum_replies_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.forum_threads
    ADD CONSTRAINT forum_threads_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.forum_votes
    ADD CONSTRAINT forum_votes_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.forum_votes
    ADD CONSTRAINT forum_votes_user_id_reply_id_key UNIQUE (user_id, reply_id);

ALTER TABLE ONLY public.mirror_metrics
    ADD CONSTRAINT mirror_metrics_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.moderation_actions
    ADD CONSTRAINT moderation_actions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.recommendation_feedback
    ADD CONSTRAINT recommendation_feedback_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.ad_experiment_analytics
    ADD CONSTRAINT unique_experiment_variant_date UNIQUE (experiment_id, variant, date);

ALTER TABLE ONLY public.user_bans
    ADD CONSTRAINT user_bans_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_clip_interactions
    ADD CONSTRAINT user_clip_interactions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_clip_interactions
    ADD CONSTRAINT user_clip_interactions_user_id_clip_id_interaction_type_key UNIQUE (user_id, clip_id, interaction_type);

ALTER TABLE ONLY public.user_presence
    ADD CONSTRAINT user_presence_pkey PRIMARY KEY (user_id, channel_id);

ALTER TABLE ONLY public.user_reputation
    ADD CONSTRAINT user_reputation_pkey PRIMARY KEY (user_id);

ALTER TABLE ONLY public.verification_audit_logs
    ADD CONSTRAINT verification_audit_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.watch_parties
    ADD CONSTRAINT watch_parties_invite_code_key UNIQUE (invite_code);

ALTER TABLE ONLY public.watch_parties
    ADD CONSTRAINT watch_parties_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.watch_party_events
    ADD CONSTRAINT watch_party_events_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.watch_party_messages
    ADD CONSTRAINT watch_party_messages_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.watch_party_participants
    ADD CONSTRAINT watch_party_participants_party_id_user_id_key UNIQUE (party_id, user_id);

ALTER TABLE ONLY public.watch_party_participants
    ADD CONSTRAINT watch_party_participants_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.watch_party_reactions
    ADD CONSTRAINT watch_party_reactions_pkey PRIMARY KEY (id);

CREATE INDEX idx_account_type_conversions_converted_at ON public.account_type_conversions USING btree (converted_at DESC);

CREATE INDEX idx_account_type_conversions_new_type ON public.account_type_conversions USING btree (new_type);

CREATE INDEX idx_account_type_conversions_user_id ON public.account_type_conversions USING btree (user_id);

CREATE INDEX idx_ad_campaign_analytics_ad_date ON public.ad_campaign_analytics USING btree (ad_id, date);

CREATE INDEX idx_ad_campaign_analytics_slot ON public.ad_campaign_analytics USING btree (slot_id) WHERE (slot_id IS NOT NULL);

CREATE INDEX idx_ad_experiment_analytics_experiment ON public.ad_experiment_analytics USING btree (experiment_id, date);

CREATE INDEX idx_ad_frequency_caps_session_ad ON public.ad_frequency_caps USING btree (session_id, ad_id) WHERE (session_id IS NOT NULL);

CREATE INDEX idx_ad_frequency_caps_user_ad ON public.ad_frequency_caps USING btree (user_id, ad_id) WHERE (user_id IS NOT NULL);

CREATE INDEX idx_ad_frequency_caps_window ON public.ad_frequency_caps USING btree (ad_id, window_type, window_start);

CREATE INDEX idx_ad_impressions_ad_id ON public.ad_impressions USING btree (ad_id);

CREATE INDEX idx_ad_impressions_created_at ON public.ad_impressions USING btree (created_at);

CREATE INDEX idx_ad_impressions_date ON public.ad_impressions USING btree (created_at);

CREATE INDEX idx_ad_impressions_experiment ON public.ad_impressions USING btree (experiment_id, experiment_variant) WHERE (experiment_id IS NOT NULL);

CREATE INDEX idx_ad_impressions_session_id ON public.ad_impressions USING btree (session_id) WHERE (session_id IS NOT NULL);

CREATE INDEX idx_ad_impressions_slot_id ON public.ad_impressions USING btree (slot_id) WHERE (slot_id IS NOT NULL);

CREATE INDEX idx_ad_impressions_user_id ON public.ad_impressions USING btree (user_id) WHERE (user_id IS NOT NULL);

CREATE INDEX idx_ad_impressions_viewable ON public.ad_impressions USING btree (ad_id, is_viewable) WHERE (is_viewable = true);

CREATE INDEX idx_ad_targeting_rules_ad_id ON public.ad_targeting_rules USING btree (ad_id);

CREATE INDEX idx_ad_targeting_rules_type ON public.ad_targeting_rules USING btree (rule_type);

CREATE INDEX idx_ads_active_dates ON public.ads USING btree (is_active, start_date, end_date) WHERE (is_active = true);

CREATE INDEX idx_ads_experiment_id ON public.ads USING btree (experiment_id) WHERE (experiment_id IS NOT NULL);

CREATE INDEX idx_ads_priority_weight ON public.ads USING btree (priority DESC, weight DESC) WHERE (is_active = true);

CREATE INDEX idx_ads_slot_id ON public.ads USING btree (slot_id) WHERE (slot_id IS NOT NULL);

CREATE INDEX idx_cdn_config_active ON public.cdn_configurations USING btree (is_active) WHERE (is_active = true);

CREATE INDEX idx_cdn_config_priority ON public.cdn_configurations USING btree (priority DESC);

CREATE INDEX idx_cdn_config_provider ON public.cdn_configurations USING btree (provider);

CREATE INDEX idx_cdn_metrics_provider ON public.cdn_metrics USING btree (provider);

CREATE INDEX idx_cdn_metrics_recorded_at ON public.cdn_metrics USING btree (recorded_at DESC);

CREATE INDEX idx_cdn_metrics_type ON public.cdn_metrics USING btree (metric_type);

CREATE INDEX idx_channel_members_channel ON public.channel_members USING btree (channel_id, role);

CREATE INDEX idx_channel_members_joined ON public.channel_members USING btree (joined_at DESC);

CREATE INDEX idx_channel_members_user ON public.channel_members USING btree (user_id);

CREATE INDEX idx_channel_moderators_channel ON public.channel_moderators USING btree (channel_id) WHERE (is_active = true);

CREATE INDEX idx_channel_moderators_moderator ON public.channel_moderators USING btree (moderator_id) WHERE (is_active = true);

CREATE UNIQUE INDEX idx_channel_moderators_unique ON public.channel_moderators USING btree (channel_id, moderator_id) WHERE (is_active = true);

CREATE INDEX idx_chat_bans_active ON public.chat_bans USING btree (channel_id, user_id, expires_at);

CREATE INDEX idx_chat_bans_channel_user ON public.chat_bans USING btree (channel_id, user_id);

CREATE INDEX idx_chat_bans_expires ON public.chat_bans USING btree (expires_at) WHERE (expires_at IS NOT NULL);

CREATE INDEX idx_chat_channels_active ON public.chat_channels USING btree (is_active) WHERE (is_active = true);

CREATE INDEX idx_chat_channels_creator ON public.chat_channels USING btree (creator_id);

CREATE INDEX idx_chat_messages_channel ON public.chat_messages USING btree (channel_id, created_at DESC);

CREATE INDEX idx_chat_messages_not_deleted ON public.chat_messages USING btree (channel_id) WHERE (is_deleted = false);

CREATE INDEX idx_chat_messages_user ON public.chat_messages USING btree (user_id);

CREATE INDEX idx_chat_moderation_log_channel ON public.chat_moderation_log USING btree (channel_id, created_at DESC);

CREATE INDEX idx_chat_moderation_log_moderator ON public.chat_moderation_log USING btree (moderator_id);

CREATE INDEX idx_chat_moderation_log_target ON public.chat_moderation_log USING btree (target_user_id);

CREATE INDEX idx_clip_mirrors_clip_id ON public.clip_mirrors USING btree (clip_id);

CREATE INDEX idx_clip_mirrors_expires_at ON public.clip_mirrors USING btree (expires_at) WHERE (expires_at IS NOT NULL);

CREATE INDEX idx_clip_mirrors_last_accessed ON public.clip_mirrors USING btree (last_accessed_at DESC);

CREATE INDEX idx_clip_mirrors_region ON public.clip_mirrors USING btree (region);

CREATE INDEX idx_clip_mirrors_status ON public.clip_mirrors USING btree (status);

CREATE INDEX idx_content_flags_created ON public.content_flags USING btree (created_at DESC);

CREATE INDEX idx_content_flags_status ON public.content_flags USING btree (status);

CREATE INDEX idx_content_flags_target ON public.content_flags USING btree (target_type, target_id);

CREATE INDEX idx_content_flags_user ON public.content_flags USING btree (user_id);

CREATE INDEX idx_forum_replies_depth ON public.forum_replies USING btree (depth, thread_id) WHERE (is_deleted = false);

CREATE INDEX idx_forum_replies_parent ON public.forum_replies USING btree (parent_reply_id);

CREATE INDEX idx_forum_replies_path ON public.forum_replies USING gist (path) WHERE (is_deleted = false);

CREATE INDEX idx_forum_replies_search ON public.forum_replies USING gin (search_vector) WHERE (is_deleted = false);

CREATE INDEX idx_forum_replies_spam ON public.forum_replies USING btree (flagged_as_spam, hidden) WHERE (is_deleted = false);

CREATE INDEX idx_forum_replies_thread ON public.forum_replies USING btree (thread_id, created_at);

CREATE INDEX idx_forum_replies_thread_parent ON public.forum_replies USING btree (thread_id, parent_reply_id, created_at) WHERE (is_deleted = false);

CREATE INDEX idx_forum_replies_user ON public.forum_replies USING btree (user_id);

CREATE INDEX idx_forum_threads_created ON public.forum_threads USING btree (created_at DESC);

CREATE INDEX idx_forum_threads_deleted ON public.forum_threads USING btree (is_deleted) WHERE (is_deleted = false);

CREATE INDEX idx_forum_threads_game ON public.forum_threads USING btree (game_id, created_at DESC) WHERE (is_deleted = false);

CREATE INDEX idx_forum_threads_locked ON public.forum_threads USING btree (locked, updated_at) WHERE (is_deleted = false);

CREATE INDEX idx_forum_threads_pinned ON public.forum_threads USING btree (pinned DESC, created_at DESC);

CREATE INDEX idx_forum_threads_search ON public.forum_threads USING gin (search_vector) WHERE (is_deleted = false);

CREATE INDEX idx_forum_threads_tags ON public.forum_threads USING gin (tags) WHERE (is_deleted = false);

CREATE INDEX idx_forum_threads_user ON public.forum_threads USING btree (user_id);

CREATE UNIQUE INDEX idx_forum_vote_counts_reply ON public.forum_vote_counts USING btree (reply_id);

CREATE INDEX idx_forum_votes_created ON public.forum_votes USING btree (created_at DESC);

CREATE INDEX idx_forum_votes_reply ON public.forum_votes USING btree (reply_id) WHERE (vote_value <> 0);

CREATE INDEX idx_forum_votes_user ON public.forum_votes USING btree (user_id);

CREATE INDEX idx_mirror_metrics_clip_id ON public.mirror_metrics USING btree (clip_id);

CREATE INDEX idx_mirror_metrics_recorded_at ON public.mirror_metrics USING btree (recorded_at DESC);

CREATE INDEX idx_mirror_metrics_region ON public.mirror_metrics USING btree (region);

CREATE INDEX idx_mirror_metrics_type ON public.mirror_metrics USING btree (metric_type);

CREATE INDEX idx_moderation_actions_created ON public.moderation_actions USING btree (created_at DESC);

CREATE INDEX idx_moderation_actions_moderator ON public.moderation_actions USING btree (moderator_id);

CREATE INDEX idx_moderation_actions_target ON public.moderation_actions USING btree (target_type, target_id);

CREATE INDEX idx_moderation_actions_type ON public.moderation_actions USING btree (action_type);

CREATE INDEX idx_participants_last_sync ON public.watch_party_participants USING btree (party_id, last_sync_at) WHERE (left_at IS NULL);

CREATE INDEX idx_participants_party ON public.watch_party_participants USING btree (party_id) WHERE (left_at IS NULL);

CREATE INDEX idx_participants_party_active ON public.watch_party_participants USING btree (party_id) WHERE (left_at IS NULL);

CREATE INDEX idx_participants_user ON public.watch_party_participants USING btree (user_id) WHERE (left_at IS NULL);

CREATE INDEX idx_parties_active ON public.watch_parties USING btree (created_at DESC) WHERE (ended_at IS NULL);

CREATE INDEX idx_parties_host ON public.watch_parties USING btree (host_user_id) WHERE (ended_at IS NULL);

CREATE INDEX idx_parties_invite ON public.watch_parties USING btree (invite_code) WHERE (ended_at IS NULL);

CREATE INDEX idx_parties_playlist ON public.watch_parties USING btree (playlist_id) WHERE ((playlist_id IS NOT NULL) AND (ended_at IS NULL));

CREATE INDEX idx_parties_visibility ON public.watch_parties USING btree (visibility) WHERE (ended_at IS NULL);

CREATE INDEX idx_party_events_party_time ON public.watch_party_events USING btree (party_id, "time" DESC);

CREATE INDEX idx_party_events_type ON public.watch_party_events USING btree (party_id, event_type);

CREATE INDEX idx_party_events_user ON public.watch_party_events USING btree (user_id) WHERE (user_id IS NOT NULL);

CREATE INDEX idx_party_messages ON public.watch_party_messages USING btree (watch_party_id, created_at DESC);

CREATE INDEX idx_party_messages_user ON public.watch_party_messages USING btree (user_id, created_at DESC);

CREATE INDEX idx_party_reactions ON public.watch_party_reactions USING btree (watch_party_id, created_at DESC);

CREATE INDEX idx_party_reactions_timestamp ON public.watch_party_reactions USING btree (watch_party_id, video_timestamp) WHERE (video_timestamp IS NOT NULL);

CREATE INDEX idx_recommendation_feedback_clip ON public.recommendation_feedback USING btree (clip_id, created_at DESC);

CREATE INDEX idx_recommendation_feedback_user_created ON public.recommendation_feedback USING btree (user_id, created_at DESC);

CREATE INDEX idx_user_bans_active ON public.user_bans USING btree (active) WHERE (active = true);

CREATE INDEX idx_user_bans_expires ON public.user_bans USING btree (expires_at) WHERE (expires_at IS NOT NULL);

CREATE INDEX idx_user_bans_user ON public.user_bans USING btree (user_id);

CREATE INDEX idx_user_interactions_clip ON public.user_clip_interactions USING btree (clip_id, interaction_type);

CREATE INDEX idx_user_interactions_type ON public.user_clip_interactions USING btree (interaction_type, "timestamp" DESC);

CREATE INDEX idx_user_interactions_user ON public.user_clip_interactions USING btree (user_id, "timestamp" DESC);

CREATE INDEX idx_user_presence_channel ON public.user_presence USING btree (channel_id, status);

CREATE INDEX idx_user_presence_last_seen ON public.user_presence USING btree (last_seen);

CREATE INDEX idx_user_presence_user ON public.user_presence USING btree (user_id);

CREATE INDEX idx_user_reputation_badge ON public.user_reputation USING btree (reputation_badge);

CREATE INDEX idx_user_reputation_score ON public.user_reputation USING btree (reputation_score DESC);

CREATE INDEX idx_verification_app_created_at ON public.creator_verification_applications USING btree (created_at DESC);

CREATE INDEX idx_verification_app_reviewed_by ON public.creator_verification_applications USING btree (reviewed_by);

CREATE INDEX idx_verification_app_status_priority ON public.creator_verification_applications USING btree (status, priority DESC, created_at);

CREATE INDEX idx_verification_app_user_id ON public.creator_verification_applications USING btree (user_id);

CREATE INDEX idx_verification_audit_audited_by ON public.verification_audit_logs USING btree (audited_by);

CREATE INDEX idx_verification_audit_created_at ON public.verification_audit_logs USING btree (created_at DESC);

CREATE INDEX idx_verification_audit_flagged ON public.verification_audit_logs USING btree (status, created_at DESC) WHERE ((status)::text = ANY ((ARRAY['flagged'::character varying, 'revoked'::character varying])::text[]));

CREATE INDEX idx_verification_audit_status ON public.verification_audit_logs USING btree (status);

CREATE INDEX idx_verification_audit_type ON public.verification_audit_logs USING btree (audit_type);

CREATE INDEX idx_verification_audit_user_id ON public.verification_audit_logs USING btree (user_id);

CREATE INDEX idx_verification_decision_application ON public.creator_verification_decisions USING btree (application_id);

CREATE INDEX idx_verification_decision_created_at ON public.creator_verification_decisions USING btree (created_at DESC);

CREATE INDEX idx_verification_decision_reviewer ON public.creator_verification_decisions USING btree (reviewer_id);

CREATE INDEX idx_watch_party_analytics_host ON public.watch_party_analytics USING btree (host_user_id);

CREATE UNIQUE INDEX idx_watch_party_analytics_party ON public.watch_party_analytics USING btree (party_id);

CREATE UNIQUE INDEX unique_ad_date_slot ON public.ad_campaign_analytics USING btree (ad_id, date, COALESCE(slot_id, ''::character varying));

CREATE UNIQUE INDEX unique_session_ad_window ON public.ad_frequency_caps USING btree (ad_id, session_id, window_type) WHERE (session_id IS NOT NULL);

CREATE UNIQUE INDEX unique_user_ad_window ON public.ad_frequency_caps USING btree (ad_id, user_id, window_type) WHERE (user_id IS NOT NULL);

CREATE UNIQUE INDEX uq_verification_app_user_pending ON public.creator_verification_applications USING btree (user_id) WHERE ((status)::text = 'pending'::text);

CREATE TRIGGER trg_forum_votes_update_counts AFTER INSERT OR DELETE OR UPDATE ON public.forum_votes FOR EACH ROW EXECUTE FUNCTION public.trigger_update_vote_counts();

CREATE TRIGGER trg_update_thread_flag_count AFTER INSERT OR DELETE OR UPDATE OF status ON public.content_flags FOR EACH ROW EXECUTE FUNCTION public.update_thread_flag_count();

CREATE TRIGGER trg_update_thread_reply_count AFTER INSERT OR DELETE OR UPDATE OF is_deleted ON public.forum_replies FOR EACH ROW EXECUTE FUNCTION public.update_thread_reply_count();

CREATE TRIGGER trg_update_thread_timestamp AFTER INSERT OR UPDATE ON public.forum_replies FOR EACH ROW EXECUTE FUNCTION public.update_thread_updated_at();

CREATE TRIGGER trg_update_user_verification AFTER UPDATE ON public.creator_verification_applications FOR EACH ROW EXECUTE FUNCTION public.update_user_verification_status();

CREATE TRIGGER trg_verification_application_reviewed BEFORE UPDATE ON public.creator_verification_applications FOR EACH ROW EXECUTE FUNCTION public.update_verification_application_reviewed();

CREATE TRIGGER update_ad_campaign_analytics_updated_at BEFORE UPDATE ON public.ad_campaign_analytics FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_ad_experiments_updated_at BEFORE UPDATE ON public.ad_experiments FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_ad_frequency_caps_updated_at BEFORE UPDATE ON public.ad_frequency_caps FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_ads_updated_at BEFORE UPDATE ON public.ads FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER update_chat_channels_updated_at BEFORE UPDATE ON public.chat_channels FOR EACH ROW EXECUTE FUNCTION public.update_chat_updated_at_column();

CREATE TRIGGER update_chat_messages_updated_at BEFORE UPDATE ON public.chat_messages FOR EACH ROW EXECUTE FUNCTION public.update_chat_updated_at_column();

CREATE TRIGGER update_user_presence_updated_at BEFORE UPDATE ON public.user_presence FOR EACH ROW EXECUTE FUNCTION public.update_chat_updated_at_column();

ALTER TABLE ONLY public.account_type_conversions
    ADD CONSTRAINT account_type_conversions_converted_by_fkey FOREIGN KEY (converted_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.account_type_conversions
    ADD CONSTRAINT account_type_conversions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.ad_campaign_analytics
    ADD CONSTRAINT ad_campaign_analytics_ad_id_fkey FOREIGN KEY (ad_id) REFERENCES public.ads(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.ad_experiment_analytics
    ADD CONSTRAINT ad_experiment_analytics_experiment_id_fkey FOREIGN KEY (experiment_id) REFERENCES public.ad_experiments(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.ad_frequency_caps
    ADD CONSTRAINT ad_frequency_caps_ad_id_fkey FOREIGN KEY (ad_id) REFERENCES public.ads(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.ad_frequency_caps
    ADD CONSTRAINT ad_frequency_caps_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.ad_frequency_limits
    ADD CONSTRAINT ad_frequency_limits_ad_id_fkey FOREIGN KEY (ad_id) REFERENCES public.ads(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.ad_impressions
    ADD CONSTRAINT ad_impressions_ad_id_fkey FOREIGN KEY (ad_id) REFERENCES public.ads(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.ad_impressions
    ADD CONSTRAINT ad_impressions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.ad_targeting_rules
    ADD CONSTRAINT ad_targeting_rules_ad_id_fkey FOREIGN KEY (ad_id) REFERENCES public.ads(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.channel_members
    ADD CONSTRAINT channel_members_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES public.chat_channels(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.channel_members
    ADD CONSTRAINT channel_members_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.channel_members
    ADD CONSTRAINT channel_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.channel_moderators
    ADD CONSTRAINT channel_moderators_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES public.chat_channels(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.channel_moderators
    ADD CONSTRAINT channel_moderators_moderator_id_fkey FOREIGN KEY (moderator_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.chat_bans
    ADD CONSTRAINT chat_bans_banned_by_fkey FOREIGN KEY (banned_by) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.chat_bans
    ADD CONSTRAINT chat_bans_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES public.chat_channels(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.chat_bans
    ADD CONSTRAINT chat_bans_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.chat_channels
    ADD CONSTRAINT chat_channels_creator_id_fkey FOREIGN KEY (creator_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.chat_messages
    ADD CONSTRAINT chat_messages_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES public.chat_channels(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.chat_messages
    ADD CONSTRAINT chat_messages_deleted_by_fkey FOREIGN KEY (deleted_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.chat_messages
    ADD CONSTRAINT chat_messages_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.chat_moderation_log
    ADD CONSTRAINT chat_moderation_log_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES public.chat_channels(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.chat_moderation_log
    ADD CONSTRAINT chat_moderation_log_moderator_id_fkey FOREIGN KEY (moderator_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.chat_moderation_log
    ADD CONSTRAINT chat_moderation_log_target_user_id_fkey FOREIGN KEY (target_user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.clip_mirrors
    ADD CONSTRAINT clip_mirrors_clip_id_fkey FOREIGN KEY (clip_id) REFERENCES public.clips(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.content_flags
    ADD CONSTRAINT content_flags_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.content_flags
    ADD CONSTRAINT content_flags_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.creator_verification_applications
    ADD CONSTRAINT creator_verification_applications_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.creator_verification_applications
    ADD CONSTRAINT creator_verification_applications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.creator_verification_decisions
    ADD CONSTRAINT creator_verification_decisions_application_id_fkey FOREIGN KEY (application_id) REFERENCES public.creator_verification_applications(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.creator_verification_decisions
    ADD CONSTRAINT creator_verification_decisions_reviewer_id_fkey FOREIGN KEY (reviewer_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.ads
    ADD CONSTRAINT fk_ads_experiment FOREIGN KEY (experiment_id) REFERENCES public.ad_experiments(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.forum_replies
    ADD CONSTRAINT forum_replies_parent_reply_id_fkey FOREIGN KEY (parent_reply_id) REFERENCES public.forum_replies(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.forum_replies
    ADD CONSTRAINT forum_replies_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES public.forum_threads(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.forum_replies
    ADD CONSTRAINT forum_replies_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.forum_threads
    ADD CONSTRAINT forum_threads_game_id_fkey FOREIGN KEY (game_id) REFERENCES public.games(id);

ALTER TABLE ONLY public.forum_threads
    ADD CONSTRAINT forum_threads_locked_by_fkey FOREIGN KEY (locked_by) REFERENCES public.users(id);

ALTER TABLE ONLY public.forum_threads
    ADD CONSTRAINT forum_threads_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.forum_votes
    ADD CONSTRAINT forum_votes_reply_id_fkey FOREIGN KEY (reply_id) REFERENCES public.forum_replies(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.forum_votes
    ADD CONSTRAINT forum_votes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.mirror_metrics
    ADD CONSTRAINT mirror_metrics_clip_id_fkey FOREIGN KEY (clip_id) REFERENCES public.clips(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.moderation_actions
    ADD CONSTRAINT moderation_actions_moderator_id_fkey FOREIGN KEY (moderator_id) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.recommendation_feedback
    ADD CONSTRAINT recommendation_feedback_clip_id_fkey FOREIGN KEY (clip_id) REFERENCES public.clips(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.recommendation_feedback
    ADD CONSTRAINT recommendation_feedback_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_bans
    ADD CONSTRAINT user_bans_banned_by_fkey FOREIGN KEY (banned_by) REFERENCES public.users(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.user_bans
    ADD CONSTRAINT user_bans_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_clip_interactions
    ADD CONSTRAINT user_clip_interactions_clip_id_fkey FOREIGN KEY (clip_id) REFERENCES public.clips(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_clip_interactions
    ADD CONSTRAINT user_clip_interactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_presence
    ADD CONSTRAINT user_presence_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES public.chat_channels(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_presence
    ADD CONSTRAINT user_presence_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.user_reputation
    ADD CONSTRAINT user_reputation_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.verification_audit_logs
    ADD CONSTRAINT verification_audit_logs_audited_by_fkey FOREIGN KEY (audited_by) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.verification_audit_logs
    ADD CONSTRAINT verification_audit_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.watch_parties
    ADD CONSTRAINT watch_parties_current_clip_id_fkey FOREIGN KEY (current_clip_id) REFERENCES public.clips(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.watch_parties
    ADD CONSTRAINT watch_parties_host_user_id_fkey FOREIGN KEY (host_user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.watch_parties
    ADD CONSTRAINT watch_parties_playlist_id_fkey FOREIGN KEY (playlist_id) REFERENCES public.playlists(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.watch_party_events
    ADD CONSTRAINT watch_party_events_party_id_fkey FOREIGN KEY (party_id) REFERENCES public.watch_parties(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.watch_party_events
    ADD CONSTRAINT watch_party_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.watch_party_messages
    ADD CONSTRAINT watch_party_messages_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.watch_party_messages
    ADD CONSTRAINT watch_party_messages_watch_party_id_fkey FOREIGN KEY (watch_party_id) REFERENCES public.watch_parties(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.watch_party_participants
    ADD CONSTRAINT watch_party_participants_party_id_fkey FOREIGN KEY (party_id) REFERENCES public.watch_parties(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.watch_party_participants
    ADD CONSTRAINT watch_party_participants_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.watch_party_reactions
    ADD CONSTRAINT watch_party_reactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.watch_party_reactions
    ADD CONSTRAINT watch_party_reactions_watch_party_id_fkey FOREIGN KEY (watch_party_id) REFERENCES public.watch_parties(id) ON DELETE CASCADE;
