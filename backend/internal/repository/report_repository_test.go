package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MockReportRepository is a mock implementation of the report repository for testing
type MockReportRepository struct {
	reports             map[uuid.UUID]*models.Report
	reportsByReportable map[string][]models.Report
	reportCountsByUser  map[uuid.UUID]int
	duplicateReports    map[string]bool
}

// NewMockReportRepository creates a new mock report repository
func NewMockReportRepository() *MockReportRepository {
	return &MockReportRepository{
		reports:             make(map[uuid.UUID]*models.Report),
		reportsByReportable: make(map[string][]models.Report),
		reportCountsByUser:  make(map[uuid.UUID]int),
		duplicateReports:    make(map[string]bool),
	}
}

func (m *MockReportRepository) CreateReport(ctx context.Context, report *models.Report) error {
	m.reports[report.ID] = report

	key := report.ReportableType + report.ReportableID.String()
	m.reportsByReportable[key] = append(m.reportsByReportable[key], *report)

	m.reportCountsByUser[report.ReporterID]++

	dupKey := report.ReporterID.String() + report.ReportableType + report.ReportableID.String()
	m.duplicateReports[dupKey] = true

	return nil
}

func (m *MockReportRepository) GetReportByID(ctx context.Context, reportID uuid.UUID) (*models.Report, error) {
	report, exists := m.reports[reportID]
	if !exists {
		return nil, ErrReportNotFound
	}
	return report, nil
}

func (m *MockReportRepository) ListReports(ctx context.Context, status, reportableType string, page, limit int) ([]models.Report, int, error) {
	var filtered []models.Report
	for _, report := range m.reports {
		if status != "" && report.Status != status {
			continue
		}
		if reportableType != "" && report.ReportableType != reportableType {
			continue
		}
		filtered = append(filtered, *report)
	}

	total := len(filtered)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		return []models.Report{}, total, nil
	}
	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

func (m *MockReportRepository) UpdateReportStatus(ctx context.Context, reportID uuid.UUID, status string, reviewerID uuid.UUID) error {
	report, exists := m.reports[reportID]
	if !exists {
		return ErrReportNotFound
	}

	report.Status = status
	report.ReviewedBy = &reviewerID
	now := time.Now()
	report.ReviewedAt = &now

	return nil
}

func (m *MockReportRepository) CheckDuplicateReport(ctx context.Context, reporterID, reportableID uuid.UUID, reportableType string) (bool, error) {
	key := reporterID.String() + reportableType + reportableID.String()
	return m.duplicateReports[key], nil
}

func (m *MockReportRepository) GetReportCountByUser(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	// Simple mock: return the count without time filtering
	return m.reportCountsByUser[userID], nil
}

func (m *MockReportRepository) GetReportsByReportable(ctx context.Context, reportableID uuid.UUID, reportableType string) ([]models.Report, error) {
	key := reportableType + reportableID.String()
	return m.reportsByReportable[key], nil
}

var ErrReportNotFound = struct {
	error
}{}

// Tests

func TestMockReportRepository_CreateReport(t *testing.T) {
	repo := NewMockReportRepository()
	ctx := context.Background()

	report := &models.Report{
		ID:             uuid.New(),
		ReporterID:     uuid.New(),
		ReportableType: "clip",
		ReportableID:   uuid.New(),
		Reason:         "spam",
		Status:         "pending",
		CreatedAt:      time.Now(),
	}

	err := repo.CreateReport(ctx, report)
	if err != nil {
		t.Fatalf("CreateReport failed: %v", err)
	}

	retrieved, err := repo.GetReportByID(ctx, report.ID)
	if err != nil {
		t.Fatalf("GetReportByID failed: %v", err)
	}

	if retrieved.ID != report.ID {
		t.Errorf("Expected report ID %s, got %s", report.ID, retrieved.ID)
	}
}

