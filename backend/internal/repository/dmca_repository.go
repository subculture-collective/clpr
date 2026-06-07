package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// DMCARepository handles database operations for DMCA notices, counter-notices, and strikes
type DMCARepository struct {
	db *pgxpool.Pool
}

// NewDMCARepository creates a new DMCA repository
func NewDMCARepository(db *pgxpool.Pool) *DMCARepository {
	return &DMCARepository{db: db}
}

// ==============================================================================
// DMCA Notice Operations
// ==============================================================================

// CreateNotice creates a new DMCA takedown notice
func (r *DMCARepository) CreateNotice(ctx context.Context, notice *models.DMCANotice) error {
	query := `
		INSERT INTO dmca_notices (
			complainant_name, complainant_email, complainant_address, complainant_phone,
			relationship, copyrighted_work_description, infringing_urls,
			good_faith_statement, accuracy_statement, signature,
			submitted_at, ip_address, user_agent, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		notice.ComplainantName,
		notice.ComplainantEmail,
		notice.ComplainantAddress,
		notice.ComplainantPhone,
		notice.Relationship,
		notice.CopyrightedWorkDescription,
		notice.InfringingURLs,
		notice.GoodFaithStatement,
		notice.AccuracyStatement,
		notice.Signature,
		notice.SubmittedAt,
		notice.IPAddress,
		notice.UserAgent,
		notice.Status,
	).Scan(&notice.ID, &notice.CreatedAt, &notice.UpdatedAt)

	return err
}

// GetNoticeByID retrieves a DMCA notice by ID
func (r *DMCARepository) GetNoticeByID(ctx context.Context, id uuid.UUID) (*models.DMCANotice, error) {
	query := `
		SELECT id, complainant_name, complainant_email, complainant_address, complainant_phone,
			relationship, copyrighted_work_description, infringing_urls,
			good_faith_statement, accuracy_statement, signature,
			submitted_at, reviewed_at, reviewed_by, status, notes,
			ip_address::text AS ip_address, user_agent, created_at, updated_at
		FROM dmca_notices
		WHERE id = $1`

	notice := &models.DMCANotice{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&notice.ID,
		&notice.ComplainantName,
		&notice.ComplainantEmail,
		&notice.ComplainantAddress,
		&notice.ComplainantPhone,
		&notice.Relationship,
		&notice.CopyrightedWorkDescription,
		&notice.InfringingURLs,
		&notice.GoodFaithStatement,
		&notice.AccuracyStatement,
		&notice.Signature,
		&notice.SubmittedAt,
		&notice.ReviewedAt,
		&notice.ReviewedBy,
		&notice.Status,
		&notice.Notes,
		&notice.IPAddress,
		&notice.UserAgent,
		&notice.CreatedAt,
		&notice.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("DMCA notice not found")
	}
	return notice, err
}

// ListNotices lists DMCA notices with pagination and filtering
func (r *DMCARepository) ListNotices(ctx context.Context, status string, page, pageSize int) ([]models.DMCANotice, int, error) {
	offset := (page - 1) * pageSize

	// Base query
	baseQuery := `FROM dmca_notices`
	whereClause := ""
	args := []interface{}{}

	if status != "" && status != "all" {
		whereClause = " WHERE status = $1"
		args = append(args, status)
	}

	// Count total
	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Fetch records
	selectQuery := `
		SELECT id, complainant_name, complainant_email, complainant_address, complainant_phone,
			relationship, copyrighted_work_description, infringing_urls,
			good_faith_statement, accuracy_statement, signature,
			submitted_at, reviewed_at, reviewed_by, status, notes,
			ip_address::text AS ip_address, user_agent, created_at, updated_at
		` + baseQuery + whereClause + `
		ORDER BY submitted_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	notices := []models.DMCANotice{}
	for rows.Next() {
		notice := models.DMCANotice{}
		err := rows.Scan(
			&notice.ID,
			&notice.ComplainantName,
			&notice.ComplainantEmail,
			&notice.ComplainantAddress,
			&notice.ComplainantPhone,
			&notice.Relationship,
			&notice.CopyrightedWorkDescription,
			&notice.InfringingURLs,
			&notice.GoodFaithStatement,
			&notice.AccuracyStatement,
			&notice.Signature,
			&notice.SubmittedAt,
			&notice.ReviewedAt,
			&notice.ReviewedBy,
			&notice.Status,
			&notice.Notes,
			&notice.IPAddress,
			&notice.UserAgent,
			&notice.CreatedAt,
			&notice.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		notices = append(notices, notice)
	}

	return notices, totalCount, rows.Err()
}

// UpdateNoticeStatus updates the status and review information of a notice
func (r *DMCARepository) UpdateNoticeStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy uuid.UUID, notes *string) error {
	query := `
		UPDATE dmca_notices
		SET status = $1, reviewed_by = $2, reviewed_at = $3, notes = $4, updated_at = NOW()
		WHERE id = $5`

	_, err := r.db.Exec(ctx, query, status, reviewedBy, time.Now(), notes, id)
	return err
}

