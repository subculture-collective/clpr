#!/bin/bash
set -euo pipefail

# Restore Drill Script
# Performs scheduled restore test to validate RTO < 1h and RPO < 15m
# Exit codes: 0 = success, 1 = drill failed

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
CLOUD_PROVIDER="${CLOUD_PROVIDER:-gcp}"
BACKUP_BUCKET="${BACKUP_BUCKET:-clpr-backups-prod}"
AZURE_STORAGE_ACCOUNT="${AZURE_STORAGE_ACCOUNT:-}"
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-clpr}"
POSTGRES_DB="${POSTGRES_DB:-clpr}"
TEST_DB_PREFIX="${TEST_DB_PREFIX:-restore_drill_test}"
RTO_TARGET_SECONDS="${RTO_TARGET_SECONDS:-3600}"  # 1 hour
RPO_TARGET_SECONDS="${RPO_TARGET_SECONDS:-900}"   # 15 minutes
DRILL_LOG="${DRILL_LOG:-/var/log/clpr/restore-drill.log}"

# Metrics
DRILL_SUCCESS=0
DRILL_TIMESTAMP=$(date +%s)
RESTORE_DURATION_SECONDS=0
RPO_SECONDS=0
CLIP_COUNT=0
USER_COUNT=0
RTO_MET=0
RPO_MET=0

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1" | tee -a "$DRILL_LOG"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$DRILL_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$DRILL_LOG"
}

# Ensure log directory exists
mkdir -p "$(dirname "$DRILL_LOG")"

echo "=== Restore Drill Started at $(date) ===" | tee -a "$DRILL_LOG"
log_info "Configuration:"
log_info "  Cloud Provider: $CLOUD_PROVIDER"
log_info "  Backup Bucket: $BACKUP_BUCKET"
log_info "  PostgreSQL Host: $POSTGRES_HOST:$POSTGRES_PORT"
log_info "  RTO Target: ${RTO_TARGET_SECONDS}s ($(($RTO_TARGET_SECONDS / 60)) minutes)"
log_info "  RPO Target: ${RPO_TARGET_SECONDS}s ($(($RPO_TARGET_SECONDS / 60)) minutes)"