func TestMockReportRepository_CheckDuplicateReport(t *testing.T) {
	repo := NewMockReportRepository()
	ctx := context.Background()

	reporterID := uuid.New()
	reportableID := uuid.New()
	reportableType := "comment"

	// No duplicate initially
	isDupe, err := repo.CheckDuplicateReport(ctx, reporterID, reportableID, reportableType)
	if err != nil {
		t.Fatalf("CheckDuplicateReport failed: %v", err)
	}
	if isDupe {
		t.Error("Expected no duplicate report initially")
	}

	// Create a report
	report := &models.Report{
		ID:             uuid.New(),
		ReporterID:     reporterID,
		ReportableType: reportableType,
		ReportableID:   reportableID,
		Reason:         "harassment",
		Status:         "pending",
		CreatedAt:      time.Now(),
	}

	err = repo.CreateReport(ctx, report)
	if err != nil {
		t.Fatalf("CreateReport failed: %v", err)
	}

	// Now should be a duplicate
	isDupe, err = repo.CheckDuplicateReport(ctx, reporterID, reportableID, reportableType)
	if err != nil {
		t.Fatalf("CheckDuplicateReport failed: %v", err)
	}
	if !isDupe {
		t.Error("Expected duplicate report after creating one")
	}
}

func TestMockReportRepository_ListReports(t *testing.T) {
	repo := NewMockReportRepository()
	ctx := context.Background()

	// Create multiple reports
	for i := 0; i < 5; i++ {
		status := "pending"
		if i%2 == 0 {
			status = "actioned"
		}

		report := &models.Report{
			ID:             uuid.New(),
			ReporterID:     uuid.New(),
			ReportableType: "clip",
			ReportableID:   uuid.New(),
			Reason:         "spam",
			Status:         status,
			CreatedAt:      time.Now(),
		}

		err := repo.CreateReport(ctx, report)
		if err != nil {
			t.Fatalf("CreateReport failed: %v", err)
		}
	}

	// List all reports
	reports, total, err := repo.ListReports(ctx, "", "", 1, 10)
	if err != nil {
		t.Fatalf("ListReports failed: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected 5 total reports, got %d", total)
	}

	if len(reports) != 5 {
		t.Errorf("Expected 5 reports in result, got %d", len(reports))
	}

	// Filter by status
	reports, _, err = repo.ListReports(ctx, "pending", "", 1, 10)
	if err != nil {
		t.Fatalf("ListReports failed: %v", err)
	}

	if len(reports) != 2 {
		t.Errorf("Expected 2 pending reports, got %d", len(reports))
	}
}

func TestMockReportRepository_UpdateReportStatus(t *testing.T) {
	repo := NewMockReportRepository()
	ctx := context.Background()

	report := &models.Report{
		ID:             uuid.New(),
		ReporterID:     uuid.New(),
		ReportableType: "user",
		ReportableID:   uuid.New(),
		Reason:         "harassment",
		Status:         "pending",
		CreatedAt:      time.Now(),
	}

	err := repo.CreateReport(ctx, report)
	if err != nil {
		t.Fatalf("CreateReport failed: %v", err)
	}

	reviewerID := uuid.New()
	err = repo.UpdateReportStatus(ctx, report.ID, "actioned", reviewerID)
	if err != nil {
		t.Fatalf("UpdateReportStatus failed: %v", err)
	}

	updated, err := repo.GetReportByID(ctx, report.ID)
	if err != nil {
		t.Fatalf("GetReportByID failed: %v", err)
	}

	if updated.Status != "actioned" {
		t.Errorf("Expected status 'actioned', got %s", updated.Status)
	}

	if updated.ReviewedBy == nil || *updated.ReviewedBy != reviewerID {
		t.Error("Expected reviewed_by to be set")
	}

	if updated.ReviewedAt == nil {
		t.Error("Expected reviewed_at to be set")
	}
}

func TestMockReportRepository_GetReportsByReportable(t *testing.T) {
	repo := NewMockReportRepository()
	ctx := context.Background()

	reportableID := uuid.New()
	reportableType := "clip"

	// Create multiple reports for the same item
	for i := 0; i < 3; i++ {
		report := &models.Report{
			ID:             uuid.New(),
			ReporterID:     uuid.New(),
			ReportableType: reportableType,
			ReportableID:   reportableID,
			Reason:         "nsfw",
			Status:         "pending",
			CreatedAt:      time.Now(),
		}

		err := repo.CreateReport(ctx, report)
		if err != nil {
			t.Fatalf("CreateReport failed: %v", err)
		}
	}

	reports, err := repo.GetReportsByReportable(ctx, reportableID, reportableType)
	if err != nil {
		t.Fatalf("GetReportsByReportable failed: %v", err)
	}

	if len(reports) != 3 {
		t.Errorf("Expected 3 reports for the item, got %d", len(reports))
	}
}
