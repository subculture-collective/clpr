CREATE TABLE recommendation_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    clip_id UUID NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
    feedback_type VARCHAR(10) NOT NULL CHECK (feedback_type IN ('positive', 'negative')),
    algorithm VARCHAR(20) CHECK (algorithm IN ('content', 'collaborative', 'hybrid', 'trending')),
    score DOUBLE PRECISION CHECK (score >= 0 AND score <= 1),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recommendation_feedback_user_created
    ON recommendation_feedback(user_id, created_at DESC);
CREATE INDEX idx_recommendation_feedback_clip
    ON recommendation_feedback(clip_id, created_at DESC);