# Function to download latest backup
download_latest_backup() {
    local backup_file="/tmp/restore-drill-$(date +%Y%m%d-%H%M%S).sql.gz"
    local latest_backup=""
    local backup_timestamp=""
    
    log_info "Finding latest backup..."
    
    if [ "$CLOUD_PROVIDER" = "gcp" ]; then
        latest_backup=$(gsutil ls -l "gs://${BACKUP_BUCKET}/database/postgres-backup-*.sql.gz" 2>/dev/null | \
            grep -v TOTAL | sort -k2 -r | head -1 | awk '{print $3}') || {
            log_error "Failed to list backups from GCS"
            return 1
        }
        
        if [ -z "$latest_backup" ]; then
            log_error "No backups found in GCS bucket"
            return 1
        fi
        
        backup_timestamp=$(gsutil ls -l "$latest_backup" | grep -v TOTAL | awk '{print $2}')
        
        log_info "Downloading backup from GCS: $latest_backup"
        gsutil cp "$latest_backup" "$backup_file" || {
            log_error "Failed to download backup from GCS"
            return 1
        }
        
    elif [ "$CLOUD_PROVIDER" = "aws" ]; then
        local latest_file=$(aws s3 ls "s3://${BACKUP_BUCKET}/database/" 2>/dev/null | \
            grep "postgres-backup-" | sort -r | head -1 | awk '{print $4}') || {
            log_error "Failed to list backups from S3"
            return 1
        }
        
        if [ -z "$latest_file" ]; then
            log_error "No backups found in S3 bucket"
            return 1
        fi
        
        latest_backup="s3://${BACKUP_BUCKET}/database/${latest_file}"
        
        local s3_info=$(aws s3 ls "s3://${BACKUP_BUCKET}/database/${latest_file}" 2>/dev/null)
        backup_timestamp=$(echo "$s3_info" | awk '{print $1" "$2}')
        
        log_info "Downloading backup from S3: $latest_backup"
        aws s3 cp "$latest_backup" "$backup_file" || {
            log_error "Failed to download backup from S3"
            return 1
        }
        
    elif [ "$CLOUD_PROVIDER" = "azure" ]; then
        if [ -z "$AZURE_STORAGE_ACCOUNT" ]; then
            log_error "AZURE_STORAGE_ACCOUNT not set"
            return 1
        fi
        
        latest_backup=$(az storage blob list \
            --account-name "${AZURE_STORAGE_ACCOUNT}" \
            --container-name "${BACKUP_BUCKET}" \
            --prefix "database/postgres-backup-" \
            --query "sort_by([].{name:name, modified:properties.lastModified}, &modified)[-1].name" \
            -o tsv 2>/dev/null) || {
            log_error "Failed to list backups from Azure"
            return 1
        }
        
        if [ -z "$latest_backup" ]; then
            log_error "No backups found in Azure blob storage"
            return 1
        fi
        
        local blob_info=$(az storage blob show \
            --account-name "${AZURE_STORAGE_ACCOUNT}" \
            --container-name "${BACKUP_BUCKET}" \
            --name "${latest_backup}" \
            --query "properties.lastModified" \
            -o tsv 2>/dev/null)
        backup_timestamp="$blob_info"
        
        log_info "Downloading backup from Azure: $latest_backup"
        az storage blob download \
            --account-name "${AZURE_STORAGE_ACCOUNT}" \
            --container-name "${BACKUP_BUCKET}" \
            --name "${latest_backup}" \
            --file "$backup_file" || {
            log_error "Failed to download backup from Azure"
            return 1
        }
    else
        log_error "Unsupported cloud provider: $CLOUD_PROVIDER"
        return 1
    fi
    
    log_info "✓ Backup downloaded: $backup_file"
    log_info "  Size: $(du -h "$backup_file" | cut -f1)"
    log_info "  Backup timestamp: $backup_timestamp"
    
    # Calculate RPO (time between backup and current time)
    if [ -n "$backup_timestamp" ]; then
        local backup_ts_epoch=$(date -d "$backup_timestamp" +%s 2>/dev/null || date -j -f "%Y-%m-%d %H:%M:%S" "$backup_timestamp" +%s 2>/dev/null || echo "0")
        if [ "$backup_ts_epoch" = "0" ] || [ -z "$backup_ts_epoch" ]; then
            log_warn "Failed to parse backup timestamp: $backup_timestamp"
            RPO_SECONDS=999999  # Very large value to indicate failure
        else
            local current_ts=$(date +%s)
            RPO_SECONDS=$((current_ts - backup_ts_epoch))
        fi
        
        log_info "  RPO (backup age): ${RPO_SECONDS}s ($(($RPO_SECONDS / 60)) minutes)"
        
        if [ "$RPO_SECONDS" -le "$RPO_TARGET_SECONDS" ]; then
            log_info "  ✓ RPO target met"
            RPO_MET=1
        else
            log_warn "  ✗ RPO target exceeded (${RPO_SECONDS}s > ${RPO_TARGET_SECONDS}s)"
            RPO_MET=0
        fi
    else
        log_warn "Backup timestamp is empty, cannot calculate RPO"
        RPO_SECONDS=999999
        RPO_MET=0
    fi
    
    # Export for use in other functions
    export BACKUP_FILE="$backup_file"
}

# Function to create test database
create_test_database() {
    local test_db="${TEST_DB_PREFIX}_$(date +%Y%m%d_%H%M%S)"
    
    log_info "Creating test database: $test_db"
    
    # Drop if exists (cleanup from previous failed runs)
    PGPASSWORD="${POSTGRES_PASSWORD}" psql \
        -h "$POSTGRES_HOST" \
        -p "$POSTGRES_PORT" \
        -U "$POSTGRES_USER" \
        -d postgres \
        -c "DROP DATABASE IF EXISTS ${test_db};" 2>/dev/null || true
    
    # Create new test database
    PGPASSWORD="${POSTGRES_PASSWORD}" psql \
        -h "$POSTGRES_HOST" \
        -p "$POSTGRES_PORT" \
        -U "$POSTGRES_USER" \
        -d postgres \
        -c "CREATE DATABASE ${test_db};" || {
        log_error "Failed to create test database"
        return 1
    }
    
    log_info "✓ Test database created: $test_db"
    
    # Export for use in other functions
    export TEST_DB="$test_db"
}

