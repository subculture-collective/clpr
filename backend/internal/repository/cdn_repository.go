package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// CDNRepository handles database operations for CDN configurations and metrics
type CDNRepository struct {
	db *pgxpool.Pool
}

// NewCDNRepository creates a new CDNRepository
func NewCDNRepository(db *pgxpool.Pool) *CDNRepository {
	return &CDNRepository{db: db}
}

// CreateConfiguration creates a new CDN configuration
func (r *CDNRepository) CreateConfiguration(ctx context.Context, config *models.CDNConfiguration) error {
	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO cdn_configurations (
			id, provider, region, is_active, priority, config, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(
		ctx, query,
		uuid.New(), config.Provider, config.Region, config.IsActive,
		config.Priority, configJSON, time.Now(), time.Now(),
	).Scan(&config.ID, &config.CreatedAt, &config.UpdatedAt)

	return err
}

// GetConfiguration retrieves a CDN configuration by ID
func (r *CDNRepository) GetConfiguration(ctx context.Context, id uuid.UUID) (*models.CDNConfiguration, error) {
	query := `
		SELECT id, provider, region, is_active, priority, config, created_at, updated_at
		FROM cdn_configurations
		WHERE id = $1
	`

	var config models.CDNConfiguration
	var configJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&config.ID, &config.Provider, &config.Region, &config.IsActive,
		&config.Priority, &configJSON, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configJSON, &config.Config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetConfigurationByProvider retrieves a CDN configuration by provider and region
func (r *CDNRepository) GetConfigurationByProvider(ctx context.Context, provider string, region *string) (*models.CDNConfiguration, error) {
	query := `
		SELECT id, provider, region, is_active, priority, config, created_at, updated_at
		FROM cdn_configurations
		WHERE provider = $1 AND (region = $2 OR (region IS NULL AND $2 IS NULL))
		AND is_active = true
		ORDER BY priority DESC
		LIMIT 1
	`

	var config models.CDNConfiguration
	var configJSON []byte

	err := r.db.QueryRow(ctx, query, provider, region).Scan(
		&config.ID, &config.Provider, &config.Region, &config.IsActive,
		&config.Priority, &configJSON, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configJSON, &config.Config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ListActiveConfigurations retrieves all active CDN configurations
func (r *CDNRepository) ListActiveConfigurations(ctx context.Context) ([]*models.CDNConfiguration, error) {
	query := `
		SELECT id, provider, region, is_active, priority, config, created_at, updated_at
		FROM cdn_configurations
		WHERE is_active = true
		ORDER BY priority DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.CDNConfiguration
	for rows.Next() {
		var config models.CDNConfiguration
		var configJSON []byte

		err := rows.Scan(
			&config.ID, &config.Provider, &config.Region, &config.IsActive,
			&config.Priority, &configJSON, &config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configJSON, &config.Config); err != nil {
			return nil, err
		}

		configs = append(configs, &config)
	}

	return configs, rows.Err()
}

// UpdateConfiguration updates a CDN configuration
func (r *CDNRepository) UpdateConfiguration(ctx context.Context, config *models.CDNConfiguration) error {
	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return err
	}

	query := `
		UPDATE cdn_configurations
		SET provider = $1, region = $2, is_active = $3, priority = $4,
			config = $5, updated_at = $6
		WHERE id = $7
	`

	_, err = r.db.Exec(
		ctx, query,
		config.Provider, config.Region, config.IsActive, config.Priority,
		configJSON, time.Now(), config.ID,
	)

	return err
}

// DeleteConfiguration deletes a CDN configuration
func (r *CDNRepository) DeleteConfiguration(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM cdn_configurations WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// CreateMetric creates a CDN metric
func (r *CDNRepository) CreateMetric(ctx context.Context, metric *models.CDNMetrics) error {
	metadataJSON, err := json.Marshal(metric.Metadata)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO cdn_metrics (
			id, provider, region, metric_type, metric_value, recorded_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.Exec(
		ctx, query,
		uuid.New(), metric.Provider, metric.Region, metric.MetricType,
		metric.MetricValue, time.Now(), metadataJSON,
	)

	return err
}

// GetMetricsSummary retrieves total aggregated metrics for a provider and time period
// This sums metric values (e.g., for bandwidth totals). For averaging metrics like latency,
// use a separate query with AVG() aggregation instead of calling this method.
func (r *CDNRepository) GetMetricsSummary(ctx context.Context, provider string, metricType string, startTime time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(metric_value), 0)
		FROM cdn_metrics
		WHERE provider = $1
			AND metric_type = $2
			AND recorded_at >= $3
	`

	var totalValue float64
	err := r.db.QueryRow(ctx, query, provider, metricType, startTime).Scan(&totalValue)
	if err != nil {
		return 0, err
	}

	return totalValue, nil
}

// GetTotalCost retrieves total CDN costs for a time period
func (r *CDNRepository) GetTotalCost(ctx context.Context, startTime time.Time, endTime time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(metric_value), 0)
		FROM cdn_metrics
		WHERE metric_type = 'cost'
			AND recorded_at >= $1
			AND recorded_at < $2
	`

	var totalCost float64
	err := r.db.QueryRow(ctx, query, startTime, endTime).Scan(&totalCost)
	if err != nil {
		return 0, err
	}

	return totalCost, nil
}

// GetCacheHitRate calculates the average cache hit rate for a provider
func (r *CDNRepository) GetCacheHitRate(ctx context.Context, provider string, startTime time.Time) (float64, error) {
	query := `
		SELECT COALESCE(AVG(metric_value), 0)
		FROM cdn_metrics
		WHERE provider = $1
			AND metric_type = 'cache_hit_rate'
			AND recorded_at >= $2
	`

	var hitRate float64
	err := r.db.QueryRow(ctx, query, provider, startTime).Scan(&hitRate)
	if err != nil {
		return 0, err
	}

	return hitRate, nil
}