// GetPendingNoticesCount returns the count of pending DMCA notices
func (r *DMCARepository) GetPendingNoticesCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM dmca_notices WHERE status = 'pending'`
	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// ==============================================================================
// DMCA Counter-Notice Operations
// ==============================================================================

// CreateCounterNotice creates a new DMCA counter-notice
func (r *DMCARepository) CreateCounterNotice(ctx context.Context, cn *models.DMCACounterNotice) error {
	query := `
		INSERT INTO dmca_counter_notices (
			dmca_notice_id, user_id, user_name, user_email, user_address, user_phone,
			removed_material_url, removed_material_description,
			good_faith_statement, consent_to_jurisdiction, consent_to_service,
			signature, submitted_at, waiting_period_ends, ip_address, user_agent, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		cn.DMCANoticeID,
		cn.UserID,
		cn.UserName,
		cn.UserEmail,
		cn.UserAddress,
		cn.UserPhone,
		cn.RemovedMaterialURL,
		cn.RemovedMaterialDescription,
		cn.GoodFaithStatement,
		cn.ConsentToJurisdiction,
		cn.ConsentToService,
		cn.Signature,
		cn.SubmittedAt,
		cn.WaitingPeriodEnds,
		cn.IPAddress,
		cn.UserAgent,
		cn.Status,
	).Scan(&cn.ID, &cn.CreatedAt, &cn.UpdatedAt)

	return err
}

// GetCounterNoticeByID retrieves a counter-notice by ID
func (r *DMCARepository) GetCounterNoticeByID(ctx context.Context, id uuid.UUID) (*models.DMCACounterNotice, error) {
	query := `
		SELECT id, dmca_notice_id, user_id, user_name, user_email, user_address, user_phone,
			removed_material_url, removed_material_description,
			good_faith_statement, consent_to_jurisdiction, consent_to_service,
			signature, submitted_at, forwarded_at, waiting_period_ends,
			status, lawsuit_filed, lawsuit_filed_at, notes,
			ip_address::text AS ip_address, user_agent, created_at, updated_at
		FROM dmca_counter_notices
		WHERE id = $1`

	cn := &models.DMCACounterNotice{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&cn.ID,
		&cn.DMCANoticeID,
		&cn.UserID,
		&cn.UserName,
		&cn.UserEmail,
		&cn.UserAddress,
		&cn.UserPhone,
		&cn.RemovedMaterialURL,
		&cn.RemovedMaterialDescription,
		&cn.GoodFaithStatement,
		&cn.ConsentToJurisdiction,
		&cn.ConsentToService,
		&cn.Signature,
		&cn.SubmittedAt,
		&cn.ForwardedAt,
		&cn.WaitingPeriodEnds,
		&cn.Status,
		&cn.LawsuitFiled,
		&cn.LawsuitFiledAt,
		&cn.Notes,
		&cn.IPAddress,
		&cn.UserAgent,
		&cn.CreatedAt,
		&cn.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("counter-notice not found")
	}
	return cn, err
}