# Function to restore backup to test database
restore_backup() {
    log_info "Starting restore operation..."
    log_info "  Source: $BACKUP_FILE"
    log_info "  Target: $TEST_DB"
    
    local restore_start=$(date +%s)
    
    # Restore backup using pg_restore
    # Using --no-owner and --no-acl to avoid permission issues during test
    PGPASSWORD="${POSTGRES_PASSWORD}" pg_restore \
        -h "$POSTGRES_HOST" \
        -p "$POSTGRES_PORT" \
        -U "$POSTGRES_USER" \
        -d "$TEST_DB" \
        -F c \
        --no-owner \
        --no-acl \
        --verbose \
        "$BACKUP_FILE" 2>&1 | tee -a "$DRILL_LOG" || {
        log_error "Failed to restore backup"
        return 1
    }
    
    local restore_end=$(date +%s)
    RESTORE_DURATION_SECONDS=$((restore_end - restore_start))
    
    log_info "✓ Restore completed"
    log_info "  Duration: ${RESTORE_DURATION_SECONDS}s ($(($RESTORE_DURATION_SECONDS / 60)) minutes)"
    
    # Check RTO
    if [ "$RESTORE_DURATION_SECONDS" -le "$RTO_TARGET_SECONDS" ]; then
        log_info "  ✓ RTO target met (${RESTORE_DURATION_SECONDS}s < ${RTO_TARGET_SECONDS}s)"
        RTO_MET=1
    else
        log_error "  ✗ RTO target exceeded (${RESTORE_DURATION_SECONDS}s > ${RTO_TARGET_SECONDS}s)"
        RTO_MET=0
    fi
}

# Function to validate restored data
validate_restored_data() {
    log_info "Validating restored data..."
    
    # Check clips table
    local clip_result=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql \
        -h "$POSTGRES_HOST" \
        -p "$POSTGRES_PORT" \
        -U "$POSTGRES_USER" \
        -d "$TEST_DB" \
        -t -c "SELECT COUNT(*) FROM clips;" 2>&1)
    
    if [ $? -ne 0 ]; then
        log_error "Failed to query clips table: $clip_result"
        return 1
    fi
    
    CLIP_COUNT=$(echo "$clip_result" | tr -d '[:space:]')
    log_info "  Clips count: $CLIP_COUNT"
    
    # Check users table
    local user_result=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql \
        -h "$POSTGRES_HOST" \
        -p "$POSTGRES_PORT" \
        -U "$POSTGRES_USER" \
        -d "$TEST_DB" \
        -t -c "SELECT COUNT(*) FROM users;" 2>&1)
    
    if [ $? -ne 0 ]; then
        log_error "Failed to query users table: $user_result"
        return 1
    fi
    
    USER_COUNT=$(echo "$user_result" | tr -d '[:space:]')
    log_info "  Users count: $USER_COUNT"
    
    # Basic sanity checks
    if [ "$CLIP_COUNT" -lt 1 ] && [ "$USER_COUNT" -lt 1 ]; then
        log_warn "Restored database appears to be empty (this may be expected for new installations)"
    fi
    
    # Check table integrity
    log_info "Checking table integrity..."
    
    local integrity_check=$(PGPASSWORD="${POSTGRES_PASSWORD}" psql \
        -h "$POSTGRES_HOST" \
        -p "$POSTGRES_PORT" \
        -U "$POSTGRES_USER" \
        -d "$TEST_DB" \
        -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>&1)
    
    if [ $? -ne 0 ]; then
        log_error "Failed to check table integrity: $integrity_check"
        return 1
    fi
    
    local table_count=$(echo "$integrity_check" | tr -d '[:space:]')
    log_info "  Tables restored: $table_count"
    
    if [ "$table_count" -lt 1 ]; then
        log_error "No tables found in restored database"
        return 1
    fi
    
    log_info "✓ Data validation passed"
    return 0
}

# Function to cleanup test database
cleanup_test_database() {
    if [ -z "$TEST_DB" ]; then
        return 0
    fi
    
    log_info "Cleaning up test database: $TEST_DB"
    
    PGPASSWORD="${POSTGRES_PASSWORD}" psql \
        -h "$POSTGRES_HOST" \
        -p "$POSTGRES_PORT" \
        -U "$POSTGRES_USER" \
        -d postgres \
        -c "DROP DATABASE IF EXISTS ${TEST_DB};" 2>/dev/null || {
        log_warn "Failed to cleanup test database (may not exist)"
    }
    
    # Cleanup backup file
    if [ -n "$BACKUP_FILE" ] && [ -f "$BACKUP_FILE" ]; then
        rm -f "$BACKUP_FILE"
        log_info "✓ Backup file cleaned up"
    fi
}

