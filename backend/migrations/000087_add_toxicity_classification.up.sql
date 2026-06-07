-- Add Toxicity Classification System for ML-Based Comment Moderation
-- Roadmap 5.0 Phase 3.3: ML-based moderation to reduce abusive comments

-- Toxicity Predictions Table
-- Stores ML model predictions for toxicity detection with metrics tracking
CREATE TABLE toxicity_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    toxic BOOLEAN NOT NULL DEFAULT FALSE,
    confidence_score DECIMAL(3,2) NOT NULL CHECK (confidence_score >= 0 AND confidence_score <= 1),
    categories JSONB, -- Detailed scores per category (e.g., {"TOXICITY": 0.85, "INSULT": 0.42})
    reason_codes TEXT[], -- Array of triggered categories (e.g., {"TOXICITY", "INSULT"})
    model_version VARCHAR(50) DEFAULT 'perspective-api-v1',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    -- Ensure one prediction per comment
    CONSTRAINT uq_toxicity_predictions_comment UNIQUE (comment_id)
);

-- Indexes for efficient querying
CREATE INDEX idx_toxicity_predictions_comment ON toxicity_predictions(comment_id);
CREATE INDEX idx_toxicity_predictions_toxic ON toxicity_predictions(toxic);
CREATE INDEX idx_toxicity_predictions_confidence ON toxicity_predictions(confidence_score DESC);
CREATE INDEX idx_toxicity_predictions_created_at ON toxicity_predictions(created_at DESC);

-- Human Review Feedback Table
-- Tracks moderator feedback on ML predictions for model improvement
CREATE TABLE toxicity_review_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prediction_id UUID NOT NULL REFERENCES toxicity_predictions(id) ON DELETE CASCADE,
    reviewer_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actual_toxic BOOLEAN NOT NULL,
    feedback_notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    -- One feedback per prediction
    CONSTRAINT uq_toxicity_feedback_prediction UNIQUE (prediction_id)
);

CREATE INDEX idx_toxicity_feedback_prediction ON toxicity_review_feedback(prediction_id);
CREATE INDEX idx_toxicity_feedback_reviewer ON toxicity_review_feedback(reviewer_id);

-- Add metadata column to moderation_queue if it doesn't exist
-- This allows storing ML-specific metadata like model version, confidence breakdown
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'moderation_queue' 
        AND column_name = 'metadata'
    ) THEN
        ALTER TABLE moderation_queue ADD COLUMN metadata JSONB;
        CREATE INDEX idx_modqueue_metadata ON moderation_queue USING GIN(metadata);
    END IF;
END $$;

-- Function to automatically create moderation queue entry for high-confidence toxic comments
-- Note: The threshold (0.85) is intentionally hardcoded here to match the default TOXICITY_THRESHOLD.
-- If you change the threshold in application config, you should also update this function
-- by running: ALTER FUNCTION auto_flag_toxic_comment() ... or by creating a new migration.
CREATE OR REPLACE FUNCTION auto_flag_toxic_comment()
RETURNS TRIGGER AS $$
DECLARE
    threshold DECIMAL(3,2) := 0.85; -- Default threshold matching TOXICITY_THRESHOLD config
BEGIN
    -- Only auto-flag if confidence is above threshold and marked as toxic
    IF NEW.toxic = TRUE AND NEW.confidence_score >= threshold THEN
        -- Insert into moderation queue if not already there
        INSERT INTO moderation_queue (
            content_type,
            content_id,
            reason,
            priority,
            status,
            auto_flagged,
            confidence_score,
            metadata
        )
        VALUES (
            'comment',
            NEW.comment_id,
            -- Determine reason from reason_codes
            CASE 
                WHEN 'SEVERE_TOXICITY' = ANY(NEW.reason_codes) THEN 'toxic'
                WHEN 'IDENTITY_ATTACK' = ANY(NEW.reason_codes) THEN 'harassment'
                WHEN 'THREAT' = ANY(NEW.reason_codes) THEN 'harassment'
                WHEN 'SEXUALLY_EXPLICIT' = ANY(NEW.reason_codes) THEN 'inappropriate'
                WHEN 'INSULT' = ANY(NEW.reason_codes) THEN 'offensive'
                WHEN 'PROFANITY' = ANY(NEW.reason_codes) THEN 'offensive'
                ELSE 'toxic'
            END,
            -- Priority based on confidence (50-100 scale)
            GREATEST(50, LEAST(100, (NEW.confidence_score * 100)::INT)),
            'pending',
            TRUE,
            NEW.confidence_score,
            jsonb_build_object(
                'model_version', NEW.model_version,
                'categories', NEW.categories,
                'reason_codes', NEW.reason_codes
            )
        )
        ON CONFLICT (content_type, content_id) WHERE status = 'pending'
        DO UPDATE SET
            confidence_score = EXCLUDED.confidence_score,
            priority = EXCLUDED.priority,
            metadata = EXCLUDED.metadata;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-flag toxic comments
CREATE TRIGGER trg_auto_flag_toxic_comment
    AFTER INSERT OR UPDATE ON toxicity_predictions
    FOR EACH ROW
    EXECUTE FUNCTION auto_flag_toxic_comment();

-- View for toxicity classification metrics
CREATE OR REPLACE VIEW toxicity_metrics AS
SELECT
    DATE(tp.created_at) as date,
    COUNT(*) as total_predictions,
    COUNT(*) FILTER (WHERE tp.toxic = true) as total_flagged_toxic,
    ROUND(AVG(tp.confidence_score), 4) as avg_confidence,
    -- Precision: TP / (TP + FP)
    ROUND(
        COUNT(*) FILTER (WHERE tp.toxic = true AND md.action = 'reject')::DECIMAL / 
        NULLIF(COUNT(*) FILTER (WHERE tp.toxic = true), 0),
        4
    ) as precision,
    -- Recall: TP / (TP + FN)
    ROUND(
        COUNT(*) FILTER (WHERE tp.toxic = true AND md.action = 'reject')::DECIMAL /
        NULLIF(
            COUNT(*) FILTER (WHERE tp.toxic = true AND md.action = 'reject') +
            COUNT(*) FILTER (WHERE tp.toxic = false AND md.action = 'reject'),
            0
        ),
        4
    ) as recall,
    -- False Positive Rate: FP / (FP + TN)
    ROUND(
        COUNT(*) FILTER (WHERE tp.toxic = true AND md.action = 'approve')::DECIMAL /
        NULLIF(
            COUNT(*) FILTER (WHERE tp.toxic = true AND md.action = 'approve') +
            COUNT(*) FILTER (WHERE tp.toxic = false AND md.action = 'approve'),
            0
        ),
        4
    ) as false_positive_rate,
    COUNT(*) FILTER (WHERE tp.toxic = true AND md.action = 'reject') as true_positives,
    COUNT(*) FILTER (WHERE tp.toxic = true AND md.action = 'approve') as false_positives,
    COUNT(*) FILTER (WHERE tp.toxic = false AND md.action = 'reject') as false_negatives,
    COUNT(*) FILTER (WHERE tp.toxic = false AND md.action = 'approve') as true_negatives
FROM toxicity_predictions tp
LEFT JOIN moderation_queue mq ON tp.comment_id = mq.content_id AND mq.content_type = 'comment'
LEFT JOIN moderation_decisions md ON mq.id = md.queue_item_id
GROUP BY DATE(tp.created_at)
ORDER BY date DESC;

-- Grant permissions
GRANT SELECT ON toxicity_predictions TO clpr;
GRANT SELECT ON toxicity_review_feedback TO clpr;
GRANT SELECT ON toxicity_metrics TO clpr;