// ListCounterNotices lists counter-notices with pagination and filtering
func (r *DMCARepository) ListCounterNotices(ctx context.Context, status string, page, pageSize int) ([]models.DMCACounterNotice, int, error) {
	offset := (page - 1) * pageSize

	baseQuery := `FROM dmca_counter_notices`
	whereClause := ""
	args := []interface{}{}

	if status != "" && status != "all" {
		whereClause = " WHERE status = $1"
		args = append(args, status)
	}

	// Count total
	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause
	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Fetch records
	selectQuery := `
		SELECT id, dmca_notice_id, user_id, user_name, user_email, user_address, user_phone,
			removed_material_url, removed_material_description,
			good_faith_statement, consent_to_jurisdiction, consent_to_service,
			signature, submitted_at, forwarded_at, waiting_period_ends,
			status, lawsuit_filed, lawsuit_filed_at, notes,
			ip_address::text AS ip_address, user_agent, created_at, updated_at
		` + baseQuery + whereClause + `
		ORDER BY submitted_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	counterNotices := []models.DMCACounterNotice{}
	for rows.Next() {
		cn := models.DMCACounterNotice{}
		err := rows.Scan(
			&cn.ID,
			&cn.DMCANoticeID,
			&cn.UserID,
			&cn.UserName,
			&cn.UserEmail,
			&cn.UserAddress,
			&cn.UserPhone,
			&cn.RemovedMaterialURL,
			&cn.RemovedMaterialDescription,
			&cn.GoodFaithStatement,
			&cn.ConsentToJurisdiction,
			&cn.ConsentToService,
			&cn.Signature,
			&cn.SubmittedAt,
			&cn.ForwardedAt,
			&cn.WaitingPeriodEnds,
			&cn.Status,
			&cn.LawsuitFiled,
			&cn.LawsuitFiledAt,
			&cn.Notes,
			&cn.IPAddress,
			&cn.UserAgent,
			&cn.CreatedAt,
			&cn.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		counterNotices = append(counterNotices, cn)
	}

	return counterNotices, totalCount, rows.Err()
}

// UpdateCounterNoticeStatus updates counter-notice status and metadata
func (r *DMCARepository) UpdateCounterNoticeStatus(ctx context.Context, id uuid.UUID, status string, notes *string) error {
	query := `
		UPDATE dmca_counter_notices
		SET status = $1, notes = $2, updated_at = NOW()
		WHERE id = $3`

	_, err := r.db.Exec(ctx, query, status, notes, id)
	return err
}

// MarkCounterNoticeForwarded marks a counter-notice as forwarded to complainant
func (r *DMCARepository) MarkCounterNoticeForwarded(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE dmca_counter_notices
		SET status = 'forwarded', forwarded_at = NOW(), updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	return err
}

// GetCounterNoticesAwaitingRestore returns counter-notices past their waiting period
func (r *DMCARepository) GetCounterNoticesAwaitingRestore(ctx context.Context) ([]models.DMCACounterNotice, error) {
	query := `
		SELECT id, dmca_notice_id, user_id, user_name, user_email, user_address, user_phone,
			removed_material_url, removed_material_description,
			good_faith_statement, consent_to_jurisdiction, consent_to_service,
			signature, submitted_at, forwarded_at, waiting_period_ends,
			status, lawsuit_filed, lawsuit_filed_at, notes,
			ip_address::text AS ip_address, user_agent, created_at, updated_at
		FROM dmca_counter_notices
		WHERE status = 'waiting'
		  AND waiting_period_ends IS NOT NULL
		  AND waiting_period_ends <= NOW()
		  AND lawsuit_filed = false
		ORDER BY waiting_period_ends ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counterNotices := []models.DMCACounterNotice{}
	for rows.Next() {
		cn := models.DMCACounterNotice{}
		err := rows.Scan(
			&cn.ID,
			&cn.DMCANoticeID,
			&cn.UserID,
			&cn.UserName,
			&cn.UserEmail,
			&cn.UserAddress,
			&cn.UserPhone,
			&cn.RemovedMaterialURL,
			&cn.RemovedMaterialDescription,
			&cn.GoodFaithStatement,
			&cn.ConsentToJurisdiction,
			&cn.ConsentToService,
			&cn.Signature,
			&cn.SubmittedAt,
			&cn.ForwardedAt,
			&cn.WaitingPeriodEnds,
			&cn.Status,
			&cn.LawsuitFiled,
			&cn.LawsuitFiledAt,
			&cn.Notes,
			&cn.IPAddress,
			&cn.UserAgent,
			&cn.CreatedAt,
			&cn.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		counterNotices = append(counterNotices, cn)
	}

	return counterNotices, rows.Err()
}

// ==============================================================================
// DMCA Strike Operations
// ==============================================================================

// CreateStrike creates a new DMCA strike for a user
func (r *DMCARepository) CreateStrike(ctx context.Context, strike *models.DMCAStrike) error {
	query := `
		INSERT INTO dmca_strikes (
			user_id, dmca_notice_id, clip_id, submission_id,
			strike_number, issued_at, expires_at, status, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		strike.UserID,
		strike.DMCANoticeID,
		strike.ClipID,
		strike.SubmissionID,
		strike.StrikeNumber,
		strike.IssuedAt,
		strike.ExpiresAt,
		strike.Status,
		strike.Notes,
	).Scan(&strike.ID, &strike.CreatedAt, &strike.UpdatedAt)

	return err
}

// GetUserActiveStrikes retrieves all active strikes for a user
func (r *DMCARepository) GetUserActiveStrikes(ctx context.Context, userID uuid.UUID) ([]models.DMCAStrike, error) {
	query := `
		SELECT id, user_id, dmca_notice_id, clip_id, submission_id,
			strike_number, issued_at, expires_at, status, removal_reason, removed_at, notes,
			created_at, updated_at
		FROM dmca_strikes
		WHERE user_id = $1 AND status = 'active'
		ORDER BY issued_at ASC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	strikes := []models.DMCAStrike{}
	for rows.Next() {
		strike := models.DMCAStrike{}
		err := rows.Scan(
			&strike.ID,
			&strike.UserID,
			&strike.DMCANoticeID,
			&strike.ClipID,
			&strike.SubmissionID,
			&strike.StrikeNumber,
			&strike.IssuedAt,
			&strike.ExpiresAt,
			&strike.Status,
			&strike.RemovalReason,
			&strike.RemovedAt,
			&strike.Notes,
			&strike.CreatedAt,
			&strike.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		strikes = append(strikes, strike)
	}

	return strikes, rows.Err()
}

// GetUserAllStrikes retrieves all strikes (active, expired, removed) for a user
func (r *DMCARepository) GetUserAllStrikes(ctx context.Context, userID uuid.UUID) ([]models.DMCAStrike, error) {
	query := `
		SELECT id, user_id, dmca_notice_id, clip_id, submission_id,
			strike_number, issued_at, expires_at, status, removal_reason, removed_at, notes,
			created_at, updated_at
		FROM dmca_strikes
		WHERE user_id = $1
		ORDER BY issued_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	strikes := []models.DMCAStrike{}
	for rows.Next() {
		strike := models.DMCAStrike{}
		err := rows.Scan(
			&strike.ID,
			&strike.UserID,
			&strike.DMCANoticeID,
			&strike.ClipID,
			&strike.SubmissionID,
			&strike.StrikeNumber,
			&strike.IssuedAt,
			&strike.ExpiresAt,
			&strike.Status,
			&strike.RemovalReason,
			&strike.RemovedAt,
			&strike.Notes,
			&strike.CreatedAt,
			&strike.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		strikes = append(strikes, strike)
	}

	return strikes, rows.Err()
}

// RemoveStrike removes a strike (marks as removed with reason)
func (r *DMCARepository) RemoveStrike(ctx context.Context, strikeID uuid.UUID, reason string) error {
	query := `
		UPDATE dmca_strikes
		SET status = 'removed', removal_reason = $1, removed_at = NOW(), updated_at = NOW()
		WHERE id = $2`

	_, err := r.db.Exec(ctx, query, reason, strikeID)
	return err
}

// ExpireOldStrikes marks strikes older than 12 months as expired
func (r *DMCARepository) ExpireOldStrikes(ctx context.Context) (int, error) {
	query := `
		UPDATE dmca_strikes
		SET status = 'expired', removal_reason = 'expired', updated_at = NOW()
		WHERE status = 'active' AND expires_at <= NOW()`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	count := result.RowsAffected()
	return int(count), nil
}

// GetUsersWithStrikes returns users with active strikes
func (r *DMCARepository) GetUsersWithStrikes(ctx context.Context, minStrikes int) ([]uuid.UUID, error) {
	query := `
		SELECT user_id
		FROM users
		WHERE dmca_strikes_count >= $1
		ORDER BY dmca_strikes_count DESC`

	rows, err := r.db.Query(ctx, query, minStrikes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs := []uuid.UUID{}
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}

// ==============================================================================
// DMCA Dashboard Statistics
// ==============================================================================

// GetDashboardStats returns statistics for admin DMCA dashboard
func (r *DMCARepository) GetDashboardStats(ctx context.Context) (*models.DMCADashboardStats, error) {
	stats := &models.DMCADashboardStats{}

	// Pending notices
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM dmca_notices WHERE status = 'pending'`).Scan(&stats.PendingNotices)
	if err != nil {
		return nil, err
	}

	// Pending counter-notices
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM dmca_counter_notices WHERE status = 'pending'`).Scan(&stats.PendingCounterNotices)
	if err != nil {
		return nil, err
	}

	// Content awaiting removal (valid notices not yet processed)
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM dmca_notices WHERE status = 'valid'`).Scan(&stats.ContentAwaitingRemoval)
	if err != nil {
		return nil, err
	}

	// Content awaiting restore (counter-notices past waiting period)
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM dmca_counter_notices
		WHERE status = 'waiting'
		  AND waiting_period_ends <= NOW()
		  AND lawsuit_filed = false`).Scan(&stats.ContentAwaitingRestore)
	if err != nil {
		return nil, err
	}

	// Users with active strikes
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE dmca_strikes_count > 0`).Scan(&stats.UsersWithActiveStrikes)
	if err != nil {
		return nil, err
	}

	// Users with 2 strikes (one away from termination)
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE dmca_strikes_count = 2`).Scan(&stats.UsersWithTwoStrikes)
	if err != nil {
		return nil, err
	}

	// Total takedowns this month
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM dmca_notices
		WHERE submitted_at >= DATE_TRUNC('month', CURRENT_DATE)`).Scan(&stats.TotalTakedownsThisMonth)
	if err != nil {
		return nil, err
	}

	// Total counter-notices this month
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM dmca_counter_notices
		WHERE submitted_at >= DATE_TRUNC('month', CURRENT_DATE)`).Scan(&stats.TotalCounterNoticesThisMonth)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