# Function to report metrics
report_metrics() {
    local pushgateway="${PROMETHEUS_PUSHGATEWAY:-}"
    
    if [ -z "$pushgateway" ]; then
        log_info "PROMETHEUS_PUSHGATEWAY not set, skipping metrics push"
        return 0
    fi
    
    log_info "Reporting metrics to $pushgateway..."
    
    cat <<METRICS | curl --data-binary @- "${pushgateway}/metrics/job/restore_drill" 2>/dev/null || log_warn "Failed to push metrics to Prometheus"
# HELP restore_drill_success Whether the last restore drill succeeded (1 = success, 0 = failure)
# TYPE restore_drill_success gauge
restore_drill_success ${DRILL_SUCCESS}
# HELP restore_drill_timestamp Unix timestamp of the last restore drill
# TYPE restore_drill_timestamp gauge
restore_drill_timestamp ${DRILL_TIMESTAMP}
# HELP restore_drill_duration_seconds Duration of the restore operation in seconds
# TYPE restore_drill_duration_seconds gauge
restore_drill_duration_seconds ${RESTORE_DURATION_SECONDS}
# HELP restore_drill_rpo_seconds Recovery Point Objective in seconds (backup age)
# TYPE restore_drill_rpo_seconds gauge
restore_drill_rpo_seconds ${RPO_SECONDS}
# HELP restore_drill_clip_count Number of clips in restored database
# TYPE restore_drill_clip_count gauge
restore_drill_clip_count ${CLIP_COUNT}
# HELP restore_drill_user_count Number of users in restored database
# TYPE restore_drill_user_count gauge
restore_drill_user_count ${USER_COUNT}
# HELP restore_drill_rto_met Whether RTO target was met (1 = met, 0 = not met)
# TYPE restore_drill_rto_met gauge
restore_drill_rto_met ${RTO_MET}
# HELP restore_drill_rpo_met Whether RPO target was met (1 = met, 0 = not met)
# TYPE restore_drill_rpo_met gauge
restore_drill_rpo_met ${RPO_MET}
METRICS
    
    log_info "✓ Metrics reported successfully"
}

# Main drill flow
main() {
    local exit_code=0
    
    # Validate required environment variables
    if [ -z "${POSTGRES_PASSWORD:-}" ]; then
        log_error "POSTGRES_PASSWORD environment variable is not set"
        DRILL_SUCCESS=0
        report_metrics
        return 1
    fi
    
    # Ensure we cleanup on exit
    trap cleanup_test_database EXIT INT TERM
    
    # Download latest backup
    if ! download_latest_backup; then
        log_error "Failed to download backup"
        DRILL_SUCCESS=0
        report_metrics
        return 1
    fi
    
    # Create test database
    if ! create_test_database; then
        log_error "Failed to create test database"
        DRILL_SUCCESS=0
        report_metrics
        return 1
    fi
    
    # Restore backup
    if ! restore_backup; then
        log_error "Failed to restore backup"
        exit_code=1
    fi
    
    # Validate restored data
    if ! validate_restored_data; then
        log_error "Failed to validate restored data"
        exit_code=1
    fi
    
    # Check if RTO and RPO targets were met
    if [ "$RTO_MET" -ne 1 ]; then
        log_error "RTO target not met"
        exit_code=1
    fi
    
    if [ "$RPO_MET" -ne 1 ]; then
        log_warn "RPO target not met (non-fatal)"
    fi
    
    # Set drill success based on exit code
    if [ $exit_code -eq 0 ]; then
        DRILL_SUCCESS=1
        log_info "✓ All restore drill checks passed"
        log_info "Summary:"
        log_info "  - Restore Duration: ${RESTORE_DURATION_SECONDS}s (RTO: ${RTO_TARGET_SECONDS}s)"
        log_info "  - Backup Age: ${RPO_SECONDS}s (RPO: ${RPO_TARGET_SECONDS}s)"
        log_info "  - Clips: $CLIP_COUNT"
        log_info "  - Users: $USER_COUNT"
        echo "=== Restore Drill SUCCEEDED ===" | tee -a "$DRILL_LOG"
    else
        DRILL_SUCCESS=0
        log_error "✗ Restore drill failed"
        echo "=== Restore Drill FAILED ===" | tee -a "$DRILL_LOG"
    fi
    
    # Report metrics
    report_metrics
    
    return $exit_code
}

# Run main function
main
