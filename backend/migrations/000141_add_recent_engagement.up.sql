-- Lifetime counters are not historical observations. Existing clips get no
-- fabricated baseline: their first subsequent observation begins tracking.
ALTER TABLE clips ADD COLUMN twitch_view_count_raw INTEGER;

CREATE TABLE engagement_tracking (
    singleton BOOLEAN PRIMARY KEY DEFAULT true CHECK (singleton),
    started_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
INSERT INTO engagement_tracking DEFAULT VALUES;

CREATE TABLE clip_engagement_state (
    clip_id UUID PRIMARY KEY REFERENCES clips(id) ON DELETE CASCADE,
    tracking_started_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    observed_at TIMESTAMPTZ,
    first_observed_at TIMESTAMPTZ,
    raw_view_count INTEGER CHECK (raw_view_count >= 0),
    next_poll_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    last_poll_error_at TIMESTAMPTZ,
    active_until TIMESTAMPTZ,
    view_gain DOUBLE PRECISION NOT NULL DEFAULT 0,
    vote_gain BIGINT NOT NULL DEFAULT 0,
    comment_gain BIGINT NOT NULL DEFAULT 0,
    favorite_gain BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX clip_engagement_poll_due ON clip_engagement_state(next_poll_at,clip_id);

CREATE TABLE clip_view_observations (
    clip_id UUID NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
    observed_at TIMESTAMPTZ NOT NULL,
    raw_view_count INTEGER NOT NULL CHECK (raw_view_count >= 0),
    interval_start TIMESTAMPTZ,
    view_gain DOUBLE PRECISION NOT NULL DEFAULT 0,
    PRIMARY KEY (clip_id,observed_at)
);
CREATE INDEX clip_view_observations_expiry ON clip_view_observations(observed_at);

CREATE TABLE clip_local_engagement (
    clip_id UUID NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
    occurred_at TIMESTAMPTZ NOT NULL,
    score_delta BIGINT NOT NULL
);
CREATE INDEX clip_local_engagement_window ON clip_local_engagement(occurred_at,clip_id);

CREATE TABLE clip_engagement_hourly (
    clip_id UUID NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
    bucket_start TIMESTAMPTZ NOT NULL,
    view_gain DOUBLE PRECISION NOT NULL DEFAULT 0,
    vote_gain BIGINT NOT NULL DEFAULT 0,
    comment_gain BIGINT NOT NULL DEFAULT 0,
    favorite_gain BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (bucket_start,clip_id)
);

CREATE TABLE engagement_generations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    published_at TIMESTAMPTZ NOT NULL,
    tracking_started_at TIMESTAMPTZ NOT NULL,
    coverage_started_at TIMESTAMPTZ NOT NULL,
    oldest_observation_at TIMESTAMPTZ,
    stale_clips BIGINT NOT NULL,
    eligible_clips BIGINT NOT NULL
);
CREATE INDEX engagement_generations_latest ON engagement_generations(published_at DESC);
CREATE TABLE engagement_rankings (
    generation_id UUID NOT NULL REFERENCES engagement_generations(id) ON DELETE CASCADE,
    period TEXT NOT NULL CHECK (period IN ('hour','day','week','month','year','all')),
    clip_id UUID NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
    score DOUBLE PRECISION NOT NULL CHECK (score > 0),
    PRIMARY KEY (generation_id,period,clip_id)
);
CREATE INDEX engagement_rankings_order ON engagement_rankings(generation_id,period,score DESC,clip_id DESC);

CREATE FUNCTION record_twitch_observation(p_clip UUID,p_count INTEGER,p_at TIMESTAMPTZ) RETURNS BOOLEAN
LANGUAGE plpgsql AS $$
DECLARE previous clip_engagement_state%ROWTYPE; gain DOUBLE PRECISION; bucket TIMESTAMPTZ; share DOUBLE PRECISION;
BEGIN
    IF p_count < 0 OR p_at > clock_timestamp() + INTERVAL '1 minute' THEN
        RAISE EXCEPTION 'invalid Twitch observation';
    END IF;
    -- Match source-mutation lock order: clip, then engagement state.
    PERFORM id FROM clips WHERE id=p_clip FOR NO KEY UPDATE;
    INSERT INTO clip_engagement_state(clip_id) VALUES(p_clip) ON CONFLICT DO NOTHING;
    SELECT * INTO previous FROM clip_engagement_state WHERE clip_id=p_clip FOR UPDATE;
    IF previous.observed_at IS NOT NULL AND p_at <= previous.observed_at THEN RETURN false; END IF;
    gain := CASE WHEN previous.observed_at IS NULL THEN 0 ELSE GREATEST(0,p_count-previous.raw_view_count) END;
    INSERT INTO clip_view_observations(clip_id,observed_at,raw_view_count,interval_start,view_gain) VALUES(p_clip,p_at,p_count,previous.observed_at,gain);
    -- Split observed growth across its actual observation interval. Period
    -- boundaries use the same uniform estimate; never assign lifetime totals.
    IF gain > 0 THEN
        FOR bucket IN SELECT generate_series(date_trunc('hour',GREATEST(previous.observed_at,p_at-INTERVAL '400 days'),'UTC'),date_trunc('hour',p_at,'UTC'),INTERVAL '1 hour') LOOP
            share := gain * GREATEST(0,EXTRACT(EPOCH FROM (LEAST(p_at,bucket+INTERVAL '1 hour')-GREATEST(previous.observed_at,bucket)))) / EXTRACT(EPOCH FROM (p_at-previous.observed_at));
            IF share > 0 THEN
                INSERT INTO clip_engagement_hourly(clip_id,bucket_start,view_gain) VALUES(p_clip,bucket,share)
                ON CONFLICT (bucket_start,clip_id) DO UPDATE SET view_gain=clip_engagement_hourly.view_gain+EXCLUDED.view_gain;
            END IF;
        END LOOP;
    END IF;
    UPDATE clip_engagement_state SET observed_at=p_at,first_observed_at=COALESCE(first_observed_at,p_at),raw_view_count=p_count,view_gain=view_gain+gain,
        active_until=CASE WHEN gain>0 THEN p_at+INTERVAL '1 day' ELSE active_until END,
        next_poll_at=p_at+CASE WHEN gain>0 OR active_until>p_at THEN INTERVAL '15 minutes' ELSE INTERVAL '1 day' END,
        last_poll_error_at=NULL WHERE clip_id=p_clip;
    RETURN true;
END $$;

CREATE FUNCTION capture_twitch_observation() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.twitch_clip_id IS NOT NULL THEN
        PERFORM record_twitch_observation(NEW.id,COALESCE(NEW.twitch_view_count_raw,NEW.view_count,0),clock_timestamp());
    END IF;
    RETURN NEW;
END $$;
CREATE TRIGGER capture_twitch_observation_insert AFTER INSERT ON clips FOR EACH ROW EXECUTE FUNCTION capture_twitch_observation();
CREATE TRIGGER capture_twitch_observation_update AFTER UPDATE OF twitch_view_count_raw ON clips FOR EACH ROW EXECUTE FUNCTION capture_twitch_observation();

CREATE FUNCTION capture_local_engagement() RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE event_at TIMESTAMPTZ:=clock_timestamp(); cid UUID; votes_delta BIGINT:=0; comments_delta BIGINT:=0; favorites_delta BIGINT:=0;
BEGIN
    IF TG_OP='DELETE' THEN cid:=OLD.clip_id; ELSE cid:=NEW.clip_id; END IF;
    IF TG_TABLE_NAME='votes' THEN
        IF TG_OP<>'DELETE' THEN votes_delta:=NEW.vote_type; END IF;
        IF TG_OP<>'INSERT' THEN votes_delta:=votes_delta-OLD.vote_type; END IF;
    ELSIF TG_TABLE_NAME='favorites' THEN
        IF TG_OP='INSERT' THEN favorites_delta:=1; ELSIF TG_OP='DELETE' THEN favorites_delta:=-1; END IF;
    ELSE
        IF TG_OP<>'DELETE' AND NOT COALESCE(NEW.is_removed,false) THEN comments_delta:=1; END IF;
        IF TG_OP<>'INSERT' AND NOT COALESCE(OLD.is_removed,false) THEN comments_delta:=comments_delta-1; END IF;
    END IF;
    IF (votes_delta=0 AND comments_delta=0 AND favorites_delta=0) OR NOT EXISTS(SELECT 1 FROM clips WHERE id=cid) THEN RETURN NULL; END IF;
    PERFORM id FROM clips WHERE id=cid FOR NO KEY UPDATE;
    IF NOT FOUND THEN RETURN NULL; END IF;
    INSERT INTO clip_engagement_state(clip_id,vote_gain,comment_gain,favorite_gain,active_until)
    VALUES(cid,votes_delta,comments_delta,favorites_delta,event_at+INTERVAL '1 day')
    ON CONFLICT(clip_id) DO UPDATE SET vote_gain=clip_engagement_state.vote_gain+EXCLUDED.vote_gain,
        comment_gain=clip_engagement_state.comment_gain+EXCLUDED.comment_gain,favorite_gain=clip_engagement_state.favorite_gain+EXCLUDED.favorite_gain,
        active_until=event_at+INTERVAL '1 day',
        next_poll_at=LEAST(clip_engagement_state.next_poll_at,clock_timestamp());
    INSERT INTO clip_engagement_hourly(clip_id,bucket_start,vote_gain,comment_gain,favorite_gain)
    VALUES(cid,date_trunc('hour',event_at,'UTC'),votes_delta,comments_delta,favorites_delta)
    ON CONFLICT(bucket_start,clip_id) DO UPDATE SET vote_gain=clip_engagement_hourly.vote_gain+EXCLUDED.vote_gain,
        comment_gain=clip_engagement_hourly.comment_gain+EXCLUDED.comment_gain,favorite_gain=clip_engagement_hourly.favorite_gain+EXCLUDED.favorite_gain;
    INSERT INTO clip_local_engagement VALUES(cid,event_at,2*votes_delta+3*comments_delta+2*favorites_delta);
    RETURN NULL;
END $$;
CREATE TRIGGER capture_vote_engagement AFTER INSERT OR UPDATE OR DELETE ON votes FOR EACH ROW EXECUTE FUNCTION capture_local_engagement();
CREATE TRIGGER capture_favorite_engagement AFTER INSERT OR DELETE ON favorites FOR EACH ROW EXECUTE FUNCTION capture_local_engagement();
CREATE TRIGGER capture_comment_engagement AFTER INSERT OR UPDATE OR DELETE ON comments FOR EACH ROW EXECUTE FUNCTION capture_local_engagement();
