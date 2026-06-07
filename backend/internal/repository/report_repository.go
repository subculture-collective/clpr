package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
)

// ReportRepository handles database operations for reports
type ReportRepository struct {
	db *pgxpool.Pool
}

// NewReportRepository creates a new report repository
func NewReportRepository(db *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{db: db}
}

// CreateReport creates a new report
func (r *ReportRepository) CreateReport(ctx context.Context, report *models.Report) error {
	query := `
		INSERT INTO reports (
			id, reporter_id, reportable_type, reportable_id,
			reason, description, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		report.ID,
		report.ReporterID,
		report.ReportableType,
		report.ReportableID,
		report.Reason,
		report.Description,
		report.Status,
		report.CreatedAt,
	)

	return err
}

// GetReportByID retrieves a report by ID
func (r *ReportRepository) GetReportByID(ctx context.Context, reportID uuid.UUID) (*models.Report, error) {
	query := `
		SELECT id, reporter_id, reportable_type, reportable_id,
			reason, description, status, reviewed_by, reviewed_at, created_at
		FROM reports
		WHERE id = $1
	`

	var report models.Report
	err := r.db.QueryRow(ctx, query, reportID).Scan(
		&report.ID,
		&report.ReporterID,
		&report.ReportableType,
		&report.ReportableID,
		&report.Reason,
		&report.Description,
		&report.Status,
		&report.ReviewedBy,
		&report.ReviewedAt,
		&report.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &report, nil
}

// ListReports retrieves reports with filtering and pagination
func (r *ReportRepository) ListReports(ctx context.Context, status, reportableType string, page, limit int) ([]models.Report, int, error) {
	// Build the WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		whereClause += fmt.Sprintf(" AND status = %s", utils.SQLPlaceholder(argIndex))
		args = append(args, status)
		argIndex++
	}

	if reportableType != "" {
		whereClause += fmt.Sprintf(" AND reportable_type = %s", utils.SQLPlaceholder(argIndex))
		args = append(args, reportableType)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM reports %s", whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT id, reporter_id, reportable_type, reportable_id,
			reason, description, status, reviewed_by, reviewed_at, created_at
		FROM reports
		%s
		ORDER BY created_at DESC
		LIMIT %s OFFSET %s
	`, whereClause, utils.SQLPlaceholder(argIndex), utils.SQLPlaceholder(argIndex+1))

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reports []models.Report
	for rows.Next() {
		var report models.Report
		err := rows.Scan(
			&report.ID,
			&report.ReporterID,
			&report.ReportableType,
			&report.ReportableID,
			&report.Reason,
			&report.Description,
			&report.Status,
			&report.ReviewedBy,
			&report.ReviewedAt,
			&report.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		reports = append(reports, report)
	}

	return reports, total, nil
}

// UpdateReportStatus updates the status of a report
func (r *ReportRepository) UpdateReportStatus(ctx context.Context, reportID uuid.UUID, status string, reviewerID uuid.UUID) error {
	query := `
		UPDATE reports
		SET status = $1, reviewed_by = $2, reviewed_at = $3
		WHERE id = $4
	`

	_, err := r.db.Exec(ctx, query, status, reviewerID, time.Now(), reportID)
	return err
}

// CheckDuplicateReport checks if a user has already reported the same item
func (r *ReportRepository) CheckDuplicateReport(ctx context.Context, reporterID, reportableID uuid.UUID, reportableType string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM reports
			WHERE reporter_id = $1
			AND reportable_id = $2
			AND reportable_type = $3
			AND status IN ('pending', 'reviewed')
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, reporterID, reportableID, reportableType).Scan(&exists)
	return exists, err
}

// GetReportCountByUser gets the number of reports submitted by a user in a time window
func (r *ReportRepository) GetReportCountByUser(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM reports
		WHERE reporter_id = $1 AND created_at >= $2
	`

	var count int
	err := r.db.QueryRow(ctx, query, userID, since).Scan(&count)
	return count, err
}

// GetReportsByReportable retrieves all reports for a specific reportable item
func (r *ReportRepository) GetReportsByReportable(ctx context.Context, reportableID uuid.UUID, reportableType string) ([]models.Report, error) {
	query := `
		SELECT id, reporter_id, reportable_type, reportable_id,
			reason, description, status, reviewed_by, reviewed_at, created_at
		FROM reports
		WHERE reportable_id = $1 AND reportable_type = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, reportableID, reportableType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []models.Report
	for rows.Next() {
		var report models.Report
		err := rows.Scan(
			&report.ID,
			&report.ReporterID,
			&report.ReportableType,
			&report.ReportableID,
			&report.Reason,
			&report.Description,
			&report.Status,
			&report.ReviewedBy,
			&report.ReviewedAt,
			&report.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, nil
}
