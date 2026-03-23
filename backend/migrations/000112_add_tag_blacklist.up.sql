-- Tag blacklist table
CREATE TABLE IF NOT EXISTS blacklisted_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern VARCHAR(255) NOT NULL UNIQUE,
    reason VARCHAR(500),
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_blacklisted_tags_pattern ON blacklisted_tags(pattern);

-- Seed common useless auto-tags
INSERT INTO blacklisted_tags (pattern, reason) VALUES
    ('short', 'Duration-based tag adds no value'),
    ('long', 'Duration-based tag adds no value'),
    ('en', 'Language code tag'),
    ('es', 'Language code tag'),
    ('fr', 'Language code tag'),
    ('de', 'Language code tag'),
    ('ja', 'Language code tag'),
    ('ko', 'Language code tag'),
    ('pt', 'Language code tag'),
    ('ru', 'Language code tag'),
    ('it', 'Language code tag'),
    ('zh', 'Language code tag'),
    ('pl', 'Language code tag'),
    ('english', 'Language name tag'),
    ('spanish', 'Language name tag'),
    ('french', 'Language name tag'),
    ('german', 'Language name tag'),
    ('japanese', 'Language name tag'),
    ('korean', 'Language name tag'),
    ('portuguese', 'Language name tag'),
    ('russian', 'Language name tag'),
    ('italian', 'Language name tag'),
    ('chinese', 'Language name tag'),
    ('polish', 'Language name tag')
ON CONFLICT (pattern) DO NOTHING;
