-- YouTube publishing of countdown renders (#79). An edition moves from
-- "rendered" to "publishing" while the worker uploads it, then to
-- "published". Each uploaded render is recorded once, so a retried edition
-- does not upload a file twice.
ALTER TABLE countdown_editions DROP CONSTRAINT countdown_editions_status_check;
ALTER TABLE countdown_editions ADD CONSTRAINT countdown_editions_status_check
    CHECK (status IN ('selected', 'fetching', 'fetched', 'rendering', 'rendered', 'publishing', 'published', 'failed'));

CREATE TABLE countdown_uploads (
    render_id UUID PRIMARY KEY REFERENCES countdown_renders(id) ON DELETE CASCADE,
    edition_id UUID NOT NULL REFERENCES countdown_editions(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('countdown', 'short')),
    clip_id UUID REFERENCES clips(id) ON DELETE SET NULL,
    youtube_video_id TEXT NOT NULL CHECK (youtube_video_id ~ '^[A-Za-z0-9_-]{6,20}$'),
    privacy_status TEXT NOT NULL CHECK (privacy_status IN ('private', 'unlisted', 'public')),
    publish_at TIMESTAMPTZ,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX countdown_uploads_edition_idx ON countdown_uploads (edition_id);

-- One YouTube playlist per series, created on its first upload.
CREATE TABLE countdown_youtube_playlists (
    series_id UUID PRIMARY KEY REFERENCES countdown_series(id) ON DELETE CASCADE,
    playlist_id TEXT NOT NULL CHECK (playlist_id ~ '^[A-Za-z0-9_-]{10,64}$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
